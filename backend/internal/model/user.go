package model

import "time"

type User struct {
	ID                    string    `json:"id"`
	Email                 string    `json:"email"`
	PasswordHash          *string   `json:"-"`
	Name                  string    `json:"name"`
	AvatarURL             *string   `json:"avatar_url"`
	Locale                string    `json:"locale"`
	TimeFormat            string    `json:"time_format"`
	ReminderMinutesBefore *int      `json:"reminder_minutes_before"`
	CreatedAt             time.Time `json:"created_at"`
}
