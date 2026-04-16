package usecases

import (
	"context"
	"sort"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type GetDailyAttendanceSummaryInput struct {
	Date        time.Time
	EmployeeUID *string
}

type EmployeeDailyAttendanceSummary struct {
	EmployeeUID     string
	EmployeeName    string
	CheckIn         *time.Time
	CheckOut        *time.Time
	WorkedHours     float64
	LateArrival     bool
	EarlyDeparture  bool
	MissingCheckIn  bool
	MissingCheckOut bool
}

type GetDailyAttendanceSummaryOutput struct {
	Date      time.Time
	Summaries []EmployeeDailyAttendanceSummary
}

type GetDailyAttendanceSummaryUseCase struct {
	db           ports.DB
	recordRepo   ports.AttendanceRecordRepository
	employeeRepo ports.EmployeeRepository
	deptRepo     ports.DepartmentRepository
	shiftRepo    ports.ShiftRepository
}

func NewGetDailyAttendanceSummaryUseCase(
	db ports.DB,
	recordRepo ports.AttendanceRecordRepository,
	employeeRepo ports.EmployeeRepository,
	deptRepo ports.DepartmentRepository,
	shiftRepo ports.ShiftRepository,
) *GetDailyAttendanceSummaryUseCase {
	return &GetDailyAttendanceSummaryUseCase{
		db:           db,
		recordRepo:   recordRepo,
		employeeRepo: employeeRepo,
		deptRepo:     deptRepo,
		shiftRepo:    shiftRepo,
	}
}

func (uc *GetDailyAttendanceSummaryUseCase) Execute(ctx context.Context, input GetDailyAttendanceSummaryInput) (*GetDailyAttendanceSummaryOutput, error) {
	records, err := uc.recordRepo.ListByDate(ctx, uc.db, input.Date, input.EmployeeUID)
	if err != nil {
		return nil, err
	}

	type accumulator struct {
		checkIn  *time.Time
		checkOut *time.Time
	}

	byEmployee := make(map[string]*accumulator)
	for _, rec := range records {
		item, ok := byEmployee[rec.EmployeeUID]
		if !ok {
			item = &accumulator{}
			byEmployee[rec.EmployeeUID] = item
		}

		switch rec.PunchType {
		case domain.AttendancePunchTypeCheckIn:
			if item.checkIn == nil || rec.PunchedAt.Before(*item.checkIn) {
				t := rec.PunchedAt
				item.checkIn = &t
			}
		case domain.AttendancePunchTypeCheckOut:
			if item.checkOut == nil || rec.PunchedAt.After(*item.checkOut) {
				t := rec.PunchedAt
				item.checkOut = &t
			}
		case domain.AttendancePunchTypeUnknown:
			if item.checkIn == nil || rec.PunchedAt.Before(*item.checkIn) {
				t := rec.PunchedAt
				item.checkIn = &t
			}
			if item.checkOut == nil || rec.PunchedAt.After(*item.checkOut) {
				t := rec.PunchedAt
				item.checkOut = &t
			}
		}
	}

	result := make([]EmployeeDailyAttendanceSummary, 0, len(byEmployee))
	for employeeUID, item := range byEmployee {
		shift, err := resolveEffectiveShift(ctx, uc.db, uc.employeeRepo, uc.deptRepo, uc.shiftRepo, employeeUID, nil)
		if err != nil {
			return nil, err
		}

		workStart, err := buildTimeOnDate(input.Date, shift.StartTime)
		if err != nil {
			return nil, ErrInvalidWorkDayStart
		}
		workEnd, err := buildTimeOnDate(input.Date, shift.EndTime)
		if err != nil {
			return nil, ErrInvalidWorkDayEnd
		}

		summary := EmployeeDailyAttendanceSummary{
			EmployeeUID:     employeeUID,
			CheckIn:         item.checkIn,
			CheckOut:        item.checkOut,
			MissingCheckIn:  item.checkIn == nil,
			MissingCheckOut: item.checkOut == nil,
		}

		if item.checkIn != nil && item.checkOut != nil && item.checkOut.After(*item.checkIn) {
			summary.WorkedHours = item.checkOut.Sub(*item.checkIn).Hours()
		}

		if item.checkIn != nil {
			lateThreshold := workStart.Add(time.Duration(shift.GraceMinutes) * time.Minute)
			summary.LateArrival = item.checkIn.After(lateThreshold)
		}

		if item.checkOut != nil {
			earlyThreshold := workEnd.Add(-time.Duration(shift.GraceMinutes) * time.Minute)
			summary.EarlyDeparture = item.checkOut.Before(earlyThreshold)
		}

		employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, employeeUID)
		if err != nil {
			return nil, err
		}
		if employee != nil {
			summary.EmployeeName = employee.Name
		}

		result = append(result, summary)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].EmployeeName == result[j].EmployeeName {
			return result[i].EmployeeUID < result[j].EmployeeUID
		}
		return result[i].EmployeeName < result[j].EmployeeName
	})

	return &GetDailyAttendanceSummaryOutput{
		Date:      input.Date,
		Summaries: result,
	}, nil
}

func buildTimeOnDate(date time.Time, hhmm string) (time.Time, error) {
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
		time.Local,
	), nil
}
