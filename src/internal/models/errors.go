package models

import "errors"

var (
	ErrInvalidEmail          = errors.New("invalid email format")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidInput          = errors.New("invalid input")
	ErrInvalidToken          = errors.New("invalid token")
	ErrExpiredToken          = errors.New("token has expired")
	ErrUnauthorized          = errors.New("unauthorized access")
	ErrForbidden             = errors.New("forbidden access")
	ErrTokenGenerationFailed = errors.New("failed to generate token")
	ErrAccountNotVerified    = errors.New("account not verified")
	ErrPasswordTooWeak       = errors.New("password is too weak")
	ErrPasswordMismatch      = errors.New("passwords do not match")
	ErrAvatarUploadFailed    = errors.New("failed to upload avatar")
)