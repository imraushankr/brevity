package repository

import (
	"context"
	"errors"
	"time"

	"github.com/imraushankr/brevity/server/src/internal/models"
	"github.com/imraushankr/brevity/server/src/internal/pkg/logger"
	"gorm.io/gorm"
)

type userRepository struct {
	db  *gorm.DB
	log logger.Logger
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db:  db,
		log: logger.Get(),
	}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	r.log.Debug("Creating new user", logger.String("email", user.Email))
	if err := user.Validate(); err != nil {
		r.log.Error("User validation failed", logger.NamedError("error", err))
		return err
	}

	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		r.log.Error("Failed to create user", logger.NamedError("error", err))
	}
	return err
}

func (r *userRepository) FindUser(ctx context.Context, identifier string) (*models.User, error) {
	r.log.Debug("Finding user by identifier", logger.String("identifier", identifier))

	var user models.User
	query := r.db.WithContext(ctx).Where("id = ? OR email = ? OR username = ?", identifier, identifier, identifier)
	err := query.First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		r.log.Debug("User not found", logger.String("identifier", identifier))
		return nil, models.ErrUserNotFound
	}
	if err != nil {
		r.log.Error("Failed to find user", logger.NamedError("error", err))
	}
	return &user, err
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	return r.FindUser(ctx, id)
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	return r.FindUser(ctx, email)
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	return r.FindUser(ctx, username)
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	r.log.Debug("Updating user", logger.String("userID", user.ID))
	err := r.db.WithContext(ctx).Save(user).Error
	if err != nil {
		r.log.Error("Failed to update user", logger.NamedError("error", err))
	}
	return err
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	r.log.Debug("Deleting user", logger.String("userID", id))
	err := r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error
	if err != nil {
		r.log.Error("Failed to delete user", logger.NamedError("error", err))
	}
	return err
}

func (r *userRepository) SaveVerificationToken(ctx context.Context, email, token string, expires time.Time) error {
	r.log.Debug("Saving verification token", logger.String("email", email))
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("email = ?", email).
		Updates(map[string]interface{}{
			"verification_token":   token,
			"verification_expires": expires,
		}).Error
	if err != nil {
		r.log.Error("Failed to save verification token", logger.NamedError("error", err))
	}
	return err
}

func (r *userRepository) VerifyUser(ctx context.Context, token string) error {
	r.log.Debug("Verifying user with token")
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("verification_token = ? AND verification_expires > ?", token, time.Now()).
		Updates(map[string]interface{}{
			"is_verified":          true,
			"verification_token":   nil,
			"verification_expires": nil,
		}).Error
	if err != nil {
		r.log.Error("Failed to verify user", logger.NamedError("error", err))
	}
	return err
}

func (r *userRepository) SaveResetToken(ctx context.Context, email, token string, expires time.Time) error {
	r.log.Debug("Saving reset token", logger.String("email", email))
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("email = ?", email).
		Updates(map[string]interface{}{
			"reset_password_token":   token,
			"reset_password_expires": expires,
		}).Error
	if err != nil {
		r.log.Error("Failed to save reset token", logger.NamedError("error", err))
	}
	return err
}

func (r *userRepository) ResetPassword(ctx context.Context, token, newPassword string) error {
	r.log.Debug("Resetting password with token")
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("reset_password_token = ? AND reset_password_expires > ?", token, time.Now()).
		Updates(map[string]interface{}{
			"password":               newPassword,
			"reset_password_token":   nil,
			"reset_password_expires": nil,
		}).Error
	if err != nil {
		r.log.Error("Failed to reset password", logger.NamedError("error", err))
	}
	return err
}

func (r *userRepository) UpdateAvatar(ctx context.Context, userID, avatarURL string) error {
	r.log.Debug("Updating user avatar",
		logger.String("userID", userID),
		logger.String("avatarURL", avatarURL))

	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("avatar", avatarURL).Error

	if err != nil {
		r.log.Error("Failed to update avatar",
			logger.NamedError("error", err),
			logger.String("userID", userID))
	}
	return err
}