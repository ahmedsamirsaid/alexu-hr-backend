package domain

import "time"

type DeviceToken struct {
	ID        int64
	UID       string
	UserUID   string
	Token     string
	Platform  string // "android" or "ios"
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewDeviceToken(userUID, token, platform string) *DeviceToken {
	return &DeviceToken{
		UID:      GenerateUID("dtkn"),
		UserUID:  userUID,
		Token:    token,
		Platform: platform,
	}
}
