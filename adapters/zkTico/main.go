package main

import (
	"context"
	"time"
)

type AttendanceEvent struct {
	UserID     int
	AttendedAt time.Time
}

type ZKDevice interface {
	Connect() error
	LiveCapture() (<-chan *AttendanceEvent, error)
	StopCapture()
}

func RunListener(ctx context.Context, device ZKDevice, handle func(*AttendanceEvent)) error {
	if err := device.Connect(); err != nil {
		return err
	}

	ch, err := device.LiveCapture()
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			device.StopCapture()
			return nil
		case event, ok := <-ch:
			if !ok {
				return nil
			}
			handle(event)
		}
	}
}