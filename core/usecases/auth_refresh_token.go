package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type RefreshTokenInput struct {
	RefreshToken string
}

type RefreshTokenOutput struct {
	AccessToken  string
	RefreshToken string
}

type RefreshTokenUseCase struct {
	db               ports.DB
	userRepo         ports.UserRepository
	roleRepo         ports.RoleRepository
	refreshTokenRepo ports.RefreshTokenRepository
	jwtService       JWTService
	refreshTokenDays int
}

func NewRefreshTokenUseCase(
	db ports.DB,
	userRepo ports.UserRepository,
	roleRepo ports.RoleRepository,
	refreshTokenRepo ports.RefreshTokenRepository,
	jwtService JWTService,
	refreshTokenDays int,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		db:               db,
		userRepo:         userRepo,
		roleRepo:         roleRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
		refreshTokenDays: refreshTokenDays,
	}
}

func (uc *RefreshTokenUseCase) Execute(ctx context.Context, input RefreshTokenInput) (*RefreshTokenOutput, error) {
	tokenHash := domain.HashToken(input.RefreshToken)

	// Validate token before starting transaction
	storedToken, err := uc.refreshTokenRepo.GetByHash(ctx, uc.db, tokenHash)
	if err != nil {
		return nil, err
	}
	if storedToken == nil || !storedToken.IsValid() {
		return nil, ErrInvalidRefreshToken
	}

	// Start transaction for token rotation
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Revoke old token (rotation)
	if err := uc.refreshTokenRepo.Revoke(ctx, tx, storedToken.ID); err != nil {
		return nil, err
	}

	// Get user
	user, err := uc.userRepo.GetByID(ctx, tx, storedToken.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.IsActive {
		return nil, ErrInvalidRefreshToken
	}

	// Load roles
	roles, err := uc.roleRepo.GetRolesForUser(ctx, tx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Roles = make([]domain.Role, len(roles))
	for i, r := range roles {
		user.Roles[i] = *r
	}

	// Generate new tokens
	accessToken, err := uc.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	newRefreshToken := domain.NewRefreshToken(user.ID, uc.refreshTokenDays)
	if err := uc.refreshTokenRepo.Create(ctx, tx, &newRefreshToken.RefreshToken); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &RefreshTokenOutput{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken.RawToken,
	}, nil
}
