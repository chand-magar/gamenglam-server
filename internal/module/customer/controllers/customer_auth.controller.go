package controllers

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/customer/models"
	"github.com/chand-magar/gamenglam-server/internal/module/customer/services"
)

type CustomerAuthHandler struct {
	svc services.CustomerAuthService
}

func NewCustomerAuthHandler(svc services.CustomerAuthService) *CustomerAuthHandler {
	return &CustomerAuthHandler{svc: svc}
}

// POST /api/v1/customers/register
func (h *CustomerAuthHandler) Register(c *gin.Context) {
	var req models.CustomerRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEmailAlreadyExists):
			ErrorResponse(c, http.StatusConflict, "An account with this email already exists.")
		default:
			ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Account created successfully.",
		"data":    resp,
	})
}

// POST /api/v1/customers/login
func (h *CustomerAuthHandler) Login(c *gin.Context) {
	var req models.CustomerLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.svc.Login(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCustomerCreds):
			ErrorResponse(c, http.StatusUnauthorized, "Invalid email or password.")
		case errors.Is(err, services.ErrCustomerNotActive):
			ErrorResponse(c, http.StatusForbidden, "Your account has been deactivated.")
		default:
			ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		}
		return
	}

	SuccessResponse(c, resp, "Login successful.")
}

// POST /api/v1/customers/refresh
func (h *CustomerAuthHandler) RefreshToken(c *gin.Context) {
	var req models.CustomerRefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.svc.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCustomerRefresh):
			ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired refresh token.")
		case errors.Is(err, services.ErrCustomerNotActive):
			ErrorResponse(c, http.StatusForbidden, "Your account has been deactivated.")
		default:
			ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		}
		return
	}

	SuccessResponse(c, resp, "Token refreshed successfully.")
}

// POST /api/v1/customers/logout
func (h *CustomerAuthHandler) Logout(c *gin.Context) {
	// Stateless JWT — client drops the token
	// Future: add token blacklist here
	SuccessResponse(c, nil, "Logged out successfully.")
}

// GET /api/v1/customers/profile
func (h *CustomerAuthHandler) GetProfile(c *gin.Context) {
	customerId, err := extractCustomerIdFromToken(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, "Unauthorized.")
		return
	}

	customer, err := h.svc.GetProfile(c.Request.Context(), customerId)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCustomerNotFound):
			ErrorResponse(c, http.StatusNotFound, "Customer not found.")
		default:
			ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		}
		return
	}

	SuccessResponse(c, customer, "Profile retrieved successfully.")
}

// ── JWT extraction ────────────────────────────────────────────────────────────

func extractCustomerIdFromToken(c *gin.Context) (uuid.UUID, error) {
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

	// Reject non-access tokens
	if tt, ok := claims["token_type"].(string); !ok || tt != "customer_access" {
		return uuid.Nil, errors.New("customer access token required")
	}

	customerIdStr, ok := claims["customer_id"].(string)
	if !ok {
		return uuid.Nil, errors.New("customer_id not in claims")
	}

	return uuid.Parse(customerIdStr)
}
