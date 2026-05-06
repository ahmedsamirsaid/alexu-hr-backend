package domain

import (
	"errors"
	"fmt"
	"time"
)

type PermissionType string

const (
	PermissionTypeMorning         PermissionType = "morning"
	PermissionTypePersonal        PermissionType = "personal"
	PermissionTypeOfficial        PermissionType = "official"
	PermissionTypeHealthInsurance PermissionType = "health_insurance"
)

func (t PermissionType) IsValid() bool {
	switch t {
	case PermissionTypeMorning, PermissionTypePersonal, PermissionTypeOfficial, PermissionTypeHealthInsurance:
		return true
	}
	return false
}

func (t PermissionType) NameAR() string {
	switch t {
	case PermissionTypeMorning:
		return "إذن صباحي"
	case PermissionTypePersonal:
		return "إذن شخصي"
	case PermissionTypeOfficial:
		return "إذن مصلحي"
	case PermissionTypeHealthInsurance:
		return "إذن تأمين صحي"
	}
	return string(t)
}

func (t PermissionType) NameEN() string {
	switch t {
	case PermissionTypeMorning:
		return "Morning Permission"
	case PermissionTypePersonal:
		return "Personal Permission"
	case PermissionTypeOfficial:
		return "Official Permission"
	case PermissionTypeHealthInsurance:
		return "Health Insurance Permission"
	}
	return string(t)
}

// HasExplicitTimeWindow returns true for types whose start/end times are set by the requester.
// Morning and personal permissions use the system-wide caps (10:00 AM and 1:00 PM) instead.
func (t PermissionType) HasExplicitTimeWindow() bool {
	return t == PermissionTypeOfficial || t == PermissionTypeHealthInsurance
}

// System-wide constants (NOT shift-relative).
const (
	PermissionMorningLatestCheckIn   = "10:00" // morning permission caps late check-in to this clock time
	PermissionPersonalEarliestLeave  = "13:00" // personal permission unlocks early check-out from this clock time
)

type PermissionRequest struct {
	ID             int64
	UID            string
	EmployeeUID    string
	Type           PermissionType
	PermissionDate time.Time
	// StartTime / EndTime are always populated, in "HH:MM" form. For morning / personal types
	// they are resolved at submission time from the employee's effective shift and the system
	// constants (10:00 latest check-in / 13:00 earliest leave). For official / health_insurance
	// they hold the caller-supplied window (validated to be inside the shift).
	StartTime          string
	EndTime            string
	Reason             *string
	SubmittedAt        time.Time
	DecidedAt          *time.Time
	ApprovalRequestUID string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func NewPermissionRequest(
	employeeUID string,
	permissionType PermissionType,
	permissionDate time.Time,
	startTime, endTime string,
	reason *string,
	approvalRequestUID string,
) *PermissionRequest {
	return &PermissionRequest{
		UID:                GenerateUID("prq"),
		EmployeeUID:        employeeUID,
		Type:               permissionType,
		PermissionDate:     permissionDate,
		StartTime:          startTime,
		EndTime:            endTime,
		Reason:             reason,
		SubmittedAt:        time.Now(),
		ApprovalRequestUID: approvalRequestUID,
	}
}

func (p *PermissionRequest) SetDecided() {
	now := time.Now()
	p.DecidedAt = &now
}

// EffectiveWindow returns the [start, end] clock window the permission covers on PermissionDate.
// shiftStart and shiftEnd are "HH:MM" formatted clock times from the employee's effective shift.
//
// - morning           → [shiftStart, max(shiftStart, 10:00)] (capped to shiftEnd)
// - personal          → [min(shiftEnd, 13:00), shiftEnd]
// - official / health → [StartTime, EndTime] (clamped to [shiftStart, shiftEnd])
func (p *PermissionRequest) EffectiveWindow(shiftStart, shiftEnd string) (time.Time, time.Time, error) {
	dayStart, err := buildPermissionTimeOnDate(p.PermissionDate, shiftStart)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("permission: invalid shift start: %w", err)
	}
	dayEnd, err := buildPermissionTimeOnDate(p.PermissionDate, shiftEnd)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("permission: invalid shift end: %w", err)
	}

	switch p.Type {
	case PermissionTypeMorning:
		cap, err := buildPermissionTimeOnDate(p.PermissionDate, PermissionMorningLatestCheckIn)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		end := cap
		if end.Before(dayStart) {
			end = dayStart
		}
		if end.After(dayEnd) {
			end = dayEnd
		}
		return dayStart, end, nil

	case PermissionTypePersonal:
		floor, err := buildPermissionTimeOnDate(p.PermissionDate, PermissionPersonalEarliestLeave)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		start := floor
		if start.After(dayEnd) {
			start = dayEnd
		}
		if start.Before(dayStart) {
			start = dayStart
		}
		return start, dayEnd, nil

	case PermissionTypeOfficial, PermissionTypeHealthInsurance:
		if p.StartTime == "" || p.EndTime == "" {
			return time.Time{}, time.Time{}, ErrPermissionWindowMissing
		}
		start, err := buildPermissionTimeOnDate(p.PermissionDate, p.StartTime)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		end, err := buildPermissionTimeOnDate(p.PermissionDate, p.EndTime)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		if start.Before(dayStart) {
			start = dayStart
		}
		if end.After(dayEnd) {
			end = dayEnd
		}
		if !end.After(start) {
			return time.Time{}, time.Time{}, ErrPermissionWindowInvalid
		}
		return start, end, nil
	}

	return time.Time{}, time.Time{}, ErrPermissionTypeInvalid
}

// TouchesShiftStart reports whether the permission window begins exactly at the shift start —
// i.e. the permission extends the late-arrival allowance.
func (p *PermissionRequest) TouchesShiftStart(shiftStart, shiftEnd string) bool {
	start, _, err := p.EffectiveWindow(shiftStart, shiftEnd)
	if err != nil {
		return false
	}
	dayStart, err := buildPermissionTimeOnDate(p.PermissionDate, shiftStart)
	if err != nil {
		return false
	}
	return !start.After(dayStart)
}

// TouchesShiftEnd reports whether the permission window ends exactly at the shift end —
// i.e. the permission extends the early-departure allowance / converts missing-out to left-early.
func (p *PermissionRequest) TouchesShiftEnd(shiftStart, shiftEnd string) bool {
	_, end, err := p.EffectiveWindow(shiftStart, shiftEnd)
	if err != nil {
		return false
	}
	dayEnd, err := buildPermissionTimeOnDate(p.PermissionDate, shiftEnd)
	if err != nil {
		return false
	}
	return !end.Before(dayEnd)
}

var (
	ErrPermissionTypeInvalid   = errors.New("permission: invalid type")
	ErrPermissionWindowMissing = errors.New("permission: time window is required for this type")
	ErrPermissionWindowInvalid = errors.New("permission: end time must be after start time")
)

// permissionLoc matches attendanceLoc in the usecases package — fixed UTC+2.
// Both must use the same zone so that shift-time comparisons are consistent.
var permissionLoc = time.FixedZone("Africa/Cairo", 2*60*60)

func buildPermissionTimeOnDate(date time.Time, hhmm string) (time.Time, error) {
	parsed, err := time.Parse("15:04", hhmm)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(
		date.Year(),
		date.Month(),
		date.Day(),
		parsed.Hour(),
		parsed.Minute(),
		0,
		0,
		permissionLoc,
	), nil
}
