package grpc

import (
	"fmt"
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/rs/zerolog/log"

	"github.com/go-payment-authorizer/internal/core/service"
	"github.com/go-payment-authorizer/internal/core/model"
	"github.com/go-payment-authorizer/internal/core/erro"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/genproto/googleapis/rpc/errdetails"

	token_proto_service "github.com/go-payment-authorizer/protogen/token"
	proto "github.com/go-payment-authorizer/protogen/token"
	//proto "github.com/eliezerraj/go-grpc-proto/protogen/token"

	go_core_observ "github.com/eliezerraj/go-core/observability"
)

var childLogger = log.With().Str("component","go-payment-authorizer").Str("package","internal.adapter.grpc").Logger()
var tracerProvider go_core_observ.TracerProvider

type AdapterGrpc struct{
	appServer 		*model.AppServer
	workerService 	*service.WorkerService
	token_proto_service.UnimplementedTokenServiceServer
}

// About create new adapter
func NewAdapterGrpc(appServer *model.AppServer, workerService *service.WorkerService) *AdapterGrpc {
	childLogger.Info().Str("func","NewAdapterGrpc").Send()

	return &AdapterGrpc{
		appServer: appServer,
		workerService: workerService,
	}
}

// About get pod data
func (a *AdapterGrpc) GetPod(ctx context.Context, podRequest *proto.PodRequest) (*proto.PodResponse, error) {
	childLogger.Info().Str("func","GetPodInfo").Send()

	// Trace
	span := tracerProvider.Span(ctx, "adpater.grpc.GetPod")
	defer span.End()

	pod := proto.Pod{	IpAddress: 	a.appServer.InfoPod.IPAddress,
						PodName: a.appServer.InfoPod.PodName,
						AvailabilityZone: a.appServer.InfoPod.AvailabilityZone,
						Host: a.appServer.Server.Port,
						Version: a.appServer.InfoPod.ApiVersion,
					}

	res_pod := &proto.PodResponse {
		Pod: &pod,
	}
	
	return res_pod, nil
}

// About get card from token
func (a *AdapterGrpc) AddPaymentToken(ctx context.Context, paymentRequest *proto.PaymentTokenRequest) (*proto.PaymentTokenResponse, error) {
	childLogger.Info().Str("func","AddPaymentToken").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("paymentRequest", paymentRequest).Send()

	// Trace
	span := tracerProvider.Span(ctx, "adpater.grpc.AddPaymentToken")
	defer span.End()

	// Prepare
	payment := model.Payment{ TokenData: paymentRequest.Payment.TokenData,
							  Terminal: paymentRequest.Payment.Terminal,	
							  Currency: paymentRequest.Payment.Currency,
							  Amount: paymentRequest.Payment.Amount,
							  CardType: paymentRequest.Payment.CardType,
							  Mcc: paymentRequest.Payment.Mcc,
							  TransactionId : &paymentRequest.Payment.TransactionId,						
							}

	// Call service
	res_payment, err := a.workerService.AddPaymentToken(ctx, payment)
	if (err != nil) {
		switch err {
		case erro.ErrCardTypeInvalid:
			s := status.New(codes.InvalidArgument, err.Error())
			s, _ = s.WithDetails(&errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "cart.type",
						Description: fmt.Sprintf("card type (%v) informed not valid", paymentRequest.Payment.CardType),
					},
				},
			})
			return nil, s.Err()		
		case erro.ErrNotFound:
			s := status.New(codes.InvalidArgument, err.Error())
			s, _ = s.WithDetails(&errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "token/terminal",
						Description: fmt.Sprintf("token (%v) or terminal (%v) informed not found", paymentRequest.Payment.TokenData, paymentRequest.Payment.Terminal),
					},
				},
			})
			return nil, s.Err()
		default:
			s := status.New(codes.Internal, err.Error())
			s, _ = s.WithDetails(&errdetails.ErrorInfo{
				Domain: "service",
				Reason: "service payment unreacheable",
			})
			return nil, s.Err()
		}
	}	

	res_payment_proto_response := &proto.PaymentTokenResponse {
		Payment: &proto.Payment{	TokenData: 	res_payment.TokenData,
									CardType:  	res_payment.CardType,
									CardModel:  res_payment.CardModel,
									CardAtc:	uint32(res_payment.CardAtc),
									Status:  	res_payment.Status,
									Currency:  	res_payment.Currency,
									Amount:  	res_payment.Amount,
									Mcc: 		res_payment.Mcc,
									Terminal: 	res_payment.Terminal,
									TransactionId: *res_payment.TransactionId,
									PaymentAt: 	timestamppb.New(res_payment.PaymentAt),
							},	
	}

	return res_payment_proto_response, nil
}

