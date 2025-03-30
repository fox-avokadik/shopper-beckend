package handler

import (
	"auth-service/internal/service"
	authProto "auth-service/proto"
	"context"
)

type UserHandler struct {
	authProto.UnimplementedAuthServiceServer
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) CreateUser(ctx context.Context, req *authProto.CreateUserRequest) (*authProto.CreateUserResponse, error) {
	return h.service.CreateUser(ctx, req)
}

func (h *UserHandler) UserAuthorize(ctx context.Context, req *authProto.UserAuthorizationRequest) (*authProto.UserAuthorizationResponse, error) {
	return h.service.UserAuthorize(ctx, req)
}

func (h *UserHandler) RefreshTokens(ctx context.Context, req *authProto.RefreshTokenRequest) (*authProto.RefreshTokensResponse, error) {
	return h.service.RefreshTokens(ctx, req)
}
