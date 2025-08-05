package storage

import (
	c "context"
	"errors"
	"l0/internal/models"
)

var (
	ErrOrderExists   = errors.New("order already exists")
	ErrOrderNotFound = errors.New("order not found")
)

type Storage interface {
	SaveOrder(c.Context, *models.Order) error
	GetOrder(c.Context, string) (*models.Order, error)
}
