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

// AuthRepository реалізує AuthRepositoryInterface
type AuthRepository struct {
	DB *gorm.DB
}

// Переконуємося, що AuthRepository реалізує інтерфейс
var _ AuthRepositoryInterface = (*AuthRepository)(nil)

// NewAuthRepository приймає *gorm.DB через DI
func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{DB: db}
}

// Register створює нового користувача
func (r *AuthRepository) Register(name, email, password string) (*models.User, string, string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", models.New("HASHING_FAILED", "failed to hash password")
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
		return nil, "", "", models.New("DATABASE_ERROR", "failed to create user")
	}

	accessToken, err := utils.GenerateAccessToken(&user)
	if err != nil {
		return nil, "", "", models.New("TOKEN_GENERATION_FAILED", "failed to generate access token")
	}

	refreshToken, err := r.generateAndStoreRefreshToken(user.ID)
	if err != nil {
		return nil, "", "", err
	}

	return &user, accessToken, refreshToken, nil
}

// Login автентифікує користувача
func (r *AuthRepository) Login(email, password string) (*models.User, string, string, error) {
	var user models.User
	err := r.DB.Raw(`
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users WHERE email = ?`, email).Scan(&user).Error

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", "", models.ErrInvalidCredentials
		}
		return nil, "", "", models.New("DATABASE_ERROR", "failed to authenticate user")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", "", models.ErrInvalidCredentials
	}

	accessToken, err := utils.GenerateAccessToken(&user)
	if err != nil {
		return nil, "", "", models.New("TOKEN_GENERATION_FAILED", "failed to generate access token")
	}

	refreshToken, err := r.generateAndStoreRefreshToken(user.ID)
	if err != nil {
		return nil, "", "", err
	}

	return &user, accessToken, refreshToken, nil
}

// RefreshToken оновлює токен доступу
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
		return "", "", models.New("USER_NOT_FOUND", "user not found")
	}

	accessToken, err := utils.GenerateAccessToken(&user)
	if err != nil {
		return "", "", models.New("TOKEN_GENERATION_FAILED", "failed to generate access token")
	}

	newRefreshToken, err := r.generateAndStoreRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

// Приватні методи
func (r *AuthRepository) generateAndStoreRefreshToken(userID uuid.UUID) (string, error) {
	token := &models.RefreshToken{
		Token:     uuid.New(),
		UserID:    userID,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := r.DB.Exec(`
		INSERT INTO refresh_tokens (token, user_id, expires_at) 
		VALUES (?, ?, ?)`,
		token.Token, token.UserID, token.ExpiresAt).Error; err != nil {
		return "", models.New("TOKEN_STORAGE_FAILED", "failed to store refresh token")
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
		return nil, models.New("DATABASE_ERROR", "failed to validate token")
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
		return models.New("TOKEN_REVOKE_FAILED", "failed to revoke token")
	}
	return nil
}
