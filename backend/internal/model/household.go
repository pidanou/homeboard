package model

import "time"

type Household struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type HouseholdMember struct {
	FamilyID  string    `json:"family_id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	Role      string    `json:"role"` // "admin" | "member"
	JoinedAt  time.Time `json:"joined_at"`
	Virtual   bool      `json:"virtual"`
}

type VirtualMember struct {
	ID            string    `json:"id"`
	FamilyID      string    `json:"family_id"`
	Name          string    `json:"name"`
	LinkedUserID  *string   `json:"linked_user_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
