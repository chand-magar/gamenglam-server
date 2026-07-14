package models

import (
	"time"

	"github.com/google/uuid"
)

// RolePagePermission represents a single page with its allowed methods
type RolePagePermission struct {
	PageId    uuid.UUID `json:"page_id"`
	PageName  string    `json:"page_name"`
	PagePath  string    `json:"page_path"`
	PageOrder uint8     `json:"page_order"`
	Methods   []string  `json:"methods"`
}

// RoleSectionPermission represents a section with its pages and permissions
type RoleSectionPermission struct {
	SectionId    uuid.UUID            `json:"section_id"`
	SectionName  string               `json:"section_name"`
	SectionPath  string               `json:"section_path"`
	SectionIcon  string               `json:"section_icon"`
	SectionOrder uint8                `json:"section_order"`
	Pages        []RolePagePermission `json:"pages"`
}

type Role struct {
	RoleNo      uint8                   `json:"role_no"`
	RoleId      uuid.UUID               `json:"role_id"`
	RoleName    string                  `json:"role_name"`
	RoleDetails []RoleSectionPermission `json:"role_details"`
	Status      StatusEnum              `json:"status"`
	CreatedAt   time.Time               `json:"created_at"`
	CreatedBy   *uint32                 `json:"created_by"`
	UpdatedAt   time.Time               `json:"updated_at"`
	UpdatedBy   *uint32                 `json:"updated_by"`
}

type CreateRoleRequest struct {
	RoleName    string                  `json:"role_name"    binding:"required,min=2,max=65"`
	RoleDetails []RoleSectionPermission `json:"role_details" binding:"omitempty"`
}

type UpdateRoleRequest struct {
	RoleName    string                  `json:"role_name"    binding:"omitempty,min=2,max=65"`
	RoleDetails []RoleSectionPermission `json:"role_details" binding:"omitempty"`
	UpdatedBy   *uint32                 `json:"updated_by"`
}
