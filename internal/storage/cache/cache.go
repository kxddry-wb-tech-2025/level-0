package cache

import (
	"container/list"
	c "context"
	"l0/internal/models"
	"l0/internal/storage"
	"sync"
	"time"
)

type cacheEntry struct {
	order   models.Order
	time    time.Time
	lruElem *list.Element
}

type Cache struct {
	mp       map[string]*cacheEntry
	mu       *sync.Mutex
	ttl      time.Duration
	stopChan chan struct{}
	limit    int
	lru      *list.List
}

func NewCache(ttl time.Duration, limit int) *Cache {
	cc := &Cache{
		mp:       make(map[string]*cacheEntry),
		mu:       new(sync.Mutex),
		ttl:      ttl,
		stopChan: make(chan struct{}),
		limit:    limit,
		lru:      list.New(),
	}

	go cc.removeExpired()
	return cc
}

func (c *Cache) SaveOrder(ctx c.Context, order *models.Order) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.mp[order.OrderUID]; ok {
		entry.order = *order
		entry.time = time.Now()
		c.lru.MoveToFront(entry.lruElem)
		return nil
	}

	elem := c.lru.PushFront(order.OrderUID)
	c.mp[order.OrderUID] = &cacheEntry{
		order:   *order,
		time:    time.Now(),
		lruElem: elem,
	}

	if c.lru.Len() > c.limit {
		c.removeLRU()
	}

	return nil
}

func (c *Cache) removeLRU() {
	back := c.lru.Back()
	if back == nil {
		return
	}
	orderId := back.Value.(string)
	c.remove(orderId)
}

func (c *Cache) remove(orderId string) {
	entry, ok := c.mp[orderId]
	if !ok {
		return
	}
	c.lru.Remove(entry.lruElem)
	delete(c.mp, orderId)
}

func (c *Cache) GetOrder(ctx c.Context, orderId string) (*models.Order, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.mp[orderId]
	if !ok {
		return nil, storage.ErrOrderNotFound
	}

	// cache invalidation
	if time.Since(entry.time) > c.ttl {
		c.remove(orderId)
		return nil, storage.ErrOrderNotFound
	}

	c.lru.MoveToFront(entry.lruElem)

	return &entry.order, nil
}

func (c *Cache) LoadOrders(ctx c.Context, orders []*models.Order) error {
	for _, order := range orders {
		err := c.SaveOrder(ctx, order)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Cache) removeExpired() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			c.mu.Lock()
			for id, entry := range c.mp {
				if now.Sub(entry.time) > c.ttl {
					c.remove(id)
				}
			}
			c.mu.Unlock()
		case <-c.stopChan:
			return
		}

	}
}

func (c *Cache) Stop() {
	close(c.stopChan)
}
