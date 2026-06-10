package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/auth/models"
)

type UserRepository interface {
	GetByProfileId(ctx context.Context, profileId uuid.UUID) (*models.User, error)
	GetByProfileNo(ctx context.Context, profileNo uint32) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}
type CredentialRepository interface {
	GetByUsername(ctx context.Context, username string) (*models.Credentials, error)
	GetByProfileNo(ctx context.Context, profileNo uint32) (*models.Credentials, error)
	UpdatePassword(ctx context.Context, profileNo uint32, hashedPassword string) error
}
