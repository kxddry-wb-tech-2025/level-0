package handlers

import (
	"context"
	"errors"
	"l0/internal/kafka"
	"l0/internal/models"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/kxddry/go-utils/pkg/logger/handlers/sl"
)

// OrderSaver can save orders
type OrderSaver interface {
	SaveOrder(context.Context, *models.Order) error
}

// Writer can write
type Writer interface {
	Write(ctx context.Context, record models.Order) error
}

// HandleSaves saves orders incoming from a Kafka-like message channel,
// in case of an error sends the order to a Dead-Letter Queue (DLQ)
func HandleSaves(ctx context.Context, log *slog.Logger, saver OrderSaver, msgCh <-chan kafka.Message, dlq Writer, commit kafka.CommitFunc, v *validator.Validate) <-chan error {
	const op = "handler.HandleSaves"
	log = log.With(slog.String("op", op))
	errCh := make(chan error, 100)
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

				err := v.Struct(o)
				if err != nil {
					var errs validator.ValidationErrors
					if errors.As(err, &errs) {
						for _, e := range errs {
							log.Error("validation error", sl.Err(e))
						}
					}
					log.Error("validation failed", sl.Err(err), slog.String("order_uid", o.OrderUID))
					select {
					case errCh <- err:
					default:
					}

					if err2 := dlq.Write(ctx, o); err2 != nil {
						log.Error("failed to send invalid order to dlq", sl.Err(err2), slog.String("order_uid", o.OrderUID))
						select {
						case errCh <- err2:
						default:
						}
					}

					if err3 := commit(ctx, msg.Raw); err3 != nil {
						log.Error("failed to commit offset after validation error", sl.Err(err3))
						select {
						case errCh <- err3:
						default:
						}
					}

					continue
				}

				err = saver.SaveOrder(ctx, &o)
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
