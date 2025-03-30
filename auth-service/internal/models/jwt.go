package models

import (
	"time"

	"github.com/golang-jwt/jwt"
)

type JWTConfig struct {
	AccessSecret  string
	AccessExpire  time.Duration
	RefreshExpire time.Duration
}

var TokenConfig = JWTConfig{
	AccessSecret:  "your-secret-key",
	AccessExpire:  15 * time.Minute,
	RefreshExpire: 7 * 24 * time.Hour,
}

type AccessClaims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}
