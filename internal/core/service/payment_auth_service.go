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

var childLogger = log.With().Str("component","go-payment-authorizer").Str("package","internal.core.service").Logger()
var tracerProvider go_core_observ.TracerProvider
var apiService go_core_api.ApiService

type WorkerService struct {
	apiService			[]model.ApiService
	workerRepository 	*database.WorkerRepository
	adapterGrpcClient	*client.AdapterGrpc
}

// About create a new worker service
func NewWorkerService(	workerRepository 	*database.WorkerRepository,
						apiService			[]model.ApiService,
						adapterGrpcClient	*client.AdapterGrpc) *WorkerService{

	childLogger.Info().Str("func","NewWorkerService").Send()

	return &WorkerService{
		workerRepository: workerRepository,
		apiService: apiService,
		adapterGrpcClient: adapterGrpcClient,
	}
}

// About handle/convert http status code
func errorStatusCode(statusCode int, serviceName string) error{
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
			err = errors.New(fmt.Sprintf("service %s in outage", serviceName))
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

	// Handle the transaction
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		s.workerRepository.DatabasePGServer.ReleaseTx(conn)
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
	transactionLimit := model.TransactionLimit{ Category: 		"CREDIT",
												CardNumber: 	payment.CardNumber,
												TransactionId: 	*payment.TransactionId,
												Mcc: 			payment.Mcc,
												Currency:		payment.Currency,
												Amount:			payment.Amount }

	// Set headers
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Request-Id": trace_id,
		"x-apigw-api-id": s.apiService[1].XApigwApiId,
		"Host": s.apiService[1].HostName,
	}
	// Prepare http client
	httpClient := go_core_api.HttpClient {
		Url: fmt.Sprintf("%v%v",s.apiService[1].Url,"/transactionLimit"),
		Method: s.apiService[1].Method,
		Timeout: 15,
		Headers: &headers,
	}

	// Call go-limit
	res_limit, statusCode, err := apiService.CallRestApi(ctx,
														httpClient, 
														transactionLimit)
	if err != nil {
		return nil, errorStatusCode(statusCode, s.apiService[1].Name)
	}

	jsonString, err  := json.Marshal(res_limit)
	if err != nil {
		return nil, errors.New(err.Error())
    }
	json.Unmarshal(jsonString, &transactionLimit)
	
	// add step 02
	stepProcess02 := model.StepProcess{	Name: fmt.Sprintf("LIMIT:%v", transactionLimit.Status),
										ProcessedAt: time.Now(),}
	list_stepProcess = append(list_stepProcess, stepProcess02)

	// ------------------------  STEP-3 ----------------------------------//
	childLogger.Info().Str("func","AddPaymentToken").Msg("===> STEP - 03 (FRAUD) <===")
	// Check Fraud

	// ------------------------  STEP-4 ----------------------------------//
	childLogger.Info().Str("func","AddPaymentToken").Msg("===> STEP - 04 (LEDGER) <===")
	// Access Account (ledger)
	moviment := model.Moviment{	AccountFrom: model.Account{AccountID: (*res_list_card)[0].AccountID}, 
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
		Timeout: 15,
		Headers: &headers,
	}

	// Call go-ledger
	_, statusCode, err = apiService.CallRestApi(ctx,
												httpClient, 
												moviment)
	if err != nil {
		return nil, errorStatusCode(statusCode, s.apiService[2].Name)
	}

	// add step 04
	stepProcess04 := model.StepProcess{	Name: "LEDGER:WITHDRAW:OK",
										ProcessedAt: time.Now(),}
	list_stepProcess = append(list_stepProcess, stepProcess04)

	// ------------------------  STEP-5 ----------------------------------//
	childLogger.Info().Str("func","AddPaymentToken").Msg("===> STEP - 05 (CARDS:ATC) <===")
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
		Timeout: 15,
		Headers: &headers,
	}

	// update card atc
	_, statusCode, err = apiService.CallRestApi(ctx,
												httpClient, 
												card)
	if err != nil {
		return nil, errorStatusCode(statusCode, s.apiService[3].Name)
	}

	stepProcess05 := model.StepProcess{	Name: "CARD-ATC:OK",
										ProcessedAt: time.Now(),}
	list_stepProcess = append(list_stepProcess, stepProcess05)
	
	childLogger.Info().Str("func","AddPaymentToken").Msg("===> STEP - (UPDATE PAYMENT) <===")
	
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

	stepProcess06 := model.StepProcess{Name: "AUTHORIZATION-GRPC:STATUS:OK",
										ProcessedAt: time.Now(),}
	list_stepProcess = append(list_stepProcess, stepProcess06)
	payment.StepProcess = &list_stepProcess

	childLogger.Info().Str("func","AddPaymentToken: ===> FINAL").Interface("payment", payment).Send()

	return &payment, nil
}