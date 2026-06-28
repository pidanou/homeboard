package service

import (
	"context"
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

func (s *PushService) Subscribe(ctx context.Context, sub *model.PushSubscription) error {
	return s.repo.Save(ctx, sub)
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
			log.Printf("push: send to %s: %v", sub.Endpoint, err)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusGone || resp.StatusCode == http.StatusNotFound {
			_ = s.repo.DeleteByEndpoint(ctx, sub.Endpoint)
		}
	}
}
