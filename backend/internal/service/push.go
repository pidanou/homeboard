package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/pidanou/homeboard/internal/model"
	"github.com/pidanou/homeboard/internal/repository"
)

type PushService struct {
	repo       repository.PushRepository
	vapidPriv  string
	vapidPub   string
	vapidEmail string
}

func NewPushService(repo repository.PushRepository, vapidPriv, vapidPub, vapidEmail string) *PushService {
	return &PushService{repo: repo, vapidPriv: vapidPriv, vapidPub: vapidPub, vapidEmail: vapidEmail}
}

func (s *PushService) Subscribe(ctx context.Context, userID, endpoint, auth, p256dh string) error {
	log.Printf("push: subscribe p256dh=%s auth=%s", p256dh, auth)
	return s.repo.Save(ctx, &model.PushSubscription{
		UserID:   userID,
		Endpoint: endpoint,
		Auth:     auth,
		P256DH:   p256dh,
	})
}

func (s *PushService) Unsubscribe(ctx context.Context, userID, endpoint string) error {
	return s.repo.Delete(ctx, userID, endpoint)
}

func (s *PushService) SendToFamily(ctx context.Context, familyID, title, body string) {
	subs, err := s.repo.ListForFamily(ctx, familyID)
	if err != nil {
		log.Printf("push: list family subs: %v", err)
		return
	}

	payload, _ := json.Marshal(map[string]string{"title": title, "body": body})

	for _, sub := range subs {
		dh, err := base64.RawURLEncoding.DecodeString(sub.P256DH)
		if err != nil {
			log.Printf("push: p256dh base64 decode error: %v (raw=%q)", err, sub.P256DH)
			continue
		}
		log.Printf("push: p256dh decoded len=%d first=%02x last=%02x", len(dh), dh[0], dh[len(dh)-1])
		resp, err := webpush.SendNotification(payload, &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpush.Keys{
				Auth:   sub.Auth,
				P256dh: sub.P256DH,
			},
		}, &webpush.Options{
			VAPIDPrivateKey: s.vapidPriv,
			VAPIDPublicKey:  s.vapidPub,
			Subscriber:      s.vapidEmail,
			TTL:             86400,
		})
		if err != nil {
			log.Printf("push: send error: %v", err)
			continue
		}
resp.Body.Close()
		if resp.StatusCode == http.StatusGone || resp.StatusCode == http.StatusNotFound {
			_ = s.repo.DeleteByEndpoint(ctx, sub.Endpoint)
		}
	}
}
