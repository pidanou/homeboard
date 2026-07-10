package model

import "time"

type OIDCIdentity struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Issuer        string    `json:"issuer"`
	Subject       string    `json:"subject"`
	Email         *string   `json:"email"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}
