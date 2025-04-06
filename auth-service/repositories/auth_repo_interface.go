package repositories

import (
	"auth-service/models"
)

type AuthRepositoryInterface interface {
	Register(name, email, password string) (*models.AuthenticationResult, error)
	Login(email, password string) (*models.AuthenticationResult, error)
	RefreshToken(tokenString string) (*models.AuthenticationResult, error)
}
