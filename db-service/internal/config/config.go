package config

import (
	"fmt"
	"os"
)

func BuildDSN() string {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	sslmode := os.Getenv("DB_SSLMODE")

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
