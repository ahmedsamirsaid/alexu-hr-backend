package main

import (
	"context"
	"log"
	"time"
)

func main() {
	mock := NewMockZKDevice()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := RunListener(ctx, mock, func(e *AttendanceEvent) {
			log.Printf("Mock event: user=%d time=%s", e.UserID, e.AttendedAt)
		})
		if err != nil {
			log.Println("listener error:", err)
		}
	}()
		
	mock.Events <- &AttendanceEvent{UserID: 1, AttendedAt: time.Now()}
	time.Sleep(time.Second)

	mock.Events <- &AttendanceEvent{UserID: 2, AttendedAt: time.Now()}
	time.Sleep(time.Second)

	cancel()
	time.Sleep(time.Second)
}