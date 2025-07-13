package models

import (
	"time"

	"github.com/teris-io/shortid"
)

var (
	urlSid, _ = shortid.New(1, shortid.DefaultABC, 3453)
)

type URL struct {
	ID          string     `json:"id" gorm:"primaryKey;type:varchar(20)"`
	OriginalURL string     `json:"original_url" validate:"required,url" gorm:"not null"`
	ShortCode   string     `json:"short_code" validate:"required,alphanum,min=3,max=10" gorm:"unique;not null"`
	UserID      string     `json:"user_id" gorm:"type:varchar(20);index"`
	User        User       `json:"-" gorm:"foreignKey:UserID"`
	Title       string     `json:"title" validate:"max=100"`
	Description string     `json:"description" validate:"max=255"`
	Clicks      int        `json:"clicks" gorm:"default:0"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `json:"-" gorm:"index"`
}

func (u *URL) BeforeCreate() error {
	id, err := urlSid.Generate()
	if err != nil {
		return err
	}
	u.ID = id
	return nil
}

type URLClick struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(20)"`
	URLID     string    `json:"url_id" gorm:"type:varchar(20);index"`
	URL       URL       `json:"-" gorm:"foreignKey:URLID"`
	IPAddress string    `json:"ip_address" gorm:"type:varchar(45)"`
	Referrer  string    `json:"referrer"`
	UserAgent string    `json:"user_agent"`
	Country   string    `json:"country" gorm:"type:varchar(2)"`
	City      string    `json:"city"`
	Device    string    `json:"device" gorm:"type:varchar(20)"`
	OS        string    `json:"os" gorm:"type:varchar(20)"`
	Browser   string    `json:"browser" gorm:"type:varchar(20)"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (uc *URLClick) BeforeCreate() error {
	id, err := urlSid.Generate()
	if err != nil {
		return err
	}
	uc.ID = id
	return nil
}

type CreateURLRequest struct {
	OriginalURL string     `json:"original_url" validate:"required,url"`
	CustomCode  string     `json:"custom_code" validate:"omitempty,alphanum,min=3,max=10"`
	Title       string     `json:"title" validate:"max=100"`
	Description string     `json:"description" validate:"max=255"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

type URLResponse struct {
	ID          string     `json:"id"`
	OriginalURL string     `json:"original_url"`
	ShortURL    string     `json:"short_url"`
	ShortCode   string     `json:"short_code"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Clicks      int        `json:"clicks"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (u *URL) Validate() error {
	return validate.Struct(u)
}

func (u *CreateURLRequest) Validate() error {
	return validate.Struct(u)
}

func (u *URL) ToResponse(baseURL string) *URLResponse {
	return &URLResponse{
		ID:          u.ID,
		OriginalURL: u.OriginalURL,
		ShortURL:    baseURL + "/" + u.ShortCode,
		ShortCode:   u.ShortCode,
		Title:       u.Title,
		Description: u.Description,
		Clicks:      u.Clicks,
		ExpiresAt:   u.ExpiresAt,
		IsActive:    u.IsActive,
		CreatedAt:   u.CreatedAt,
	}
}