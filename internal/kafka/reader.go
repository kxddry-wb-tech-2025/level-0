package kafka

import (
	c "context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"l0/internal/config"
	"l0/internal/models"
)

type Reader[T models.Order] struct {
	r *kafka.Reader
}

func NewReader[T models.Order](cfg config.ReaderConfig, brokers []string) Reader[T] {
	var startOffset int64
	switch cfg.StartOffset {
	case "earliest":
		startOffset = kafka.FirstOffset
	case "latest":
		startOffset = kafka.LastOffset
	default:
		startOffset = kafka.LastOffset // fallback
	}

	return Reader[T]{
		kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			GroupID:        cfg.GroupID,
			Topic:          cfg.Topic,
			MinBytes:       cfg.MinBytes,
			MaxBytes:       cfg.MaxBytes,
			CommitInterval: cfg.CommitInterval,
			StartOffset:    startOffset,
		}),
	}
}

func (r Reader[T]) Messages(ctx c.Context) (<-chan T, <-chan error) {
	msgCh := make(chan T)
	errCh := make(chan error)

	go func() {
		defer close(msgCh)
		defer close(errCh)

		for {
			m, err := r.r.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				errCh <- err
				return
			}

			var record T
			if err = json.Unmarshal(m.Value, &record); err != nil {
				errCh <- err
				continue
			}
			select {
			case <-ctx.Done():
				return
			case msgCh <- record:
				if err = r.r.CommitMessages(ctx, m); err != nil {
					errCh <- err
					return
				}
			}
		}
	}()

	return msgCh, errCh
}
