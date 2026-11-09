package models

import (
	"time"

	"github.com/google/uuid"
)

// ── Subscriber ────────────────────────────────────────────────────────────────

type Subscriber struct {
	SubscriberNo uint32    `json:"-"`
	SubscriberId uuid.UUID `json:"subscriber_id"`
	Email        string    `json:"email"`
	Name         string    `json:"name,omitempty"`
	Status       string    `json:"status"`
	SubscribedAt time.Time `json:"subscribed_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateSubscriberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name"`
}

type SubscriberListRequest struct {
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
	Search string `form:"search"`
	Status string `form:"status"`
}

type SubscriberListResponse struct {
	Items      []Subscriber `json:"items"`
	Total      int          `json:"total"`
	Page       int          `json:"page"`
	Limit      int          `json:"limit"`
	TotalPages int          `json:"total_pages"`
}

// ── Campaign ─────────────────────────────────────────────────────────────────

type Campaign struct {
	CampaignNo      uint32     `json:"-"`
	CampaignId      uuid.UUID  `json:"campaign_id"`
	Subject         string     `json:"subject"`
	Body            string     `json:"body"`
	SentStatus      string     `json:"sent_status"`
	RecipientCount  int        `json:"recipient_count"`
	SentAt          *time.Time `json:"sent_at"`
	ErrorMessage    string     `json:"error_message,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type SendCampaignRequest struct {
	Subject string `json:"subject" binding:"required,min=3"`
	Body    string `json:"body"    binding:"required,min=10"`
}
