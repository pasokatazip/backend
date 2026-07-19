package notification

import (
	"context"
	"encoding/json"
	"fmt"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases"
)

type WebPushSender struct {
	vapidPublicKey  string
	vapidPrivateKey string
	subject         string
	ttl             int
}

type WebPushSenderConfig struct {
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	Subject         string
	TTL             int
}

func NewWebPushSender(config WebPushSenderConfig) (*WebPushSender, error) {
	if config.VAPIDPublicKey == "" || config.VAPIDPrivateKey == "" || config.Subject == "" {
		return nil, fmt.Errorf("%w: vapid public key, private key, and subject are required", domain.ErrValidation)
	}
	if config.TTL <= 0 {
		config.TTL = 60
	}

	return &WebPushSender{
		vapidPublicKey:  config.VAPIDPublicKey,
		vapidPrivateKey: config.VAPIDPrivateKey,
		subject:         config.Subject,
		ttl:             config.TTL,
	}, nil
}

func (s *WebPushSender) Send(ctx context.Context, subscription json.RawMessage, payload usecases.NotificationPayload) error {
	var sub webpush.Subscription
	if err := json.Unmarshal(subscription, &sub); err != nil {
		return fmt.Errorf("%w: decode web push subscription: %v", domain.ErrValidation, err)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("%w: encode web push payload: %v", domain.ErrInternal, err)
	}

	resp, err := webpush.SendNotificationWithContext(ctx, body, &sub, &webpush.Options{
		Subscriber:      s.subject,
		VAPIDPublicKey:  s.vapidPublicKey,
		VAPIDPrivateKey: s.vapidPrivateKey,
		TTL:             s.ttl,
	})
	if err != nil {
		return fmt.Errorf("%w: send web push: %v", domain.ErrExternalService, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: web push failed: status=%d", domain.ErrExternalService, resp.StatusCode)
	}

	return nil
}
