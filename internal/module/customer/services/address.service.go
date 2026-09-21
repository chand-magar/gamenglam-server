package services

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/chand-magar/gamenglam-server/internal/module/customer/models"
	"github.com/chand-magar/gamenglam-server/internal/module/customer/repository"
)

var (
	ErrAddressNotFound = errors.New("address not found")
)

type addressService struct {
	repo repository.AddressRepository
}

func NewAddressService(repo repository.AddressRepository) AddressService {
	return &addressService{repo: repo}
}

func (s *addressService) FindAllByCustomer(ctx context.Context, customerNo uint32) ([]models.Address, error) {
	addresses, err := s.repo.GetAllByCustomerNo(ctx, customerNo)
	if err != nil {
		log.Printf("addressService.FindAllByCustomer error: %v", err)
		return nil, fmt.Errorf("addressService.FindAllByCustomer: %w", err)
	}
	return addresses, nil
}

func (s *addressService) FindOneByNo(ctx context.Context, addressNo uint32) (*models.Address, error) {
	address, err := s.repo.GetByNo(ctx, addressNo)
	if err != nil {
		log.Printf("addressService.FindOneByNo error: %v", err)
		return nil, fmt.Errorf("addressService.FindOneByNo: %w", err)
	}
	if address == nil {
		return nil, ErrAddressNotFound
	}
	return address, nil
}

func (s *addressService) Create(ctx context.Context, req *models.CreateAddressRequest) (*models.Address, error) {
	address := &models.Address{
		CustomerNo:   req.CustomerNo,
		Label:        req.Label,
		AddressLine1: req.AddressLine1,
		AddressLine2: req.AddressLine2,
		City:         req.City,
		State:        req.State,
		Country:      req.Country,
		PostalCode:   req.PostalCode,
		IsDefault:    req.IsDefault,
	}
	if address.Label == "" {
		address.Label = "Home"
	}

	created, err := s.repo.Save(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("addressService.Create: %w", err)
	}
	return created, nil
}

func (s *addressService) Update(ctx context.Context, Id uuid.UUID, req *models.UpdateAddressRequest) (*models.Address, error) {
	existing, err := s.repo.GetById(ctx, Id)
	if err != nil {
		return nil, fmt.Errorf("addressService.Update: %w", err)
	}
	if existing == nil {
		return nil, ErrAddressNotFound
	}

	updated, err := s.repo.Modify(ctx, Id, req)
	if err != nil {
		return nil, fmt.Errorf("addressService.Update: %w", err)
	}
	return updated, nil
}

func (s *addressService) Delete(ctx context.Context, Id uuid.UUID) error {
	existing, err := s.repo.GetById(ctx, Id)
	if err != nil {
		return fmt.Errorf("addressService.Delete: %w", err)
	}
	if existing == nil {
		return ErrAddressNotFound
	}

	if err := s.repo.Remove(ctx, Id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAddressNotFound
		}
		return fmt.Errorf("addressService.Delete: %w", err)
	}
	return nil
}

func (s *addressService) FindOneById(ctx context.Context, Id uuid.UUID) (*models.Address, error) {
	address, err := s.repo.GetById(ctx, Id)
	if err != nil {
		return nil, fmt.Errorf("addressService.FindOneById: %w", err)
	}
	if address == nil {
		return nil, ErrAddressNotFound
	}
	return address, nil
}
