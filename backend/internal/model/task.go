package model

import "time"

type Task struct {
	ID          string     `json:"id"`
	FamilyID    string     `json:"family_id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Status      string     `json:"status"`    // todo | done
	Important   bool       `json:"important"` // flag for important tasks
	AssignedTo  *string    `json:"assigned_to,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	CategoryID  *string    `json:"category_id,omitempty"`
	ManualOrder *int       `json:"manual_order,omitempty"`
	CreatedBy   string     `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
