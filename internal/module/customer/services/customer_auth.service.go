package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/chand-magar/gamenglam-server/internal/module/customer/models"
	"github.com/chand-magar/gamenglam-server/internal/module/customer/repository"
)

// ── Sentinel errors ───────────────────────────────────────────────────────────

// ── Sentinel errors ───────────────────────────────────────────────────────────
// ErrCustomerNotFound and ErrEmailAlreadyExists are declared in customer.service.go

var (
	ErrInvalidCustomerCreds   = errors.New("invalid email or password")
	ErrCustomerNotActive      = errors.New("customer account is not active")
	ErrInvalidCustomerRefresh = errors.New("invalid or expired refresh token")
)

// ── Token config ──────────────────────────────────────────────────────────────

const (
	customerAccessExpiry  = 24 * time.Hour
	customerRefreshExpiry = 7 * 24 * time.Hour
)

type customerTokenType string

const (
	customerTokenAccess  customerTokenType = "customer_access"
	customerTokenRefresh customerTokenType = "customer_refresh"
)

// ── JWT claims ────────────────────────────────────────────────────────────────

type customerJWTClaims struct {
	CustomerId uuid.UUID         `json:"customer_id"`
	CustomerNo uint32            `json:"customer_no"`
	TokenType  customerTokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// ── Service ───────────────────────────────────────────────────────────────────

type customerAuthService struct {
	customerRepo repository.CustomerRepository
	credRepo     repository.CredentialRepository
}

func NewCustomerAuthService(
	customerRepo repository.CustomerRepository,
	credRepo repository.CredentialRepository,
) CustomerAuthService {
	return &customerAuthService{
		customerRepo: customerRepo,
		credRepo:     credRepo,
	}
}

// ── Register ──────────────────────────────────────────────────────────────────

func (s *customerAuthService) Register(ctx context.Context, req *models.CustomerRegisterRequest) (*models.CustomerAuthResponse, error) {
	// Check duplicate email
	existing, err := s.customerRepo.GetByEmail(ctx, req.EmailId)
	if err != nil {
		return nil, fmt.Errorf("customerAuthService.Register: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("customerAuthService.Register: hash password: %w", err)
	}

	// Save customer
	customer, err := s.customerRepo.Save(ctx, &models.Customer{
		CustomerId: uuid.New(),
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		EmailId:    req.EmailId,
		MobileNo:   req.MobileNo,
		Gender:     req.Gender,
	})
	if err != nil {
		return nil, fmt.Errorf("customerAuthService.Register: save customer: %w", err)
	}

	// Save credential — username = email
	_, err = s.credRepo.Save(ctx, &models.Credentials{
		CredentialId: uuid.New(),
		CustomerNo:   customer.CustomerNo,
		Username:     req.EmailId,
		Password:     string(hashed),
	})
	if err != nil {
		return nil, fmt.Errorf("customerAuthService.Register: save credential: %w", err)
	}

	// Generate tokens
	accessToken, err := s.generateToken(customer, customerTokenAccess, customerAccessExpiry)
	if err != nil {
		return nil, fmt.Errorf("customerAuthService.Register: access token: %w", err)
	}
	refreshToken, err := s.generateToken(customer, customerTokenRefresh, customerRefreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("customerAuthService.Register: refresh token: %w", err)
	}

	return &models.CustomerAuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		Customer:     customer,
	}, nil
}

// ── Login ─────────────────────────────────────────────────────────────────────

func (s *customerAuthService) Login(ctx context.Context, req *models.CustomerLoginRequest) (*models.CustomerAuthResponse, error) {
	// Find credential by email
	cred, err := s.credRepo.GetByEmail(ctx, req.EmailId)
	if err != nil {
		return nil, fmt.Errorf("customerAuthService.Login: %w", err)
	}
	if cred == nil {
		return nil, ErrInvalidCustomerCreds
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(cred.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCustomerCreds
	}

	// Load customer
	customer, err := s.customerRepo.GetByCustomerNo(ctx, cred.CustomerNo)
	if err != nil {
		return nil, fmt.Errorf("customerAuthService.Login: %w", err)
	}
	if customer == nil {
		return nil, ErrCustomerNotFound
	}
	if customer.Status != models.StatusActive {
		return nil, ErrCustomerNotActive
	}

	// Generate tokens
	accessToken, err := s.generateToken(customer, customerTokenAccess, customerAccessExpiry)
	if err != nil {
		return nil, fmt.Errorf("customerAuthService.Login: access token: %w", err)
	}
	refreshToken, err := s.generateToken(customer, customerTokenRefresh, customerRefreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("customerAuthService.Login: refresh token: %w", err)
	}

	return &models.CustomerAuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		Customer:     customer,
	}, nil
}

// ── Refresh Token ─────────────────────────────────────────────────────────────

func (s *customerAuthService) RefreshToken(ctx context.Context, req *models.CustomerRefreshTokenRequest) (*models.CustomerRefreshResponse, error) {
	claims, err := s.parseToken(req.RefreshToken)
	if err != nil {
		return nil, ErrInvalidCustomerRefresh
	}

	// Must be a customer refresh token
	tt, ok := claims["token_type"].(string)
	if !ok || tt != string(customerTokenRefresh) {
		return nil, ErrInvalidCustomerRefresh
	}

	customerIdStr, ok := claims["customer_id"].(string)
	if !ok {
		return nil, ErrInvalidCustomerRefresh
	}

	customerId, err := uuid.Parse(customerIdStr)
	if err != nil {
		return nil, ErrInvalidCustomerRefresh
	}

	customer, err := s.customerRepo.GetById(ctx, customerId)
	if err != nil {
		return nil, fmt.Errorf("customerAuthService.RefreshToken: %w", err)
	}
	if customer == nil {
		return nil, ErrCustomerNotFound
	}
	if customer.Status != models.StatusActive {
		return nil, ErrCustomerNotActive
	}

	// Rotate tokens
	accessToken, err := s.generateToken(customer, customerTokenAccess, customerAccessExpiry)
	if err != nil {
		return nil, fmt.Errorf("customerAuthService.RefreshToken: access token: %w", err)
	}
	refreshToken, err := s.generateToken(customer, customerTokenRefresh, customerRefreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("customerAuthService.RefreshToken: refresh token: %w", err)
	}

	return &models.CustomerRefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}, nil
}

// ── GetProfile ────────────────────────────────────────────────────────────────

func (s *customerAuthService) GetProfile(ctx context.Context, customerId uuid.UUID) (*models.Customer, error) {
	customer, err := s.customerRepo.GetById(ctx, customerId)
	if err != nil {
		return nil, fmt.Errorf("customerAuthService.GetProfile: %w", err)
	}
	if customer == nil {
		return nil, ErrCustomerNotFound
	}
	return customer, nil
}

// ── JWT helpers ───────────────────────────────────────────────────────────────

func (s *customerAuthService) generateToken(customer *models.Customer, tt customerTokenType, expiry time.Duration) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "changeme-use-env-in-production"
	}

	claims := customerJWTClaims{
		CustomerId: customer.CustomerId,
		CustomerNo: customer.CustomerNo,
		TokenType:  tt,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   customer.CustomerId.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func (s *customerAuthService) parseToken(tokenStr string) (jwt.MapClaims, error) {
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
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	return claims, nil
}
