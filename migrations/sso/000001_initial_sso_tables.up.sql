CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    avatar TEXT[] DEFAULT ARRAY['https://i.pinimg.com/736x/4e/df/5b/4edf5b10482c6cf5c53bb6f0b919c3fa.jpg'],
    email VARCHAR(255) UNIQUE NOT NULL CHECK (email ~* '^[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$'),
    phone_number VARCHAR(17) UNIQUE NOT NULL,
    profile_status VARCHAR(100),
    description TEXT,
    risk_profile INT CHECK (risk_profile >= 0 AND risk_profile <= 100) DEFAULT 33,
    is_private BOOLEAN DEFAULT FALSE,
    is_activated BOOLEAN DEFAULT FALSE,
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_name ON users(name);
CREATE INDEX IF NOT EXISTS idx_user_email ON users(email);

CREATE TABLE IF NOT EXISTS verification_codes (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    verification_code VARCHAR(6) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP DEFAULT NOW() + INTERVAL '5 minutes',
    used_at TIMESTAMP,

    CONSTRAINT fk_verification_codes_users
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_verification_codes_user_id ON verification_codes(user_id);
CREATE INDEX IF NOT EXISTS idx_verification_codes_code ON verification_codes(verification_code);
CREATE INDEX IF NOT EXISTS idx_verification_codes_expires ON verification_codes(expires_at);

CREATE TABLE IF NOT EXISTS user_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    access_expires_at TIMESTAMP NOT NULL,
    refresh_expires_at TIMESTAMP NOT NULL,
    
    CONSTRAINT fk_user_tokens_users
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- Исправлено: индексы для таблицы user_tokens вместо sessions
CREATE INDEX IF NOT EXISTS idx_user_tokens_id ON user_tokens(id);
CREATE INDEX IF NOT EXISTS idx_user_tokens_user_id ON user_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_user_tokens_access_token ON user_tokens(access_token);
CREATE INDEX IF NOT EXISTS idx_user_tokens_refresh_token ON user_tokens(refresh_token);
CREATE INDEX IF NOT EXISTS idx_user_tokens_refresh_expires ON user_tokens(refresh_expires_at);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    used BOOLEAN DEFAULT FALSE,
    expires_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_password_reset_tokens_token ON password_reset_tokens(token);
CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);

CREATE TABLE IF NOT EXISTS passport_photos (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    photo_path VARCHAR(500) NOT NULL,
    uploaded_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT fk_passport_photos_users
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_passport_photos_user_id ON passport_photos(user_id);

-- Функции и триггеры
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
    
CREATE OR REPLACE FUNCTION cleanup_expired_verification_codes()
RETURNS void AS $$
BEGIN
    DELETE FROM verification_codes 
    WHERE expires_at < NOW() 
       OR used_at IS NOT NULL;
END;
$$ LANGUAGE plpgsql;

-- Исправлено: функция для user_tokens вместо sessions
CREATE OR REPLACE FUNCTION cleanup_expired_user_tokens()
RETURNS void AS $$
BEGIN
    DELETE FROM user_tokens 
    WHERE refresh_expires_at < NOW();
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION cleanup_unactivated_users()
RETURNS void AS $$
BEGIN
    DELETE FROM users 
    WHERE is_activated = false 
      AND created_at < NOW() - INTERVAL '2 hours';
END;
$$ LANGUAGE plpgsql;
