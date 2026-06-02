package models

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	CustomerNo    	uint32     `json:"customer_no"`
	CustomerId    	uuid.UUID  `json:"customer_id"`
	FullName 		string     `json:"fullname"`
	EmailId      	string     `json:"email_id"`
	Gender       	*string    `json:"gender"`
	Dob          	*time.Time `json:"dob"`
	MobileNo     	*string    `json:"mobile_no"`
	Address      	*string    `json:"address"`
	Status       	StatusEnum `json:"status"`
	CreatedAt    	time.Time  `json:"created_at"`
	CreatedBy    	*uint32    `json:"created_by"`
	UpdatedAt    	time.Time  `json:"updated_at"`
	UpdatedBy    	*uint32    `json:"updated_by"`
}

type CreateCustomerRequest struct {
	FullName 		string  `json:"fullname" binding:"required,min=2,max=65"`
	EmailId      	string  `json:"email_id"      binding:"required,email"`
	MobileNo     	*string `json:"mobile_no"     binding:"omitempty,max=15"`
	Gender       	*string `json:"gender"        binding:"omitempty"`
	Password     	string  `json:"password"      binding:"required,min=8"`
}

type UpdateCustomerRequest struct {
	FullName 		string  `json:"user_fullname" binding:"omitempty,min=2,max=65"`
	EmailId      	string  `json:"email_id"      binding:"omitempty,email"`
	Gender       	*string `json:"gender"        binding:"omitempty"`
	MobileNo     	*string `json:"mobile_no"     binding:"omitempty,max=15"`
	Address      	*string `json:"address"       binding:"omitempty"`
	UpdatedBy    	*uint32 `json:"updated_by"`
}