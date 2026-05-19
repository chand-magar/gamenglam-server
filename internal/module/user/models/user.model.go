package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ProfileNo    uint32     `json:"profile_no"`
	ProfileId    uuid.UUID  `json:"profile_id"`
	RoleNo       uint32     `json:"role_no"`
	UserFullName string     `json:"user_fullname"`
	EmailId      string     `json:"email_id"`
	Gender       *string    `json:"gender"`
	Dob          *time.Time `json:"dob"`
	MobileNo     *string    `json:"mobile_no"`
	Address      *string    `json:"address"`
	XApiKey      *string    `json:"x_api_key"`
	SecretKey    *string    `json:"secret_key"`
	Status       StatusEnum `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	CreatedBy    *uint32    `json:"created_by"`
	UpdatedAt    time.Time  `json:"updated_at"`
	UpdatedBy    *uint32    `json:"updated_by"`
}

type CreateUserRequest struct {
	UserFullName string  `json:"user_fullname" binding:"required,min=2,max=65"`
	EmailId      string  `json:"email_id"      binding:"required,email"`
	MobileNo     *string `json:"mobile_no"     binding:"omitempty,max=15"`
	Gender       *string `json:"gender"        binding:"omitempty"`
	RoleNo       uint32  `json:"role_no"       binding:"required"`
	Password     string  `json:"password"      binding:"required,min=8"`
}

type UpdateUserRequest struct {
	UserFullName string  `json:"user_fullname" binding:"omitempty,min=2,max=65"`
	EmailId      string  `json:"email_id"      binding:"omitempty,email"`
	Gender       *string `json:"gender"        binding:"omitempty"`
	MobileNo     *string `json:"mobile_no"     binding:"omitempty,max=15"`
	Address      *string `json:"address"       binding:"omitempty"`
	UpdatedBy    *uint32 `json:"updated_by"`
}