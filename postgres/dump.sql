-- Активуйте розширення для генерації UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Створення таблиці users
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name TEXT NOT NULL,
  email TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Індекс для швидкого пошуку за email
CREATE INDEX idx_users_email ON users (email);

-- Опціонально: тригер для автоматичного оновлення updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Створення таблиці refresh_tokens
CREATE TABLE refresh_tokens (
  token UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
  is_revoked BOOLEAN DEFAULT FALSE
);

-- Індекс для швидкого пошуку за user_id
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);

-- Індекс для expires_at для швидкої перевірки застарілих токенів
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens (expires_at);

-- Активуйте розширення pg_cron
CREATE EXTENSION IF NOT EXISTS "pg_cron";

-- Автоматичне видалення записів, де is_revoked = TRUE, кожну годину
SELECT cron.schedule(
  '0 * * * *',  -- кожну годину
  $$DELETE FROM refresh_tokens WHERE is_revoked = TRUE$$
);
