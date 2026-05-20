package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/auth/models"
	"github.com/chand-magar/gamenglam-server/internal/module/auth/services"
)

type UserHandler struct {
	svc services.UserService
}

func NewUserHandler(svc services.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) FindAll(c *gin.Context) {
	users, err := h.svc.FindAll(c.Request.Context())
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	SuccessResponse(c, users, "Records retrieved successfully.")
}

func (h *UserHandler) FindOne(c *gin.Context) {
	Id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Record not found.")
		return
	}

	user, err := h.svc.FindOne(c.Request.Context(), Id)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			ErrorResponse(c, http.StatusNotFound, "Record not found.")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	SuccessResponse(c, user, "Record retrieved successfully.")
}

func (h *UserHandler) Create(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, services.ErrEmailAlreadyExists) {
			ErrorResponse(c, http.StatusConflict, "Email already exists")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Failed to create the record.")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Record created successfully.",
		"data":    user,
	})
}

func (h *UserHandler) Update(c *gin.Context) {
	Id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid ID provided.")
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.svc.Update(c.Request.Context(), Id, &req)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			ErrorResponse(c, http.StatusNotFound, "Record not found.")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Failed to update the record.")
		return
	}

	SuccessResponse(c, user, "Record updated successfully.")
}

func (h *UserHandler) Delete(c *gin.Context) {
	Id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid ID provided.")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), Id); err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			ErrorResponse(c, http.StatusNotFound, "Record not found.")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Failed to delete the record.")
		return
	}

	SuccessResponse(c, nil, "Record deleted successfully.")
}