package handlers

import (
	"context"
	"l0/internal/models"
)

type OrderSaver interface {
	SaveOrder(context.Context, *models.Order) error
}

type Writer interface {
	Write(ctx context.Context, record models.Order) error
}

// HandleSaves saves orders incoming from a Kafka-like message channel,
// in case of an error sends the order to a Dead-Letter Queue (DLQ)
func HandleSaves(ctx context.Context, saver OrderSaver, msgCh <-chan models.Order, dlq Writer) <-chan error {
	errCh := make(chan error)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case o, ok := <-msgCh:
				if !ok {
					return
				}
				err := saver.SaveOrder(ctx, &o)
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					err2 := dlq.Write(ctx, o)
					if err2 != nil {
						select {
						case errCh <- err2:
						case <-ctx.Done():
						}
					}
					select {
					case errCh <- err:
					case <-ctx.Done():
					}
				}
			}
		}
	}()
	return errCh
}
