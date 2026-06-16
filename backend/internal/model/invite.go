package model

import "time"

type Invite struct {
	Token      string     `json:"token"`
	FamilyID   string     `json:"family_id"`
	FamilyName string     `json:"family_name,omitempty"`
	CreatedBy  string     `json:"created_by"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
	UsedAt     *time.Time `json:"used_at,omitempty"`
}
