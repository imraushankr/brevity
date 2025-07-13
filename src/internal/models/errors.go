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
)