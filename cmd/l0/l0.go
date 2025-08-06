package main

import (
	"context"
	"errors"
	initCfg "github.com/kxddry/go-utils/pkg/config"
	initLog "github.com/kxddry/go-utils/pkg/logger"
	"github.com/kxddry/go-utils/pkg/logger/handlers/sl"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"l0/internal/config"
	"l0/internal/handlers"
	"l0/internal/kafka"
	"l0/internal/models"
	"l0/internal/storage/cache"
	"l0/internal/storage/postgres"
	"net/http"
)

func main() {
	ctx := context.Background()
	var cfg config.Config
	initCfg.MustParseConfig(&cfg)
	log := initLog.SetupLogger(cfg.Env)
	log.Debug("debug enabled")

	st, err := postgres.NewStorage(cfg.Storage)
	if err != nil {
		panic(err)
	}
	cacher := cache.NewCache()

	err = handlers.LoadCache(ctx, cacher, st)
	if err != nil {
		panic(err)
	}

	kr := kafka.NewReader[models.Order](cfg.Kafka.Reader, cfg.Kafka.Brokers)
	dlq := kafka.NewWriter[models.Order](cfg.Kafka.Writer, cfg.Kafka.Brokers)
	msgCh, errCh := kr.Messages(ctx)
	saveErrCh := handlers.HandleSaves(ctx, st, msgCh, dlq)

	handlers.HandleErrors(ctx, log, errCh)
	handlers.HandleErrors(ctx, log, saveErrCh)

	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	e.GET("/order/:id", handlers.GetOrderHandler(st, cacher))

	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("failed to start", sl.Err(err))
	}
}
