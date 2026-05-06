package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type permissionAdjuster struct {
	permissions []*domain.PermissionRequest
	shift       *domain.Shift
}

// newEmptyPermissionAdjuster returns an adjuster that never applies any adjustment.
// Use as the default until callers thread the permission repository through.
func newEmptyPermissionAdjuster() *permissionAdjuster {
	return &permissionAdjuster{}
}

func newPermissionAdjuster(permissions []*domain.PermissionRequest, shift *domain.Shift) *permissionAdjuster {
	return &permissionAdjuster{permissions: permissions, shift: shift}
}

// newPermissionAdjusterForDate fetches approved permissions for the given (employee, date).
// Safe to call with permRepo == nil — returns an empty adjuster.
func newPermissionAdjusterForDate(ctx context.Context, q ports.Querier, permRepo ports.PermissionRequestRepository, employeeUID string, date time.Time, shift *domain.Shift) (*permissionAdjuster, error) {
	if permRepo == nil {
		return newEmptyPermissionAdjuster(), nil
	}
	rows, err := permRepo.ListApprovedForEmployeeOnDate(ctx, q, employeeUID, date)
	if err != nil {
		return nil, err
	}
	return newPermissionAdjuster(rows, shift), nil
}

// PermissionUIDs returns the UIDs of all permissions backing this day, for UI surfacing.
func (a *permissionAdjuster) PermissionUIDs() []string {
	if a == nil {
		return nil
	}
	out := make([]string, 0, len(a.permissions))
	for _, p := range a.permissions {
		out = append(out, p.UID)
	}
	return out
}

func (a *permissionAdjuster) LateThreshold(defaultThreshold time.Time) time.Time {
	if a == nil || a.shift == nil || len(a.permissions) == 0 {
		return defaultThreshold
	}
	threshold := defaultThreshold
	for _, p := range a.permissions {
		switch p.Type {
		case domain.PermissionTypeMorning:
			cap, err := buildTimeOnDate(p.PermissionDate, domain.PermissionMorningLatestCheckIn)
			if err == nil && cap.After(threshold) {
				threshold = cap
			}
		case domain.PermissionTypeOfficial, domain.PermissionTypeHealthInsurance:
			if !p.TouchesShiftStart(a.shift.StartTime, a.shift.EndTime) {
				continue
			}
			_, end, err := p.EffectiveWindow(a.shift.StartTime, a.shift.EndTime)
			if err == nil && end.After(threshold) {
				threshold = end
			}
		}
	}
	return threshold
}

func (a *permissionAdjuster) EarlyThreshold(defaultThreshold time.Time) time.Time {
	if a == nil || a.shift == nil || len(a.permissions) == 0 {
		return defaultThreshold
	}
	threshold := defaultThreshold
	for _, p := range a.permissions {
		switch p.Type {
		case domain.PermissionTypePersonal:
			floor, err := buildTimeOnDate(p.PermissionDate, domain.PermissionPersonalEarliestLeave)
			if err == nil && floor.Before(threshold) {
				threshold = floor
			}
		case domain.PermissionTypeOfficial, domain.PermissionTypeHealthInsurance:
			if !p.TouchesShiftEnd(a.shift.StartTime, a.shift.EndTime) {
				continue
			}
			start, _, err := p.EffectiveWindow(a.shift.StartTime, a.shift.EndTime)
			if err == nil && start.Before(threshold) {
				threshold = start
			}
		}
	}
	return threshold
}

func (a *permissionAdjuster) MissingCheckOutHandling() (suppress bool, syntheticEarlyAt *time.Time) {
	if a == nil || a.shift == nil || len(a.permissions) == 0 {
		return false, nil
	}

	var earliest *time.Time
	for _, p := range a.permissions {
		var start time.Time
		switch p.Type {
		case domain.PermissionTypePersonal:
			s, _, err := p.EffectiveWindow(a.shift.StartTime, a.shift.EndTime)
			if err != nil {
				continue
			}
			start = s
		case domain.PermissionTypeOfficial, domain.PermissionTypeHealthInsurance:
			if !p.TouchesShiftEnd(a.shift.StartTime, a.shift.EndTime) {
				continue
			}
			s, _, err := p.EffectiveWindow(a.shift.StartTime, a.shift.EndTime)
			if err != nil {
				continue
			}
			start = s
		default:
			continue
		}
		if earliest == nil || start.Before(*earliest) {
			s := start
			earliest = &s
		}
	}
	if earliest == nil {
		return false, nil
	}
	return true, earliest
}
