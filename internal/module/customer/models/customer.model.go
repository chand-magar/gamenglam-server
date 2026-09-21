package models

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	CustomerNo uint32     `json:"-"` // ← hide
	CustomerId uuid.UUID  `json:"customer_id"`
	FirstName  string     `json:"first_name"`
	LastName   string     `json:"last_name"`
	EmailId    string     `json:"email_id"`
	Gender     *string    `json:"gender"`
	Dob        *time.Time `json:"dob"`
	MobileNo   *string    `json:"mobile_no"`
	Status     StatusEnum `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type CreateCustomerRequest struct {
	FirstName string  `json:"first_name" binding:"required,min=2,max=65"`
	LastName  string  `json:"last_name" binding:"required,min=2,max=65"`
	EmailId   string  `json:"email_id"      binding:"required,email"`
	MobileNo  *string `json:"mobile_no"     binding:"omitempty,max=15"`
	Gender    *string `json:"gender"        binding:"omitempty"`
	Password  string  `json:"password"      binding:"required,min=8"`
}

type UpdateCustomerRequest struct {
	FirstName string  `json:"first_name" binding:"omitempty,min=2,max=65"`
	LastName  string  `json:"last_name" binding:"omitempty,min=2,max=65"`
	EmailId   string  `json:"email_id"      binding:"omitempty,email"`
	Gender    *string `json:"gender"        binding:"omitempty"`
	MobileNo  *string `json:"mobile_no"     binding:"omitempty,max=15"`
}
