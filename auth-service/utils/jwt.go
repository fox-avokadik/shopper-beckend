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

func GenerateAccessToken(user *models.User) (string, error) {
	expirationTime := AccessTokenExpiry()
	claims := &Claims{
		UserID: user.ID.String(),
		Email:  user.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func GenerateRefreshToken(userID uuid.UUID) (*models.RefreshToken, error) {
	return &models.RefreshToken{
		Token:     generateTokenString(),
		UserID:    userID,
		ExpiresAt: RefreshTokenExpiry(),
	}, nil

}

func generateTokenString() uuid.UUID {
	return uuid.New()
}
