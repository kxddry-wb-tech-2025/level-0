package handlers

import (
	"context"
	"l0/internal/kafka"
	"l0/internal/models"
	"log/slog"

	"github.com/kxddry/go-utils/pkg/logger/handlers/sl"
)

type OrderSaver interface {
	SaveOrder(context.Context, *models.Order) error
}

type Writer interface {
	Write(ctx context.Context, record models.Order) error
}

// HandleSaves saves orders incoming from a Kafka-like message channel,
// in case of an error sends the order to a Dead-Letter Queue (DLQ)
func HandleSaves(ctx context.Context, log *slog.Logger, saver OrderSaver, msgCh <-chan kafka.Message, dlq Writer, commit kafka.CommitFunc) <-chan error {
	const op = "handler.HandleSaves"
	log = log.With(slog.String("op", op))
	errCh := make(chan error)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgCh:
				if !ok {
					log.Info("closed channel")
					return
				}
				o := msg.Value
				log.Debug("got message", slog.String("uid", o.OrderUID))
				err := saver.SaveOrder(ctx, &o)
				if err != nil {
					log.Error("failed to save order", sl.Err(err))
					if ctx.Err() != nil {
						return
					}
					err2 := dlq.Write(ctx, o)
					if err2 != nil {
						log.Error("failed to send to dlq", sl.Err(err2))
						select {
						case errCh <- err2:
						case <-ctx.Done():
						}
					}
					select {
					case errCh <- err:
					case <-ctx.Done():
					}
					continue
				}
				if err := commit(ctx, msg.Raw); err != nil {
					log.Error("failed to commit", sl.Err(err))
				} else {
					log.Debug("saved order", slog.String("uid", o.OrderUID))
				}
			}
		}
	}()
	return errCh
}
