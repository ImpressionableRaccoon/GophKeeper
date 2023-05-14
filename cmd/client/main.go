package main

import (
	"context"
	"crypto/rsa"
	"flag"
	"fmt"
	"os/signal"
	"strings"
	"syscall"

	"github.com/chzyer/readline"
	"go.uber.org/zap"

	"github.com/ImpressionableRaccoon/GophKeeper/internal/dataverse"
	"github.com/ImpressionableRaccoon/GophKeeper/internal/grpc/keeper"
	"github.com/ImpressionableRaccoon/GophKeeper/internal/keys"
	"github.com/ImpressionableRaccoon/GophKeeper/internal/service"
)

var (
	err    error
	logger *zap.Logger
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func init() {
	config := zap.NewProductionConfig()
	config.Encoding = "console"
	config.EncoderConfig.TimeKey = ""
	config.EncoderConfig.CallerKey = ""

	logger, err = config.Build()
	if err != nil {
		panic(fmt.Errorf("error create logger: %w", err))
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	logger.Info("build info",
		zap.String("version", buildVersion),
		zap.String("date", buildDate),
		zap.String("commit", buildCommit),
	)

	var (
		serverAddress string
		keyPath       string
	)

	flag.StringVar(&serverAddress, "a", "", "server address")
	flag.StringVar(&keyPath, "k", "", "user key path")

	flag.Parse()

	if serverAddress == "" {
		logger.Warn("server address not defined, run in offline mode (use -help for more info)")
	}

	var c *keeper.Client
	if serverAddress != "" {
		c, err = keeper.NewClient(serverAddress)
		if err != nil {
			logger.Panic("error create client", zap.Error(err))
		}
		logger.Info("server connection established")
		defer func() {
			closeErr := c.Close()
			if closeErr != nil {
				logger.Error("client close failed", zap.Error(closeErr))
				return
			}
			logger.Info("server connection closed")
		}()
	}

	if keyPath == "" {
		l, err = readline.NewEx(&readline.Config{Prompt: "Are you want to generate a new key [Y/n]: "})
		if err != nil {
			logger.Panic("init readline failed", zap.Error(err))
		}

		line, err = l.Readline()
		if err != nil || (err == nil && line != "Y" && line != "") {
			logger.Info("Use flag -k to specify RSA key file path")
			return
		}

		err = l.Close()
		if err != nil {
			logger.Panic("readline close failed", zap.Error(err))
		}
	}

	var privateKey *rsa.PrivateKey
	if keyPath == "" {
		var fileName string
		privateKey, fileName, err = keys.GenRSAKey(ctx)
		if err != nil {
			logger.Panic("gen new key failed", zap.Error(err))
		}
		logger.Info("key generated successfully", zap.String("file name", fileName))
	} else {
		privateKey, err = keys.LoadRSAKey(ctx, keyPath)
		if err != nil {
			logger.Panic("load key failed", zap.Error(err))
		}
		logger.Info("load key ok")
	}

	var s *service.Service
	s, err = service.New(c, privateKey)

	l, err = readline.NewEx(readlineCfg)
	if err != nil {
		logger.Panic("init readline failed", zap.Error(err))
	}
	defer func() {
		err = l.Close()
		if err != nil {
			logger.Error("readline close failed", zap.Error(err))
		}
	}()

	work(s)
}

func work(s *service.Service) {
	for {
		line, err = l.Readline()
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		var resp string
		switch {
		case line == "get":
			_, _ = fmt.Fprintln(l.Stderr(), "Usage: get {id}")
		case strings.HasPrefix(line, "get "):
			resp, err = s.Get(line[4:])
			if err != nil {
				logger.Error("get method returned an error", zap.Error(err))
				continue
			}
			_, _ = fmt.Fprintln(l.Stderr(), resp)
		case line == "add":
			_, _ = fmt.Fprint(l.Stderr(), "Usage: add {type}\n\n")
			_, _ = fmt.Fprintln(l.Stderr(), dataverse.Description)
		case strings.HasPrefix(line, "add "):
			resp, err = s.Add(line[4:])
			if err != nil {
				logger.Error("add method returned an error", zap.Error(err))
				continue
			}
			_, _ = fmt.Fprintln(l.Stderr(), resp)
		case line == "all":
			resp, err = s.All()
			if err != nil {
				logger.Error("all method returned an error", zap.Error(err))
				continue
			}
			_, _ = fmt.Fprintln(l.Stderr(), resp)
		case line == "delete":
			_, _ = fmt.Fprintln(l.Stderr(), "Usage: delete {id}")
		case strings.HasPrefix(line, "delete "):
			resp, err = s.Delete(line[7:])
			if err != nil {
				logger.Error("delete method returned an error", zap.Error(err))
				continue
			}
			_, _ = fmt.Fprintln(l.Stderr(), resp)
		default:
			usage(l.Stderr())
		}
	}
}
