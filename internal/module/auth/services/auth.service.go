package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"

	"github.com/chand-magar/gamenglam-server/internal/module/auth/models"
	"github.com/chand-magar/gamenglam-server/internal/module/auth/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserNotActive      = errors.New("user account is not active")
	ErrInvalidOTP         = errors.New("invalid or expired OTP")
	ErrInvalidPassword    = errors.New("current password is incorrect")
	ErrUserNotFound       = errors.New("user not found")
)

// ── OTP store — in-memory with mutex ─────────────────────────────────────────

type otpEntry struct {
	otp       string
	expiresAt time.Time
}

var (
	otpStore   = make(map[string]otpEntry)
	otpStoreMu sync.Mutex
)

// ── Service ───────────────────────────────────────────────────────────────────

type authService struct {
	credRepo repository.CredentialRepository
	userRepo repository.UserRepository
}

func NewAuthService(
	credRepo repository.CredentialRepository,
	userRepo repository.UserRepository,
) AuthService {
	return &authService{
		credRepo: credRepo,
		userRepo: userRepo,
	}
}

// ── Login ─────────────────────────────────────────────────────────────────────

func (s *authService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	cred, err := s.credRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("authService.Login: %w", err)
	}
	if cred == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(cred.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	user, err := s.userRepo.GetByProfileNo(ctx, cred.ProfileNo)
	if err != nil {
		return nil, fmt.Errorf("authService.Login: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if user.Status != models.StatusActive {
		return nil, ErrUserNotActive
	}

	token, err := generateJWT(user)
	if err != nil {
		return nil, fmt.Errorf("authService.Login: generate token: %w", err)
	}

	return &models.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		User:        user,
	}, nil
}

// ── Forgot Password ───────────────────────────────────────────────────────────

func (s *authService) ForgotPassword(ctx context.Context, req *models.ForgotPasswordRequest) error {
	user, err := s.userRepo.GetByEmail(ctx, req.EmailId)
	if err != nil {
		return fmt.Errorf("authService.ForgotPassword: %w", err)
	}
	if user == nil {
		// Don't reveal if email exists
		return nil
	}

	otp, err := generateOTP()
	if err != nil {
		return fmt.Errorf("authService.ForgotPassword: generate OTP: %w", err)
	}

	otpStoreMu.Lock()
	otpStore[req.EmailId] = otpEntry{
		otp:       otp,
		expiresAt: time.Now().Add(10 * time.Minute),
	}
	otpStoreMu.Unlock()

	if err := sendOTPEmail(req.EmailId, user.UserFullName, otp); err != nil {
		log.Printf("authService.ForgotPassword: send email: %v", err)
		return fmt.Errorf("authService.ForgotPassword: send email: %w", err)
	}

	log.Printf("authService.ForgotPassword: OTP sent to %s", req.EmailId)
	return nil
}

// ── Verify OTP ────────────────────────────────────────────────────────────────

func (s *authService) VerifyOTP(ctx context.Context, req *models.VerifyOTPRequest) error {
	otpStoreMu.Lock()
	entry, exists := otpStore[req.EmailId]
	otpStoreMu.Unlock()

	if !exists || entry.otp != req.OTP || time.Now().After(entry.expiresAt) {
		return ErrInvalidOTP
	}

	return nil
}

// ── Reset Password ────────────────────────────────────────────────────────────

func (s *authService) ResetPassword(ctx context.Context, req *models.ResetPasswordRequest) error {
	otpStoreMu.Lock()
	entry, exists := otpStore[req.EmailId]
	otpStoreMu.Unlock()

	if !exists || entry.otp != req.OTP || time.Now().After(entry.expiresAt) {
		return ErrInvalidOTP
	}

	user, err := s.userRepo.GetByEmail(ctx, req.EmailId)
	if err != nil {
		return fmt.Errorf("authService.ResetPassword: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("authService.ResetPassword: hash password: %w", err)
	}

	if err := s.credRepo.UpdatePassword(ctx, user.ProfileNo, string(hashed)); err != nil {
		return fmt.Errorf("authService.ResetPassword: %w", err)
	}

	// Invalidate OTP after successful reset
	otpStoreMu.Lock()
	delete(otpStore, req.EmailId)
	otpStoreMu.Unlock()

	return nil
}

// ── Change Password ───────────────────────────────────────────────────────────

func (s *authService) ChangePassword(ctx context.Context, profileId uuid.UUID, req *models.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByProfileId(ctx, profileId)
	if err != nil {
		return fmt.Errorf("authService.ChangePassword: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	cred, err := s.credRepo.GetByProfileNo(ctx, user.ProfileNo)
	if err != nil {
		return fmt.Errorf("authService.ChangePassword: %w", err)
	}
	if cred == nil {
		return ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(cred.Password), []byte(req.CurrentPassword)); err != nil {
		return ErrInvalidPassword
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("authService.ChangePassword: hash password: %w", err)
	}

	if err := s.credRepo.UpdatePassword(ctx, user.ProfileNo, string(hashed)); err != nil {
		return fmt.Errorf("authService.ChangePassword: %w", err)
	}

	return nil
}

// ── Internal helpers ──────────────────────────────────────────────────────────

// getUserByProfileNo fetches user by profile_no via email workaround
// since UserRepository only exposes GetByProfileId and GetByEmail
func (s *authService) getUserByProfileNo(ctx context.Context, profileNo uint32) (*models.User, error) {
	cred, err := s.credRepo.GetByProfileNo(ctx, profileNo)
	if err != nil || cred == nil {
		return nil, err
	}
	// We don't have email here — add GetByProfileNo to UserRepository instead
	return nil, nil
}

// ── JWT ───────────────────────────────────────────────────────────────────────

type jwtClaims struct {
	ProfileId uuid.UUID `json:"profile_id"`
	ProfileNo uint32    `json:"profile_no"`
	RoleNo    uint32    `json:"role_no"`
	jwt.RegisteredClaims
}

func generateJWT(user *models.User) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "changeme-use-env-in-production"
	}

	claims := jwtClaims{
		ProfileId: user.ProfileId,
		ProfileNo: user.ProfileNo,
		RoleNo:    user.RoleNo,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ProfileId.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ── OTP ───────────────────────────────────────────────────────────────────────

func generateOTP() (string, error) {
	const digits = "0123456789"
	otp := make([]byte, 6)
	for i := range otp {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		otp[i] = digits[n.Int64()]
	}
	return string(otp), nil
}

// ── Email ─────────────────────────────────────────────────────────────────────

func sendOTPEmail(toEmail, fullName, otp string) error {
	host := os.Getenv("SMTP_HOST")
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("SMTP_FROM")

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Your Password Reset OTP")
	m.SetBody("text/html", fmt.Sprintf(`
		<h2>Password Reset Request</h2>
		<p>Hi %s,</p>
		<p>Your OTP for password reset is:</p>
		<h1 style="letter-spacing:8px; color:#333;">%s</h1>
		<p>This OTP expires in <strong>10 minutes</strong>.</p>
		<p>If you did not request this, please ignore this email.</p>
	`, fullName, otp))

	d := gomail.NewDialer(host, 587, username, password)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("sendOTPEmail: %w", err)
	}

	return nil
}
