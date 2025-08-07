package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"l0/internal/config"
	"l0/internal/kafka"
	"l0/internal/models"
	"log"
	"net/http"

	initCfg "github.com/kxddry/go-utils/pkg/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Sender is a service for uploading orders and sending them to Kafka
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var cfg config.Kafka
	initCfg.MustParseConfig(&cfg)
	e := echo.New()
	kw := kafka.NewWriter[models.Order](cfg.Writer, cfg.Brokers)

	e.POST("/save", func(c echo.Context) error {
		body := c.Request().Body
		bytes, _ := io.ReadAll(body)
		var order models.Order
		if err := json.Unmarshal(bytes, &order); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		err := kw.Write(ctx, order)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.String(http.StatusOK, "sent order")
	})

	e.GET("/save", func(c echo.Context) error {
		return c.String(http.StatusMethodNotAllowed, "use POST to send an order with JSON")
	})

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	if err := e.Start(":8085"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
