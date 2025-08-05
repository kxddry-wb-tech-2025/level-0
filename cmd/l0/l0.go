package main

import (
	initCfg "github.com/kxddry/go-utils/pkg/config"
	initLog "github.com/kxddry/go-utils/pkg/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"l0/internal/config"
	"l0/internal/handlers"
	"l0/internal/storage/cache"
	"l0/internal/storage/postgres"
	"net/http"
)

func main() {
	var cfg config.Config
	initCfg.MustParseConfig(&cfg)
	log := initLog.SetupLogger(cfg.Env)
	log.Debug("debug enabled")

	st, err := postgres.NewStorage(cfg.Storage)
	if err != nil {
		panic(st)
	}
	cacher := cache.NewCache()
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
}
