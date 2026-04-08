package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type DeviceTokenRepository interface {
	// Create inserts a new device token. If token already exists, updates the user_uid and updated_at.
	Create(ctx context.Context, q Querier, token *domain.DeviceToken) error

	// GetByUID returns a device token by its UID.
	GetByUID(ctx context.Context, q Querier, uid string) (*domain.DeviceToken, error)

	// GetByToken returns a device token by its FCM token string.
	GetByToken(ctx context.Context, q Querier, token string) (*domain.DeviceToken, error)

	// GetByUserUID returns all device tokens for a user.
	GetByUserUID(ctx context.Context, q Querier, userUID string) ([]*domain.DeviceToken, error)

	// Delete removes a device token by UID.
	Delete(ctx context.Context, q Querier, uid string) error

	// DeleteByToken removes a device token by its FCM token string.
	DeleteByToken(ctx context.Context, q Querier, token string) error

	// DeleteAllForUser removes all device tokens for a user.
	DeleteAllForUser(ctx context.Context, q Querier, userUID string) error
}
