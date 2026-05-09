package usecases

import (
	"context"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type LogoutInput struct {
	RefreshToken string
}

type LogoutUseCase struct {
	db               ports.DB
	refreshTokenRepo ports.RefreshTokenRepository
	userRepo         ports.UserRepository
	auditor          audit.Auditor
}

func NewLogoutUseCase(
	db ports.DB,
	refreshTokenRepo ports.RefreshTokenRepository,
	userRepo ports.UserRepository,
	auditor audit.Auditor,
) *LogoutUseCase {
	return &LogoutUseCase{
		db:               db,
		refreshTokenRepo: refreshTokenRepo,
		userRepo:         userRepo,
		auditor:          auditor,
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

	// Get user info for audit log (skip admin)
	user, err := uc.userRepo.GetByID(ctx, uc.db, storedToken.UserID)
	if err == nil && user != nil && user.Phone != "+201000000000" {
		actorUID := user.UID // Use user UID as actor if no employee UID
		if user.EmployeeUID != nil {
			actorUID = *user.EmployeeUID
		}
		actionParams := map[string]interface{}{
			"Phone": user.Phone,
		}
		uc.auditor.Actor(actorUID).
			Did(audit.ActionLogout).
			On(audit.EntityUser, user.UID).
			WithMeta("action_key", "audit.sentence.user_logout").
			WithMeta("action_params", actionParams).
			Save(ctx)
	}

	return uc.refreshTokenRepo.Revoke(ctx, uc.db, storedToken.ID)
}
