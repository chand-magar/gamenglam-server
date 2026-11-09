package controllers

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/auth/models"
	"github.com/chand-magar/gamenglam-server/internal/module/auth/services"
)

type AuthHandler struct {
	svc services.AuthService
}

func NewAuthHandler(svc services.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.svc.Login(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			ErrorResponse(c, http.StatusUnauthorized, "Invalid username or password.")
		case errors.Is(err, services.ErrUserNotActive):
			ErrorResponse(c, http.StatusForbidden, "Account is inactive. Contact administrator.")
		default:
			ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		}
		return
	}

	SuccessResponse(c, resp, "Login successful.")
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.ForgotPassword(c.Request.Context(), &req); err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	// Always return success — don't reveal if email exists
	SuccessResponse(c, nil, "If this email is registered, an OTP has been sent.")
}

func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req models.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.VerifyOTP(c.Request.Context(), &req); err != nil {
		if errors.Is(err, services.ErrInvalidOTP) {
			ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired OTP.")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	SuccessResponse(c, nil, "OTP verified successfully.")
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.ResetPassword(c.Request.Context(), &req); err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidOTP):
			ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired OTP.")
		case errors.Is(err, services.ErrUserNotFound):
			ErrorResponse(c, http.StatusNotFound, "User not found.")
		default:
			ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		}
		return
	}

	SuccessResponse(c, nil, "Password reset successfully.")
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	profileId, err := extractProfileIdFromToken(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, "Unauthorized.")
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), profileId, &req); err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidPassword):
			ErrorResponse(c, http.StatusUnauthorized, "Current password is incorrect.")
		case errors.Is(err, services.ErrUserNotFound):
			ErrorResponse(c, http.StatusNotFound, "User not found.")
		default:
			ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		}
		return
	}

	SuccessResponse(c, nil, "Password changed successfully.")
}

// ── JWT extraction ────────────────────────────────────────────────────────────

func extractProfileIdFromToken(c *gin.Context) (uuid.UUID, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return uuid.Nil, errors.New("missing token")
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "changeme-use-env-in-production"
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid claims")
	}

	// Reject refresh tokens on protected endpoints
	if tt, ok := claims["token_type"].(string); !ok || tt != "access" {
		return uuid.Nil, errors.New("access token required")
	}

	profileIdStr, ok := claims["profile_id"].(string)
	if !ok {
		return uuid.Nil, errors.New("profile_id not in claims")
	}

	return uuid.Parse(profileIdStr)
}

func (h *AuthHandler) Me(c *gin.Context) {
	profileId, err := extractProfileIdFromToken(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, "Unauthorized.")
		return
	}

	user, err := h.svc.GetMe(c.Request.Context(), profileId)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}
	if user == nil {
		ErrorResponse(c, http.StatusNotFound, "User not found.")
		return
	}

	SuccessResponse(c, user, "Profile retrieved successfully.")
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.svc.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidRefreshToken):
			ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired refresh token.")
		case errors.Is(err, services.ErrUserNotActive):
			ErrorResponse(c, http.StatusForbidden, "Account is inactive.")
		case errors.Is(err, services.ErrUserNotFound):
			ErrorResponse(c, http.StatusNotFound, "User not found.")
		default:
			ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		}
		return
	}

	SuccessResponse(c, resp, "Token refreshed successfully.")
}
