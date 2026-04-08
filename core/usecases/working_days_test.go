package usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
)

type mockWeekendConfigRepo struct {
	weekendDays []int
}

func (m *mockWeekendConfigRepo) List(ctx context.Context, q ports.Querier) ([]*domain.WeekendConfig, error) {
	return nil, nil
}

func (m *mockWeekendConfigRepo) GetWeekendDays(ctx context.Context, q ports.Querier) ([]int, error) {
	return m.weekendDays, nil
}

type mockHolidayInstanceRepo struct {
	holidays []*domain.HolidayInstance
}

func (m *mockHolidayInstanceRepo) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.HolidayInstance, error) {
	return nil, nil
}

func (m *mockHolidayInstanceRepo) GetByDefinitionAndYear(ctx context.Context, q ports.Querier, definitionID int64, year int) (*domain.HolidayInstance, error) {
	return nil, nil
}

func (m *mockHolidayInstanceRepo) ListByYear(ctx context.Context, q ports.Querier, year int) ([]*domain.HolidayInstance, error) {
	return m.holidays, nil
}

func (m *mockHolidayInstanceRepo) ListByDateRange(ctx context.Context, q ports.Querier, start, end time.Time) ([]*domain.HolidayInstance, error) {
	var result []*domain.HolidayInstance
	for _, h := range m.holidays {
		if !h.ObservedDate.Before(start) && !h.ObservedDate.After(end) {
			result = append(result, h)
		}
	}
	return result, nil
}

func (m *mockHolidayInstanceRepo) Create(ctx context.Context, q ports.Querier, instance *domain.HolidayInstance) error {
	return nil
}

func (m *mockHolidayInstanceRepo) Update(ctx context.Context, q ports.Querier, instance *domain.HolidayInstance) error {
	return nil
}

func TestWorkingDaysCalculator_CalculateWorkingDays(t *testing.T) {
	tests := []struct {
		name        string
		start       time.Time
		end         time.Time
		weekendDays []int
		holidays    []*domain.HolidayInstance
		expected    int
	}{
		{
			name:        "single working day",
			start:       time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC), // Monday
			end:         time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC),
			weekendDays: []int{5, 6}, // Friday, Saturday (Egyptian weekend)
			holidays:    nil,
			expected:    1,
		},
		{
			name:        "full work week (Sun-Thu)",
			start:       time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC), // Sunday
			end:         time.Date(2025, 1, 9, 0, 0, 0, 0, time.UTC), // Thursday
			weekendDays: []int{5, 6},                                 // Friday, Saturday
			holidays:    nil,
			expected:    5,
		},
		{
			name:        "week with weekend",
			start:       time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC),  // Sunday
			end:         time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC), // Saturday
			weekendDays: []int{5, 6},                                  // Friday, Saturday
			holidays:    nil,
			expected:    5,
		},
		{
			name:        "week with holiday",
			start:       time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC), // Sunday
			end:         time.Date(2025, 1, 9, 0, 0, 0, 0, time.UTC), // Thursday
			weekendDays: []int{5, 6},
			holidays: []*domain.HolidayInstance{
				{ObservedDate: time.Date(2025, 1, 7, 0, 0, 0, 0, time.UTC)}, // Tuesday holiday
			},
			expected: 4,
		},
		{
			name:        "two consecutive days",
			start:       time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC), // Monday
			end:         time.Date(2025, 1, 7, 0, 0, 0, 0, time.UTC), // Tuesday
			weekendDays: []int{5, 6},
			holidays:    nil,
			expected:    2,
		},
		{
			name:        "weekend days only",
			start:       time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC), // Friday
			end:         time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC), // Saturday
			weekendDays: []int{5, 6},
			holidays:    nil,
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weekendRepo := &mockWeekendConfigRepo{weekendDays: tt.weekendDays}
			holidayRepo := &mockHolidayInstanceRepo{holidays: tt.holidays}

			calc := usecases.NewWorkingDaysCalculator(weekendRepo, holidayRepo)

			result, err := calc.CalculateWorkingDays(context.Background(), nil, tt.start, tt.end)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("expected %d working days, got %d", tt.expected, result)
			}
		})
	}
}
