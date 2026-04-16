package main

import (
	"context"
	"testing"
	"time"
)

func TestRunListener_ProcessesEvents(t *testing.T) {
	mock := NewMockZKDevice()

	var got []*AttendanceEvent

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)

	go func() {
		done <- RunListener(ctx, mock, func(e *AttendanceEvent) {
			got = append(got, e)
			if len(got) == 2 {
				cancel()
			}
		})
	}()

	mock.Events <- &AttendanceEvent{
		UserID:     101,
		AttendedAt: time.Date(2026, 4, 11, 9, 0, 0, 0, time.UTC),
	}

	mock.Events <- &AttendanceEvent{
		UserID:     102,
		AttendedAt: time.Date(2026, 4, 11, 9, 5, 0, 0, time.UTC),
	}

	err := <-done
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mock.Connected {
		t.Fatal("expected device to connect")
	}

	if !mock.Stopped {
		t.Fatal("expected device to stop on context cancel")
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 events, got %d", len(got))
	}

	if got[0].UserID != 101 {
		t.Fatalf("expected first user 101, got %d", got[0].UserID)
	}
}