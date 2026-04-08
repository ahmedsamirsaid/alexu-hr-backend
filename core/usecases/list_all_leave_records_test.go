package usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
)

type mockLeaveTypeRepoForAll struct {
	leaveTypes []*domain.LeaveType
}

func (m *mockLeaveTypeRepoForAll) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.LeaveType, error) {
	for _, lt := range m.leaveTypes {
		if lt.ID == id {
			return lt, nil
		}
	}
	return nil, nil
}

func (m *mockLeaveTypeRepoForAll) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.LeaveType, error) {
	for _, lt := range m.leaveTypes {
		if lt.UID == uid {
			return lt, nil
		}
	}
	return nil, nil
}

func (m *mockLeaveTypeRepoForAll) GetByCode(ctx context.Context, q ports.Querier, code string) (*domain.LeaveType, error) {
	for _, lt := range m.leaveTypes {
		if lt.Code == code {
			return lt, nil
		}
	}
	return nil, nil
}

func (m *mockLeaveTypeRepoForAll) List(ctx context.Context, q ports.Querier, activeOnly bool) ([]*domain.LeaveType, error) {
	return m.leaveTypes, nil
}

func (m *mockLeaveTypeRepoForAll) SetActive(ctx context.Context, q ports.Querier, uid string, isActive bool) error {
	return nil
}

type mockLeaveRecordRepoForAll struct {
	records []*ports.LeaveRecordWithEmployee
	total   int
}

func (m *mockLeaveRecordRepoForAll) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.LeaveRecord, error) {
	return nil, nil
}

func (m *mockLeaveRecordRepoForAll) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.LeaveRecord, error) {
	return nil, nil
}

func (m *mockLeaveRecordRepoForAll) Create(ctx context.Context, q ports.Querier, record *domain.LeaveRecord) error {
	return nil
}

func (m *mockLeaveRecordRepoForAll) ListByEmployee(ctx context.Context, q ports.Querier, employeeID int64) ([]*domain.LeaveRecord, error) {
	return nil, nil
}

func (m *mockLeaveRecordRepoForAll) ListByEmployeePaginated(ctx context.Context, q ports.Querier, employeeID int64, limit, offset int) ([]*domain.LeaveRecord, error) {
	return nil, nil
}

func (m *mockLeaveRecordRepoForAll) CountByEmployee(ctx context.Context, q ports.Querier, employeeID int64) (int, error) {
	return 0, nil
}

func (m *mockLeaveRecordRepoForAll) ListByEmployeeAndDateRange(ctx context.Context, q ports.Querier, employeeID int64, start, end time.Time) ([]*domain.LeaveRecord, error) {
	return nil, nil
}

func (m *mockLeaveRecordRepoForAll) ListByEmployeeAndType(ctx context.Context, q ports.Querier, employeeID, leaveTypeID int64) ([]*domain.LeaveRecord, error) {
	return nil, nil
}

func (m *mockLeaveRecordRepoForAll) ListAllPaginated(ctx context.Context, q ports.Querier, filter ports.ListAllLeaveRecordsFilter, limit, offset int) ([]*ports.LeaveRecordWithEmployee, error) {
	return m.records, nil
}

func (m *mockLeaveRecordRepoForAll) CountAll(ctx context.Context, q ports.Querier, filter ports.ListAllLeaveRecordsFilter) (int, error) {
	return m.total, nil
}

func (m *mockLeaveRecordRepoForAll) CountOnLeaveToday(ctx context.Context, q ports.Querier, date time.Time) (int, error) {
	return 0, nil
}

func TestListAllLeaveRecordsUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	leaveType := &domain.LeaveType{
		ID:     1,
		UID:    "ltype_casual",
		Code:   "CASUAL",
		NameEN: "Casual Leave",
		NameAR: "العارضة",
	}

	mockDB := &mockDB{tx: &mockTx{}}
	leaveTypeRepo := &mockLeaveTypeRepoForAll{
		leaveTypes: []*domain.LeaveType{leaveType},
	}

	now := time.Now()
	leaveRecordRepo := &mockLeaveRecordRepoForAll{
		records: []*ports.LeaveRecordWithEmployee{
			{
				LeaveRecord: &domain.LeaveRecord{
					ID:          1,
					UID:         "lr_001",
					LeaveTypeID: 1,
					StartDate:   now,
					EndDate:     now,
					Days:        1,
					RecordedAt:  now,
				},
				EmployeeUID:  "emp_001",
				EmployeeName: "أحمد محمد",
			},
			{
				LeaveRecord: &domain.LeaveRecord{
					ID:          2,
					UID:         "lr_002",
					LeaveTypeID: 1,
					StartDate:   now.AddDate(0, 0, 1),
					EndDate:     now.AddDate(0, 0, 2),
					Days:        2,
					RecordedAt:  now,
				},
				EmployeeUID:  "emp_002",
				EmployeeName: "فاطمة حسن",
			},
		},
		total: 2,
	}

	uc := usecases.NewListAllLeaveRecordsUseCase(mockDB, leaveTypeRepo, leaveRecordRepo)

	t.Run("returns all leave records with employee info", func(t *testing.T) {
		input := usecases.ListAllLeaveRecordsInput{
			Page:     1,
			PageSize: 10,
		}

		output, err := uc.Execute(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if output.Total != 2 {
			t.Errorf("expected total 2, got %d", output.Total)
		}

		if len(output.Records) != 2 {
			t.Errorf("expected 2 records, got %d", len(output.Records))
		}

		if output.Records[0].EmployeeName != "أحمد محمد" {
			t.Errorf("expected employee name 'أحمد محمد', got '%s'", output.Records[0].EmployeeName)
		}

		if output.Records[0].LeaveTypeNameAR != "العارضة" {
			t.Errorf("expected leave type name 'العارضة', got '%s'", output.Records[0].LeaveTypeNameAR)
		}
	})

	t.Run("defaults page to 1 when less than 1", func(t *testing.T) {
		input := usecases.ListAllLeaveRecordsInput{
			Page:     0,
			PageSize: 10,
		}

		output, err := uc.Execute(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if output.Page != 1 {
			t.Errorf("expected page 1, got %d", output.Page)
		}
	})

	t.Run("defaults pageSize to 10 when less than 1", func(t *testing.T) {
		input := usecases.ListAllLeaveRecordsInput{
			Page:     1,
			PageSize: 0,
		}

		output, err := uc.Execute(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if output.PageSize != 10 {
			t.Errorf("expected pageSize 10, got %d", output.PageSize)
		}
	})

	t.Run("caps pageSize at 100", func(t *testing.T) {
		input := usecases.ListAllLeaveRecordsInput{
			Page:     1,
			PageSize: 200,
		}

		output, err := uc.Execute(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if output.PageSize != 100 {
			t.Errorf("expected pageSize 100, got %d", output.PageSize)
		}
	})

	t.Run("returns error for invalid leave type UID", func(t *testing.T) {
		invalidUID := "ltype_invalid"
		input := usecases.ListAllLeaveRecordsInput{
			LeaveTypeUID: &invalidUID,
		}

		_, err := uc.Execute(ctx, input)
		if err != usecases.ErrLeaveTypeNotFound {
			t.Errorf("expected ErrLeaveTypeNotFound, got %v", err)
		}
	})
}
