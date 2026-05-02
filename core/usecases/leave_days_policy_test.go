package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type mockWeekendConfigRepoForLeaveDays struct {
	weekendDays []int
}

func (m *mockWeekendConfigRepoForLeaveDays) List(ctx context.Context, q ports.Querier) ([]*domain.WeekendConfig, error) {
	return nil, nil
}

func (m *mockWeekendConfigRepoForLeaveDays) GetWeekendDays(ctx context.Context, q ports.Querier) ([]int, error) {
	return m.weekendDays, nil
}

type mockHolidayDefinitionRepoForLeaveDays struct {
	holidays []*domain.HolidayDefinition
}

func (m *mockHolidayDefinitionRepoForLeaveDays) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.HolidayDefinition, error) {
	return nil, nil
}

func (m *mockHolidayDefinitionRepoForLeaveDays) GetByCode(ctx context.Context, q ports.Querier, code string) (*domain.HolidayDefinition, error) {
	return nil, nil
}

func (m *mockHolidayDefinitionRepoForLeaveDays) GetByDate(ctx context.Context, q ports.Querier, date time.Time) ([]*domain.HolidayDefinition, error) {
	return m.holidays, nil
}

func (m *mockHolidayDefinitionRepoForLeaveDays) ListAllByDateRange(ctx context.Context, q ports.Querier, start, end time.Time) ([]*domain.HolidayDefinition, error) {
	return m.ListByDateRange(ctx, q, start, end)
}

func (m *mockHolidayDefinitionRepoForLeaveDays) ListByDateRangeForDepartment(ctx context.Context, q ports.Querier, start, end time.Time, departmentUID string) ([]*domain.HolidayDefinition, error) {
	return m.ListByDateRange(ctx, q, start, end)
}

func (m *mockHolidayDefinitionRepoForLeaveDays) ListByDateRange(ctx context.Context, q ports.Querier, start, end time.Time) ([]*domain.HolidayDefinition, error) {
	var result []*domain.HolidayDefinition
	for _, h := range m.holidays {
		if !h.Date.Before(start) && !h.Date.After(end) {
			result = append(result, h)
		}
	}
	return result, nil
}

func (m *mockHolidayDefinitionRepoForLeaveDays) List(ctx context.Context, q ports.Querier) ([]*domain.HolidayDefinition, error) {
	return m.holidays, nil
}

func (m *mockHolidayDefinitionRepoForLeaveDays) Create(ctx context.Context, q ports.Querier, def *domain.HolidayDefinition) error {
	return nil
}

func (m *mockHolidayDefinitionRepoForLeaveDays) Update(ctx context.Context, q ports.Querier, def *domain.HolidayDefinition) error {
	return nil
}

func (m *mockHolidayDefinitionRepoForLeaveDays) Delete(ctx context.Context, q ports.Querier, id int64) error {
	return nil
}

func TestCalculateLeaveDaysPolicy(t *testing.T) {
	weekendRepo := &mockWeekendConfigRepoForLeaveDays{weekendDays: []int{5, 6}}
	holidayRepo := &mockHolidayDefinitionRepoForLeaveDays{
		holidays: []*domain.HolidayDefinition{
			{Date: time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC)},
		},
	}
	calc := NewWorkingDaysCalculator(weekendRepo, holidayRepo)

	start := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		leaveType *domain.LeaveType
		expected  int
	}{
		{name: "casual uses working days", leaveType: &domain.LeaveType{Code: "CASUAL"}, expected: 0},
		{name: "regular uses working days", leaveType: &domain.LeaveType{Code: "REGULAR"}, expected: 0},
		{name: "other leave types use calendar days", leaveType: &domain.LeaveType{Code: "SPECIAL"}, expected: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calculateLeaveDays(context.Background(), nil, calc, tt.leaveType, start, end)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Fatalf("result = %d, want %d", result, tt.expected)
			}
		})
	}
}
