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
}

// NewUserService creates a new user service instance
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
	}
}

func (s *userService) Register(ctx context.Context, user *models.User) error {
	logger.Info("Registering new user", 
		logger.String("email", user.Email),
		logger.String("username", user.Username))

	// Check if email exists
	if _, err := s.userRepo.FindByEmail(ctx, user.Email); err == nil {
		logger.Warn("Email already exists", logger.String("email", user.Email))
		return models.ErrEmailAlreadyExists
	} else if !errors.Is(err, models.ErrUserNotFound) {
		logger.Error("Error checking email existence", logger.ErrorField(err))
		return fmt.Errorf("error checking email existence: %w", err)
	}

	// Check if username exists
	if _, err := s.userRepo.FindByUsername(ctx, user.Username); err == nil {
		logger.Warn("Username already exists", logger.String("username", user.Username))
		return models.ErrUsernameAlreadyExists
	} else if !errors.Is(err, models.ErrUserNotFound) {
		logger.Error("Error checking username existence", logger.ErrorField(err))
		return fmt.Errorf("error checking username existence: %w", err)
	}

	// Hash password
	hashedPassword, err := auth.EncryptPassword(user.Password)
	if err != nil {
		logger.Error("Password hashing failed", logger.ErrorField(err))
		return fmt.Errorf("password hashing failed: %w", err)
	}
	user.Password = hashedPassword

	// Create user
	if err := s.userRepo.Create(ctx, user); err != nil {
		logger.Error("User creation failed", 
			logger.ErrorField(err),
			logger.String("email", user.Email))
		return fmt.Errorf("user creation failed: %w", err)
	}

	// Generate and save verification token
	token, err := s.auth.GenerateVerificationToken(user.ID)
	if err != nil {
		logger.Error("Verification token generation failed", 
			logger.ErrorField(err),
			logger.String("userID", user.ID))
		return fmt.Errorf("verification token generation failed: %w", err)
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	if err := s.userRepo.SaveVerificationToken(ctx, user.Email, token, expiresAt); err != nil {
		logger.Error("Failed to save verification token", 
			logger.ErrorField(err),
			logger.String("email", user.Email))
		return fmt.Errorf("failed to save verification token: %w", err)
	}

	// Send verification email
	verificationLink := fmt.Sprintf("%s/verify-email?token=%s", s.cfg.App.BaseURL, token)
	if err := s.email.SendVerificationEmail(user.Email, verificationLink); err != nil {
		logger.Error("Failed to send verification email", 
			logger.ErrorField(err),
			logger.String("email", user.Email))
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	logger.Info("User registered successfully", 
		logger.String("email", user.Email),
		logger.String("userID", user.ID))
	return nil
}

func (s *userService) Login(ctx context.Context, email, password string) (*models.User, string, error) {
	logger.Info("Login attempt", logger.String("email", email))

	user, err := s.userRepo.FindByEmail(ctx, email)
	switch {
	case errors.Is(err, models.ErrUserNotFound):
		logger.Warn("User not found during login", logger.String("email", email))
		return nil, "", models.ErrInvalidCredentials
	case err != nil:
		logger.Error("Failed to find user during login", 
			logger.ErrorField(err),
			logger.String("email", email))
		return nil, "", fmt.Errorf("failed to find user: %w", err)
	}

	if !user.IsVerified {
		logger.Warn("Account not verified attempt", 
			logger.String("email", email),
			logger.String("userID", user.ID))
		return nil, "", models.ErrAccountNotVerified
	}

	if err := auth.IsPasswordCorrect(password, user.Password); err != nil {
		logger.Warn("Invalid password attempt", 
			logger.String("email", email),
			logger.String("userID", user.ID))
		return nil, "", models.ErrInvalidCredentials
	}

	accessToken, err := s.auth.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		logger.Error("Access token generation failed", 
			logger.ErrorField(err),
			logger.String("userID", user.ID))
		return nil, "", fmt.Errorf("access token generation failed: %w", err)
	}

	user.Sanitize()
	logger.Info("Login successful", 
		logger.String("email", email),
		logger.String("userID", user.ID))
	return user, accessToken, nil
}

func (s *userService) FindUser(ctx context.Context, identifier string) (*models.User, error) {
	logger.Debug("Finding user", logger.String("identifier", identifier))

	user, err := s.userRepo.FindUser(ctx, identifier)
	if err != nil {
		logger.Error("Failed to find user", 
			logger.ErrorField(err),
			logger.String("identifier", identifier))
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	user.Sanitize()
	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, user *models.User) error {
	logger.Info("Updating user", 
		logger.String("userID", user.ID),
		logger.String("email", user.Email))

	if err := user.Validate(); err != nil {
		logger.Warn("User validation failed", 
			logger.ErrorField(err),
			logger.String("userID", user.ID))
		return fmt.Errorf("validation failed: %w", err)
	}
	
	if err := s.userRepo.Update(ctx, user); err != nil {
		logger.Error("Failed to update user", 
			logger.ErrorField(err),
			logger.String("userID", user.ID))
		return err
	}
	
	logger.Info("User updated successfully", logger.String("userID", user.ID))
	return nil
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	logger.Info("Deleting user", logger.String("userID", id))

	if err := s.userRepo.Delete(ctx, id); err != nil {
		logger.Error("Failed to delete user", 
			logger.ErrorField(err),
			logger.String("userID", id))
		return err
	}
	
	logger.Info("User deleted successfully", logger.String("userID", id))
	return nil
}

func (s *userService) VerifyEmail(ctx context.Context, token string) error {
	logger.Info("Verifying email with token")

	if err := s.userRepo.VerifyUser(ctx, token); err != nil {
		logger.Error("Email verification failed", 
			logger.ErrorField(err),
			logger.String("token", token))
		return err
	}
	
	logger.Info("Email verified successfully")
	return nil
}

func (s *userService) InitiatePasswordReset(ctx context.Context, email string) error {
	logger.Info("Initiating password reset", logger.String("email", email))

	user, err := s.userRepo.FindByEmail(ctx, email)
	if errors.Is(err, models.ErrUserNotFound) {
		logger.Debug("Password reset requested for non-existent email", 
			logger.String("email", email))
		return nil // Don't reveal non-existent emails
	} else if err != nil {
		logger.Error("Failed to find user for password reset", 
			logger.ErrorField(err),
			logger.String("email", email))
		return fmt.Errorf("failed to find user: %w", err)
	}

	resetToken, err := s.auth.GeneratePasswordResetToken(user.ID)
	if err != nil {
		logger.Error("Reset token generation failed", 
			logger.ErrorField(err),
			logger.String("userID", user.ID))
		return fmt.Errorf("reset token generation failed: %w", err)
	}

	expiresAt := time.Now().Add(15 * time.Minute)
	if err := s.userRepo.SaveResetToken(ctx, user.Email, resetToken, expiresAt); err != nil {
		logger.Error("Failed to save reset token", 
			logger.ErrorField(err),
			logger.String("email", user.Email))
		return fmt.Errorf("failed to save reset token: %w", err)
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.cfg.App.BaseURL, resetToken)
	if err := s.email.SendPasswordResetEmail(user.Email, resetLink); err != nil {
		logger.Error("Failed to send password reset email", 
			logger.ErrorField(err),
			logger.String("email", user.Email))
		return err
	}

	logger.Info("Password reset initiated successfully", 
		logger.String("email", email),
		logger.String("userID", user.ID))
	return nil
}

func (s *userService) CompletePasswordReset(ctx context.Context, token, newPassword string) error {
	logger.Info("Completing password reset")

	hashedPassword, err := auth.EncryptPassword(newPassword)
	if err != nil {
		logger.Error("Password hashing failed during reset", logger.ErrorField(err))
		return fmt.Errorf("password hashing failed: %w", err)
	}

	if err := s.userRepo.ResetPassword(ctx, token, hashedPassword); err != nil {
		logger.Error("Password reset failed", 
			logger.ErrorField(err),
			logger.String("token", token))
		return err
	}

	logger.Info("Password reset completed successfully")
	return nil
}

func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	logger.Debug("Refreshing token")

	claims, err := s.auth.VerifyRefreshToken(refreshToken)
	if err != nil {
		logger.Error("Refresh token verification failed", logger.ErrorField(err))
		return "", fmt.Errorf("refresh token verification failed: %w", err)
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserId)
	if err != nil {
		logger.Error("Failed to find user during token refresh", 
			logger.ErrorField(err),
			logger.String("userID", claims.UserId))
		return "", fmt.Errorf("failed to find user: %w", err)
	}

	accessToken, err := s.auth.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		logger.Error("Access token generation failed during refresh", 
			logger.ErrorField(err),
			logger.String("userID", user.ID))
		return "", fmt.Errorf("access token generation failed: %w", err)
	}

	logger.Debug("Token refreshed successfully", logger.String("userID", user.ID))
	return accessToken, nil
}

func (s *userService) UploadAvatar(ctx context.Context, userID string, file multipart.File, header *multipart.FileHeader) (string, error) {
	logger.Info("Uploading avatar", logger.String("userID", userID))

	// Verify user exists
	if _, err := s.userRepo.FindByID(ctx, userID); err != nil {
		logger.Error("User verification failed for avatar upload", 
			logger.ErrorField(err),
			logger.String("userID", userID))
		return "", fmt.Errorf("user verification failed: %w", err)
	}

	avatarURL, err := s.storage.UploadFile(ctx, file, header, "avatars", userID)
	if err != nil {
		logger.Error("Avatar upload failed", 
			logger.ErrorField(err),
			logger.String("userID", userID))
		return "", fmt.Errorf("avatar upload failed: %w", err)
	}

	if err := s.userRepo.UpdateAvatar(ctx, userID, avatarURL); err != nil {
		logger.Error("Failed to update avatar URL", 
			logger.ErrorField(err),
			logger.String("userID", userID),
			logger.String("avatarURL", avatarURL))
		return "", fmt.Errorf("failed to update avatar URL: %w", err)
	}

	logger.Info("Avatar uploaded successfully", 
		logger.String("userID", userID),
		logger.String("avatarURL", avatarURL))
	return avatarURL, nil
}