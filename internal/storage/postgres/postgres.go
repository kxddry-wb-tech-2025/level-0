package postgres

import (
	c "context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"l0/internal/config"
	"l0/internal/models"
)

func fmterr(op string, err error) error {
	return fmt.Errorf("%s: %w", op, err)
}

type Storage struct {
	db *sql.DB
}

func (s *Storage) begin(ctx c.Context) (*sql.Tx, error) {
	return s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
}

func (s *Storage) SaveOrder(ctx c.Context, order *models.Order) (err error) {
	const op = "storage.postgres.SaveOrder"
	tx, err := s.begin(ctx)
	if err != nil {
		return fmterr(op, err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	// check if user exists
	var (
		uid uint
	)
	err = tx.QueryRow(`INSERT INTO users (customer_id, name, phone, email) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING RETURNING id`,
		order.CustomerID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Email).Scan(&uid)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRow(`SELECT id FROM users WHERE customer_id = $1`, order.CustomerID).Scan(&uid)
	}
	if err != nil {
		return fmterr(op, err)
	}

	// check address
	var addrId uint
	err = tx.QueryRow(`INSERT INTO addresses (customer_id, zip, city, address, region) VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING RETURNING id`,
		order.CustomerID, order.Delivery.Zip, order.Delivery.City, order.Delivery.Address, order.Delivery.Region).Scan(&addrId)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRow(`SELECT id FROM addresses WHERE customer_id = $1 AND zip = $2 AND city = $3 AND address = $4 AND region = $5`,
			order.CustomerID, order.Delivery.Zip, order.Delivery.City, order.Delivery.Address, order.Delivery.Region).Scan(&addrId)
	}
	if err != nil {
		return fmterr(op, err)
	}

	// check correlation between user and address
	_, err = tx.Exec(`INSERT INTO users_addresses (user_id, address_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, uid, addrId)
	if err != nil {
		return fmterr(op, err)
	}

	// create payment
	p := order.Payment
	_, err = tx.Exec(`INSERT INTO payments 
    (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) ON CONFLICT DO NOTHING`,
		p.Transaction, p.RequestID, p.Currency, p.Provider, p.Amount, p.PaymentDT, p.Bank, p.DeliveryCost, p.GoodsTotal, p.CustomFee)
	if err != nil {
		return fmterr(op, err)
	}

	// create order
	_, err = tx.Exec(`INSERT INTO orders 
    (order_uid, track_number, entry, delivery, payment, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) ON CONFLICT DO NOTHING`,
		order.OrderUID, order.TrackNumber, order.Entry, addrId, p.Transaction, order.Locale, order.InternalSignature, order.CustomerID, order.DeliveryService,
		order.ShardKey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		return fmterr(op, err)
	}

	// check all items for existence
	items := order.Items
	for _, item := range items {
		// ensure each item exists
		_, err = tx.Exec(`INSERT INTO items (nm_id, chrt_id, price, name, size, brand) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING`,
			item.NmID, item.ChrtID, item.Price, item.Name, item.Size, item.Brand)
		if err != nil {
			return fmterr(op, err)
		}

		_, err = tx.Exec(`INSERT INTO order_items (order_uid, item_id, rid, track_number, sale, total_price, status) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING`,
			order.OrderUID, item.NmID, item.RID, order.TrackNumber, item.Sale, item.TotalPrice, item.Status)
		if err != nil {
			return fmterr(op, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmterr(op, err)
	}
	return nil
}

func (s *Storage) GetOrder(ctx c.Context, orderUid string) (_ *models.Order, err error) {
	const op = "storage.postgres.GetOrder"
	order := models.Order{}
	tx, err := s.begin(ctx)
	if err != nil {
		return nil, fmterr(op, err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	var (
		deliveryId  uint
		transaction string
	)
	err = tx.QueryRow(`SELECT order_uid, track_number, entry, delivery, payment, locale, internal_signature,
       customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders WHERE order_uid = $1`, orderUid).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &deliveryId, &transaction, &order.Locale, &order.InternalSignature, &order.CustomerID, &order.DeliveryService, &order.ShardKey, &order.SmID, &order.DateCreated, &order.OofShard,
	)
	if err != nil {
		return nil, fmterr(op, err)
	}

	// payment
	err = tx.QueryRow(`SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		FROM payments WHERE transaction = $1`, transaction).Scan(
		&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency, &order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDT, &order.Payment.Bank,
		&order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee)
	if err != nil {
		return nil, fmterr(op, err)
	}

	// delivery
	var d models.Delivery
	err = tx.QueryRow(`SELECT u.name, u.phone, a.zip, a.city, a.address, a.region, u.email
				FROM addresses a
						 JOIN users u ON u.customer_id = a.customer_id
				WHERE a.id = $1;
`, deliveryId).Scan(&d.Name, &d.Phone, &d.Zip, &d.City, &d.Address, &d.Region, &d.Email)
	if err != nil {
		return nil, fmterr(op, err)
	}
	order.Delivery = d

	// get all items out
	items, err := tx.Query(`SELECT  i.chrt_id, oi.track_number, i.price, oi.rid, i.name, oi.sale, i.size, oi.total_price, i.nm_id, i.brand, oi.status
				FROM order_items oi
				JOIN items i ON i.nm_id = oi.item_id
				WHERE oi.track_number = $1;
	`, order.TrackNumber)
	if err != nil {
		return nil, fmterr(op, err)
	}
	defer items.Close()
	for items.Next() {
		item := models.Item{}
		err = items.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status)
		if err != nil {
			return nil, fmterr(op, err)
		}
		order.Items = append(order.Items, item)
	}
	if err := items.Err(); err != nil {
		return nil, fmterr(op, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmterr(op, err)
	}
	return &order, nil
}

func NewStorage(s config.Storage) (*Storage, error) {
	const op = "storage.postgres.NewStorage"
	url := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		s.User, s.Password, s.Host, s.Port, s.DBName, s.SSLMode,
	)
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, fmterr(op, err)
	}
	return &Storage{db: db}, db.Ping()
}
