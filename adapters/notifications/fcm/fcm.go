package fcm

import (
	"context"
	"log/slog"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"

	"github.com/banumusa/backend/core/ports"
)

// FCMNotificationService implements NotificationService using Firebase Admin SDK.
type FCMNotificationService struct {
	client          *messaging.Client
	deviceTokenRepo ports.DeviceTokenRepository
	db              ports.Querier
}

// Config holds the configuration for the FCM notification service.
type Config struct {
	ServiceAccountJSONPath string
	DeviceTokenRepo        ports.DeviceTokenRepository
	DB                     ports.Querier
}

// NewFCMNotificationService creates a new FCM notification service.
func NewFCMNotificationService(ctx context.Context, cfg Config) (*FCMNotificationService, error) {
	opt := option.WithCredentialsFile(cfg.ServiceAccountJSONPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	return &FCMNotificationService{
		client:          client,
		deviceTokenRepo: cfg.DeviceTokenRepo,
		db:              cfg.DB,
	}, nil
}

// SendToUser sends a push notification to all devices registered to a user.
func (s *FCMNotificationService) SendToUser(userUID, title, body string, data ports.NotificationData) (int, error) {
	tokens, err := s.deviceTokenRepo.GetByUserUID(context.Background(), s.db, userUID)
	if err != nil {
		slog.Error("fcm_notification.SendToUser.get_tokens", "error", err, "user_uid", userUID)
		return 0, err
	}

	if len(tokens) == 0 {
		slog.Debug("fcm_notification.SendToUser.no_tokens", "user_uid", userUID)
		return 0, nil
	}

	successCount := 0
	for _, token := range tokens {
		err := s.sendToToken(token.Token, title, body, data)
		if err != nil {
			// Check if token is invalid and should be deleted
			if messaging.IsUnregistered(err) {
				slog.Debug("fcm_notification.SendToUser.stale_token",
					"user_uid", userUID,
					"token_uid", token.UID,
				)
				if delErr := s.deviceTokenRepo.DeleteByToken(context.Background(), s.db, token.Token); delErr != nil {
					slog.Error("fcm_notification.SendToUser.delete_stale_token", "error", delErr, "token_uid", token.UID)
				}
			} else {
				slog.Warn("fcm_notification.SendToUser.send_failed",
					"error", err,
					"user_uid", userUID,
					"token_uid", token.UID,
				)
			}
			continue
		}
		successCount++
	}

	return successCount, nil
}

// SendToUsers sends a push notification to multiple users.
func (s *FCMNotificationService) SendToUsers(userUIDs []string, title, body string, data ports.NotificationData) (int, error) {
	totalSuccess := 0
	for _, userUID := range userUIDs {
		count, err := s.SendToUser(userUID, title, body, data)
		if err != nil {
			slog.Warn("fcm_notification.SendToUsers.user_failed", "error", err, "user_uid", userUID)
			continue
		}
		totalSuccess += count
	}
	return totalSuccess, nil
}

func (s *FCMNotificationService) sendToToken(token, title, body string, data ports.NotificationData) error {
	msg := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Sound: "default",
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Category: "APPROVAL_CATEGORY",
					Sound:    "default",
				},
			},
		},
	}

	_, err := s.client.Send(context.Background(), msg)
	return err
}
