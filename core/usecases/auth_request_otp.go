package usecases

import (
	"context"
	"errors"
	"log/slog"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type RequestOTPInput struct {
	Phone string
}

type RequestOTPOutput struct {
	// OTP is only returned in dev mode for testing
	OTP string `json:"-"`
}

var ErrFailedToSendOTP = errors.New("failed_to_send_OTP")

type RequestOTPUseCase struct {
	db       ports.DB
	otpRepo  ports.OTPRepository
	userRepo ports.UserRepository
}

func NewRequestOTPUseCase(
	db ports.DB,
	otpRepo ports.OTPRepository,
	userRepo ports.UserRepository,
) *RequestOTPUseCase {
	return &RequestOTPUseCase{
		db:       db,
		otpRepo:  otpRepo,
		userRepo: userRepo,
	}
}

func (uc *RequestOTPUseCase) Execute(ctx context.Context, input RequestOTPInput) (*RequestOTPOutput, error) {
	user, err := uc.userRepo.GetByPhone(ctx, uc.db, input.Phone)
	if err != nil {
		return nil, ErrFailedToSendOTP
	}

	if user == nil {
		slog.Error("auth_request_otp.Execute.user_not_found", "phone", input.Phone)
		return nil, ErrFailedToSendOTP
	}

	otp := domain.NewOTPCode(input.Phone)

	if err := uc.otpRepo.Create(ctx, uc.db, otp); err != nil {
		return nil, ErrFailedToSendOTP
	}
	// In production, this would send the OTP via SMS
	// For now, just log it
	slog.Info("auth_request_otp.Execute.otp_generated", "otp", otp.Code, "phone", input.Phone)

	return &RequestOTPOutput{
		OTP: otp.Code,
	}, nil
}
