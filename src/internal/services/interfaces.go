package services

import (
	"context"
	"mime/multipart"

	"github.com/imraushankr/brevity/server/src/internal/models"
)

// UserService defines all user-related business operations
type UserService interface {
	// User Management
	Register(ctx context.Context, user *models.User) error
	Login(ctx context.Context, email, password string) (*models.User, string, error)
	FindUser(ctx context.Context, identifier string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id string) error

	// Email Verification
	VerifyEmail(ctx context.Context, token string) error

	// Password Management
	InitiatePasswordReset(ctx context.Context, email string) error
	CompletePasswordReset(ctx context.Context, token, newPassword string) error

	// Token Management
	RefreshToken(ctx context.Context, refreshToken string) (string, error)

	// Avatar Management
	UploadAvatar(ctx context.Context, userID string, file multipart.File, header *multipart.FileHeader) (string, error)
}