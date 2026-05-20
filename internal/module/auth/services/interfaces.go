package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/auth/models"
)

type UserService interface {
	FindAll(ctx context.Context) ([]models.User, error)
	FindOne(ctx context.Context, Id uuid.UUID) (*models.User, error)
	Create(ctx context.Context, req *models.CreateUserRequest) (*models.User, error)
	Update(ctx context.Context, Id uuid.UUID, req *models.UpdateUserRequest) (*models.User, error)
	Delete(ctx context.Context, Id uuid.UUID) error
}

type RoleService interface {
	FindAll(ctx context.Context) ([]models.Role, error)
	FindOne(ctx context.Context, Id uuid.UUID) (*models.Role, error)
	Create(ctx context.Context, req *models.CreateRoleRequest) (*models.Role, error)
	Update(ctx context.Context, Id uuid.UUID, req *models.UpdateRoleRequest) (*models.Role, error)
	Delete(ctx context.Context, Id uuid.UUID) error
}

type SectionService interface {
	FindAll(ctx context.Context) ([]models.Section, error)
	FindOne(ctx context.Context, Id uuid.UUID) (*models.Section, error)
	Create(ctx context.Context, req *models.CreateSectionRequest) (*models.Section, error)
	Update(ctx context.Context, Id uuid.UUID, req *models.UpdateSectionRequest) (*models.Section, error)
	Delete(ctx context.Context, Id uuid.UUID) error
}

type PageService interface {
	FindAll(ctx context.Context) ([]models.Page, error)
	FindOne(ctx context.Context, Id uuid.UUID) (*models.Page, error)
	Create(ctx context.Context, req *models.CreatePageRequest) (*models.Page, error)
	Update(ctx context.Context, Id uuid.UUID, req *models.UpdatePageRequest) (*models.Page, error)
	Delete(ctx context.Context, Id uuid.UUID) error
}