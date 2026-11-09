package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/notification/models"
	"github.com/chand-magar/gamenglam-server/internal/module/notification/services"
)

type SubscriberHandler struct {
	svc services.SubscriberService
}

func NewSubscriberHandler(svc services.SubscriberService) *SubscriberHandler {
	return &SubscriberHandler{svc: svc}
}

// GET /api/v1/newsletter/subscribers
func (h *SubscriberHandler) FindAll(c *gin.Context) {
	req := &models.SubscriberListRequest{}
	_ = c.ShouldBindQuery(req)
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}

	result, err := h.svc.GetAll(c.Request.Context(), req)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to fetch subscribers.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// POST /api/v1/newsletter/subscribe  (public)
func (h *SubscriberHandler) Subscribe(c *gin.Context) {
	var req models.CreateSubscriberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	sub, err := h.svc.Subscribe(c.Request.Context(), &req)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to subscribe.")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Subscribed successfully.", "data": sub})
}

// POST /api/v1/newsletter/subscribers  (admin — manually add subscriber)
func (h *SubscriberHandler) AddSubscriber(c *gin.Context) {
	var req models.CreateSubscriberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	sub, err := h.svc.Subscribe(c.Request.Context(), &req)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to add subscriber.")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": sub})
}

// DELETE /api/v1/newsletter/subscribers/:id
func (h *SubscriberHandler) Unsubscribe(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid subscriber ID.")
		return
	}

	if err := h.svc.Unsubscribe(c.Request.Context(), id); err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to unsubscribe.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Unsubscribed."})
}
