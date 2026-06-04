package models

import (
	"time"

	"github.com/google/uuid"
)

type Credentials struct {
	CredentialNo 	uint32     `json:"credential_no"`
	CredentialId 	uuid.UUID  `json:"credential_id"`
	CustomerNo    	uint32     `json:"customer_no"`
	Username     	string     `json:"username"`
	Password     	string     `json:"password"`
	Status      	StatusEnum `json:"status"`
	CreatedAt    	time.Time  `json:"created_at"`
	UpdatedAt    	time.Time  `json:"updated_at"`
}

type CreateCredentialRequest struct {
	Password     	string  `json:"password"      binding:"required,min=8"`
}

type UpdateCredentialRequest struct {
	EmailId      	string  `json:"email_id"      binding:"omitempty,email"`
	Password     	string  `json:"password"      binding:"required,min=8"`
	UpdatedBy    	*uint32 `json:"updated_by"`
}