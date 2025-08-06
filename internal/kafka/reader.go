package kafka

import (
	c "context"
	"encoding/json"
	"l0/internal/config"
	"l0/internal/models"

	"github.com/segmentio/kafka-go"
)

type Reader struct {
	r *kafka.Reader
}

type Message struct {
	Value models.Order
	Raw   kafka.Message
}

type CommitFunc func(ctx c.Context, m kafka.Message) error

func NewReader(cfg config.ReaderConfig, brokers []string) Reader {
	var startOffset int64
	switch cfg.StartOffset {
	case "earliest":
		startOffset = kafka.FirstOffset
	case "latest":
		startOffset = kafka.LastOffset
	default:
		startOffset = kafka.LastOffset // fallback
	}

	return Reader{
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

func (r Reader) Messages(ctx c.Context) (<-chan Message, <-chan error, CommitFunc) {
	msgCh := make(chan Message)
	errCh := make(chan error, 1)
	commit := func(ctx c.Context, m kafka.Message) error {
		return r.r.CommitMessages(ctx, m)
	}

	go func() {
		defer close(msgCh)
		defer close(errCh)

		for {
			m, err := r.r.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				select {
				case errCh <- err:
				case <-ctx.Done():
				}
				return
			}

			var order models.Order
			if err := json.Unmarshal(m.Value, &order); err != nil {
				select {
				case errCh <- err:
				case <-ctx.Done():
				}
				continue
			}

			select {
			case msgCh <- Message{Value: order, Raw: m}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return msgCh, errCh, commit
}
