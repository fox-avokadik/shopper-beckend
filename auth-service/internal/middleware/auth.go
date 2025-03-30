package middleware

import (
	"auth-service/internal/models"
	"context"
	"strings"

	"auth-service/internal/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type AuthInterceptor struct {
	userRepo      *repository.UserRepository
	noAuthMethods map[string]bool
}

func NewAuthInterceptor(userRepo *repository.UserRepository) *AuthInterceptor {
	return &AuthInterceptor{
		userRepo: userRepo,
		noAuthMethods: map[string]bool{
			"/auth.AuthService/CreateUser":    true,
			"/auth.AuthService/UserAuthorize": true,
			"/auth.AuthService/RefreshTokens": true,
		},
	}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if i.noAuthMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		userID, err := i.authorize(ctx)
		if err != nil {
			return nil, err
		}

		ctx = context.WithValue(ctx, "user_id", userID)

		return handler(ctx, req)
	}
}

func (i *AuthInterceptor) authorize(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", models.ErrInternal
	}

	values := md["authorization"]
	if len(values) == 0 {
		return "", models.ErrUnauthorized
	}

	authHeader := values[0]
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", models.ErrInvalidAccessToken
	}

	token := parts[1]
	userID, err := i.userRepo.ValidateAccessToken(token)
	if err != nil {
		return "", models.ErrInvalidAccessToken
	}

	return userID, nil
}
