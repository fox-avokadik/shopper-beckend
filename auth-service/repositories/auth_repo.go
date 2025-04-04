package repositories

import (
	"auth-service/models"
	"auth-service/utils"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"strings"
	"time"
)

type AuthRepository struct {
	DB *gorm.DB
}

var _ AuthRepositoryInterface = (*AuthRepository)(nil)

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{DB: db}
}

func (r *AuthRepository) Register(name, email, password string) (*models.User, string, string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", models.New("hashing_failed", "failed to hash password")
	}

	var user models.User
	err = r.DB.Raw(`
		INSERT INTO users (name, email, password_hash) 
		VALUES (?, ?, ?) 
		RETURNING id, name, email, created_at, updated_at`,
		name, email, string(passwordHash)).Scan(&user).Error

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, "", "", models.ErrUserExists
		}
		return nil, "", "", models.New("database_error", "failed to create user")
	}

	accessToken, err := utils.GenerateAccessToken(&user)
	if err != nil {
		return nil, "", "", models.New("token_generate_failed", "failed to generate access token")
	}

	refreshToken, err := r.generateAndStoreRefreshToken(user.ID)
	if err != nil {
		return nil, "", "", err
	}

	return &user, accessToken, refreshToken, nil
}

func (r *AuthRepository) Login(email, password string) (*models.User, string, string, error) {
	var user models.User
	err := r.DB.Raw(`
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users WHERE email = ?`, email).Scan(&user).Error

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", "", models.ErrInvalidCredentials
		}
		return nil, "", "", models.New("database_error", "failed to authenticate user")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", "", models.ErrInvalidCredentials
	}

	accessToken, err := utils.GenerateAccessToken(&user)
	if err != nil {
		return nil, "", "", models.New("token_generate_failed", "failed to generate access token")
	}

	refreshToken, err := r.generateAndStoreRefreshToken(user.ID)
	if err != nil {
		return nil, "", "", err
	}

	return &user, accessToken, refreshToken, nil
}

func (r *AuthRepository) RefreshToken(tokenString string) (string, string, error) {
	token, err := uuid.Parse(tokenString)
	if err != nil {
		return "", "", models.ErrInvalidInput
	}

	refreshToken, err := r.validateRefreshToken(token)
	if err != nil {
		return "", "", err
	}

	if err := r.revokeRefreshToken(token); err != nil {
		return "", "", err
	}

	var user models.User
	if err := r.DB.Raw("SELECT id, name, email FROM users WHERE id = ?", refreshToken.UserID).Scan(&user).Error; err != nil {
		return "", "", models.ErrInvalidCredentials
	}

	accessToken, err := utils.GenerateAccessToken(&user)
	if err != nil {
		return "", "", models.New("token_generate_failed", "failed to generate access token")
	}

	newRefreshToken, err := r.generateAndStoreRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

func (r *AuthRepository) generateAndStoreRefreshToken(userID uuid.UUID) (string, error) {
	token, _ := utils.GenerateRefreshToken(userID)

	if err := r.DB.Exec(`
		INSERT INTO refresh_tokens (token, user_id, expires_at) 
		VALUES (?, ?, ?)`,
		token.Token, token.UserID, token.ExpiresAt).Error; err != nil {
		return "", models.New("token_storage_failed", "failed to store refresh token")
	}

	return token.Token.String(), nil
}

func (r *AuthRepository) validateRefreshToken(token uuid.UUID) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	err := r.DB.Raw(`
		SELECT token, user_id, created_at, expires_at, is_revoked
		FROM refresh_tokens WHERE token = ?`, token).Scan(&refreshToken).Error

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrTokenNotFound
		}
		return nil, models.New("database_error", "failed to validate token")
	}

	if refreshToken.IsRevoked {
		return nil, models.ErrTokenRevoked
	}

	if time.Now().After(refreshToken.ExpiresAt) {
		return nil, models.ErrTokenExpired
	}

	return &refreshToken, nil
}

func (r *AuthRepository) revokeRefreshToken(token uuid.UUID) error {
	if err := r.DB.Exec(`
		UPDATE refresh_tokens 
		SET is_revoked = true 
		WHERE token = ?`, token).Error; err != nil {
		return models.New("token_revoke_failed", "failed to revoke token")
	}
	return nil
}
