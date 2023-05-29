package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/ImpressionableRaccoon/GophKeeper/internal/grpc/interceptor"
	"github.com/ImpressionableRaccoon/GophKeeper/internal/grpc/keeper"
	"github.com/ImpressionableRaccoon/GophKeeper/internal/storage"
	pb "github.com/ImpressionableRaccoon/GophKeeper/proto"
)

var (
	logger *zap.Logger
	err    error
)

func init() {
	logger, err = zap.NewProduction(zap.AddStacktrace(zapcore.PanicLevel))
	if err != nil {
		panic(fmt.Errorf("error create logger: %w", err))
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	var (
		serverAddress string
		dsn           string
		tlsCert       string
		tlsKey        string
	)

	flag.StringVar(&serverAddress, "a", ":3200", "server address")
	flag.StringVar(&dsn, "d", "", "postgres dsn")
	flag.StringVar(&tlsCert, "c", "", "TLS server cert path")
	flag.StringVar(&tlsKey, "k", "", "TLS server key path")
	flag.Parse()

	if dsn == "" {
		dsn = os.Getenv("POSTGRES_DSN")
	}

	if serverAddress == "" || dsn == "" {
		flag.PrintDefaults()
		return
	}

	logger.Info("configuration",
		zap.String("serverAddress", serverAddress),
		zap.String("dsn", dsn),
		zap.String("tlsCert", tlsCert),
		zap.String("tlsKey", tlsKey),
	)

	var tlsCredentials credentials.TransportCredentials
	if tlsCert == "" || tlsKey == "" {
		logger.Warn("TLS certificates is empty, use it for security!")
	} else {
		tlsCredentials, err = loadTLSCredentials(tlsCert, tlsKey)
		if err != nil {
			logger.Panic("cannot load TLS credentials", zap.Error(err))
		}
	}

	var s *storage.ServerStorage
	s, err = storage.NewServerStorage(dsn)
	if err != nil {
		logger.Panic("error create server storage", zap.Error(err))
	}
	defer func() {
		closeErr := s.Close()
		if closeErr != nil {
			logger.Error("error closing storage", zap.Error(closeErr))
			return
		}
		logger.Info("storage closed")
	}()

	var ln net.Listener
	ln, err = net.Listen("tcp", serverAddress)
	if err != nil {
		logger.Panic("error listen server address", zap.Error(err))
	}

	var g *grpc.Server
	if tlsCredentials != nil {
		g = grpc.NewServer(
			grpc.Creds(tlsCredentials),
			grpc.ChainUnaryInterceptor(
				logging.UnaryServerInterceptor(
					interceptor.Logger(logger),
					logging.WithLogOnEvents(
						logging.StartCall,
						logging.FinishCall,
					),
				),
			),
		)
	} else {
		g = grpc.NewServer()
	}

	pb.RegisterKeeperServer(g, keeper.NewServer(s))
	go func() {
		logger.Info("starting server")
		serverErr := g.Serve(ln)
		if serverErr != nil {
			logger.Panic("grpc server failed", zap.Error(serverErr))
		}
	}()

	<-ctx.Done()
	logger.Info("ctx done")

	g.GracefulStop()
	logger.Info("grpc server stopped")
}

func loadTLSCredentials(cert, key string) (credentials.TransportCredentials, error) {
	var serverCert tls.Certificate
	serverCert, err = tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.NoClientCert,
	}

	return credentials.NewTLS(config), nil
}
