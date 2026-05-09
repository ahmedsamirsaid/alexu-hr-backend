package ports

// NotificationData contains additional data to include in the push notification payload.
type NotificationData map[string]string

// LocalizableString holds the Arabic and English variants of a dynamic notification param.
// The NotificationService implementation resolves it to the recipient's preferred language
// before applying it to the template, so callers never need to do language selection themselves.
type LocalizableString struct {
	Ar string
	En string
}

// NotificationService defines the interface for sending push notifications.
// Callers pass i18n translation keys (e.g. "notification.leave_approved.title") so that
// the implementation can translate them to each recipient's preferred language.
type NotificationService interface {
	// SendToUser sends a localized push notification to all devices registered to a user.
	// titleKey and bodyKey are i18n translation keys; params is the optional template data.
	// The implementation looks up the user's preferred language and translates accordingly.
	// Returns the number of successful deliveries and any error encountered.
	SendToUser(userUID, titleKey, bodyKey string, params map[string]interface{}, data NotificationData) (int, error)

	// SendToUsers sends localized push notifications to multiple users.
	// Each user's preferred language is used for translation.
	// Returns the total number of successful deliveries and any error encountered.
	SendToUsers(userUIDs []string, titleKey, bodyKey string, params map[string]interface{}, data NotificationData) (int, error)
}
