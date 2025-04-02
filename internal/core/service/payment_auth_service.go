package service

import(
	"context"
	"github.com/rs/zerolog/log"

	"github.com/go-payment-authorizer/internal/adapter/database"
	"github.com/go-payment-authorizer/internal/adapter/grpc/client"
	"github.com/go-payment-authorizer/internal/core/model"
	"github.com/go-payment-authorizer/internal/core/erro"

	go_core_observ "github.com/eliezerraj/go-core/observability"
)

var childLogger = log.With().Str("component","go-payment-authorizer").Str("package","internal.core.service").Logger()
var tracerProvider go_core_observ.TracerProvider

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

// About create a tokenization data
func (s * WorkerService) AddPaymentToken(ctx context.Context, payment model.Payment) (*model.Payment, error){
	childLogger.Info().Str("func","AddPaymentToken").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("payment", payment).Send()

	// Trace
	span := tracerProvider.Span(ctx, "service.AddPaymentToken")

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

	// STEP-1 Get a card from token (call token-grpc)
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
	payment.FkCardId = (*res_list_card)[0].ID
	payment.CardNumber = (*res_list_card)[0].CardNumber
	payment.CardModel = (*res_list_card)[0].Model
	payment.CardAtc = (*res_list_card)[0].Atc
	payment.CardType = (*res_list_card)[0].Type
	payment.FkTerminalId = res_terminal.ID
	payment.Status = "AUTHORIZATION-PENDING:GRPC"

	childLogger.Info().Str("func","=====================>").Interface("payment", payment).Send()

	res_payment, err := s.workerRepository.AddPayment(ctx, tx, &payment)
	if err != nil {
		return nil, err
	}

	// STEP-2 Check the limits

	// STEP-3 Check Fraud

	// STEP-4 Access Account (ledger)

	// update status payment
	res_update, err := s.workerRepository.UpdatePayment(ctx, tx, *res_payment)
	if err != nil {
		return nil, err
	}
	if res_update == 0 {
		err = erro.ErrUpdate
		return nil, err
	}

	return &payment, nil
}
