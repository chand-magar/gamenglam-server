package models

import (
	"time"

	"github.com/google/uuid"
)

type Section struct {
	SectionNo    	uint8      		`json:"section_no"`
	SectionId    	uuid.UUID  		`json:"section_id"`
	SectionName  	string     		`json:"section_name"`
	SectionPath  	string     		`json:"section_path"`
	SectionIcon  	string     		`json:"section_icon"`
	SectionOrder 	uint8      		`json:"section_order"`
	Status       StatusEnum `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	CreatedBy    *uint32    `json:"created_by"`
	UpdatedAt    time.Time  `json:"updated_at"`
	UpdatedBy    *uint32    `json:"updated_by"`
}

type CreateSectionRequest struct {
	SectionName    	string           `json:"section_name"   	binding:"required,min=2,max=65"`
	SectionPath		string           `json:"section_path"    	binding:"required,min=2,max=65"`
	SectionIcon		string           `json:"section_icon"    	binding:"required,min=2,max=65"`
	SectionOrder	uint8           `json:"section_order"`
}

type UpdateSectionRequest struct {
	SectionName    	string           `json:"section_name"		binding:"omitempty,min=2,max=65"`
	SectionPath		string           `json:"section_path"    	binding:"required,min=2,max=65"`
	SectionIcon		string           `json:"section_icon"   	binding:"required,min=2,max=65"`
	SectionOrder	uint8           `json:"section_order"`
	UpdatedBy   	*uint32          `json:"updated_by"`
}