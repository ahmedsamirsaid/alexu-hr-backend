package usecases

import (
	"context"
	"errors"

	"github.com/banumusa/backend/core/ports"
)

var ErrTestPushUserNotFound = errors.New("user not found")

type SendTestPushNotificationInput struct {
	UserUID string
	Title   string
	Body    string
	Data    ports.NotificationData
}

type SendTestPushNotificationOutput struct {
	SentCount int `json:"sentCount"`
}

type SendTestPushNotificationUseCase struct {
	db                  ports.DB
	userRepo            ports.UserRepository
	notificationService ports.NotificationService
}

func NewSendTestPushNotificationUseCase(
	db ports.DB,
	userRepo ports.UserRepository,
	notificationService ports.NotificationService,
) *SendTestPushNotificationUseCase {
	return &SendTestPushNotificationUseCase{
		db:                  db,
		userRepo:            userRepo,
		notificationService: notificationService,
	}
}

func (uc *SendTestPushNotificationUseCase) Execute(ctx context.Context, input SendTestPushNotificationInput) (*SendTestPushNotificationOutput, error) {
	user, err := uc.userRepo.GetByUID(ctx, uc.db, input.UserUID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrTestPushUserNotFound
	}

	title := input.Title
	if title == "" {
		title = "Test Push"
	}
	body := input.Body
	if body == "" {
		body = "Manual push notification test"
	}

	data := input.Data
	if data == nil {
		data = ports.NotificationData{}
	}
	data["type"] = "manual_test_push"

	sentCount, err := uc.notificationService.SendToUser(user.UID, title, body, nil, data)
	if err != nil {
		return nil, err
	}

	return &SendTestPushNotificationOutput{SentCount: sentCount}, nil
}
