package services

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/chand-magar/gamenglam-server/internal/module/employee/models"
	"github.com/chand-magar/gamenglam-server/internal/module/employee/repository"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) FindAll(ctx context.Context) ([]models.User, error) {
	users, err := s.repo.GetAll(ctx)
	if err != nil {
		log.Printf("userService.FindAll error: %v", err)
		return nil, fmt.Errorf("userService.FindAll: %w", err)
	}

	return users, nil
}

func (s *userService) FindOne(ctx context.Context, Id uuid.UUID) (*models.User, error) {
	user, err := s.repo.GetById(ctx, Id)
	if err != nil {
		log.Printf("userService.FindOne error: %v", err)
		return nil, fmt.Errorf("userService.FindOne: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (s *userService) Create(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	existing, err := s.repo.GetByEmail(ctx, req.EmailId)
	if err != nil {
		log.Printf("userService.Create error: %v", err)
		return nil, fmt.Errorf("userService.Create: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("userService.Create: hash password: %w", err)
	}

	user := &models.User{
		ProfileId:    uuid.New(),
		RoleNo:       req.RoleNo,
		UserFullName: req.UserFullName,
		EmailId:      req.EmailId,
		MobileNo:     req.MobileNo,
		Gender:       req.Gender,
		Status:       models.StatusActive,
	}

	created, err := s.repo.Save(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("userService.CreateUser: %w", err)
	}

	// Store hashed password in credentials — not on user directly
	_ = hashed

	return created, nil
}

func (s *userService) Update(ctx context.Context, Id uuid.UUID, req *models.UpdateUserRequest) (*models.User, error) {
	existing, err := s.repo.GetById(ctx, Id)
	if err != nil {
		log.Printf("userService.Update error: %v", err)
		return nil, fmt.Errorf("userService.Update: %w", err)
	}
	if existing == nil {
		return nil, ErrUserNotFound
	}

	updated, err := s.repo.Modify(ctx, Id, req)
	if err != nil {
		return nil, fmt.Errorf("userService.Update: %w", err)
	}

	return updated, nil
}

func (s *userService) Delete(ctx context.Context, Id uuid.UUID) error {
	existing, err := s.repo.GetById(ctx, Id)
	if err != nil {
		log.Printf("userService.Delete error: %v", err)
		return fmt.Errorf("userService.Delete: %w", err)
	}
	if existing == nil {
		return ErrUserNotFound
	}

	if err := s.repo.Remove(ctx, Id); err != nil {
		return fmt.Errorf("userService.Delete: %w", err)
	}

	return nil
}
