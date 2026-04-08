package usecases

import (
	"context"
	"errors"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

var (
	ErrInvalidPlatform = errors.New("invalid platform, must be 'android' or 'ios'")
	ErrInvalidToken    = errors.New("token is required")
)

type RegisterDeviceTokenInput struct {
	UserUID  string
	Token    string
	Platform string
}

type RegisterDeviceTokenOutput struct {
	UID string
}

type RegisterDeviceTokenUseCase struct {
	deviceTokenRepo ports.DeviceTokenRepository
	db              ports.DB
}

func NewRegisterDeviceTokenUseCase(deviceTokenRepo ports.DeviceTokenRepository, db ports.DB) *RegisterDeviceTokenUseCase {
	return &RegisterDeviceTokenUseCase{
		deviceTokenRepo: deviceTokenRepo,
		db:              db,
	}
}

func (uc *RegisterDeviceTokenUseCase) Execute(ctx context.Context, input RegisterDeviceTokenInput) (*RegisterDeviceTokenOutput, error) {
	if input.Token == "" {
		return nil, ErrInvalidToken
	}

	if input.Platform != "android" && input.Platform != "ios" {
		return nil, ErrInvalidPlatform
	}

	// Check if token already exists
	existing, err := uc.deviceTokenRepo.GetByToken(ctx, uc.db, input.Token)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		// Token exists - update it (might be a different user now)
		existing.UserUID = input.UserUID
		if err := uc.deviceTokenRepo.Create(ctx, uc.db, existing); err != nil {
			return nil, err
		}
		return &RegisterDeviceTokenOutput{UID: existing.UID}, nil
	}

	// Create new token
	token := domain.NewDeviceToken(input.UserUID, input.Token, input.Platform)
	if err := uc.deviceTokenRepo.Create(ctx, uc.db, token); err != nil {
		return nil, err
	}

	return &RegisterDeviceTokenOutput{UID: token.UID}, nil
}
