package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/customer/models"
)

type CustomerRepository interface {
	GetAll(ctx context.Context) ([]models.Customer, error)
	GetById(ctx context.Context, Id uuid.UUID) (*models.Customer, error)
	GetByEmail(ctx context.Context, email string) (*models.Customer, error)
	Save(ctx context.Context, customer *models.Customer) (*models.Customer, error)
	Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateCustomerRequest) (*models.Customer, error)
	Remove(ctx context.Context, Id uuid.UUID) error
}