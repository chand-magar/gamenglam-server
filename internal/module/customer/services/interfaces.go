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

type AddressService interface {
	FindAllByCustomer(ctx context.Context, customerNo uint32) ([]models.Address, error)
	FindOneByNo(ctx context.Context, addressNo uint32) (*models.Address, error)
	FindOneById(ctx context.Context, Id uuid.UUID) (*models.Address, error) // ← add this
	Create(ctx context.Context, req *models.CreateAddressRequest) (*models.Address, error)
	Update(ctx context.Context, Id uuid.UUID, req *models.UpdateAddressRequest) (*models.Address, error)
	Delete(ctx context.Context, Id uuid.UUID) error
}

// CustomerAuthService handles customer-facing registration, login and token management
type CustomerAuthService interface {
	Register(ctx context.Context, req *models.CustomerRegisterRequest) (*models.CustomerAuthResponse, error)
	Login(ctx context.Context, req *models.CustomerLoginRequest) (*models.CustomerAuthResponse, error)
	RefreshToken(ctx context.Context, req *models.CustomerRefreshTokenRequest) (*models.CustomerRefreshResponse, error)
	GetProfile(ctx context.Context, customerId uuid.UUID) (*models.Customer, error)
}
