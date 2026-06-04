package models

import (
	"time"

	"github.com/google/uuid"
)

type Address struct {
	AddressNo    uint32    `json:"address_no"`
	AddressId    uuid.UUID `json:"address_id"`
	CustomerNo   uint32    `json:"customer_no"`
	Label        string    `json:"label"`
	AddressLine1 string    `json:"address_line1"`
	AddressLine2 *string   `json:"address_line2"`
	City         string    `json:"city"`
	State        string    `json:"state"`
	Country      string    `json:"country"`
	PostalCode   *string   `json:"postal_code"`
	IsDefault    bool      `json:"is_default"`
	Status       StatusEnum `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateAddressRequest struct {
	CustomerNo   uint32  `json:"customer_no"   binding:"required"`
	Label        string  `json:"label"         binding:"omitempty,max=65"`
	AddressLine1 string  `json:"address_line1" binding:"required,max=255"`
	AddressLine2 *string `json:"address_line2" binding:"omitempty,max=255"`
	City         string  `json:"city"          binding:"required,max=100"`
	State        string  `json:"state"         binding:"required,max=100"`
	Country      string  `json:"country"       binding:"omitempty,max=100"`
	PostalCode   *string `json:"postal_code"   binding:"omitempty,max=20"`
	IsDefault    bool    `json:"is_default"`
}

type UpdateAddressRequest struct {
	Label        string  `json:"label"         binding:"omitempty,max=65"`
	AddressLine1 string  `json:"address_line1" binding:"omitempty,max=255"`
	AddressLine2 *string `json:"address_line2" binding:"omitempty,max=255"`
	City         string  `json:"city"          binding:"omitempty,max=100"`
	State        string  `json:"state"         binding:"omitempty,max=100"`
	Country      string  `json:"country"       binding:"omitempty,max=100"`
	PostalCode   *string `json:"postal_code"   binding:"omitempty,max=20"`
	IsDefault    *bool   `json:"is_default"`
}