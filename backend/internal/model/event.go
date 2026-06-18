package model

import "time"

type Event struct {
	ID          string    `json:"id"`
	FamilyID    string    `json:"family_id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Location    string    `json:"location,omitempty"`
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
	AllDay      bool      `json:"all_day"`
	AttendeeIDs []string  `json:"attendee_ids,omitempty"`
	CategoryID  *string   `json:"category_id,omitempty"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
