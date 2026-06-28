package model

type PushSubscription struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Endpoint string `json:"endpoint"`
	Auth     string `json:"auth"`
	P256DH   string `json:"p256dh"`
}
