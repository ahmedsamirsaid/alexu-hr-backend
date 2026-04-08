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

// SendToUser logs the notification but does not send it.
func (s *NoopNotificationService) SendToUser(userUID, title, body string, data ports.NotificationData) (int, error) {
	slog.Info("noop_notification.SendToUser",
		"user_uid", userUID,
		"title", title,
		"body", body,
		"data", data,
	)
	return 0, nil
}

// SendToUsers logs the notification but does not send it.
func (s *NoopNotificationService) SendToUsers(userUIDs []string, title, body string, data ports.NotificationData) (int, error) {
	slog.Info("noop_notification.SendToUsers",
		"user_uids", userUIDs,
		"title", title,
		"body", body,
		"data", data,
	)
	return 0, nil
}
