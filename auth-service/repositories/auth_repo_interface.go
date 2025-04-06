package repositories

import (
	"auth-service/models"
)

type AuthRepositoryInterface interface {
	Register(name, email, password string) (*models.User, string, *models.RefreshToken, error)
	Login(email, password string) (*models.User, string, *models.RefreshToken, error)
	RefreshToken(tokenString string) (string, *models.RefreshToken, error)
}
