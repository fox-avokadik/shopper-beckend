package models

import (
	"github.com/google/uuid"
	"time"
)

type RefreshToken struct {
	Token     uuid.UUID `json:"token"`
	UserID    uuid.UUID `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	IsRevoked bool      `json:"isRevoked"`
}
