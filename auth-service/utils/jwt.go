package utils

import (
	"auth-service/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"os"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))

type Claims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	jwt.StandardClaims
}

func GenerateAccessToken(user *models.User) (string, int64, error) {
	expirationTime := AccessTokenExpiry()
	claims := &Claims{
		UserID: user.ID.String(),
		Email:  user.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	return tokenString, expirationTime.Unix(), err
}

func GenerateRefreshToken(userID uuid.UUID) (*models.RefreshToken, error) {
	expiresAt := RefreshTokenExpiry()
	return &models.RefreshToken{
		Token:     generateTokenString(),
		UserID:    userID,
		ExpiresAt: expiresAt,
	}, nil
}

func generateTokenString() uuid.UUID {
	return uuid.New()
}
