package services

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/auth/models"
	"github.com/chand-magar/gamenglam-server/internal/module/auth/repository"
)

var (
	ErrRoleNotFound = errors.New("role not found")
)

type roleService struct {
	repo repository.RoleRepository
}

func NewRoleService(repo repository.RoleRepository) RoleService {
	return &roleService{repo: repo}
}

func (s *roleService) FindAll(ctx context.Context) ([]models.Role, error) {
	data, err := s.repo.GetAll(ctx)
	if err != nil {
		log.Printf("roleService.FindAll error: %v", err)
		return nil, fmt.Errorf("roleService.FindAll: %w", err)
	}

	return data, nil
}

func (s *roleService) FindOne(ctx context.Context, Id uuid.UUID) (*models.Role, error) {
	data, err := s.repo.GetById(ctx, Id)
	if err != nil {
		log.Printf("roleService.FindOne error: %v", err)
		return nil, fmt.Errorf("roleService.FindOne: %w", err)
	}
	if data == nil {
		return nil, ErrRoleNotFound
	}

	return data, nil
}

func (s *roleService) Create(ctx context.Context, req *models.CreateRoleRequest) (*models.Role, error) {
	data := &models.Role{
		RoleId:      uuid.New(),
		RoleName:    req.RoleName,
		RoleDetails: req.RoleDetails,
		Status:      models.StatusActive,
	}

	created, err := s.repo.Save(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("roleService.Create: %w", err)
	}

	return created, nil
}

func (s *roleService) Update(ctx context.Context, Id uuid.UUID, req *models.UpdateRoleRequest) (*models.Role, error) {
	existing, err := s.repo.GetById(ctx, Id)
	if err != nil {
		return nil, fmt.Errorf("roleService.Update: %w", err)
	}
	if existing == nil {
		return nil, ErrRoleNotFound
	}

	updated, err := s.repo.Modify(ctx, Id, req)
	if err != nil {
		return nil, fmt.Errorf("roleService.Update: %w", err)
	}

	return updated, nil
}

func (s *roleService) Delete(ctx context.Context, Id uuid.UUID) error {
	existing, err := s.repo.GetById(ctx, Id)
	if err != nil {
		return fmt.Errorf("roleService.Delete: %w", err)
	}
	if existing == nil {
		return ErrRoleNotFound
	}

	if err := s.repo.Remove(ctx, Id); err != nil {
		return fmt.Errorf("roleService.Delete: %w", err)
	}

	return nil
}