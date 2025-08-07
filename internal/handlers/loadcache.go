package handlers

import (
	"context"
	"fmt"
	"l0/internal/models"
)

type Database interface {
	OrderGetter
	AllOrders(ctx context.Context) ([]*models.Order, error)
}

func LoadCache(ctx context.Context, cacher Cacher, db Database) error {
	const op = "handlers.LoadCache"

	orders, err := db.AllOrders(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = cacher.LoadOrders(ctx, orders)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
