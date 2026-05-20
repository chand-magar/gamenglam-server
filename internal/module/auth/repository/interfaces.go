package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/auth/models"
)

type UserRepository interface {
	GetAll(ctx context.Context) ([]models.User, error)
	GetById(ctx context.Context, Id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Save(ctx context.Context, user *models.User) (*models.User, error)
	Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateUserRequest) (*models.User, error)
	Remove(ctx context.Context, Id uuid.UUID) error
}

type RoleRepository interface {
	GetAll(ctx context.Context) ([]models.Role, error)
	GetById(ctx context.Context, Id uuid.UUID) (*models.Role, error)
	Save(ctx context.Context, role *models.Role) (*models.Role, error)
	Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateRoleRequest) (*models.Role, error)
	Remove(ctx context.Context, Id uuid.UUID) error
}

type SectionRepository interface {
	GetAll(ctx context.Context) ([]models.Section, error)
	GetById(ctx context.Context, Id uuid.UUID) (*models.Section, error)
	Save(ctx context.Context, section *models.Section) (*models.Section, error)
	Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateSectionRequest) (*models.Section, error)
	Remove(ctx context.Context, Id uuid.UUID) error
}

type PageRepository interface {
	GetAll(ctx context.Context) ([]models.Page, error)
	GetById(ctx context.Context, Id uuid.UUID) (*models.Page, error)
	Save(ctx context.Context, page *models.Page) (*models.Page, error)
	Modify(ctx context.Context, Id uuid.UUID, req *models.UpdatePageRequest) (*models.Page, error)
	Remove(ctx context.Context, Id uuid.UUID) error
}