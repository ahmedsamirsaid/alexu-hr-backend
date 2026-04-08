package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type LogoutInput struct {
	RefreshToken string
}

type LogoutUseCase struct {
	db               ports.DB
	refreshTokenRepo ports.RefreshTokenRepository
}

func NewLogoutUseCase(
	db ports.DB,
	refreshTokenRepo ports.RefreshTokenRepository,
) *LogoutUseCase {
	return &LogoutUseCase{
		db:               db,
		refreshTokenRepo: refreshTokenRepo,
	}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, input LogoutInput) error {
	tokenHash := domain.HashToken(input.RefreshToken)

	storedToken, err := uc.refreshTokenRepo.GetByHash(ctx, uc.db, tokenHash)
	if err != nil {
		return err
	}
	if storedToken == nil {
		return nil // Already logged out or invalid token
	}

	return uc.refreshTokenRepo.Revoke(ctx, uc.db, storedToken.ID)
}
