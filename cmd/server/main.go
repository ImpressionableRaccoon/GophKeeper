package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os/signal"
	"syscall"

	"github.com/ImpressionableRaccoon/GophKeeper/internal/grpc/keeper"
	"github.com/ImpressionableRaccoon/GophKeeper/internal/storage"
	pb "github.com/ImpressionableRaccoon/GophKeeper/proto"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

var logger *zap.Logger

func init() {
	var err error
	logger, err = zap.NewProduction(zap.AddStacktrace(zapcore.PanicLevel))
	if err != nil {
		panic(fmt.Errorf("error create logger: %w", err))
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	serverAddress := ""
	dsn := ""

	flag.StringVar(&serverAddress, "a", ":3200", "server address")
	flag.StringVar(&dsn, "d", "", "postgres dsn")
	flag.Parse()

	if serverAddress == "" || dsn == "" {
		flag.PrintDefaults()
		return
	}

	logger.Info("configuration",
		zap.String("serverAddress", serverAddress),
		zap.String("dsn", dsn),
	)

	s, err := storage.NewServerStorage(dsn)
	if err != nil {
		panic(fmt.Errorf("error create server storage: %w", err))
	}
	defer func() {
		closeErr := s.Close()
		if closeErr != nil {
			logger.Error("error closing storage", zap.Error(closeErr))
			return
		}
		logger.Info("storage closed")
	}()

	ln, err := net.Listen("tcp", serverAddress)
	if err != nil {
		panic(fmt.Errorf("error listen server address: %w", err))
	}

	g := grpc.NewServer()
	pb.RegisterKeeperServer(g, keeper.NewServer(s))
	go func() {
		logger.Info("starting server")
		if g.Serve(ln) != nil {
			panic(err)
		}
	}()

	<-ctx.Done()
	logger.Info("ctx done")

	g.GracefulStop()
	logger.Info("grpc server stopped")
}
