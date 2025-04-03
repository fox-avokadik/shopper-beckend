package models

import (
	"github.com/google/uuid"
	"time"
)

type RefreshToken struct {
	Token     uuid.UUID `json:"token"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IsRevoked bool      `json:"is_revoked"`
}
