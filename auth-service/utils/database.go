package utils

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"time"
)

func BuildDSN() string {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	sslmode := os.Getenv("DB_SSLMODE")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	if sslmode == "" {
		sslmode = "disable"
	}

	if password != "" {
		return fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC connect_timeout=10",
			host, port, user, password, dbname, sslmode,
		)
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s sslmode=%s TimeZone=UTC connect_timeout=10",
		host, port, user, dbname, sslmode,
	)
}

func NewGormDB() (*gorm.DB, error) {
	dsn := BuildDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Exec(`
        DROP TRIGGER IF EXISTS update_users_updated_at ON users;
        DROP TRIGGER IF EXISTS token_expiry_trigger ON refresh_tokens;
        DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;
        DROP TRIGGER IF EXISTS update_user_roles_updated_at ON user_roles;
        DROP TRIGGER IF EXISTS user_default_role_trigger ON users;
    `).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("помилка видалення тригерів: %v", err)
	}

	if err := tx.Exec(`
        CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
        CREATE EXTENSION IF NOT EXISTS "pg_cron";
    `).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("помилка створення розширень: %v", err)
	}

	if err := tx.Exec(`
        CREATE TABLE IF NOT EXISTS roles (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            name TEXT UNIQUE NOT NULL,
            description TEXT,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );

        CREATE TABLE IF NOT EXISTS users (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            name TEXT NOT NULL,
            email TEXT UNIQUE NOT NULL,
            password_hash TEXT NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );

        CREATE TABLE IF NOT EXISTS user_roles (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(user_id, role_id)
        );
        
        CREATE TABLE IF NOT EXISTS refresh_tokens (
            token UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
            is_revoked BOOLEAN DEFAULT FALSE
        );
    `).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("помилка створення таблиць: %v", err)
	}

	if err := tx.Exec(`
        CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
        CREATE INDEX IF NOT EXISTS idx_roles_name ON roles (name);
        CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles (user_id);
        CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles (role_id);
        CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens (user_id);
        CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_revoked ON refresh_tokens (expires_at) 
            WHERE is_revoked = FALSE;
    `).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("помилка створення індексів: %v", err)
	}

	if err := tx.Exec(`
        CREATE OR REPLACE FUNCTION update_updated_at_column()
        RETURNS TRIGGER AS $func$
        BEGIN
            NEW.updated_at = CURRENT_TIMESTAMP;
            RETURN NEW;
        END;
        $func$ LANGUAGE plpgsql;
        
        CREATE OR REPLACE FUNCTION revoke_expired_tokens()
        RETURNS BIGINT AS $func$
        DECLARE
            tokens_revoked BIGINT;
        BEGIN
            WITH updated AS (
                UPDATE refresh_tokens
                SET is_revoked = TRUE
                WHERE expires_at < NOW() 
                AND is_revoked = FALSE
                RETURNING 1
            )
            SELECT COUNT(*) INTO tokens_revoked FROM updated;
            RETURN tokens_revoked;
        END;
        $func$ LANGUAGE plpgsql;
        
        CREATE OR REPLACE FUNCTION check_token_expiry()
        RETURNS TRIGGER AS $func$
        BEGIN
            IF NEW.expires_at < NOW() THEN
                NEW.is_revoked := TRUE;
            END IF;
            RETURN NEW;
        END;
        $func$ LANGUAGE plpgsql;

        CREATE OR REPLACE FUNCTION ensure_user_has_default_role()
        RETURNS TRIGGER AS $func$
        DECLARE
            user_role_id UUID;
            default_role_id UUID;
        BEGIN
            SELECT role_id INTO user_role_id FROM user_roles WHERE user_id = NEW.id LIMIT 1;
            
            IF user_role_id IS NULL THEN
                SELECT id INTO default_role_id FROM roles WHERE name = 'USER' LIMIT 1;
                
                IF default_role_id IS NOT NULL THEN
                    INSERT INTO user_roles (user_id, role_id)
                    VALUES (NEW.id, default_role_id);
                END IF;
            END IF;
            
            RETURN NEW;
        END;
        $func$ LANGUAGE plpgsql;

        CREATE OR REPLACE FUNCTION user_has_role(p_user_id UUID, p_role_name TEXT)
        RETURNS BOOLEAN AS $func$
        BEGIN
            RETURN EXISTS (
                SELECT 1
                FROM user_roles ur
                JOIN roles r ON ur.role_id = r.id
                WHERE ur.user_id = p_user_id
                AND r.name = p_role_name
            );
        END;
        $func$ LANGUAGE plpgsql;
    `).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("помилка створення функцій: %v", err)
	}

	if err := tx.Exec(`
        CREATE TRIGGER update_users_updated_at
            BEFORE UPDATE ON users
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
            
        CREATE TRIGGER update_roles_updated_at
            BEFORE UPDATE ON roles
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
        
        CREATE TRIGGER update_user_roles_updated_at
            BEFORE UPDATE ON user_roles
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
            
        CREATE TRIGGER token_expiry_trigger
            BEFORE INSERT OR UPDATE ON refresh_tokens
            FOR EACH ROW
            EXECUTE FUNCTION check_token_expiry();

        CREATE TRIGGER user_default_role_trigger
            AFTER INSERT ON users
            FOR EACH ROW
            EXECUTE FUNCTION ensure_user_has_default_role();
    `).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("помилка створення тригерів: %v", err)
	}

	if err := tx.Exec(`
        INSERT INTO roles (name, description)
        VALUES 
            ('ADMIN', 'Administrator with full system access'),
            ('USER', 'Regular user with limited access')
        ON CONFLICT (name) DO UPDATE 
        SET description = EXCLUDED.description,
            updated_at = CURRENT_TIMESTAMP;
    `).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("помилка додавання базових ролей: %v", err)
	}

	if err := tx.Exec(`
        DO $$
        BEGIN
            IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_cron') THEN
                DELETE FROM cron.job WHERE command IN (
                    'SELECT revoke_expired_tokens()',
                    'DELETE FROM refresh_tokens WHERE is_revoked = TRUE'
                );

                PERFORM cron.schedule(
                    'revoke-expired-tokens',
                    '*/30 * * * *',
                    'SELECT revoke_expired_tokens()'
                );

                PERFORM cron.schedule(
                    'cleanup-revoked-tokens',
                    '0 3 * * *',
                    'DELETE FROM refresh_tokens WHERE is_revoked = TRUE'
                );
            END IF;
        END $$;
    `).Error; err != nil {
		log.Printf("Попередження: не вдалося налаштувати pg_cron: %v", err)
	}

	return tx.Commit().Error
}
