package main

import (
	"auth-service/handlers"
	"auth-service/repositories"
	"auth-service/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	utils.InitEnv()

	db, err := utils.NewGormDB()
	if err != nil {
		panic("failed to connect database")
	}

	if err := utils.AutoMigrate(db); err != nil {
		panic("failed to migrate database")
	}

	authRepo := repositories.NewAuthRepository(db)

	authHandler := handlers.NewAuthHandler(authRepo)

	r := gin.Default()
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	r.POST("/refresh", authHandler.RefreshToken)

	errServer := r.Run(":8080")
	if errServer != nil {
		panic("failed to run server")
	}
}
