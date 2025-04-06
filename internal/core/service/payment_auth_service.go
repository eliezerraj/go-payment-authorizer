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
func errorStatusCode(statusCode int) error{
	var err error
	switch statusCode {
	case http.StatusUnauthorized:
		err = erro.ErrUnauthorized
	case http.StatusForbidden:
		err = erro.ErrHTTPForbiden
	case http.StatusNotFound:
		err = erro.ErrNotFound
	default:
		err = erro.ErrServer
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

	// Get terminal
	terminal := model.Terminal{Name: payment.Terminal}
	res_terminal, err := s.workerRepository.GetTerminal(ctx, terminal)
	if err != nil {
		return nil, err
	}

	// STEP-1
	childLogger.Info().Str("func","AddPaymentToken").Msg("===> STEP - 01 <===")
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
	payment.Status = "AUTHORIZATION-PENDING:GRPC"

	// create a payment
	res_payment, err := s.workerRepository.AddPayment(ctx, tx, &payment)
	if err != nil {
		return nil, err
	}
	
	// Create a StepProcess
	list_stepProcess := []model.StepProcess{}
	stepProcess01 := model.StepProcess{Name: "AUTHORIZATION-PENDING:GRPC",
										ProcessedAt: time.Now(),}
	list_stepProcess = append(list_stepProcess, stepProcess01)
	payment.StepProcess = &list_stepProcess

	// STEP-2
	childLogger.Info().Str("func","AddPaymentToken").Msg("===> STEP - 02 <===")
	// Check the limits
	transactionLimit := model.TransactionLimit{
		Category: 		"CREDIT",
		CardNumber: 	payment.CardNumber,
		TransactionId: 	*payment.TransactionId,
		Mcc: 			payment.Mcc,
		Currency:		payment.Currency,
		Amount:			payment.Amount,
	}

	// Call go-limit
	res_limit, statusCode, err := apiService.CallApi(ctx,
													s.apiService[1].Url + "/transactionLimit",
													s.apiService[1].Method,
													&s.apiService[1].Header_x_apigw_api_id,
													nil,
													&trace_id, 
													transactionLimit)
	if err != nil {
		return nil, errorStatusCode(statusCode)
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

	//childLogger.Info().Str("func","AddPaymentToken: ===> STEP-2").Interface("transactionLimit", transactionLimit).Send()

	// STEP-3
	childLogger.Info().Str("func","AddPaymentToken").Msg("===> STEP - 03 <===")
	// Check Fraud

	// STEP-4
	childLogger.Info().Str("func","AddPaymentToken").Msg("===> STEP - 04 <===")
	// Access Account (ledger)
	moviment := model.Moviment{	AccountID: (*res_list_card)[0].AccountID,
								Type: "WITHDRAW",
								Currency: payment.Currency,
								Amount: payment.Amount }
	_, statusCode, err = apiService.CallApi(ctx,
											s.apiService[2].Url + "/movimentTransaction",
											s.apiService[2].Method,
											&s.apiService[2].Header_x_apigw_api_id,
											nil,
											&trace_id, 
											moviment)
	if err != nil {
		return nil, errorStatusCode(statusCode)
	}

	// add step 04
	stepProcess04 := model.StepProcess{	Name: "LEDGER:WITHDRAW:OK",
										ProcessedAt: time.Now(),}
	list_stepProcess = append(list_stepProcess, stepProcess04)

	// STEP 05
	childLogger.Info().Str("func","AddPaymentToken").Msg("===> STEP - 05 <===")
	// update card atc
	_, statusCode, err = apiService.CallApi(ctx,
											s.apiService[3].Url + "/atc",
											s.apiService[3].Method,
											&s.apiService[3].Header_x_apigw_api_id,
											nil,
											&trace_id, 
											card)
	if err != nil {
		return nil, errorStatusCode(statusCode)
	}

	stepProcess05 := model.StepProcess{	Name: "CARD-ATC:OK",
										ProcessedAt: time.Now(),}
	list_stepProcess = append(list_stepProcess, stepProcess05)
	
	childLogger.Info().Str("func","AddPaymentToken").Msg("===> STEP - UPDATE PAYMENT <===")
	// update status payment
	res_update, err := s.workerRepository.UpdatePayment(ctx, tx, *res_payment)
	if err != nil {
		return nil, err
	}
	if res_update == 0 {
		err = erro.ErrUpdate
		return nil, err
	}

	childLogger.Info().Str("func","AddPaymentToken: ===> FINAL").Interface("payment", payment).Send()

	return &payment, nil
}