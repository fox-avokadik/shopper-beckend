package models

import "fmt"

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func New(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

var (
	ErrUserExists         = New("USER_EXISTS", "user already exists")
	ErrInvalidCredentials = New("INVALID_CREDENTIALS", "invalid credentials")
	ErrTokenExpired       = New("TOKEN_EXPIRED", "token expired")
	ErrTokenRevoked       = New("TOKEN_REVOKED", "token revoked")
	ErrTokenNotFound      = New("TOKEN_NOT_FOUND", "token not found")
	ErrInvalidInput       = New("INVALID_INPUT", "invalid input data")
	ErrInternalServer     = New("INTERNAL_ERROR", "internal server error")
)
