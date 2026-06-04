package models

type StatusEnum string

const (
	StatusActive   StatusEnum = "A"
	StatusInactive StatusEnum = "I"
	StatusDeleted  StatusEnum = "D"
)
