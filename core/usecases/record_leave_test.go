package usecases_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
)

type mockDB struct {
	tx *mockTx
}

func (m *mockDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (ports.Tx, error) {
	return m.tx, nil
}

func (m *mockDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}

func (m *mockDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return nil
}

func (m *mockDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}

type mockTx struct {
	committed  bool
	rolledBack bool
}

func (m *mockTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}

func (m *mockTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return nil
}

func (m *mockTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}

func (m *mockTx) Commit() error {
	m.committed = true
	return nil
}

func (m *mockTx) Rollback() error {
	m.rolledBack = true
	return nil
}

type mockEmployeeRepo struct {
	employee *domain.Employee
}

func (m *mockEmployeeRepo) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Employee, error) {
	return m.employee, nil
}

func (m *mockEmployeeRepo) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.Employee, error) {
	if m.employee != nil && m.employee.UID == uid {
		return m.employee, nil
	}
	return nil, nil
}

func (m *mockEmployeeRepo) Create(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
	return nil
}

func (m *mockEmployeeRepo) Update(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
	return nil
}

func (m *mockEmployeeRepo) List(ctx context.Context, q ports.Querier, filter *ports.EmployeeListFilter) ([]*domain.Employee, error) {
	if m.employee != nil {
		return []*domain.Employee{m.employee}, nil
	}
	return nil, nil
}

func (m *mockEmployeeRepo) ExistingGovernmentIDs(ctx context.Context, q ports.Querier, governmentIDs []string) ([]string, error) {
	return nil, nil
}

func (m *mockEmployeeRepo) ExistingMobiles(ctx context.Context, q ports.Querier, mobiles []string) ([]string, error) {
	return nil, nil
}

func (m *mockEmployeeRepo) ExistingUniversityIDs(ctx context.Context, q ports.Querier, universityIDs []string) ([]string, error) {
	return nil, nil
}

func (m *mockEmployeeRepo) Count(ctx context.Context, q ports.Querier) (int, error) {
	if m.employee != nil {
		return 1, nil
	}
	return 0, nil
}

type mockLeaveTypeRepo struct {
	leaveType *domain.LeaveType
}

func (m *mockLeaveTypeRepo) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.LeaveType, error) {
	return m.leaveType, nil
}

func (m *mockLeaveTypeRepo) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.LeaveType, error) {
	if m.leaveType != nil && m.leaveType.UID == uid {
		return m.leaveType, nil
	}
	return nil, nil
}

func (m *mockLeaveTypeRepo) GetByCode(ctx context.Context, q ports.Querier, code string) (*domain.LeaveType, error) {
	return m.leaveType, nil
}

func (m *mockLeaveTypeRepo) GetSubLeaveTypeByUID(ctx context.Context, q ports.Querier, uid string) (*domain.SubLeaveType, error) {
	return nil, nil
}

func (m *mockLeaveTypeRepo) List(ctx context.Context, q ports.Querier, activeOnly bool) ([]*domain.LeaveType, error) {
	if m.leaveType != nil {
		return []*domain.LeaveType{m.leaveType}, nil
	}
	return nil, nil
}

func (m *mockLeaveTypeRepo) ListSubLeaveTypesByLeaveTypeUID(ctx context.Context, q ports.Querier, leaveTypeUID string) ([]*domain.SubLeaveType, error) {
	return nil, nil
}

func (m *mockLeaveTypeRepo) SetActive(ctx context.Context, q ports.Querier, uid string, isActive bool) error {
	return nil
}

type mockLeaveBalanceRepo struct {
	balance        *domain.LeaveBalance
	createdBalance *domain.LeaveBalance
}

func (m *mockLeaveBalanceRepo) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.LeaveBalance, error) {
	return m.balance, nil
}

func (m *mockLeaveBalanceRepo) GetByEmployeeAndTypeAndYear(ctx context.Context, q ports.Querier, employeeID, leaveTypeID int64, year int) (*domain.LeaveBalance, error) {
	if m.balance != nil && m.balance.EmployeeID == employeeID && m.balance.LeaveTypeID == leaveTypeID && m.balance.Year == year {
		return m.balance, nil
	}
	return nil, nil
}

func (m *mockLeaveBalanceRepo) Create(ctx context.Context, q ports.Querier, balance *domain.LeaveBalance) error {
	balance.ID = 1
	m.createdBalance = balance
	m.balance = balance
	return nil
}

func (m *mockLeaveBalanceRepo) Update(ctx context.Context, q ports.Querier, balance *domain.LeaveBalance) error {
	m.balance = balance
	return nil
}

func (m *mockLeaveBalanceRepo) ListByEmployee(ctx context.Context, q ports.Querier, employeeID int64) ([]*domain.LeaveBalance, error) {
	if m.balance != nil {
		return []*domain.LeaveBalance{m.balance}, nil
	}
	return nil, nil
}

func (m *mockLeaveBalanceRepo) ListByEmployeeAndYear(ctx context.Context, q ports.Querier, employeeID int64, year int) ([]*domain.LeaveBalance, error) {
	if m.balance != nil && m.balance.Year == year {
		return []*domain.LeaveBalance{m.balance}, nil
	}
	return nil, nil
}

type mockLeaveRecordRepo struct {
	records []*domain.LeaveRecord
}

func (m *mockLeaveRecordRepo) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.LeaveRecord, error) {
	return nil, nil
}

func (m *mockLeaveRecordRepo) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.LeaveRecord, error) {
	return nil, nil
}

func (m *mockLeaveRecordRepo) Create(ctx context.Context, q ports.Querier, record *domain.LeaveRecord) error {
	record.ID = int64(len(m.records) + 1)
	m.records = append(m.records, record)
	return nil
}

func (m *mockLeaveRecordRepo) ListByEmployee(ctx context.Context, q ports.Querier, employeeID int64) ([]*domain.LeaveRecord, error) {
	return m.records, nil
}

func (m *mockLeaveRecordRepo) ListByEmployeeAndDateRange(ctx context.Context, q ports.Querier, employeeID int64, start, end time.Time) ([]*domain.LeaveRecord, error) {
	return m.records, nil
}

func (m *mockLeaveRecordRepo) ListByEmployeeAndType(ctx context.Context, q ports.Querier, employeeID, leaveTypeID int64) ([]*domain.LeaveRecord, error) {
	return m.records, nil
}

func (m *mockLeaveRecordRepo) ListByEmployeePaginated(ctx context.Context, q ports.Querier, employeeID int64, limit, offset int) ([]*domain.LeaveRecord, error) {
	if offset >= len(m.records) {
		return nil, nil
	}
	end := offset + limit
	if end > len(m.records) {
		end = len(m.records)
	}
	return m.records[offset:end], nil
}

func (m *mockLeaveRecordRepo) CountByEmployee(ctx context.Context, q ports.Querier, employeeID int64) (int, error) {
	return len(m.records), nil
}

func (m *mockLeaveRecordRepo) ListAllPaginated(ctx context.Context, q ports.Querier, filter ports.ListAllLeaveRecordsFilter, limit, offset int) ([]*ports.LeaveRecordWithEmployee, error) {
	return nil, nil
}

func (m *mockLeaveRecordRepo) CountAll(ctx context.Context, q ports.Querier, filter ports.ListAllLeaveRecordsFilter) (int, error) {
	return 0, nil
}

func (m *mockLeaveRecordRepo) CountOnLeaveToday(ctx context.Context, q ports.Querier, date time.Time) (int, error) {
	return 0, nil
}

func (m *mockLeaveRecordRepo) HasLeaveOnDate(ctx context.Context, q ports.Querier, employeeID int64, date time.Time) (bool, error) {
	return false, nil
}

type mockBalanceTxRepo struct {
	transactions []*domain.LeaveBalanceTransaction
}

func (m *mockBalanceTxRepo) Create(ctx context.Context, q ports.Querier, tx *domain.LeaveBalanceTransaction) error {
	tx.ID = int64(len(m.transactions) + 1)
	m.transactions = append(m.transactions, tx)
	return nil
}

func (m *mockBalanceTxRepo) ListByBalance(ctx context.Context, q ports.Querier, balanceID int64) ([]*domain.LeaveBalanceTransaction, error) {
	return m.transactions, nil
}

type mockLeaveSync struct {
	syncedRecords []*domain.LeaveRecord
}

func (m *mockLeaveSync) SyncLeaveRecord(ctx context.Context, record *domain.LeaveRecord, employee *domain.Employee, leaveType *domain.LeaveType) error {
	m.syncedRecords = append(m.syncedRecords, record)
	return nil
}

func intPtr(i int) *int {
	return &i
}

func TestRecordLeaveUseCase_Execute(t *testing.T) {
	employee := &domain.Employee{
		ID:  1,
		UID: "emp_123",
	}

	leaveType := &domain.LeaveType{
		ID:             1,
		UID:            "lt_casual",
		Code:           "CASUAL",
		DefaultBalance: 7,
		MaxConsecutive: intPtr(2),
	}

	tests := []struct {
		name          string
		input         usecases.RecordLeaveInput
		employee      *domain.Employee
		leaveType     *domain.LeaveType
		balance       *domain.LeaveBalance
		expectedErr   error
		expectedDays  int
		expectedCount int
	}{
		{
			name: "successful single day leave",
			input: usecases.RecordLeaveInput{
				EmployeeUID:  "emp_123",
				LeaveTypeUID: "lt_casual",
				StartDate:    time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC), // Monday
				EndDate:      time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC),
			},
			employee:  employee,
			leaveType: leaveType,
			balance: &domain.LeaveBalance{
				ID:          1,
				EmployeeID:  1,
				LeaveTypeID: 1,
				Year:        2025,
				TotalDays:   7,
				UsedDays:    0,
			},
			expectedErr:   nil,
			expectedDays:  1,
			expectedCount: 1,
		},
		{
			name: "successful two day leave",
			input: usecases.RecordLeaveInput{
				EmployeeUID:  "emp_123",
				LeaveTypeUID: "lt_casual",
				StartDate:    time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC), // Monday
				EndDate:      time.Date(2025, 1, 7, 0, 0, 0, 0, time.UTC), // Tuesday
			},
			employee:  employee,
			leaveType: leaveType,
			balance: &domain.LeaveBalance{
				ID:          1,
				EmployeeID:  1,
				LeaveTypeID: 1,
				Year:        2025,
				TotalDays:   7,
				UsedDays:    0,
			},
			expectedErr:   nil,
			expectedDays:  2,
			expectedCount: 1,
		},
		{
			name: "exceeds max consecutive days",
			input: usecases.RecordLeaveInput{
				EmployeeUID:  "emp_123",
				LeaveTypeUID: "lt_casual",
				StartDate:    time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC), // Monday
				EndDate:      time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC), // Wednesday (3 days)
			},
			employee:  employee,
			leaveType: leaveType,
			balance: &domain.LeaveBalance{
				ID:          1,
				EmployeeID:  1,
				LeaveTypeID: 1,
				Year:        2025,
				TotalDays:   7,
				UsedDays:    0,
			},
			expectedErr: usecases.ErrExceedsConsecutiveDays,
		},
		{
			name: "insufficient balance",
			input: usecases.RecordLeaveInput{
				EmployeeUID:  "emp_123",
				LeaveTypeUID: "lt_casual",
				StartDate:    time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC),
				EndDate:      time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC),
			},
			employee:  employee,
			leaveType: leaveType,
			balance: &domain.LeaveBalance{
				ID:          1,
				EmployeeID:  1,
				LeaveTypeID: 1,
				Year:        2025,
				TotalDays:   7,
				UsedDays:    7, // All used up
			},
			expectedErr: usecases.ErrInsufficientBalance,
		},
		{
			name: "employee not found",
			input: usecases.RecordLeaveInput{
				EmployeeUID:  "emp_unknown",
				LeaveTypeUID: "lt_casual",
				StartDate:    time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC),
				EndDate:      time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC),
			},
			employee:    nil,
			leaveType:   leaveType,
			expectedErr: usecases.ErrEmployeeNotFound,
		},
		{
			name: "leave type not found",
			input: usecases.RecordLeaveInput{
				EmployeeUID:  "emp_123",
				LeaveTypeUID: "lt_unknown",
				StartDate:    time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC),
				EndDate:      time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC),
			},
			employee:    employee,
			leaveType:   nil,
			expectedErr: usecases.ErrLeaveTypeNotFound,
		},
		{
			name: "invalid date range",
			input: usecases.RecordLeaveInput{
				EmployeeUID:  "emp_123",
				LeaveTypeUID: "lt_casual",
				StartDate:    time.Date(2025, 1, 7, 0, 0, 0, 0, time.UTC),
				EndDate:      time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC), // End before start
			},
			employee:    employee,
			leaveType:   leaveType,
			expectedErr: usecases.ErrInvalidDateRange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTx := &mockTx{}
			db := &mockDB{tx: mockTx}
			employeeRepo := &mockEmployeeRepo{employee: tt.employee}
			leaveTypeRepo := &mockLeaveTypeRepo{leaveType: tt.leaveType}
			leaveBalanceRepo := &mockLeaveBalanceRepo{balance: tt.balance}
			leaveRecordRepo := &mockLeaveRecordRepo{}
			balanceTxRepo := &mockBalanceTxRepo{}
			weekendRepo := &mockWeekendConfigRepo{weekendDays: []int{5, 6}}
			holidayRepo := &mockHolidayDefinitionRepo{}
			leaveSync := &mockLeaveSync{}

			workingDaysCalc := usecases.NewWorkingDaysCalculator(weekendRepo, holidayRepo)

			uc := usecases.NewRecordLeaveUseCase(
				db,
				employeeRepo,
				leaveTypeRepo,
				leaveBalanceRepo,
				leaveRecordRepo,
				balanceTxRepo,
				workingDaysCalc,
				leaveSync,
			)

			output, err := uc.Execute(context.Background(), tt.input)

			if tt.expectedErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedErr)
				}
				if !errors.Is(err, tt.expectedErr) {
					t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(output.Records) != tt.expectedCount {
				t.Errorf("expected %d records, got %d", tt.expectedCount, len(output.Records))
			}

			if tt.expectedCount > 0 && output.Records[0].Days != tt.expectedDays {
				t.Errorf("expected %d days, got %d", tt.expectedDays, output.Records[0].Days)
			}

			if !mockTx.committed {
				t.Error("expected transaction to be committed")
			}
		})
	}
}

func TestRecordLeaveUseCase_YearBoundarySplit(t *testing.T) {
	employee := &domain.Employee{
		ID:  1,
		UID: "emp_123",
	}

	leaveType := &domain.LeaveType{
		ID:             1,
		UID:            "lt_annual",
		Code:           "ANNUAL",
		DefaultBalance: 21,
		MaxConsecutive: nil, // No limit
	}

	mockTx := &mockTx{}
	db := &mockDB{tx: mockTx}
	employeeRepo := &mockEmployeeRepo{employee: employee}
	leaveTypeRepo := &mockLeaveTypeRepo{leaveType: leaveType}
	leaveBalanceRepo := &mockLeaveBalanceRepo{} // Will create balances on demand
	leaveRecordRepo := &mockLeaveRecordRepo{}
	balanceTxRepo := &mockBalanceTxRepo{}
	weekendRepo := &mockWeekendConfigRepo{weekendDays: []int{5, 6}}
	holidayRepo := &mockHolidayDefinitionRepo{}
	leaveSync := &mockLeaveSync{}

	workingDaysCalc := usecases.NewWorkingDaysCalculator(weekendRepo, holidayRepo)

	uc := usecases.NewRecordLeaveUseCase(
		db,
		employeeRepo,
		leaveTypeRepo,
		leaveBalanceRepo,
		leaveRecordRepo,
		balanceTxRepo,
		workingDaysCalc,
		leaveSync,
	)

	input := usecases.RecordLeaveInput{
		EmployeeUID:  "emp_123",
		LeaveTypeUID: "lt_annual",
		StartDate:    time.Date(2025, 12, 29, 0, 0, 0, 0, time.UTC), // Monday
		EndDate:      time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),   // Friday next year
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should create two records - one for 2025 and one for 2026
	if len(output.Records) != 2 {
		t.Errorf("expected 2 records for year boundary split, got %d", len(output.Records))
	}

	// Verify the records are in different years
	if len(output.Records) >= 2 {
		if output.Records[0].StartDate.Year() != 2025 {
			t.Errorf("first record should be in 2025, got %d", output.Records[0].StartDate.Year())
		}
		if output.Records[1].StartDate.Year() != 2026 {
			t.Errorf("second record should be in 2026, got %d", output.Records[1].StartDate.Year())
		}
	}
}
