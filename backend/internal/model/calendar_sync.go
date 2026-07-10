package model

import "time"

type CalendarExportToken struct {
	Token     string    `json:"token"`
	FamilyID  string    `json:"family_id"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

type CalendarSubscription struct {
	ID            string     `json:"id"`
	FamilyID      string     `json:"family_id"`
	Name          string     `json:"name"`
	URL           string     `json:"url"`
	CreatedBy     string     `json:"created_by"`
	CreatedAt     time.Time  `json:"created_at"`
	LastSyncedAt  *time.Time `json:"last_synced_at,omitempty"`
	LastSyncError *string    `json:"last_sync_error,omitempty"`
}
