package models

import (
	"time"

	"github.com/google/uuid"
)

type Credentials struct {
	CredentialNo uint32     `json:"credential_no"`
	CredentialId uuid.UUID  `json:"credential_id"`
	ProfileNo    uint32     `json:"profile_no"`
	Username     string     `json:"username"`
	Password     string     `json:"password"`
	Status       StatusEnum `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	CreatedBy    uint32     `json:"created_by"`
	UpdatedAt    time.Time  `json:"updated_at"`
	UpdatedBy    uint32     `json:"updated_by"`
}