package utils

import "time"

const (
	AccessTokenLifetime  = 15 * time.Minute
	RefreshTokenLifetime = 7 * 24 * time.Hour
)

func TokenExpiryTime(duration time.Duration) time.Time {
	return time.Now().Add(duration).UTC()
}

func AccessTokenExpiry() time.Time {
	return TokenExpiryTime(AccessTokenLifetime)
}

func RefreshTokenExpiry() time.Time {
	return TokenExpiryTime(RefreshTokenLifetime)
}
