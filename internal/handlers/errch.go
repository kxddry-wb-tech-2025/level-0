package handlers

import (
	"context"
	"github.com/kxddry/go-utils/pkg/logger/handlers/sl"
	"log/slog"
)

func HandleErrors(ctx context.Context, log *slog.Logger, errCh <-chan error) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case err, ok := <-errCh:
				if !ok {
					return
				}
				log.Error("found error", sl.Err(err))
			}
		}
	}()
}
