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

type AddressRepository interface {
	GetAllByCustomerNo(ctx context.Context, customerNo uint32) ([]models.Address, error)
	GetByNo(ctx context.Context, addressNo uint32) (*models.Address, error)
	GetById(ctx context.Context, Id uuid.UUID) (*models.Address, error)
	Save(ctx context.Context, address *models.Address) (*models.Address, error)
	Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateAddressRequest) (*models.Address, error)
	Remove(ctx context.Context, Id uuid.UUID) error
}