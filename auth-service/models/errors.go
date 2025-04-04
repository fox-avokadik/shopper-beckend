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
	ErrUserExists         = New("user_already_exists", "user already exists")
	ErrInvalidCredentials = New("invalid_credentials", "invalid credentials")
	ErrTokenExpired       = New("token_expired", "token expired")
	ErrTokenRevoked       = New("token_revoked", "token revoked")
	ErrTokenNotFound      = New("token_not_found", "token not found")
	ErrInvalidInput       = New("invalid_input", "invalid input data")
	ErrInternalServer     = New("internal_server_error", "internal server error")
)
