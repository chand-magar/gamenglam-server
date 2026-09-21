package customer

import (
	"context"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/customer/services"
	orderservices "github.com/chand-magar/gamenglam-server/internal/module/order/services"
)

type customerAdapter struct {
	svc services.AddressService
}

func (a *customerAdapter) GetAddressByNo(ctx context.Context, addressNo uint32) (*orderservices.CustomerAddress, error) {
	address, err := a.svc.FindOneByNo(ctx, addressNo)
	if err != nil || address == nil {
		return nil, err
	}
	return &orderservices.CustomerAddress{
		AddressNo:    address.AddressNo,
		AddressLine1: address.AddressLine1,
		AddressLine2: address.AddressLine2,
		City:         address.City,
		State:        address.State,
		Country:      address.Country,
		PostalCode:   address.PostalCode,
	}, nil
}

func (a *customerAdapter) GetAddressByID(ctx context.Context, addressId uuid.UUID) (*orderservices.CustomerAddress, error) {
	address, err := a.svc.FindOneById(ctx, addressId)
	if err != nil || address == nil {
		return nil, err
	}
	return &orderservices.CustomerAddress{
		AddressNo:    address.AddressNo, // kept for internal DB use
		AddressLine1: address.AddressLine1,
		AddressLine2: address.AddressLine2,
		City:         address.City,
		State:        address.State,
		Country:      address.Country,
		PostalCode:   address.PostalCode,
	}, nil
}
