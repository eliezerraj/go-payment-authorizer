package main

import(
	"fmt"
	"time"
	"context"
	
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/go-payment-authorizer/internal/infra/configuration"
	"github.com/go-payment-authorizer/internal/core/model"
	"github.com/go-payment-authorizer/internal/core/service"
	"github.com/go-payment-authorizer/internal/infra/server"
	"github.com/go-payment-authorizer/internal/adapter/database"
	go_grpc_client_worker "github.com/eliezerraj/go-core/grpc"	
	adapter_grpc_client "github.com/go-payment-authorizer/internal/adapter/grpc/client"

	go_core_pg "github.com/eliezerraj/go-core/database/pg"  
	grpc_adapter "github.com/go-payment-authorizer/internal/adapter/grpc/server"
)

var(
	logLevel = 	zerolog.InfoLevel // zerolog.InfoLevel zerolog.DebugLevel
	appServer	model.AppServer
	databaseConfig go_core_pg.DatabaseConfig
	databasePGServer go_core_pg.DatabasePGServer
	goCoreGrpcClientWorker go_grpc_client_worker.GrpcClientWorker
	childLogger = log.With().Str("component","go-payment-authorizer").Str("package", "main").Logger()
)

// About initialize the enviroment var
func init(){
	childLogger.Info().Str("func","init").Send()

	zerolog.SetGlobalLevel(logLevel)

	infoPod, server := configuration.GetInfoPod()
	configOTEL 		:= configuration.GetOtelEnv()
	databaseConfig 	:= configuration.GetDatabaseEnv()
	apiService 	:= configuration.GetEndpointEnv() 

	appServer.InfoPod = &infoPod
	appServer.Server = &server
	appServer.DatabaseConfig = &databaseConfig
	appServer.ConfigOTEL = &configOTEL
	appServer.ApiService = apiService
}

func main()  {
	childLogger.Info().Str("func","main").Interface("appServer :",appServer).Send()

	ctx := context.Background()

	// Open Database
	count := 1
	var err error
	for {
		databasePGServer, err = databasePGServer.NewDatabasePGServer(ctx, *appServer.DatabaseConfig)
		if err != nil {
			if count < 3 {
				childLogger.Error().Err(err).Msg("error open database... trying again !!")
			} else {
				childLogger.Error().Err(err).Msg("fatal error open Database aborting")
				panic(err)
			}
			time.Sleep(3 * time.Second) //backoff
			count = count + 1
			continue
		}
		break
	}

	// Open client GRPC channel
	goCoreGrpcClientWorker, err  := goCoreGrpcClientWorker.StartGrpcClient(appServer.ApiService[0].Url)
	if err != nil {
		childLogger.Error().Err(err).Msg(fmt.Sprintf("erro connect to grpc server : %v %v",appServer.ApiService[0].Name, appServer.ApiService[0].Url ))
		panic(3)
	}
	// test connection+
	err = goCoreGrpcClientWorker.TestConnection(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg(fmt.Sprintf("erro connect to grpc server : %v %v",appServer.ApiService[0].Name, appServer.ApiService[0].Url ))
		panic(3)
	} else {
		childLogger.Info().Msg("gprc channel openned sucessfull")
	}

	// create and wire
	adapterGrpcClient := adapter_grpc_client.NewAdapterGrpc(goCoreGrpcClientWorker)
	database := database.NewWorkerRepository(&databasePGServer)
	workerService := service.NewWorkerService(database, appServer.ApiService, adapterGrpcClient)
	adapterGrpc := grpc_adapter.NewAdapterGrpc(&appServer, workerService)
	workerServer := server.NewWorkerServer(adapterGrpc)

	// start grpc server
	workerServer.StartGrpcServer(ctx, &appServer)
}