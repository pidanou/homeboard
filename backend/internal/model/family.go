package model

import "time"

type Family struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type FamilyMember struct {
	FamilyID  string    `json:"family_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"` // "admin" | "member"
	JoinedAt  time.Time `json:"joined_at"`
}
