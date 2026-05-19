package models

type StatusEnum string

const (
	StatusActive   StatusEnum = "A"
	StatusInactive StatusEnum = "I"
	StatusDeleted  StatusEnum = "D"
)

type ProductStatus string

const (
	ProductStatusDraft    ProductStatus = "draft"
	ProductStatusActive   ProductStatus = "active"
	ProductStatusArchived ProductStatus = "archived"
)

type AttributeType string

const (
	AttributeTypeSelect      AttributeType = "select"
	AttributeTypeMultiSelect AttributeType = "multi_select"
	AttributeTypeBoolean     AttributeType = "boolean"
	AttributeTypeText        AttributeType = "text"
	AttributeTypeNumber      AttributeType = "number"
	AttributeTypeColor       AttributeType = "color"
)