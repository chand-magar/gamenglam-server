package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	// Both live inside the module now — no flat internal/models or internal/services
	"github.com/chand-magar/gamenglam-server/internal/module/auth/models"
	"github.com/chand-magar/gamenglam-server/internal/module/auth/services"
)

type PageHandler struct {
	svc services.PageService
}

func NewPageHandler(svc services.PageService) *PageHandler {
	return &PageHandler{svc: svc}
}

func (h *PageHandler) FindAll(c *gin.Context) {
	datas, err := h.svc.FindAll(c.Request.Context())
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}
	SuccessResponse(c, datas, "Records retrieved successfully.")
}

func (h *PageHandler) FindOne(c *gin.Context) {
	Id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid ID provided.")
		return
	}

	data, err := h.svc.FindOne(c.Request.Context(), Id)
	if err != nil {
		if errors.Is(err, services.ErrPageNotFound) {
			ErrorResponse(c, http.StatusNotFound, "Record not found.")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}
	SuccessResponse(c, data, "Record retrieved successfully.")
}

func (h *PageHandler) Create(c *gin.Context) {
	var req models.CreatePageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	data, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to create the record.")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Record created successfully.",
		"data":    data,
	})
}

func (h *PageHandler) Update(c *gin.Context) {
	Id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid ID provided.")
		return
	}

	var req models.UpdatePageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	data, err := h.svc.Update(c.Request.Context(), Id, &req)
	if err != nil {
		if errors.Is(err, services.ErrPageNotFound) {
			ErrorResponse(c, http.StatusNotFound, "Record not found.")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Failed to update the record.")
		return
	}
	SuccessResponse(c, data, "Record updated successfully.")
}

func (h *PageHandler) Delete(c *gin.Context) {
	Id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid ID provided.")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), Id); err != nil {
		if errors.Is(err, services.ErrPageNotFound) {
			ErrorResponse(c, http.StatusNotFound, "Record not found.")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Failed to delete the record.")
		return
	}
	SuccessResponse(c, nil, "Record deleted successfully.")
}