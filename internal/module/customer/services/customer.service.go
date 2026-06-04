package services

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/chand-magar/gamenglam-server/internal/module/customer/models"
	"github.com/chand-magar/gamenglam-server/internal/module/customer/repository"
)

var (
	ErrCustomerNotFound   = errors.New("Customer not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type customerService struct {
	repo repository.CustomerRepository
}

func NewCustomerService(repo repository.CustomerRepository) CustomerService {
	return &customerService{repo: repo}
}

func (s *customerService) FindAll(ctx context.Context) ([]models.Customer, error) {
	customers, err := s.repo.GetAll(ctx)
	if err != nil {
		log.Printf("customerService.FindAll error: %v", err)
		return nil, fmt.Errorf("customerService.FindAll: %w", err)
	}

	return customers, nil
}

func (s *customerService) FindOne(ctx context.Context, Id uuid.UUID) (*models.Customer, error) {
	customer, err := s.repo.GetById(ctx, Id)
	if err != nil {
		log.Printf("customerService.FindOne error: %v", err)
		return nil, fmt.Errorf("customerService.FindOne: %w", err)
	}
	if customer == nil {
		return nil, ErrCustomerNotFound
	}

	return customer, nil
}

func (s *customerService) Create(ctx context.Context, req *models.CreateCustomerRequest) (*models.Customer, error) {
	existing, err := s.repo.GetByEmail(ctx, req.EmailId)
	if err != nil {
		log.Printf("customerService.Create error: %v", err)
		return nil, fmt.Errorf("customerService.Create: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("customerService.Create: hash password: %w", err)
	}

	customer := &models.Customer{
		CustomerId:  uuid.New(),
		FirstName: 	req.FirstName,
		LastName: 	req.LastName,
		EmailId:    req.EmailId,
		MobileNo:   req.MobileNo,
		Gender:     req.Gender,
		Status:     models.StatusActive,
	}

	created, err := s.repo.Save(ctx, customer)
	if err != nil {
		return nil, fmt.Errorf("customerService.CreateCustomer: %w", err)
	}

	// Store hashed password in credentials — not on customer directly
	_ = hashed

	return created, nil
}

func (s *customerService) Update(ctx context.Context, Id uuid.UUID, req *models.UpdateCustomerRequest) (*models.Customer, error) {
	existing, err := s.repo.GetById(ctx, Id)
	if err != nil {
		log.Printf("customerService.Update error: %v", err)
		return nil, fmt.Errorf("customerService.Update: %w", err)
	}
	if existing == nil {
		return nil, ErrCustomerNotFound
	}

	updated, err := s.repo.Modify(ctx, Id, req)
	if err != nil {
		return nil, fmt.Errorf("customerService.Update: %w", err)
	}

	return updated, nil
}

func (s *customerService) Delete(ctx context.Context, Id uuid.UUID) error {
	existing, err := s.repo.GetById(ctx, Id)
	if err != nil {
		log.Printf("customerService.Delete error: %v", err)
		return fmt.Errorf("customerService.Delete: %w", err)
	}
	if existing == nil {
		return ErrCustomerNotFound
	}

	if err := s.repo.Remove(ctx, Id); err != nil {
		return fmt.Errorf("customerService.Delete: %w", err)
	}

	return nil
}