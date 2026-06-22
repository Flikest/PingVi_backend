CREATE TABLE IF NOT EXISTS makers (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    rating BIGINT,
    
    CONSTRAINT fk_makers_users
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_makers_user_id ON makers(user_id);
CREATE INDEX IF NOT EXISTS idx_makers_rating ON makers(rating);

CREATE TABLE IF NOT EXISTS takers (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    rating BIGINT,
    
    CONSTRAINT fk_takers_users
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_takers_user_id ON takers(user_id);
CREATE INDEX IF NOT EXISTS idx_takers_rating ON takers(rating);

CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY,
    maker_id UUID NOT NULL,
    asset VARCHAR(6) NOT NULL,
    fiat CHAR(1)[] NOT NULL,
    qty DECIMAL(15, 2) NOT NULL,
    price DECIMAL(15, 2) NOT NULL,
    min_price DECIMAL(15, 2) NOT NULL,
    max_price DECIMAL(15, 2) NOT NULL,
    payment_methods TEXT[] NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP,

    CONSTRAINT fk_orders_makers
        FOREIGN KEY (maker_id)
        REFERENCES makers(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_orders_maker_id ON orders(maker_id);
CREATE INDEX IF NOT EXISTS idx_orders_asset ON orders(asset);
CREATE INDEX IF NOT EXISTS idx_orders_fiat ON orders USING gin(fiat);
CREATE INDEX IF NOT EXISTS idx_orders_price ON orders(price);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);
CREATE INDEX IF NOT EXISTS idx_orders_asset_fiat ON orders(asset, fiat);

CREATE TABLE IF NOT EXISTS bids (
    id UUID PRIMARY KEY,
    taker_id UUID NOT NULL,
    order_id UUID NOT NULL,
    qty DECIMAL(15, 2) NOT NULL,
    fiat VARCHAR(3) NOT NULL,
    total DECIMAL(15, 2) NOT NULL,
    payment_methods TEXT[] NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT fk_bids_takers
        FOREIGN KEY (taker_id)
        REFERENCES takers(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_bids_orders
        FOREIGN KEY (order_id)
        REFERENCES orders(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_bids_taker_id ON bids(taker_id);
CREATE INDEX IF NOT EXISTS idx_bids_order_id ON bids(order_id);
CREATE INDEX IF NOT EXISTS idx_bids_total ON bids(total);
CREATE INDEX IF NOT EXISTS idx_bids_created_at ON bids(created_at);
