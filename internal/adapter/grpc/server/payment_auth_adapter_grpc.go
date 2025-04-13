package server

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
	"google.golang.org/grpc/metadata"
	"google.golang.org/genproto/googleapis/rpc/errdetails"

	token_proto_service "github.com/go-payment-authorizer/protogen/token"
	proto "github.com/go-payment-authorizer/protogen/token"
	//proto "github.com/eliezerraj/go-grpc-proto/protogen/token"

	go_core_observ "github.com/eliezerraj/go-core/observability"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
)

var childLogger = log.With().Str("component","go-payment-authorizer").Str("package","internal.adapter.grpc.server").Logger()
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
	childLogger.Info().Str("func","AddPaymentToken").Interface("paymentRequest", paymentRequest).Send()

	// get span trace-id
	otel.SetTextMapPropagator(xray.Propagator{})
	md, _ := metadata.FromIncomingContext(ctx)
	ctx = otel.GetTextMapPropagator().Extract(ctx, go_core_observ.MetadataCarrier{md})

	// Trace
	span := tracerProvider.Span(ctx, "adpater.grpc.AddPaymentToken")
	defer span.End()

	// get request-id
	header, _ := metadata.FromIncomingContext(ctx)
	if len(header.Get("trace-request-id")) > 0 {
		ctx = context.WithValue(ctx, "trace-request-id", header.Get("trace-request-id")[0])
	}

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

	res_list_step_proto := []*proto.Step{}
	for _, v := range *res_payment.StepProcess {
		step_proto := proto.Step{StepProcess: v.Name,
								 ProcessedAt: timestamppb.New(v.ProcessedAt)}
		res_list_step_proto = append(res_list_step_proto, &step_proto)
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
									CreatedAt:  timestamppb.New(res_payment.CreatedAt),
							},	
		Steps: res_list_step_proto,						
	}

	childLogger.Info().Str("func","AddPaymentToken").Interface("=====> res_payment_proto_response", res_payment_proto_response).Send()

	return res_payment_proto_response, nil
}

