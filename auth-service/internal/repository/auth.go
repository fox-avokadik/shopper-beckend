package repository

import (
	"auth-service/internal/models"
	"auth-service/internal/models/requestes"
	authProto "auth-service/proto"
	"context"
	dbProto "db-service/proto"
	"encoding/json"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"time"
)

type UserRepository struct {
	dbClient dbProto.DatabaseServiceClient
}

func NewUserRepository(dbClient dbProto.DatabaseServiceClient) *UserRepository {
	return &UserRepository{dbClient: dbClient}
}

func (r *UserRepository) CreateUser(ctx context.Context, argv *requestes.UserRegistrationData, hashedPassword string) (*authProto.CreateUserResponse, error) {
	query := `
	   INSERT INTO users (name, email, password_hash)
	   VALUES ($1, $2, $3)
	   RETURNING id, email, name
	`

	params := []*dbProto.QueryParam{
		{Value: &dbProto.QueryParam_StrValue{StrValue: argv.Name}},
		{Value: &dbProto.QueryParam_StrValue{StrValue: argv.Email}},
		{Value: &dbProto.QueryParam_StrValue{StrValue: hashedPassword}},
	}

	response, err := r.dbClient.ExecuteQuery(ctx, &dbProto.ExecuteQueryRequest{
		Query:  query,
		Params: params,
	})

	if err != nil {
		return nil, models.ErrInternal
	}

	if response.Error != "" {
		return nil, models.ErrEmailExists
	}

	var result struct {
		Id    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.Unmarshal([]byte(response.Result), &result); err != nil {
		return nil, models.ErrInternal
	}

	return &authProto.CreateUserResponse{
		Id:    result.Id,
		Name:  result.Name,
		Email: result.Email,
	}, nil
}

func (r *UserRepository) UserAuthorize(ctx context.Context, argv *requestes.UserAuthenticationData) (*authProto.UserAuthorizationResponse, error) {
	query := `
		SELECT id, password_hash
		FROM users WHERE email = $1
	`

	params := []*dbProto.QueryParam{
		{Value: &dbProto.QueryParam_StrValue{StrValue: argv.Email}},
	}

	response, err := r.dbClient.ExecuteQuery(ctx, &dbProto.ExecuteQueryRequest{
		Query:  query,
		Params: params,
	})

	if err != nil {
		return nil, models.ErrInternal
	}

	var result struct {
		Id       string `json:"id"`
		Password string `json:"password_hash"`
	}

	if err := json.Unmarshal([]byte(response.Result), &result); err != nil {
		return nil, models.ErrInternal
	}

	if err := models.ComparePassword(result.Password, argv.Password); err != nil {
		return nil, models.ErrInvalidPassword
	}

	accessToken, accessExp, err := r.generateAccessToken(result.Id)
	if err != nil {
		return nil, models.ErrInternal
	}

	refreshToken := uuid.New()
	refreshExp := time.Now().Add(models.TokenConfig.RefreshExpire).Unix()

	if err := r.storeRefreshToken(ctx, result.Id, refreshToken, refreshExp); err != nil {
		return nil, models.ErrInternal
	}

	return &authProto.UserAuthorizationResponse{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken.String(),
		AccessTokenExpiresAt:  accessExp,
		RefreshTokenExpiresAt: refreshExp,
	}, nil
}

func (r *UserRepository) storeRefreshToken(ctx context.Context, userID string, refreshToken uuid.UUID, expiresAt int64) error {
	query := `
		INSERT INTO refresh_tokens (token, user_id, expires_at)
		VALUES ($1, $2, $3)
	`

	params := []*dbProto.QueryParam{
		{Value: &dbProto.QueryParam_StrValue{StrValue: refreshToken.String()}},
		{Value: &dbProto.QueryParam_StrValue{StrValue: userID}},
		{Value: &dbProto.QueryParam_StrValue{StrValue: time.Unix(expiresAt, 0).Format(time.RFC3339)}},
	}

	_, err := r.dbClient.ExecuteQuery(ctx, &dbProto.ExecuteQueryRequest{
		Query:  query,
		Params: params,
	})

	return err
}

func (r *UserRepository) RefreshTokens(ctx context.Context, refreshToken string) (*authProto.RefreshTokensResponse, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, models.ErrUnauthorized
	}

	query := `
		SELECT user_id, expires_at, is_revoked 
		FROM refresh_tokens 
		WHERE token = $1 AND user_id = $2
	`

	params := []*dbProto.QueryParam{
		{Value: &dbProto.QueryParam_StrValue{StrValue: refreshToken}},
		{Value: &dbProto.QueryParam_StrValue{StrValue: userID}},
	}

	response, err := r.dbClient.ExecuteQuery(ctx, &dbProto.ExecuteQueryRequest{
		Query:  query,
		Params: params,
	})

	if err != nil {
		return nil, models.ErrInternal
	}

	var tokenData struct {
		UserID    string `json:"user_id"`
		ExpiresAt string `json:"expires_at"`
		IsRevoked bool   `json:"is_revoked"`
	}

	if err := json.Unmarshal([]byte(response.Result), &tokenData); err != nil {
		return nil, models.ErrInternal
	}

	if tokenData.UserID == "" {
		return nil, models.ErrInvalidRefreshToken
	}

	if tokenData.IsRevoked {
		return nil, models.ErrRevokedRefreshToken
	}

	expiresAt, err := time.Parse(time.RFC3339, tokenData.ExpiresAt)
	if err != nil {
		return nil, models.ErrInternal
	}

	if time.Now().After(expiresAt) {
		return nil, models.ErrExpiredRefreshToken
	}

	accessToken, accessExp, err := r.generateAccessToken(tokenData.UserID)
	if err != nil {
		return nil, models.ErrInternal
	}

	newRefreshToken := uuid.New()
	newRefreshExp := time.Now().Add(models.TokenConfig.RefreshExpire).Unix()

	if err := r.updateRefreshTokens(ctx, refreshToken, tokenData.UserID, newRefreshToken, newRefreshExp); err != nil {
		return nil, models.ErrInternal
	}

	return &authProto.RefreshTokensResponse{
		AccessToken:           accessToken,
		RefreshToken:          newRefreshToken.String(),
		AccessTokenExpiresAt:  accessExp,
		RefreshTokenExpiresAt: newRefreshExp,
	}, nil
}

func (r *UserRepository) updateRefreshTokens(ctx context.Context, oldToken string, userID string, newToken uuid.UUID, newExpiresAt int64) error {
	revokeQuery := `
		UPDATE refresh_tokens 
		SET is_revoked = true 
		WHERE token = $1
	`

	revokeParams := []*dbProto.QueryParam{
		{Value: &dbProto.QueryParam_StrValue{StrValue: oldToken}},
	}

	_, err := r.dbClient.ExecuteQuery(ctx, &dbProto.ExecuteQueryRequest{
		Query:  revokeQuery,
		Params: revokeParams,
	})
	if err != nil {
		return err
	}

	return r.storeRefreshToken(ctx, userID, newToken, newExpiresAt)
}

// Additions tools

func (r *UserRepository) generateAccessToken(userID string) (string, int64, error) {
	expiresAt := time.Now().Add(models.TokenConfig.AccessExpire).Unix()

	claims := &models.AccessClaims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(models.TokenConfig.AccessSecret))

	return accessToken, expiresAt, err
}

func (r *UserRepository) ValidateAccessToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(models.TokenConfig.AccessSecret), nil
	})

	if err != nil {
		return "", models.ErrInvalidAccessToken
	}

	claims, ok := token.Claims.(*models.AccessClaims)
	if !ok || !token.Valid {
		return "", models.ErrInvalidAccessToken
	}

	return claims.UserID, nil
}
