package utils

import (
	"github.com/joho/godotenv"
	"log"
)

func InitEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}
