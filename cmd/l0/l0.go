package main

import (
	"context"
	"errors"
	"l0/internal/config"
	"l0/internal/handlers"
	"l0/internal/kafka"
	"l0/internal/storage/cache"
	"l0/internal/storage/postgres"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	initCfg "github.com/kxddry/go-utils/pkg/config"
	initLog "github.com/kxddry/go-utils/pkg/logger"
	"github.com/kxddry/go-utils/pkg/logger/handlers/sl"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var cfg config.Config
	initCfg.MustParseConfig(&cfg)
	log := initLog.SetupLogger(cfg.Env)
	log.Debug("debug enabled")

	st, err := postgres.NewStorage(cfg.Storage)
	if err != nil {
		panic(err)
	}
	cacher := cache.NewCache(cfg.Cache.TTL, cfg.Cache.Limit)

	err = handlers.LoadCache(ctx, cacher, st)
	if err != nil {
		panic(err)
	}

	kr := kafka.NewReader(cfg.Kafka.Reader, cfg.Kafka.Brokers)
	dlq := kafka.NewWriter(cfg.Kafka.Writer, cfg.Kafka.Brokers)
	msgCh, errCh, commitFunc := kr.Messages(ctx)
	saveErrCh := handlers.HandleSaves(ctx, log, st, msgCh, dlq, commitFunc)

	handlers.HandleErrors(ctx, log, errCh)
	handlers.HandleErrors(ctx, log, saveErrCh)

	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowMethods: []string{http.MethodGet}, // only GET allowed
	}))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{Timeout: cfg.Server.Timeout}))

	srv := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      e,
		ReadTimeout:  cfg.Server.Timeout,
		WriteTimeout: cfg.Server.Timeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	e.GET("/order/:id", handlers.GetOrderHandler(st, cacher))

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("failed to start", sl.Err(err))
	}

	// graceful shutdown
	down := make(chan os.Signal, 1)
	signal.Notify(down, os.Signal(syscall.SIGTERM), os.Signal(syscall.SIGINT))
	<-down
	log.Info("shutting down")

}
