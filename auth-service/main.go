package main

import (
	"auth-service/handlers"
	"auth-service/repositories"
	"auth-service/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	utils.InitEnv()

	// Ініціалізація бази даних
	db, err := utils.NewGormDB()
	if err != nil {
		panic("failed to connect database")
	}

	// Автоміграція таблиць
	if err := utils.AutoMigrate(db); err != nil {
		panic("failed to migrate database")
	}

	// Створення репозиторію з DI
	authRepo := repositories.NewAuthRepository(db)

	// Ініціалізація хендлерів з DI
	authHandler := handlers.NewAuthHandler(authRepo)

	// Налаштування маршрутів Gin
	r := gin.Default()
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	r.POST("/refresh", authHandler.RefreshToken)

	// Запуск сервера
	r.Run(":8080")
}
