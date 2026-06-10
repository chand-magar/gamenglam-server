package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/auth/models"
)

type AuthService interface {
	Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error)
	ForgotPassword(ctx context.Context, req *models.ForgotPasswordRequest) error
	VerifyOTP(ctx context.Context, req *models.VerifyOTPRequest) error
	ResetPassword(ctx context.Context, req *models.ResetPasswordRequest) error
	ChangePassword(ctx context.Context, profileId uuid.UUID, req *models.ChangePasswordRequest) error
}
