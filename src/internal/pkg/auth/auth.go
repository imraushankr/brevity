package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/imraushankr/brevity/server/src/configs"
	"github.com/imraushankr/brevity/server/src/internal/models"
)

type Auth struct {
	cfg *configs.JWTConfig
}

func NewAuth(cfg *configs.JWTConfig) *Auth {
	return &Auth{cfg: cfg}
}

type Claims struct {
	UserId string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (a *Auth) GenerateAccessToken(userId, role string) (string, error) {
	claims := &Claims{
		UserId: userId,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(a.cfg.AccessTokenExpiry) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    a.cfg.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.cfg.AccessTokenSecret))
}

func (a *Auth) GenerateRefreshToken(userId, role string) (string, error) {
	claims := &Claims{
		UserId: userId,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(a.cfg.RefreshTokenExpiry) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    a.cfg.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.cfg.RefreshTokenSecret))
}

func (a *Auth) GenerateTokens(userId, role string) (*Tokens, error) {
	accessToken, err := a.GenerateAccessToken(userId, role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := a.GenerateRefreshToken(userId, role)
	if err != nil {
		return nil, err
	}

	return &Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *Auth) GenerateVerificationToken(userId string) (string, error) {
	claims := &Claims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 hours expiry
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    a.cfg.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.cfg.AccessTokenSecret)) // Using access token secret for verification
}

func (a *Auth) GeneratePasswordResetToken(userId string) (string, error) {
	claims := &Claims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // 15 minutes expiry
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    a.cfg.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.cfg.ResetTokenSecret))
}

func (a *Auth) VerifyAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, models.ErrInvalidToken
		}
		return []byte(a.cfg.AccessTokenSecret), nil
	})

	if err != nil {
		if err == jwt.ErrTokenExpired {
			return nil, models.ErrExpiredToken
		}
		return nil, models.ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, models.ErrInvalidToken
}

func (a *Auth) VerifyRefreshToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, models.ErrInvalidToken
		}
		return []byte(a.cfg.RefreshTokenSecret), nil
	})

	if err != nil {
		if err == jwt.ErrTokenExpired {
			return nil, models.ErrExpiredToken
		}
		return nil, models.ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, models.ErrInvalidToken
}

func (a *Auth) VerifyPasswordResetToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, models.ErrInvalidToken
		}
		return []byte(a.cfg.ResetTokenSecret), nil
	})

	if err != nil {
		if err == jwt.ErrTokenExpired {
			return nil, models.ErrExpiredToken
		}
		return nil, models.ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, models.ErrInvalidToken
}