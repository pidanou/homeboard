package model

import "time"

type Task struct {
	ID         string     `json:"id"`
	FamilyID   string     `json:"family_id"`
	Title      string     `json:"title"`
	Status     string     `json:"status"` // todo | done
	AssignedTo *string    `json:"assigned_to,omitempty"`
	StartDate  *time.Time `json:"start_date,omitempty"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	CreatedBy  string     `json:"created_by"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
