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
// It accepts i18n translation keys and translates them to each user's preferred language.
type FCMNotificationService struct {
	client          *messaging.Client
	deviceTokenRepo ports.DeviceTokenRepository
	userRepo        ports.UserRepository
	i18n            ports.I18nService
	db              ports.Querier
}

// Config holds the configuration for the FCM notification service.
type Config struct {
	ServiceAccountJSONPath string
	DeviceTokenRepo        ports.DeviceTokenRepository
	UserRepo               ports.UserRepository
	I18nService            ports.I18nService
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
		userRepo:        cfg.UserRepo,
		i18n:            cfg.I18nService,
		db:              cfg.DB,
	}, nil
}

// SendToUser sends a localized push notification to all devices registered to a user.
// titleKey and bodyKey are i18n translation keys; the user's preferred language is used for translation.
func (s *FCMNotificationService) SendToUser(userUID, titleKey, bodyKey string, params map[string]interface{}, data ports.NotificationData) (int, error) {
	// Look up user's preferred language
	locale := s.userLocale(userUID)

	// Translate notification text
	title := s.translate(locale, titleKey, params)
	body := s.translate(locale, bodyKey, params)

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

// SendToUsers sends localized push notifications to multiple users.
// Each user's preferred language is resolved individually.
func (s *FCMNotificationService) SendToUsers(userUIDs []string, titleKey, bodyKey string, params map[string]interface{}, data ports.NotificationData) (int, error) {
	totalSuccess := 0
	for _, userUID := range userUIDs {
		count, err := s.SendToUser(userUID, titleKey, bodyKey, params, data)
		if err != nil {
			slog.Warn("fcm_notification.SendToUsers.user_failed", "error", err, "user_uid", userUID)
			continue
		}
		totalSuccess += count
	}
	return totalSuccess, nil
}

// userLocale looks up the user's preferred language; defaults to "ar" if unset or not found.
func (s *FCMNotificationService) userLocale(userUID string) string {
	user, err := s.userRepo.GetByUID(context.Background(), s.db, userUID)
	if err != nil || user == nil {
		return "ar"
	}
	if user.PreferredLanguage == "" {
		return "ar"
	}
	return user.PreferredLanguage
}

// translate resolves a translation key with optional params for the given locale.
func (s *FCMNotificationService) translate(locale, key string, params map[string]interface{}) string {
	if len(params) > 0 {
		return s.i18n.TLocaleWithParams(locale, key, params)
	}
	return s.i18n.TLocale(locale, key)
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
				ChannelID: "APPROVAL_CHANNEL",
				Sound:     "default",
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
