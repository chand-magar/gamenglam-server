package models

import (
	"time"

	"github.com/google/uuid"
)

// ── User ──────────────────────────────────────────────────────────────────────

type User struct {
	ProfileNo    uint32     `json:"profile_no"`
	ProfileId    uuid.UUID  `json:"profile_id"`
	RoleNo       uint32     `json:"role_no"`
	UserFullName string     `json:"user_fullname"`
	EmailId      string     `json:"email_id"`
	Gender       *string    `json:"gender"`
	Dob          *time.Time `json:"dob"`
	MobileNo     *string    `json:"mobile_no"`
	Address      *string    `json:"address"`
	XApiKey      *string    `json:"x_api_key"`
	SecretKey    *string    `json:"secret_key"`
	Status       StatusEnum `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	CreatedBy    *uint32    `json:"created_by"`
	UpdatedAt    time.Time  `json:"updated_at"`
	UpdatedBy    *uint32    `json:"updated_by"`
}

// ── Credentials ───────────────────────────────────────────────────────────────

type Credentials struct {
	CredentialNo uint32     `json:"credential_no"`
	CredentialId uuid.UUID  `json:"credential_id"`
	ProfileNo    uint32     `json:"profile_no"`
	Username     string     `json:"username"`
	Password     string     `json:"-"`
	Status       StatusEnum `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	CreatedBy    *uint32    `json:"created_by"`
	UpdatedAt    time.Time  `json:"updated_at"`
	UpdatedBy    *uint32    `json:"updated_by"`
}

// ── Request / Response ────────────────────────────────────────────────────────
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	User         *User  `json:"user"`
}
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}
type ForgotPasswordRequest struct {
	EmailId string `json:"email_id" binding:"required,email"`
}
type VerifyOTPRequest struct {
	EmailId string `json:"email_id" binding:"required,email"`
	OTP     string `json:"otp"      binding:"required,len=6"`
}
type ResetPasswordRequest struct {
	EmailId     string `json:"email_id"     binding:"required,email"`
	OTP         string `json:"otp"          binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password"     binding:"required,min=8"`
}
