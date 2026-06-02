package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/customer/models"
)

type CustomerService interface {
	FindAll(ctx context.Context) ([]models.Customer, error)
	FindOne(ctx context.Context, Id uuid.UUID) (*models.Customer, error)
	Create(ctx context.Context, req *models.CreateCustomerRequest) (*models.Customer, error)
	Update(ctx context.Context, Id uuid.UUID, req *models.UpdateCustomerRequest) (*models.Customer, error)
	Delete(ctx context.Context, Id uuid.UUID) error
}