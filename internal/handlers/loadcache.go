package handlers

import (
	"context"
	"fmt"
)

type Database interface {
	OrderGetter
	AllOrderUIDs(ctx context.Context) ([]string, error)
}

func LoadCache(ctx context.Context, cacher Cacher, db Database) error {
	const op = "handlers.LoadCache"
	uids, err := db.AllOrderUIDs(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	for _, uid := range uids {
		order, err := db.GetOrder(ctx, uid)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		err = cacher.SaveOrder(ctx, order)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	return nil
}
