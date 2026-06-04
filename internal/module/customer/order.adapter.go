package customer

import (
	"context"

	orderservices "github.com/chand-magar/gamenglam-server/internal/module/order/services"
	"github.com/chand-magar/gamenglam-server/internal/module/customer/services"
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