package usecases

import (
	"context"
	"errors"

	"github.com/banumusa/backend/core/ports"
)

var ErrDeviceTokenNotFound = errors.New("device token not found")
var ErrNotTokenOwner = errors.New("not authorized to delete this token")

type UnregisterDeviceTokenInput struct {
	UID     string
	UserUID string // Used to verify ownership
}

type UnregisterDeviceTokenUseCase struct {
	deviceTokenRepo ports.DeviceTokenRepository
	db              ports.DB
}

func NewUnregisterDeviceTokenUseCase(deviceTokenRepo ports.DeviceTokenRepository, db ports.DB) *UnregisterDeviceTokenUseCase {
	return &UnregisterDeviceTokenUseCase{
		deviceTokenRepo: deviceTokenRepo,
		db:              db,
	}
}

func (uc *UnregisterDeviceTokenUseCase) Execute(ctx context.Context, input UnregisterDeviceTokenInput) error {
	// Get the token first to verify ownership
	token, err := uc.deviceTokenRepo.GetByUID(ctx, uc.db, input.UID)
	if err != nil {
		return err
	}

	if token == nil {
		return ErrDeviceTokenNotFound
	}

	// Verify ownership
	if token.UserUID != input.UserUID {
		return ErrNotTokenOwner
	}

	return uc.deviceTokenRepo.Delete(ctx, uc.db, input.UID)
}
