package usecases

import (
	"errors"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
)

func TestValidateLeaveRequestRecordingDeadline(t *testing.T) {
	two := 2
	thirty := 30
	now := time.Date(2026, 4, 27, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		leaveType *domain.LeaveType
		startDate time.Time
		wantErr   error
	}{
		{
			name: "casual leave allows submission on same day",
			leaveType: &domain.LeaveType{
				Code:                  leaveTypeCodeCasual,
				RecordingDeadlineDays: &two,
			},
			startDate: time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "casual leave allows backdating by two days",
			leaveType: &domain.LeaveType{
				Code:                  leaveTypeCodeCasual,
				RecordingDeadlineDays: &two,
			},
			startDate: time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "casual leave rejects backdating by more than two days",
			leaveType: &domain.LeaveType{
				Code:                  leaveTypeCodeCasual,
				RecordingDeadlineDays: &two,
			},
			startDate: time.Date(2026, 4, 24, 0, 0, 0, 0, time.UTC),
			wantErr:   ErrLeaveRequestOutsideDeadline,
		},
		{
			name: "regular leave allows when exactly at deadline days ahead",
			leaveType: &domain.LeaveType{
				Code:                  leaveTypeCodeRegular,
				RecordingDeadlineDays: &two,
			},
			startDate: time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "regular leave allows further than deadline days ahead",
			leaveType: &domain.LeaveType{
				Code:                  leaveTypeCodeRegular,
				RecordingDeadlineDays: &two,
			},
			startDate: time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "regular leave rejects when less than deadline days ahead",
			leaveType: &domain.LeaveType{
				Code:                  leaveTypeCodeRegular,
				RecordingDeadlineDays: &two,
			},
			startDate: time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC),
			wantErr:   ErrLeaveRequestOutsideDeadline,
		},
		{
			name: "other leave types use configured deadline for future requests",
			leaveType: &domain.LeaveType{
				Code:                  "SICK",
				RecordingDeadlineDays: &thirty,
			},
			startDate: time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "other leave types reject when less than configured deadline ahead",
			leaveType: &domain.LeaveType{
				Code:                  "SICK",
				RecordingDeadlineDays: &thirty,
			},
			startDate: time.Date(2026, 5, 26, 0, 0, 0, 0, time.UTC),
			wantErr:   ErrLeaveRequestOutsideDeadline,
		},
		{
			name: "non casual leave rejects same day submission",
			leaveType: &domain.LeaveType{
				Code:                  leaveTypeCodeRegular,
				RecordingDeadlineDays: &two,
			},
			startDate: time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
			wantErr:   ErrLeaveRequestOutsideDeadline,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLeaveRequestRecordingDeadline(tt.leaveType, tt.startDate, now)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("validateLeaveRequestRecordingDeadline() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateLeaveRequestRecordingDeadline_ReturnsUserFacingMessage(t *testing.T) {
	two := 2
	now := time.Date(2026, 4, 27, 10, 30, 0, 0, time.UTC)
	leaveType := &domain.LeaveType{
		Code:                  leaveTypeCodeRegular,
		RecordingDeadlineDays: &two,
	}

	err := validateLeaveRequestRecordingDeadline(leaveType, time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC), now)
	if !errors.Is(err, ErrLeaveRequestOutsideDeadline) {
		t.Fatalf("validateLeaveRequestRecordingDeadline() error = %v, want %v", err, ErrLeaveRequestOutsideDeadline)
	}

	want := "You must record this leave at least 2 days before the start date."
	if err == nil || err.Error() != want {
		t.Fatalf("validateLeaveRequestRecordingDeadline() message = %v, want %q", err, want)
	}
}
