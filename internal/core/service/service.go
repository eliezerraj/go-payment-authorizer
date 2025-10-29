package service

import(
	"fmt"
	"time"
	"context"
	"errors"
	"net/http"
	"encoding/json"
	"github.com/rs/zerolog/log"

	"github.com/go-payment-authorizer/internal/adapter/database"
	"github.com/go-payment-authorizer/internal/adapter/grpc/client"
	"github.com/go-payment-authorizer/internal/core/model"
	"github.com/go-payment-authorizer/internal/core/erro"

	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_api "github.com/eliezerraj/go-core/api"
)

var (
	childLogger = log.With().Str("component","go-payment-authorizer").Str("package","internal.core.service").Logger()
	tracerProvider go_core_observ.TracerProvider
	apiService go_core_api.ApiService
)

type WorkerService struct {
	goCoreRestApiService	go_core_api.ApiService
	apiService			[]model.ApiService
	workerRepository 	*database.WorkerRepository
	adapterGrpcClient	*client.AdapterGrpc
}

// About create a new worker service
func NewWorkerService(	goCoreRestApiService	go_core_api.ApiService,	
						workerRepository 	*database.WorkerRepository,
						apiService			[]model.ApiService,
						adapterGrpcClient	*client.AdapterGrpc) *WorkerService{

	childLogger.Info().Str("func","NewWorkerService").Send()

	return &WorkerService{
		goCoreRestApiService: goCoreRestApiService,
		workerRepository: workerRepository,
		apiService: apiService,
		adapterGrpcClient: adapterGrpcClient,
	}
}

// About handle/convert http status code
func errorStatusCode(statusCode int, serviceName string,  msg_err error) error{
	childLogger.Info().Str("func","errorStatusCode").Interface("serviceName", serviceName).Interface("statusCode", statusCode).Send()
	var err error
	switch statusCode {
		case http.StatusUnauthorized:
			err = erro.ErrUnauthorized
		case http.StatusForbidden:
			err = erro.ErrHTTPForbiden
		case http.StatusNotFound:
			err = erro.ErrNotFound
		default:
			err = errors.New(fmt.Sprintf("service %s in outage => cause error: %s", serviceName, msg_err.Error() ))
		}
	return err
}

// About create a tokenization data
func (s * WorkerService) AddPaymentToken(ctx context.Context, payment model.Payment) (*model.Payment, error){
	childLogger.Info().Str("func","AddPaymentToken").Interface("trace-request-id", ctx.Value("trace-request-id")).Interface("payment", payment).Send()

	// Trace
	span := tracerProvider.Span(ctx, "service.AddPaymentToken")
	trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))

	// Get the database connection
	tx, conn, err := s.workerRepository.DatabasePGServer.StartTx(ctx)
	if err != nil {
		return nil, err
	}
	defer s.workerRepository.DatabasePGServer.ReleaseTx(conn)

	// Handle the transaction
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		span.End()
	}()

	// Businness rule
	if (payment.CardType != "CREDIT") && (payment.CardType != "DEBIT") {
		return nil, erro.ErrCardTypeInvalid
	}
	if payment.TransactionId == nil  {
		return nil, erro.ErrTransactioInvalid
	}
	
	// Get terminal
	terminal := model.Terminal{Name: payment.Terminal}
	res_terminal, err := s.workerRepository.GetTerminal(ctx, terminal)
	if err != nil {
		return nil, err
	}

	// ------------------------  STEP-1 ----------------------------------//
	childLogger.Info().Str("func","AddPaymentToken").Msg("===> STEP - 01 (PAYMENT) <===")
	// Get a card from token (call token-grpc)

	card := model.Card{	TokenData: 	payment.TokenData,
						Type:		payment.CardType}
	res_list_card, err := s.adapterGrpcClient.GetCardTokenGrpc(ctx, card)
	if err != nil {
		return nil, err
	}
	if len(*res_list_card) == 0 {
		return nil, erro.ErrNotFound
	}
	
	// Prepare payment
	card.CardNumber = (*res_list_card)[0].CardNumber
	payment.FkCardId = (*res_list_card)[0].ID
	payment.CardNumber = (*res_list_card)[0].CardNumber
	payment.CardModel = (*res_list_card)[0].Model
	payment.CardAtc = (*res_list_card)[0].Atc
	payment.CardType = (*res_list_card)[0].Type
	payment.FkTerminalId = res_terminal.ID
	payment.RequestId = &trace_id
	payment.Status = "AUTHORIZATION-GRPC:PENDING"

	// create a payment
	res_payment, err := s.workerRepository.AddPayment(ctx, tx, &payment)
	if err != nil {
		return nil, err
	}
	payment.ID = res_payment.ID // Set PK
	payment.CreatedAt = res_payment.CreatedAt
	
	// Create a StepProcess
	list_stepProcess := []model.StepProcess{}
	stepProcess01 := model.StepProcess{Name: "AUTHORIZATION-GRPC:STATUS:PENDING",
										ProcessedAt: time.Now(),}
	list_stepProcess = append(list_stepProcess, stepProcess01)
	payment.StepProcess = &list_stepProcess

	// ------------------------  STEP-2 ----------------------------------//
	childLogger.Info().Str("func","AddPaymentToken").Msg("===> STEP - 02 (LIMIT) <===")
	// Check the limits

	limit := model.Limit{ 	TransactionId: *payment.TransactionId,
						  	Key: 	payment.CardNumber,
							TypeLimit: "CREDIT",
							OrderLimit: "MCC:" + payment.Mcc,
							Amount:	payment.Amount,
							Quantity: 1}

	// Set headers
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Request-Id": trace_id,
		"x-apigw-api-id": s.apiService[1].XApigwApiId,
		"Host": s.apiService[1].HostName,
	}
	// Prepare http client
	httpClient := go_core_api.HttpClient {
		Url: fmt.Sprintf("%v%v",s.apiService[1].Url,"/checkLimitTransaction"),
		Method: s.apiService[1].Method,
		Timeout: s.apiService[1].HttpTimeout,
		Headers: &headers,
	}

	// Call go-limit
	res_limit, statusCode, err := apiService.CallRestApiV1(	ctx,
															s.goCoreRestApiService.Client,
															httpClient, 
															limit)
	if err != nil {
		return nil, errorStatusCode(statusCode, s.apiService[1].Name, err)
	}

	list_limit_transaction := []model.LimitTransaction{}
	jsonString, err  := json.Marshal(res_limit)
	if err != nil {
		return nil, errors.New(err.Error())
    }
	json.Unmarshal(jsonString, &list_limit_transaction)
	
	var list_status = []string{}
	for _, val := range list_limit_transaction {
		list_status = append(list_status, val.Status)
	}

	// add step 02
	stepProcess02 := model.StepProcess{	Name: fmt.Sprintf("LIMIT:%v", list_status),
										ProcessedAt: time.Now(),}
	list_stepProcess = append(list_stepProcess, stepProcess02)

	// ------------------------  STEP-3 ----------------------------------//
	childLogger.Info().Str("func","AddPaymentToken").Msg("===> STEP - 03 (FRAUD) <===")
	// Check Fraud

	
	// ------------------------  STEP-4 ----------------------------------//
	childLogger.Info().Str("func","AddPaymentToken").Msg("===> STEP - 04 (ACCOUNT) <===")
	// Check Account

		// prepare headers
	headers = map[string]string{
		"Content-Type":  	"application/json;charset=UTF-8",
		"X-Request-Id": 	trace_id,
		"x-apigw-api-id": 	s.apiService[4].XApigwApiId,
		"Host": 			s.apiService[4].HostName,
	}
	httpClient = go_core_api.HttpClient {
		Url: 	fmt.Sprintf("%v%v%v", s.apiService[4].Url, "/getId/", (*res_list_card)[0].FkAccountID) ,
		Method: s.apiService[4].Method,
		Timeout: s.apiService[4].HttpTimeout,
		Headers: &headers,
	}

	res_payload, statusCode, err := apiService.CallRestApiV1(ctx,
															s.goCoreRestApiService.Client,
															httpClient, 
															nil)
	if err != nil {
		return nil, errorStatusCode(statusCode, s.apiService[4].Name, err)
	}

	jsonString, err  = json.Marshal(res_payload)
	if err != nil {
		return nil, errors.New(err.Error())
    }
	var account_from_parsed model.Account
	json.Unmarshal(jsonString, &account_from_parsed)

	stepProcess04 := model.StepProcess{	Name: "ACCOUNT-FROM:OK",
										ProcessedAt: time.Now(),}

	list_stepProcess = append(list_stepProcess, stepProcess04)

	// ------------------------  STEP-5 ----------------------------------//
	childLogger.Info().Str("func","AddPaymentToken").Msg("===> STEP - 05 (LEDGER) <===")
	// Access Account (ledger)
	moviment := model.Moviment{	AccountFrom: model.Account{AccountID: account_from_parsed.AccountID}, 
								Type: "WITHDRAW",
								Currency: payment.Currency,
								Amount: payment.Amount }

	// Set headers
	headers = map[string]string{
		"Content-Type":  "application/json;charset=UTF-8",
		"X-Request-Id": trace_id,
		"x-apigw-api-id": s.apiService[2].XApigwApiId,
		"Host": s.apiService[2].HostName,
	}
	// prepare http client
	httpClient = go_core_api.HttpClient {
		Url: 	s.apiService[2].Url + "/movimentTransaction",
		Method: s.apiService[2].Method,
		Timeout: s.apiService[2].HttpTimeout,
		Headers: &headers,
	}

	// Call go-ledger
	_, statusCode, err = apiService.CallRestApiV1(ctx,
												s.goCoreRestApiService.Client,
												httpClient, 
												moviment)
	if err != nil {
		return nil, errorStatusCode(statusCode, s.apiService[2].Name, err)
	}

	stepProcess05 := model.StepProcess{	Name: "LEDGER:WITHDRAW:OK",
										ProcessedAt: time.Now(),}
	list_stepProcess = append(list_stepProcess, stepProcess05)

	// ------------------------  STEP-6 ----------------------------------//
	childLogger.Info().Str("func","AddPaymentToken").Msg("===> STEP - 06 (CARDS:ATC) <===")
	// Set headers
	
	headers = map[string]string{
		"Content-Type":  "application/json;charset=UTF-8",
		"X-Request-Id": trace_id,
		"x-apigw-api-id": s.apiService[3].XApigwApiId,
		"Host": s.apiService[3].HostName,
	}
	// prepare http client
	httpClient = go_core_api.HttpClient {
		Url: 	s.apiService[3].Url + "/atc",
		Method: s.apiService[3].Method,
		Timeout: s.apiService[3].HttpTimeout,
		Headers: &headers,
	}

	// update card atc
	_, statusCode, err = apiService.CallRestApiV1(ctx,
												s.goCoreRestApiService.Client,
												httpClient, 
												card)
	if err != nil {
		return nil, errorStatusCode(statusCode, s.apiService[3].Name, err)
	}

	stepProcess06 := model.StepProcess{	Name: "CARD-ATC:OK",
										ProcessedAt: time.Now(),}
	list_stepProcess = append(list_stepProcess, stepProcess06)
	
	// ------------------------  STEP-7 ----------------------------------//	
	childLogger.Info().Str("func","AddPaymentToken").Msg("===> STEP 7- (UPDATE PAYMENT) <===")
	
	// update status payment
	payment.Status = "AUTHORIZATION-GRPC:OK"
	res_update, err := s.workerRepository.UpdatePayment(ctx, tx, payment)
	if err != nil {
		return nil, err
	}
	if res_update == 0 {
		err = erro.ErrUpdate
		return nil, err
	}

	stepProcess07 := model.StepProcess{Name: "AUTHORIZATION-GRPC:STATUS:OK",
										ProcessedAt: time.Now(),}
	list_stepProcess = append(list_stepProcess, stepProcess07)
	payment.StepProcess = &list_stepProcess

		// ------------------------ Final ----------------------------------//	
	childLogger.Info().Str("func","AddPaymentToken: ===> FINAL").Interface("payment", payment).Send()

	return &payment, nil
}