package usecases

import (
	"context"
	"strings"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type PermissionTypeEligibility struct {
	Type          domain.PermissionType
	AllowedDates  []time.Time
	WeeklyCapHit  bool
	BlockReason   string // empty when eligible
	EffectiveStart string // HH:MM (effective window start for the *first* allowed date)
	EffectiveEnd   string // HH:MM
}

type GetPermissionEligibilityOutput struct {
	ShiftStart    string // HH:MM
	ShiftEnd      string // HH:MM
	GraceMinutes  int
	Eligibility   map[domain.PermissionType]PermissionTypeEligibility
}

type GetPermissionEligibilityUseCase struct {
	db             ports.DB
	employeeRepo   ports.EmployeeRepository
	deptRepo       ports.DepartmentRepository
	shiftRepo      ports.ShiftRepository
	permissionRepo ports.PermissionRequestRepository
	weekendRepo    ports.WeekendConfigRepository
	location       *time.Location
}

func NewGetPermissionEligibilityUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	deptRepo ports.DepartmentRepository,
	shiftRepo ports.ShiftRepository,
	permissionRepo ports.PermissionRequestRepository,
	weekendRepo ports.WeekendConfigRepository,
	timezone string,
) *GetPermissionEligibilityUseCase {
	loc, err := time.LoadLocation(strings.TrimSpace(timezone))
	if err != nil {
		loc = time.FixedZone("Africa/Cairo", 2*60*60)
	}
	return &GetPermissionEligibilityUseCase{
		db:             db,
		employeeRepo:   employeeRepo,
		deptRepo:       deptRepo,
		shiftRepo:      shiftRepo,
		permissionRepo: permissionRepo,
		weekendRepo:    weekendRepo,
		location:       loc,
	}
}

func (uc *GetPermissionEligibilityUseCase) Execute(ctx context.Context, employeeUID string) (*GetPermissionEligibilityOutput, error) {
	employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, employeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	shift, err := resolveEffectiveShift(ctx, uc.db, uc.employeeRepo, uc.deptRepo, uc.shiftRepo, employee.UID, employee.DepartmentUID)
	if err != nil {
		return nil, err
	}

	now := time.Now().In(uc.location)
	today := normalizeDateOnly(now)
	tomorrow := today.AddDate(0, 0, 1)
	cutoff13 := time.Date(today.Year(), today.Month(), today.Day(), 13, 0, 0, 0, uc.location)

	out := &GetPermissionEligibilityOutput{
		ShiftStart:   shift.StartTime,
		ShiftEnd:     shift.EndTime,
		GraceMinutes: shift.GraceMinutes,
		Eligibility:  map[domain.PermissionType]PermissionTypeEligibility{},
	}

	// Working-week filter: skip weekends so the calendar offers only days the employee actually
	// works. Holidays still aren't filtered (they're rare and don't change the validation rule).
	weekendDays, err := uc.weekendRepo.GetWeekendDays(ctx, uc.db)
	if err != nil {
		return nil, err
	}
	weekendSet := map[int]struct{}{}
	for _, d := range weekendDays {
		weekendSet[d] = struct{}{}
	}
	isWorkingDay := func(d time.Time) bool {
		_, isWeekend := weekendSet[int(d.Weekday())]
		return !isWeekend
	}

	// Per-type allowed dates (deadline rules)
	typeAllowedDates := map[domain.PermissionType][]time.Time{}
	// Morning: any future working day within a 60-day horizon. The 60 cap is generous; HR can
	// raise it later by tweaking this constant.
	typeAllowedDates[domain.PermissionTypeMorning] = workingDaysFrom(tomorrow, 60, isWorkingDay)
	if now.Before(cutoff13) && isWorkingDay(today) {
		typeAllowedDates[domain.PermissionTypePersonal] = []time.Time{today}
	} else {
		typeAllowedDates[domain.PermissionTypePersonal] = []time.Time{}
	}
	if isWorkingDay(today) {
		typeAllowedDates[domain.PermissionTypeOfficial] = []time.Time{today}
	} else {
		typeAllowedDates[domain.PermissionTypeOfficial] = []time.Time{}
	}
	hi := []time.Time{}
	if isWorkingDay(today) {
		hi = append(hi, today)
	}
	if isWorkingDay(tomorrow) {
		hi = append(hi, tomorrow)
	}
	typeAllowedDates[domain.PermissionTypeHealthInsurance] = hi

	// Compute effective window for each type using shift
	for _, t := range []domain.PermissionType{
		domain.PermissionTypeMorning,
		domain.PermissionTypePersonal,
		domain.PermissionTypeOfficial,
		domain.PermissionTypeHealthInsurance,
	} {
		dates := typeAllowedDates[t]

		entry := PermissionTypeEligibility{
			Type:         t,
			AllowedDates: dates,
		}

		if len(dates) == 0 {
			entry.BlockReason = "deadline_passed"
		}

		// Weekly cap check (against today's week — caller picks a date but we surface a warning)
		referenceDate := today
		if len(dates) > 0 {
			referenceDate = dates[0]
		}

		if t != domain.PermissionTypeOfficial {
			weekStart, weekEnd := WeekBoundsForDate(weekendDays, referenceDate)
			statuses := []domain.ApprovalRequestStatus{domain.ApprovalRequestStatusPending, domain.ApprovalRequestStatusApproved}

			var capTypes []domain.PermissionType
			switch t {
			case domain.PermissionTypeHealthInsurance:
				capTypes = []domain.PermissionType{domain.PermissionTypeHealthInsurance}
			case domain.PermissionTypeMorning, domain.PermissionTypePersonal:
				capTypes = []domain.PermissionType{domain.PermissionTypeMorning, domain.PermissionTypePersonal}
			}

			count, err := uc.permissionRepo.CountInWeek(ctx, uc.db, employee.UID, capTypes, weekStart, weekEnd, statuses, nil)
			if err != nil {
				return nil, err
			}
			if count >= 1 {
				entry.WeeklyCapHit = true
				if entry.BlockReason == "" {
					entry.BlockReason = "weekly_cap_reached"
				}
			}
		}

		// Effective window for the first allowed date — used to display "Morning: 09:00→10:00"
		if len(dates) > 0 {
			start, end := computeImplicitWindow(t, shift.StartTime, shift.EndTime)
			entry.EffectiveStart = start
			entry.EffectiveEnd = end
		}

		out.Eligibility[t] = entry
	}
	return out, nil
}

// workingDaysFrom returns the next `n` working days starting at `start`, skipping anything the
// caller's predicate rejects (typically weekends). The horizon is bounded so we don't loop
// forever in degenerate weekend configs.
func workingDaysFrom(start time.Time, n int, isWorking func(time.Time) bool) []time.Time {
	out := make([]time.Time, 0, n)
	cursor := start
	maxScan := n * 3
	for len(out) < n && maxScan > 0 {
		if isWorking(cursor) {
			out = append(out, cursor)
		}
		cursor = cursor.AddDate(0, 0, 1)
		maxScan--
	}
	return out
}

// computeImplicitWindow returns the visible "HH:MM" window for the implicit (no-window) types,
// or the shift bounds for explicit-window types so the UI can render the draggable handles.
func computeImplicitWindow(t domain.PermissionType, shiftStart, shiftEnd string) (string, string) {
	switch t {
	case domain.PermissionTypeMorning:
		return shiftStart, domain.PermissionMorningLatestCheckIn
	case domain.PermissionTypePersonal:
		return domain.PermissionPersonalEarliestLeave, shiftEnd
	case domain.PermissionTypeOfficial, domain.PermissionTypeHealthInsurance:
		return shiftStart, shiftEnd
	}
	return shiftStart, shiftEnd
}
