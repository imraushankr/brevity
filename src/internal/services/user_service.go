package services

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/imraushankr/brevity/server/src/configs"
	"github.com/imraushankr/brevity/server/src/internal/models"
	"github.com/imraushankr/brevity/server/src/internal/pkg/auth"
	"github.com/imraushankr/brevity/server/src/internal/pkg/email"
	"github.com/imraushankr/brevity/server/src/internal/pkg/logger"
	"github.com/imraushankr/brevity/server/src/internal/pkg/storage"
	"github.com/imraushankr/brevity/server/src/internal/repository"
)

// userService implements UserService interface
type userService struct {
	userRepo repository.UserRepository
	auth     *auth.Auth
	email    *email.EmailService
	cfg      *configs.Config
	storage  storage.Storage
	log      logger.Logger
}

// NewUserService creates a new user service instance
// func NewUserService(
// 	userRepo repository.UserRepository,
// 	auth *auth.Auth,
// 	email *email.EmailService,
// 	cfg *configs.Config,
// 	storage storage.Storage,
// ) UserService {
// 	return &userService{
// 		userRepo: userRepo,
// 		auth:     auth,
// 		email:    email,
// 		cfg:      cfg,
// 		storage:  storage,
// 		log:      logger.Get(),
// 	}
// }

func NewUserService(
	userRepo repository.UserRepository,
	auth *auth.Auth,
	email *email.EmailService,
	cfg *configs.Config,
	storage storage.Storage,
) UserService {
	return &userService{
		userRepo: userRepo,
		auth:     auth,
		email:    email,
		cfg:      cfg,
		storage:  storage,
		log:      logger.Get(),
	}
}

func (s *userService) Register(ctx context.Context, user *models.User) error {
	s.log.Info("Registering new user",
		logger.String("email", user.Email),
		logger.String("username", user.Username))

	// Check if email exists
	if _, err := s.userRepo.FindByEmail(ctx, user.Email); err == nil {
		s.log.Warn("Email already exists", logger.String("email", user.Email))
		return models.ErrEmailAlreadyExists
	} else if !errors.Is(err, models.ErrUserNotFound) {
		s.log.Error("Error checking email existence", logger.NamedError("error", err))
		return fmt.Errorf("error checking email existence: %w", err)
	}

	// Check if username exists
	if _, err := s.userRepo.FindByUsername(ctx, user.Username); err == nil {
		s.log.Warn("Username already exists", logger.String("username", user.Username))
		return models.ErrUsernameAlreadyExists
	} else if !errors.Is(err, models.ErrUserNotFound) {
		s.log.Error("Error checking username existence", logger.NamedError("error", err))
		return fmt.Errorf("error checking username existence: %w", err)
	}

	// Hash password
	hashedPassword, err := auth.EncryptPassword(user.Password)
	if err != nil {
		s.log.Error("Password hashing failed", logger.NamedError("error", err))
		return fmt.Errorf("password hashing failed: %w", err)
	}
	user.Password = hashedPassword

	// Create user
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.log.Error("User creation failed",
			logger.NamedError("error", err),
			logger.String("email", user.Email))
		return fmt.Errorf("user creation failed: %w", err)
	}

	// Generate and save verification token
	token, err := s.auth.GenerateVerificationToken(user.ID)
	if err != nil {
		s.log.Error("Verification token generation failed",
			logger.NamedError("error", err),
			logger.String("userID", user.ID))
		return fmt.Errorf("verification token generation failed: %w", err)
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	if err := s.userRepo.SaveVerificationToken(ctx, user.Email, token, expiresAt); err != nil {
		s.log.Error("Failed to save verification token",
			logger.NamedError("error", err),
			logger.String("email", user.Email))
		return fmt.Errorf("failed to save verification token: %w", err)
	}

	// Send verification email
	verificationLink := fmt.Sprintf("%s/api/v1/auth/verify-email?token=%s", s.cfg.App.BaseURL, token)
	if err := s.email.SendVerificationEmail(user.Email, verificationLink); err != nil {
		s.log.Error("Failed to send verification email",
			logger.NamedError("error", err),
			logger.String("email", user.Email))
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	s.log.Info("User registered successfully",
		logger.String("email", user.Email),
		logger.String("userID", user.ID))
	return nil
}

func (s *userService) Login(ctx context.Context, email, password string) (*models.User, string, error) {
	s.log.Info("Login attempt", logger.String("email", email))

	user, err := s.userRepo.FindByEmail(ctx, email)
	switch {
	case errors.Is(err, models.ErrUserNotFound):
		s.log.Warn("User not found during login", logger.String("email", email))
		return nil, "", models.ErrInvalidCredentials
	case err != nil:
		s.log.Error("Failed to find user during login",
			logger.NamedError("error", err),
			logger.String("email", email))
		return nil, "", fmt.Errorf("failed to find user: %w", err)
	}

	if !user.IsVerified {
		s.log.Warn("Account not verified attempt",
			logger.String("email", email),
			logger.String("userID", user.ID))
		return nil, "", models.ErrAccountNotVerified
	}

	if err := auth.IsPasswordCorrect(password, user.Password); err != nil {
		s.log.Warn("Invalid password attempt",
			logger.String("email", email),
			logger.String("userID", user.ID))
		return nil, "", models.ErrInvalidCredentials
	}

	accessToken, err := s.auth.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		s.log.Error("Access token generation failed",
			logger.NamedError("error", err),
			logger.String("userID", user.ID))
		return nil, "", fmt.Errorf("access token generation failed: %w", err)
	}

	user.Sanitize()
	s.log.Info("Login successful",
		logger.String("email", email),
		logger.String("userID", user.ID))
	return user, accessToken, nil
}

func (s *userService) FindUser(ctx context.Context, identifier string) (*models.User, error) {
	s.log.Debug("Finding user", logger.String("identifier", identifier))

	user, err := s.userRepo.FindUser(ctx, identifier)
	if err != nil {
		s.log.Error("Failed to find user",
			logger.NamedError("error", err),
			logger.String("identifier", identifier))
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	user.Sanitize()
	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, user *models.User) error {
	s.log.Info("Updating user",
		logger.String("userID", user.ID),
		logger.String("email", user.Email))

	if err := user.Validate(); err != nil {
		s.log.Warn("User validation failed",
			logger.NamedError("error", err),
			logger.String("userID", user.ID))
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.log.Error("Failed to update user",
			logger.NamedError("error", err),
			logger.String("userID", user.ID))
		return err
	}

	s.log.Info("User updated successfully", logger.String("userID", user.ID))
	return nil
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	s.log.Info("Deleting user", logger.String("userID", id))

	if err := s.userRepo.Delete(ctx, id); err != nil {
		s.log.Error("Failed to delete user",
			logger.NamedError("error", err),
			logger.String("userID", id))
		return err
	}

	s.log.Info("User deleted successfully", logger.String("userID", id))
	return nil
}

func (s *userService) VerifyEmail(ctx context.Context, token string) error {
	s.log.Info("Verifying email with token")

	if err := s.userRepo.VerifyUser(ctx, token); err != nil {
		s.log.Error("Email verification failed",
			logger.NamedError("error", err),
			logger.String("token", token))
		return err
	}

	s.log.Info("Email verified successfully")
	return nil
}

func (s *userService) InitiatePasswordReset(ctx context.Context, email string) error {
	s.log.Info("Initiating password reset", logger.String("email", email))

	user, err := s.userRepo.FindByEmail(ctx, email)
	if errors.Is(err, models.ErrUserNotFound) {
		s.log.Debug("Password reset requested for non-existent email",
			logger.String("email", email))
		return nil // Don't reveal non-existent emails
	} else if err != nil {
		s.log.Error("Failed to find user for password reset",
			logger.NamedError("error", err),
			logger.String("email", email))
		return fmt.Errorf("failed to find user: %w", err)
	}

	resetToken, err := s.auth.GeneratePasswordResetToken(user.ID)
	if err != nil {
		s.log.Error("Reset token generation failed",
			logger.NamedError("error", err),
			logger.String("userID", user.ID))
		return fmt.Errorf("reset token generation failed: %w", err)
	}

	expiresAt := time.Now().Add(15 * time.Minute)
	if err := s.userRepo.SaveResetToken(ctx, user.Email, resetToken, expiresAt); err != nil {
		s.log.Error("Failed to save reset token",
			logger.NamedError("error", err),
			logger.String("email", user.Email))
		return fmt.Errorf("failed to save reset token: %w", err)
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.cfg.App.BaseURL, resetToken)
	if err := s.email.SendPasswordResetEmail(user.Email, resetLink); err != nil {
		s.log.Error("Failed to send password reset email",
			logger.NamedError("error", err),
			logger.String("email", user.Email))
		return err
	}

	s.log.Info("Password reset initiated successfully",
		logger.String("email", email),
		logger.String("userID", user.ID))
	return nil
}

func (s *userService) CompletePasswordReset(ctx context.Context, token, newPassword string) error {
	s.log.Info("Completing password reset")

	hashedPassword, err := auth.EncryptPassword(newPassword)
	if err != nil {
		s.log.Error("Password hashing failed during reset", logger.NamedError("error", err))
		return fmt.Errorf("password hashing failed: %w", err)
	}

	if err := s.userRepo.ResetPassword(ctx, token, hashedPassword); err != nil {
		s.log.Error("Password reset failed",
			logger.NamedError("error", err),
			logger.String("token", token))
		return err
	}

	s.log.Info("Password reset completed successfully")
	return nil
}

func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	s.log.Debug("Refreshing token")

	claims, err := s.auth.VerifyRefreshToken(refreshToken)
	if err != nil {
		s.log.Error("Refresh token verification failed", logger.NamedError("error", err))
		return "", fmt.Errorf("refresh token verification failed: %w", err)
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserId)
	if err != nil {
		s.log.Error("Failed to find user during token refresh",
			logger.NamedError("error", err),
			logger.String("userID", claims.UserId))
		return "", fmt.Errorf("failed to find user: %w", err)
	}

	accessToken, err := s.auth.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		s.log.Error("Access token generation failed during refresh",
			logger.NamedError("error", err),
			logger.String("userID", user.ID))
		return "", fmt.Errorf("access token generation failed: %w", err)
	}

	s.log.Debug("Token refreshed successfully", logger.String("userID", user.ID))
	return accessToken, nil
}

func (s *userService) UploadAvatar(ctx context.Context, userID string, file multipart.File, header *multipart.FileHeader) (string, error) {
	s.log.Info("Uploading avatar", logger.String("userID", userID))

	// Verify user exists
	if _, err := s.userRepo.FindByID(ctx, userID); err != nil {
		s.log.Error("User verification failed for avatar upload",
			logger.NamedError("error", err),
			logger.String("userID", userID))
		return "", fmt.Errorf("user verification failed: %w", err)
	}

	avatarURL, err := s.storage.UploadFile(ctx, file, header, "avatars", userID)
	if err != nil {
		s.log.Error("Avatar upload failed",
			logger.NamedError("error", err),
			logger.String("userID", userID))
		return "", fmt.Errorf("avatar upload failed: %w", err)
	}

	if err := s.userRepo.UpdateAvatar(ctx, userID, avatarURL); err != nil {
		s.log.Error("Failed to update avatar URL",
			logger.NamedError("error", err),
			logger.String("userID", userID),
			logger.String("avatarURL", avatarURL))
		return "", fmt.Errorf("failed to update avatar URL: %w", err)
	}

	s.log.Info("Avatar uploaded successfully",
		logger.String("userID", userID),
		logger.String("avatarURL", avatarURL))
	return avatarURL, nil
}
