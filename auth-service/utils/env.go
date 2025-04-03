package utils

import (
	"github.com/joho/godotenv"
	"log"
)

func InitEnv() {
	// Завантажуємо змінні з .env файла
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}
