package ports

// NotificationData contains additional data to include in the push notification payload.
type NotificationData map[string]string

// NotificationService defines the interface for sending push notifications.
type NotificationService interface {
	// SendToUser sends a push notification to all devices registered to a user.
	// Returns the number of successful deliveries and any error encountered.
	SendToUser(userUID, title, body string, data NotificationData) (int, error)

	// SendToUsers sends a push notification to multiple users.
	// Returns the total number of successful deliveries and any error encountered.
	SendToUsers(userUIDs []string, title, body string, data NotificationData) (int, error)
}
