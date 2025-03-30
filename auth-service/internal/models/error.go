package models

type APIError struct {
	HTTPCode int    `json:"-"`
	Code     string `json:"Code"`
	Message  string `json:"message"`
}

func (e *APIError) Error() string {
	return e.Message
}

var (
	ErrEmailExists         = &APIError{409, "email_already_exists", "Email already registered"}
	ErrUserNotFound        = &APIError{404, "user_not_found", "User not found"}
	ErrInvalidPassword     = &APIError{401, "invalid_password", "Invalid password"}
	ErrExpiredRefreshToken = &APIError{401, "invalid_refresh_token", "Invalid refresh token"}
	ErrRevokedRefreshToken = &APIError{401, "revoked_refresh_token", "Revoked refresh token"}
	ErrInvalidRefreshToken = &APIError{401, "invalid_refresh_token", "Invalid refresh token"}
	ErrInvalidAccessToken  = &APIError{401, "invalid_access_token", "Invalid access token"}
	ErrUnauthorized        = &APIError{401, "user_unauthorized", "User unauthorized"}
	ErrInternal            = &APIError{500, "internal_error", "Internal server error"}

	MessageToError = map[string]*APIError{
		ErrEmailExists.Message:         ErrEmailExists,
		ErrUserNotFound.Message:        ErrUserNotFound,
		ErrInvalidPassword.Message:     ErrInvalidPassword,
		ErrExpiredRefreshToken.Message: ErrExpiredRefreshToken,
		ErrRevokedRefreshToken.Message: ErrRevokedRefreshToken,
		ErrInvalidRefreshToken.Message: ErrInvalidRefreshToken,
		ErrInvalidAccessToken.Message:  ErrInvalidAccessToken,
		ErrUnauthorized.Message:        ErrUnauthorized,
		ErrInternal.Message:            ErrInternal,
	}
)
