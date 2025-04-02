package client

import (
	"fmt"
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"
	"github.com/go-payment-authorizer/internal/core/model"

	"google.golang.org/grpc/metadata"
	go_core_observ 		"github.com/eliezerraj/go-core/observability"
	go_grpc_client "github.com/eliezerraj/go-core/grpc"	
	proto "github.com/go-payment-authorizer/protogen/token"

	//proto "github.com/eliezerraj/go-grpc-proto/protogen/token"
)

var childLogger = log.With().Str("component","go-payment-authorizer").Str("package","internal.adapter.grpc.client").Logger()
var tracerProvider go_core_observ.TracerProvider
var tokenServiceClient	proto.TokenServiceClient

type AdapterGrpc struct {
	grpcClientWorker	*go_grpc_client.GrpcClientWorker
	serviceClient		proto.TokenServiceClient
}

// About create a new worker service
func NewAdapterGrpc( grpcClientWorker	*go_grpc_client.GrpcClientWorker ) *AdapterGrpc{
	childLogger.Info().Str("func","NewAdapterGrpc").Send()

	// Create a client
	serviceClient := proto.NewTokenServiceClient(grpcClientWorker.GrcpClient)

	return &AdapterGrpc{
		grpcClientWorker: grpcClientWorker,
		serviceClient:	serviceClient,
	}
}

// About get gprc server information pod 
func (a *AdapterGrpc) GetCardTokenGrpc(ctx context.Context, card model.Card) (*[]model.Card, error){
	childLogger.Info().Str("func","GetCardTokenGrpc").Interface("trace-request-id", ctx.Value("trace-request-id")).Interface("card",card).Send()

	// Trace
	span := tracerProvider.Span(ctx, "adapter.GetCardTokenGrpc")
	defer span.End()
		
	// Prepare to receive proto data
	cardProto := proto.Card{ TokenData: card.TokenData}
	cardTokenRequest := &proto.CardTokenRequest{Card: &cardProto}

	// Set header for observability
	header := metadata.New(map[string]string{ "trace-request-id": fmt.Sprintf("%s",ctx.Value("trace-request-id")) })
	ctx = metadata.NewOutgoingContext(ctx, header)

	// request the data from grpc
	res_cardTokenResponse, err := a.serviceClient.GetCardToken(ctx, cardTokenRequest)
	if err != nil {
	  	return nil, err
	}

	// convert proto to json
	response_str, err := a.grpcClientWorker.ProtoToJSON(res_cardTokenResponse)
	if err != nil {
		return nil, err
  	}
		  
	// convert json to struct
	var res_protoJson map[string]interface{}
	err = json.Unmarshal([]byte(response_str), &res_protoJson)
	if err != nil {
		return nil, err
	}

	var list_cards []model.Card
	if _, ok := res_protoJson["cards"].([]interface{}); ok {
		for _, v := range res_protoJson["cards"].([]interface{}) {
			res_card := model.Card{}
			jsonString, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			json.Unmarshal(jsonString, &res_card)
			list_cards = append(list_cards, res_card)
		}
		
	} else {
		list_cards = append(list_cards, model.Card{})
	}
	return &list_cards, nil
}
