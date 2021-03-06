package main

import (
	"context"
	"github.com/blendle/zapdriver"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/petomalina/xrpc/examples/api"
	"github.com/petomalina/xrpc/pkg/multiplexer"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"os"
)

func main() {
	ctx := context.Background()

	// create the zap logger for future use
	config := zapdriver.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger, err := config.Build(zapdriver.WrapCore(
		zapdriver.ReportAllErrors(true),
		zapdriver.ServiceName("Echo"),
	))
	if err != nil {
		panic(err)
	}

	// create and register the grpc server
	grpcServer := grpc.NewServer()
	echoSvc := &EchoService{logger}
	api.RegisterEchoServiceServer(grpcServer, echoSvc)
	reflection.Register(grpcServer)

	// create the grpc-gateway server and register to grpc server
	gwmux := runtime.NewServeMux()
	err = api.RegisterEchoServiceHandlerFromEndpoint(ctx, gwmux, ":"+os.Getenv("PORT"), []grpc.DialOption{grpc.WithInsecure()})
	if err != nil {
		logger.Fatal("gw: failed to register: %v", zap.Error(err))
	}

	// make multiplexer
	multiplexer.Make(nil,
		// filters all application/grpc messages into the grpc server
		multiplexer.GRPCHandler(grpcServer),
		// defaults all other messages into the http multiplexer
		multiplexer.HTTPHandler(gwmux),
	)
}

type EchoService struct {
	*zap.Logger
}

func (e *EchoService) Call(ctx context.Context, m *api.EchoMessage) (*api.EchoMessage, error) {
	headers := metautils.ExtractIncoming(ctx)

	e.Info("A new message was received",
		zap.String("method", "EchoService.SubmitCall"),
		zap.Any("headers", headers),
		zap.Any("message", m),
	)

	return m, nil
}
