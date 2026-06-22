package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateUserRequest struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
	Password    string `json:"password"`
}

type UpdateUserRequest struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Avatar        []string  `json:"avatar"`
	Email         string    `json:"email"`
	PhoneNumber   string    `json:"phone_number"`
	ProfileStatus string    `json:"profile_status"`
	Description   string    `json:"description"`
	RiskProfile   int       `json:"risk_profile"`
	IsPrivate     bool      `json:"is_private"`
}

type UpdatePasswordRequest struct {
	UserID      uuid.UUID `json:"user_id"`
	OldPassword string    `json:"old_password"`
	NewPassword string    `json:"new_password"`
}

type LogInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Device   string `json:"device"`
	Lat      string `json:"lat"`
	Long     string `json:"long"`
}

type LogInResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Jwt2FAToken  string `json:"jwt_2fa_token"`
}

type TwoFaRequest struct {
	Jwt2FAToken string `json:"jwt_2fa_token"`
	Code        string `json:"code"`
}

type LogoutRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Device string    `json:"device"`
}

type RecoverPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type PasswordResetToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Token     string    `json:"token"`
	Used      bool      `json:"used"`
	ExpiresAt time.Time `json:"expires_at"`
}

type User struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	Avatar           []string  `json:"avatar"`
	Email            string    `json:"email"`
	PhoneNumber      string    `json:"phone_number"`
	ProfileStatus    string    `json:"profile_status"`
	Description      string    `json:"description"`
	RiskProfile      int       `json:"risk_profile"`
	IsPrivate        bool      `json:"is_private"`
	IsActivated      bool      `json:"is_activated"`
	TwoFactorEnabled bool      `json:"two_factor_enabled"`
	PasswordHash     string    `json:"password_hash"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Session struct {
	ID               uuid.UUID `json:"id"`
	UserID           uuid.UUID `json:"user_id"`
	UserDevice       string    `json:"user_device"`
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	CreatedAt        time.Time `json:"created_at"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
	Locale           string    `json:"locale"`
	LocaleImgUrl     string    `json:"country_img_url"`
}
