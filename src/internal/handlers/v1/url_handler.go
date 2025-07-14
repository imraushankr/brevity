package v1

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imraushankr/brevity/server/src/internal/models"
	"github.com/imraushankr/brevity/server/src/internal/pkg/logger"
	"github.com/imraushankr/brevity/server/src/internal/services"
	"github.com/imraushankr/brevity/server/src/internal/utils"
)

type UserHandler struct {
	userService services.UserService
	log         logger.Logger
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		log:         logger.Get(),
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with email, username and password
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Register request"
// @Success 201 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /v1/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	startTime := time.Now()
	h.log.Info("Handling user registration request")

	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("Invalid registration request",
			logger.NamedError("error", err),
			logger.Any("request", c.Request.Body))
		utils.APIError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user := &models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	if err := h.userService.Register(c.Request.Context(), user); err != nil {
		switch {
		case errors.Is(err, models.ErrEmailAlreadyExists):
			h.log.Warn("Registration failed - email exists",
				logger.String("email", req.Email))
			utils.APIError(c, http.StatusConflict, "Email already exists")
		case errors.Is(err, models.ErrUsernameAlreadyExists):
			h.log.Warn("Registration failed - username exists",
				logger.String("username", req.Username))
			utils.APIError(c, http.StatusConflict, "Username already exists")
		default:
			h.log.Error("Registration failed",
				logger.NamedError("error", err),
				logger.String("email", req.Email))
			utils.APIError(c, http.StatusInternalServerError, "Registration failed")
		}
		return
	}

	h.log.Info("User registered successfully",
		logger.String("email", req.Email),
		logger.String("username", req.Username),
		logger.Duration("duration", time.Since(startTime)))

	utils.APISuccess(c, http.StatusCreated, models.MessageResponse{
		Message: "Registration successful. Please check your email for verification.",
	})
}

// Login godoc
// @Summary Login a user
// @Description Login with email and password
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login request"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /v1/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	startTime := time.Now()
	h.log.Info("Handling login request")

	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("Invalid login request",
			logger.NamedError("error", err))
		utils.APIError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user, accessToken, err := h.userService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrInvalidCredentials):
			h.log.Warn("Login failed - invalid credentials",
				logger.String("email", req.Email))
			utils.APIError(c, http.StatusUnauthorized, "Invalid credentials")
		case errors.Is(err, models.ErrAccountNotVerified):
			h.log.Warn("Login failed - account not verified",
				logger.String("email", req.Email))
			utils.APIError(c, http.StatusForbidden, "Account not verified")
		default:
			h.log.Error("Login failed",
				logger.NamedError("error", err),
				logger.String("email", req.Email))
			utils.APIError(c, http.StatusInternalServerError, "Login failed")
		}
		return
	}

	h.log.Info("Login successful",
		logger.String("email", req.Email),
		logger.String("userID", user.ID),
		logger.Duration("duration", time.Since(startTime)))

	utils.APISuccess(c, http.StatusOK, models.LoginResponse{
		User:         *user,
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	})
}

// GetUserProfile godoc
// @Summary Get user profile
// @Description Get user profile by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Security BearerAuth
// @Success 200 {object} models.UserProfileResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /v1/users/{id} [get]
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	startTime := time.Now()
	userID := c.Param("id")
	h.log.Info("Fetching user profile",
		logger.String("userID", userID))

	user, err := h.userService.FindUser(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			h.log.Warn("User not found",
				logger.String("userID", userID))
			utils.APIError(c, http.StatusNotFound, "User not found")
			return
		}
		h.log.Error("Failed to fetch user profile",
			logger.NamedError("error", err),
			logger.String("userID", userID))
		utils.APIError(c, http.StatusInternalServerError, "Failed to fetch user profile")
		return
	}

	h.log.Info("User profile fetched successfully",
		logger.String("userID", userID),
		logger.Duration("duration", time.Since(startTime)))

	utils.APISuccess(c, http.StatusOK, models.UserProfileResponse{User: *user})
}

// UpdateUserProfile godoc
// @Summary Update user profile
// @Description Update user profile information
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body models.UpdateProfileRequest true "Update request"
// @Security BearerAuth
// @Success 200 {object} models.UserProfileResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /v1/users/{id} [put]
func (h *UserHandler) UpdateUserProfile(c *gin.Context) {
	startTime := time.Now()
	userID := c.Param("id")
	h.log.Info("Updating user profile",
		logger.String("userID", userID))

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("Invalid update profile request",
			logger.NamedError("error", err),
			logger.String("userID", userID))
		utils.APIError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user := &models.User{
		ID:        userID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	if err := h.userService.UpdateUser(c.Request.Context(), user); err != nil {
		h.log.Error("Failed to update user profile",
			logger.NamedError("error", err),
			logger.String("userID", userID))
		utils.APIError(c, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	updatedUser, err := h.userService.FindUser(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("Failed to fetch updated user profile",
			logger.NamedError("error", err),
			logger.String("userID", userID))
		utils.APIError(c, http.StatusInternalServerError, "Failed to fetch updated profile")
		return
	}

	h.log.Info("User profile updated successfully",
		logger.String("userID", userID),
		logger.Duration("duration", time.Since(startTime)))

	utils.APISuccess(c, http.StatusOK, models.UserProfileResponse{User: *updatedUser})
}

// VerifyEmail godoc
// @Summary Verify user email
// @Description Verify user email with token
// @Tags users
// @Accept json
// @Produce json
// @Param token query string true "Verification token"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /v1/verify-email [get]
func (h *UserHandler) VerifyEmail(c *gin.Context) {
	startTime := time.Now()
	token := c.Query("token")
	h.log.Info("Verifying email",
		logger.String("token", token))

	if token == "" {
		h.log.Warn("Empty verification token")
		utils.APIError(c, http.StatusBadRequest, "Verification token is required")
		return
	}

	if err := h.userService.VerifyEmail(c.Request.Context(), token); err != nil {
		h.log.Error("Email verification failed",
			logger.NamedError("error", err),
			logger.String("token", token))
		utils.APIError(c, http.StatusBadRequest, "Invalid or expired verification token")
		return
	}

	h.log.Info("Email verified successfully",
		logger.Duration("duration", time.Since(startTime)))

	utils.APISuccess(c, http.StatusOK, models.MessageResponse{
		Message: "Email verified successfully",
	})
}

// InitiatePasswordReset godoc
// @Summary Initiate password reset
// @Description Initiate password reset process
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.PasswordResetRequest true "Password reset request"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /v1/password-reset [post]
func (h *UserHandler) InitiatePasswordReset(c *gin.Context) {
	startTime := time.Now()
	h.log.Info("Initiating password reset")

	var req models.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("Invalid password reset request",
			logger.NamedError("error", err))
		utils.APIError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.userService.InitiatePasswordReset(c.Request.Context(), req.Email); err != nil {
		h.log.Error("Failed to initiate password reset",
			logger.NamedError("error", err),
			logger.String("email", req.Email))
		utils.APIError(c, http.StatusInternalServerError, "Failed to initiate password reset")
		return
	}

	h.log.Info("Password reset initiated successfully",
		logger.String("email", req.Email),
		logger.Duration("duration", time.Since(startTime)))

	utils.APISuccess(c, http.StatusOK, models.MessageResponse{
		Message: "If the email exists, a password reset link has been sent",
	})
}

// CompletePasswordReset godoc
// @Summary Complete password reset
// @Description Complete password reset process
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.CompletePasswordResetRequest true "Complete password reset request"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /v1/password-reset/complete [post]
func (h *UserHandler) CompletePasswordReset(c *gin.Context) {
	startTime := time.Now()
	h.log.Info("Completing password reset")

	var req models.CompletePasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("Invalid complete password reset request",
			logger.NamedError("error", err))
		utils.APIError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.userService.CompletePasswordReset(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		h.log.Error("Failed to complete password reset",
			logger.NamedError("error", err))
		utils.APIError(c, http.StatusBadRequest, "Invalid or expired reset token")
		return
	}

	h.log.Info("Password reset completed successfully",
		logger.Duration("duration", time.Since(startTime)))

	utils.APISuccess(c, http.StatusOK, models.MessageResponse{
		Message: "Password reset successfully",
	})
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Refresh expired access token
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} models.RefreshTokenResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /v1/refresh-token [post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	startTime := time.Now()
	h.log.Info("Refreshing token")

	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("Invalid refresh token request",
			logger.NamedError("error", err))
		utils.APIError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	accessToken, err := h.userService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.log.Error("Failed to refresh token",
			logger.NamedError("error", err))
		utils.APIError(c, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	h.log.Info("Token refreshed successfully",
		logger.Duration("duration", time.Since(startTime)))

	utils.APISuccess(c, http.StatusOK, models.RefreshTokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   3600,
	})
}

// UploadAvatar godoc
// @Summary Upload user avatar
// @Description Upload or update user avatar image
// @Tags users
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "User ID"
// @Param avatar formData file true "Avatar image"
// @Security BearerAuth
// @Success 200 {object} models.UploadAvatarResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /v1/users/{id}/avatar [post]
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	startTime := time.Now()
	userID := c.Param("id")
	h.log.Info("Uploading avatar",
		logger.String("userID", userID))

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		h.log.Warn("Invalid avatar upload request",
			logger.NamedError("error", err),
			logger.String("userID", userID))
		utils.APIError(c, http.StatusBadRequest, "Avatar file is required")
		return
	}
	defer file.Close()

	avatarURL, err := h.userService.UploadAvatar(c.Request.Context(), userID, file, header)
	if err != nil {
		h.log.Error("Failed to upload avatar",
			logger.NamedError("error", err),
			logger.String("userID", userID))
		utils.APIError(c, http.StatusInternalServerError, "Failed to upload avatar")
		return
	}

	h.log.Info("Avatar uploaded successfully",
		logger.String("userID", userID),
		logger.String("avatarURL", avatarURL),
		logger.Duration("duration", time.Since(startTime)))

	utils.APISuccess(c, http.StatusOK, models.UploadAvatarResponse{
		AvatarURL: avatarURL,
	})
}