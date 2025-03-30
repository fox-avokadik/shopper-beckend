package service

import (
	"auth-service/internal/models"
	"auth-service/internal/models/requestes"
	"auth-service/internal/repository"
	authProto "auth-service/proto"
	"context"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(ctx context.Context, req *authProto.CreateUserRequest) (*authProto.CreateUserResponse, error) {
	hashedPassword, err := models.HashPassword(req.Password)
	if err != nil {
		return nil, models.ErrInternal
	}

	argv := &requestes.UserRegistrationData{
		Name:  req.Name,
		Email: req.Email,
	}

	return s.repo.CreateUser(ctx, argv, hashedPassword)
}

func (s *UserService) UserAuthorize(ctx context.Context, req *authProto.UserAuthorizationRequest) (*authProto.UserAuthorizationResponse, error) {
	argv := &requestes.UserAuthenticationData{
		Email:    req.Email,
		Password: req.Password,
	}

	return s.repo.UserAuthorize(ctx, argv)
}

func (s *UserService) RefreshTokens(ctx context.Context, req *authProto.RefreshTokenRequest) (*authProto.RefreshTokensResponse, error) {
	return s.repo.RefreshTokens(ctx, req.RefreshToken)
}
