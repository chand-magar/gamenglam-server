package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/notification/models"
	"github.com/chand-magar/gamenglam-server/internal/module/notification/services"
)

type CampaignHandler struct {
	svc services.CampaignService
}

func NewCampaignHandler(svc services.CampaignService) *CampaignHandler {
	return &CampaignHandler{svc: svc}
}

// GET /api/v1/newsletter/campaigns
func (h *CampaignHandler) FindAll(c *gin.Context) {
	page, _  := strconv.Atoi(c.DefaultQuery("page",  "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	items, total, totalPages, err := h.svc.GetAll(c.Request.Context(), page, limit)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to fetch campaigns.")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"items":       items,
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
		},
	})
}

// POST /api/v1/newsletter/campaigns/send
func (h *CampaignHandler) Send(c *gin.Context) {
	var req models.SendCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	campaign, err := h.svc.Send(c.Request.Context(), &req)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to send campaign.")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Campaign processed.", "data": campaign})
}
