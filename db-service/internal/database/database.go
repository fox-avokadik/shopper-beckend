package database

import (
	"fmt"
	"sync"
	"time"

	"db-service/internal/config"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db   *gorm.DB
	once sync.Once
)

func InitDB() error {
	var initErr error
	once.Do(func() {
		err := godotenv.Load()
		if err != nil {
			initErr = fmt.Errorf("error loading .env file: %w", err)
			return
		}

		dsn := config.BuildDSN()

		db, initErr = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
			NowFunc: func() time.Time {
				return time.Now().UTC()
			},
		})

		if initErr != nil {
			initErr = fmt.Errorf("failed to connect to database: %w", initErr)
			return
		}

		sqlDB, err := db.DB()
		if err != nil {
			initErr = fmt.Errorf("failed to configure database pool: %w", err)
			return
		}

		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(25)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
	})

	return initErr
}

func GetDB() *gorm.DB {
	return db
}
