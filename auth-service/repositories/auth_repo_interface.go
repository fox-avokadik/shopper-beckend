package repositories

import (
	"auth-service/models"
)

// AuthRepositoryInterface визначає контракт для репозиторію аутентифікації
type AuthRepositoryInterface interface {
	Register(name, email, password string) (*models.User, string, string, error)
	Login(email, password string) (*models.User, string, string, error)
	RefreshToken(tokenString string) (string, string, error)
}
