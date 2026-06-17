package model

import "time"

type List struct {
	ID        string    `json:"id"`
	FamilyID  string    `json:"family_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type ListItem struct {
	ID        string     `json:"id"`
	ListID    string     `json:"list_id"`
	Name      string     `json:"name"`
	Checked   bool       `json:"checked"`
	CreatedAt time.Time  `json:"created_at"`
	CheckedAt *time.Time `json:"checked_at"`
}
