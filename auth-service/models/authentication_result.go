package models

type AuthenticationResult struct {
	User              *User
	AccessToken       string
	AccessTokenExpiry int64
	RefreshToken      *RefreshToken
}

type AuthenticationResponse struct {
	User            *User  `json:"user,omitempty"`
	AccessToken     string `json:"accessToken"`
	AccessTokenExp  int64  `json:"accessTokenExp"`
	RefreshTokenExp int64  `json:"refreshTokenExp"`
}
