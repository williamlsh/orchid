package auth

import "errors"

// Errors for token authentication.
var (
	ErrTokenInvalid             = errors.New("Invalid token")
	ErrTokenExpired             = errors.New("Token expired")
	ErrAccessTokenInvalid       = errors.New("Invalid access token")
	ErrAccessTokenExpired       = errors.New("Access token expired")
	ErrRefreshTokenInvalid      = errors.New("Invalid refresh token")
	ErrRefreshTokenExpired      = errors.New("Refresh token expired")
	ErrEmailInvalid             = errors.New("Invalid email")
	ErrPreviouslySignnedOutUser = errors.New("Previously signned out user")
)
