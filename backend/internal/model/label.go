package model

import "time"

type Label struct {
	ID        string    `json:"id"`
	FamilyID  string    `json:"family_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}
