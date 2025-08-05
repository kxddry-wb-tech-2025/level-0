package handlers

import (
	"context"
	"github.com/labstack/echo/v4"
	"l0/internal/models"
	"net/http"
)

type OrderGetter interface {
	GetOrder(context.Context, string) (*models.Order, error)
}
type OrderSaver interface {
	SaveOrder(context.Context, *models.Order) error
}

type Cacher interface {
	OrderGetter
	OrderSaver
}

func GetOrderHandler(getter OrderGetter, cacher Cacher) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		id := c.Param("id")
		if id == "" {
			return echo.ErrNotFound
		}

		cache, err := cacher.GetOrder(ctx, id)
		if err == nil {
			return c.JSON(http.StatusOK, cache)
		}

		order, err := getter.GetOrder(ctx, id)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		cacher.SaveOrder(ctx, order)
		return c.JSON(http.StatusOK, order)
	}
}
