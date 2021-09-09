package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"os"
	"phanpy/core"
)

func createLogger(env string) (*zap.Logger, error) {
	if env == "prod" {
		return zap.NewProduction()
	} else {
		return zap.NewDevelopment()
	}
}

func main() {
	ctx := context.Background()
	rootLogger, _ := createLogger(os.Getenv("ENV"))
	logger := rootLogger.Sugar()
	defer func() {
		err := rootLogger.Sync()
		if err != nil {
			fmt.Printf("Unable to close logger %v", err)
		}
	}()

	if err := core.InitDB(ctx, logger.Named("DB")); err != nil {
		logger.Panic(err)
	}
	defer core.CloseDB()

	listenAddress := ":8080"
	if len(os.Args) > 1 {
		listenAddress = os.Args[1]
	}

	logger.Info("Now booting server on address ", listenAddress)
	if err := http.ListenAndServe(listenAddress, core.Routes(logger.Named("routes"))); err != nil {
		logger.Error(err)
	}
}
