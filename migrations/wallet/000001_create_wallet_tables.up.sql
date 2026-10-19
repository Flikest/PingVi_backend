CREATE TABLE IF NOT EXISTS wallets (
    user_id      UUID PRIMARY KEY,
    evm_address  TEXT NOT NULL UNIQUE,
    sol_address  TEXT NOT NULL UNIQUE,
    btc_address  TEXT NOT NULL UNIQUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_cex_exchange_keys (
    user_id                 UUID        NOT NULL REFERENCES wallets(user_id) ON DELETE CASCADE,
    cex_name                VARCHAR(10) NOT NULL,
    encrypted_api_key       TEXT        NOT NULL,
    encrypted_secret_key    TEXT        NOT NULL,
    encrypted_passphrase    TEXT,
    encrypted_dek           TEXT        NOT NULL,
    at                      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, cex_name)
);

CREATE INDEX IF NOT EXISTS user_cex_exchange_keys_cex_name_idx
    ON user_cex_exchange_keys (cex_name);

CREATE TABLE IF NOT EXISTS snapshots (
    user_id UUID  NOT NULL REFERENCES wallets(user_id) ON DELETE CASCADE,
    account_type  VARCHAR(6) NOT NULL,
    balance_value NUMERIC NOT NULL,
    at            TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, at)
)


CREATE TABLE IF NOT EXISTS alchemy_keys (
    id SERIAL PRIMARY KEY,
    api_key TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS alchemy_keys_idx ON alchemy_keys (id);
