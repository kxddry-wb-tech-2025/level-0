CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       customer_id VARCHAR(255) NOT NULL UNIQUE,
                       name VARCHAR(255) NOT NULL,
                       phone VARCHAR(16) NOT NULL UNIQUE,
                       email VARCHAR(255) NOT NULL UNIQUE
);
CREATE TABLE addresses (
                           id SERIAL PRIMARY KEY,
                           customer_id VARCHAR(255) REFERENCES users(customer_id),
                           zip          VARCHAR(20),
                           city         VARCHAR(100),
                           address      TEXT,
                           region       VARCHAR(100),
    UNIQUE(customer_id, zip, city, address, region)
);
CREATE TABLE users_addresses (
                                 user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
                                 address_id INTEGER REFERENCES addresses(id),
    PRIMARY KEY(user_id, address_id)
);

CREATE TABLE payments (
                          transaction VARCHAR(255) PRIMARY KEY,
                          request_id VARCHAR(255),
                          currency VARCHAR(3) NOT NULL,
                          provider VARCHAR(100) NOT NULL,
                          amount INTEGER NOT NULL,
                          payment_dt BIGINT NOT NULL,
                          bank VARCHAR(100) NOT NULL,
                          delivery_cost INTEGER,
                          goods_total INTEGER,
                          custom_fee INTEGER
);

CREATE TABLE orders (
                        order_uid         VARCHAR(255) PRIMARY KEY,
                        track_number      VARCHAR(255) NOT NULL UNIQUE,
                        entry             VARCHAR(50),
                        delivery INTEGER REFERENCES addresses(id) ON DELETE CASCADE,
                        payment VARCHAR(255) REFERENCES payments(transaction) ON DELETE CASCADE,
                        locale VARCHAR(8),
                        internal_signature TEXT,
                        customer_id VARCHAR(255) REFERENCES users(customer_id) ON DELETE CASCADE,
                        delivery_service VARCHAR(50),
                        shardkey TEXT,
                        sm_id INTEGER,
                        date_created TIMESTAMP NOT NULL,
                        oof_shard TEXT
);

CREATE TABLE items (
                       nm_id INTEGER PRIMARY KEY,
                       chrt_id INTEGER NOT NULL,
                       price INTEGER,
                       name VARCHAR(255),
                       size VARCHAR(50),
                       brand VARCHAR(255)
);


CREATE TABLE order_items (
                             order_uid VARCHAR(255) REFERENCES orders(order_uid) ON DELETE CASCADE,
                             item_id INTEGER REFERENCES items(nm_id),
                             rid VARCHAR(255) NOT NULL,
                             track_number VARCHAR REFERENCES orders(track_number),
                             sale INTEGER,
                             total_price INTEGER,
                             status INTEGER NOT NULL,
                             PRIMARY KEY(order_uid, item_id)
);
