package handlers

import (
	"auth-service/models"
	"auth-service/repositories"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type AuthHandler struct {
	authRepo repositories.AuthRepositoryInterface
}

func NewAuthHandler(authRepo repositories.AuthRepositoryInterface) *AuthHandler {
	return &AuthHandler{authRepo: authRepo}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrInvalidInput)
		return
	}

	user, accessToken, refreshToken, err := h.authRepo.Register(input.Name, input.Email, input.Password)
	if err != nil {
		c.JSON(getStatusCode(err), err)
		return
	}

	setRefreshTokenCookie(c, refreshToken)
	c.JSON(http.StatusCreated, gin.H{
		"user":        user,
		"accessToken": accessToken,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrInvalidInput)
		return
	}

	user, accessToken, refreshToken, err := h.authRepo.Login(input.Email, input.Password)
	if err != nil {
		c.JSON(getStatusCode(err), err)
		return
	}

	setRefreshTokenCookie(c, refreshToken)
	c.JSON(http.StatusOK, gin.H{
		"user":        user,
		"accessToken": accessToken,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	token, err := c.Cookie("refreshToken")
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.New("token_required", "refresh token required"))
		return
	}

	accessToken, newRefreshToken, err := h.authRepo.RefreshToken(token)
	if err != nil {
		c.JSON(getStatusCode(err), err)
		return
	}

	setRefreshTokenCookie(c, newRefreshToken)
	c.JSON(http.StatusOK, gin.H{
		"accessToken": accessToken,
	})
}

func setRefreshTokenCookie(c *gin.Context, token *models.RefreshToken) {
	unixExpiry := token.ExpiresAt.Unix()
	secondsUntilExpiry := int(unixExpiry - time.Now().Unix())

	c.SetCookie(
		"refreshToken",
		token.Token.String(),
		secondsUntilExpiry,
		"/",
		"",
		false,
		true,
	)
}

func getStatusCode(err error) int {
	switch err.(*models.AppError).Code {
	case "user_already_exists", "invalid_credentials":
		return http.StatusConflict
	case "token_not_found":
		return http.StatusNotFound
	case "token_revoked", "token_expired":
		return http.StatusForbidden
	case "invalid_input":
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
