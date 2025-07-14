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

// package models

// import "errors"

// var (
// 	ErrUserNotFound            = errors.New("user not found")
// 	ErrEmailAlreadyExists      = errors.New("email already exists")
// 	ErrUsernameAlreadyExists   = errors.New("username already exists")
// 	ErrInvalidCredentials      = errors.New("invalid credentials")
// 	ErrAccountNotVerified      = errors.New("account not verified")
// 	ErrInvalidToken            = errors.New("invalid token")
// 	ErrTokenExpired            = errors.New("token expired")
// 	ErrPasswordTooWeak         = errors.New("password must be at least 8 characters long")
// 	ErrInvalidEmail            = errors.New("invalid email format")
// 	ErrInvalidUsername         = errors.New("username must be 3-30 characters and alphanumeric")
// 	ErrUnauthorized            = errors.New("unauthorized")
// 	ErrForbidden               = errors.New("forbidden")
// 	ErrInternalServerError     = errors.New("internal server error")
// 	ErrInvalidRequest          = errors.New("invalid request")
// 	ErrResourceNotFound        = errors.New("resource not found")
// 	ErrTooManyRequests         = errors.New("too many requests")
// 	ErrServiceUnavailable      = errors.New("service unavailable")
// 	ErrConflict                = errors.New("conflict")
// 	ErrUnprocessableEntity     = errors.New("unprocessable entity")
// 	ErrNotImplemented          = errors.New("not implemented")
// 	ErrBadGateway              = errors.New("bad gateway")
// 	ErrGatewayTimeout          = errors.New("gateway timeout")
// 	ErrHTTPVersionNotSupported = errors.New("http version not supported")
// )