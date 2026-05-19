package models

import (
	"time"

	"github.com/google/uuid"
)

type Page struct {

	PageNo    		uint8      `json:"page_no"`
	PageId    		uuid.UUID  `json:"page_id"`
	SectionNo   	uint8      `json:"section_no"`
	PageName  		string     `json:"page_name"`
	PagePath  		string     `json:"page_path"`
	PageOrder 		uint8      `json:"page_order"`
	Status      	StatusEnum `json:"status"`
	CreatedAt   	time.Time  `json:"created_at"`
	CreatedBy   	*uint32    `json:"created_by"`
	UpdatedAt   	time.Time  `json:"updated_at"`
	UpdatedBy   	*uint32    `json:"updated_by"`
}


type CreatePageRequest struct {

	SectionNo	 	uint8		`json:"section_no"`
	PageName     	string 		`json:"page_name"  	binding:"required,min=2,max=65"`
	PagePath	 	string		`json:"page_path" 	binding:"required,min=2,max=65"`
	PageOrder 		uint8  		`json:"page_order"`
}

type UpdatePageRequest struct {

	SectionNo	 	uint8		`json:"section_no"`
	PageName    	string  	`json:"page_name"	binding:"omitempty,min=2,max=65"`
	PagePath		string  	`json:"page_path" 	binding:"required,min=2,max=65"`
	PageOrder		uint8   	`json:"page_order"`
	UpdatedBy   	*uint32 	`json:"updated_by"`
}