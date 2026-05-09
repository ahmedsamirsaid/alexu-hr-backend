package noop

import (
	"log/slog"

	"github.com/banumusa/backend/core/ports"
)

// NoopNotificationService is a no-operation notification service for development and testing.
type NoopNotificationService struct{}

// NewNoopNotificationService creates a new no-op notification service.
func NewNoopNotificationService() *NoopNotificationService {
	return &NoopNotificationService{}
}

// SendToUser logs the notification keys but does not send it.
func (s *NoopNotificationService) SendToUser(userUID, titleKey, bodyKey string, params map[string]interface{}, data ports.NotificationData) (int, error) {
	slog.Info("noop_notification.SendToUser",
		"user_uid", userUID,
		"title_key", titleKey,
		"body_key", bodyKey,
		"params", params,
		"data", data,
	)
	return 0, nil
}

// SendToUsers logs the notification keys but does not send it.
func (s *NoopNotificationService) SendToUsers(userUIDs []string, titleKey, bodyKey string, params map[string]interface{}, data ports.NotificationData) (int, error) {
	slog.Info("noop_notification.SendToUsers",
		"user_uids", userUIDs,
		"title_key", titleKey,
		"body_key", bodyKey,
		"params", params,
		"data", data,
	)
	return 0, nil
}
