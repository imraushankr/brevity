package models

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/teris-io/shortid"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

var (
	validate = validator.New()
	sid, _   = shortid.New(1, shortid.DefaultABC, 2342)
)

type User struct {
	ID         string `json:"id" gorm:"primaryKey;type:varchar(20)"`
	FirstName  string `json:"first_name" validate:"required,min=2,max=50"`
	LastName   string `json:"last_name" validate:"required,min=2,max=50"`
	Username   string `json:"username" validate:"required,min=3,max=30,alphanum" gorm:"unique"`
	Role       Role   `json:"role" validate:"required,oneof=admin user manager" gorm:"type:varchar(20)"`
	Email      string `json:"email" validate:"required,email" gorm:"unique"`
	Phone      string `json:"phone" validate:"required,min=10,max=15"`
	Avatar     string `json:"avatar" gorm:"default:https://res.cloudinary.com/cloud-alpha/image/upload/v1739464346/Tours/user_oxe2tu.png"`
	Password   string `json:"-" validate:"required,min=8"`
	IsActive   bool   `json:"is_active" gorm:"default:true"`
	IsVerified bool   `json:"is_verified" gorm:"default:false"`

	VerificationToken   string     `json:"-" gorm:"type:varchar(255)"`
	VerificationExpires *time.Time `json:"-" gorm:"type:timestamp"`

	ResetPasswordToken   string     `json:"-" gorm:"type:varchar(255)"`
	ResetPasswordExpires *time.Time `json:"-" gorm:"type:timestamp"`

	RefreshToken string `json:"-" gorm:"-:all"`

	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `json:"-" gorm:"index"`
}

func (u *User) BeforeCreate() error {
	id, err := sid.Generate()
	if err != nil {
		return err
	}
	u.ID = id
	return nil
}

func (u *User) Validate() error {
	return validate.Struct(u)
}

func (u *User) Sanitize() {
	u.Password = ""
	u.RefreshToken = ""
	u.ResetPasswordToken = ""
	u.VerificationToken = ""
}

func (u *User) GenerateVerificationToken(token string, expires time.Time) {
	u.VerificationToken = token
	u.VerificationExpires = &expires
}

func (u *User) ClearVerificationToken() {
	u.VerificationToken = ""
	u.VerificationExpires = nil
}

func (u *User) GenerateResetToken(token string, expires time.Time) {
	u.ResetPasswordToken = token
	u.ResetPasswordExpires = &expires
}

func (u *User) ClearResetToken() {
	u.ResetPasswordToken = ""
	u.ResetPasswordExpires = nil
}