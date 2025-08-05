package cache

import (
	c "context"
	"l0/internal/models"
	"l0/internal/storage"
)

type Cache struct {
	mp map[string]models.Order
}

func NewCache() *Cache {
	return &Cache{mp: make(map[string]models.Order)}
}

func (c *Cache) SaveOrder(ctx c.Context, order *models.Order) error {
	c.mp[order.OrderUID] = *order
	return nil
}

func (c *Cache) GetOrder(ctx c.Context, orderId string) (*models.Order, error) {
	value, ok := c.mp[orderId]
	if !ok {
		return nil, storage.ErrOrderNotFound
	}
	return &value, nil
}
