package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/imraushankr/brevity/server/src/configs"
	"github.com/imraushankr/brevity/server/src/internal/pkg/logger"
)

// Storage defines the interface for file storage operations
type Storage interface {
	UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string, publicID string) (string, error)
	DeleteFile(ctx context.Context, publicID string) error
}

// CloudinaryStorage implements Storage for Cloudinary
type CloudinaryStorage struct {
	cld   *cloudinary.Cloudinary
	cfg   *configs.CloudinaryConfig
	local *LocalStorage
}

// LocalStorage implements Storage for local filesystem
type LocalStorage struct {
	uploadDir string
	baseURL   string
}

// NewStorage creates a new storage instance based on configuration
func NewStorage(cfg *configs.Config) (Storage, error) {
	// Always initialize local storage as fallback
	local := &LocalStorage{
		uploadDir: cfg.App.UploadDir,
		baseURL:   cfg.App.BaseURL,
	}

	// Initialize Cloudinary if configured
	if cfg.Cloudinary.CloudName != "" {
		cld, err := cloudinary.NewFromParams(
			cfg.Cloudinary.CloudName,
			cfg.Cloudinary.APIKey,
			cfg.Cloudinary.APISecret,
		)
		if err != nil {
			logger.Error("Failed to initialize Cloudinary, falling back to local storage",
				logger.ErrorField(err))
			return local, nil
		}

		logger.Info("Cloudinary storage initialized")
		return &CloudinaryStorage{
			cld:   cld,
			cfg:   &cfg.Cloudinary,
			local: local,
		}, nil
	}

	logger.Info("Using local storage")
	return local, nil
}

// UploadFile uploads a file to Cloudinary with fallback to local storage
func (cs *CloudinaryStorage) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string, publicID string) (string, error) {
	// Reset file reader after potential previous reads
	if _, err := file.Seek(0, 0); err != nil {
		return "", fmt.Errorf("failed to reset file reader: %w", err)
	}

	// Try uploading to Cloudinary
	uploadResult, err := cs.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:   folder,
		PublicID: publicID,
	})
	if err != nil {
		logger.Warn("Failed to upload to Cloudinary, falling back to local storage",
			logger.ErrorField(err))

		// Fallback to local storage
		return cs.local.UploadFile(ctx, file, header, folder, publicID)
	}

	logger.Info("File uploaded to Cloudinary",
		logger.String("url", uploadResult.SecureURL),
		logger.String("public_id", uploadResult.PublicID))

	return uploadResult.SecureURL, nil
}

// DeleteFile deletes a file from Cloudinary
func (cs *CloudinaryStorage) DeleteFile(ctx context.Context, publicID string) error {
	_, err := cs.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from Cloudinary: %w", err)
	}
	return nil
}

// UploadFile saves a file to the local filesystem
func (ls *LocalStorage) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string, publicID string) (string, error) {
	// Reset file reader after potential previous reads
	if _, err := file.Seek(0, 0); err != nil {
		return "", fmt.Errorf("failed to reset file reader: %w", err)
	}

	// Create folder if it doesn't exist
	fullPath := filepath.Join(ls.uploadDir, folder)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Create .gitkeep file if it doesn't exist
	gitKeepPath := filepath.Join(fullPath, ".gitkeep")
	if _, err := os.Stat(gitKeepPath); os.IsNotExist(err) {
		if _, err := os.Create(gitKeepPath); err != nil {
			logger.Warn("Failed to create .gitkeep file", logger.ErrorField(err))
		}
	}

	// Generate filename if publicID is empty
	if publicID == "" {
		publicID = fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(header.Filename))
	} else {
		publicID += filepath.Ext(header.Filename)
	}

	// Create destination file
	filePath := filepath.Join(fullPath, publicID)
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Generate URL
	url := fmt.Sprintf("%s/uploads/%s/%s", strings.TrimRight(ls.baseURL, "/"), folder, publicID)

	logger.Info("File saved locally",
		logger.String("path", filePath),
		logger.String("url", url))

	return url, nil
}

// DeleteFile removes a file from the local filesystem
func (ls *LocalStorage) DeleteFile(ctx context.Context, publicID string) error {
	filePath := filepath.Join(ls.uploadDir, publicID)
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete local file: %w", err)
	}
	return nil
}
