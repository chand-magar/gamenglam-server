package models

import (
	"time"

	"github.com/google/uuid"
)

// RolePermission represents a single permission entry in role_details
type RolePermission struct {
	PageId     uuid.UUID `json:"page_id"`
	PageName   string    `json:"page_name"`
	CanView    bool      `json:"can_view"`
	CanCreate  bool      `json:"can_create"`
	CanUpdate  bool      `json:"can_update"`
	CanDelete  bool      `json:"can_delete"`
}

type Role struct {
	RoleNo      uint8            `json:"role_no"`
	RoleId      uuid.UUID        `json:"role_id"`
	RoleName    string           `json:"role_name"`
	RoleDetails []RolePermission `json:"role_details"`
	Status      StatusEnum       `json:"status"`
	CreatedAt   time.Time        `json:"created_at"`
	CreatedBy   *uint32          `json:"created_by"`
	UpdatedAt   time.Time        `json:"updated_at"`
	UpdatedBy   *uint32          `json:"updated_by"`
}

type CreateRoleRequest struct {
	RoleName    string           `json:"role_name"    binding:"required,min=2,max=65"`
	RoleDetails []RolePermission `json:"role_details" binding:"omitempty"`
}

type UpdateRoleRequest struct {
	RoleName    string           `json:"role_name"    binding:"omitempty,min=2,max=65"`
	RoleDetails []RolePermission `json:"role_details" binding:"omitempty"`
	UpdatedBy   *uint32          `json:"updated_by"`
}