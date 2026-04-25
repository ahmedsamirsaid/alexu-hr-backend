package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

var (
	ErrEmployeeNotFound                     = errors.New("employee not found")
	ErrLeaveTypeNotFound                    = errors.New("leave type not found")
	ErrSubLeaveTypeNotFound                 = errors.New("sub leave type not found")
	ErrSubLeaveTypeDoesNotBelongToLeaveType = errors.New("sub leave type does not belong to leave type")
	ErrInsufficientBalance                  = errors.New("insufficient leave balance")
	ErrExceedsConsecutiveDays               = errors.New("exceeds maximum consecutive days")
	ErrRecordingDeadlinePassed              = errors.New("recording deadline has passed")
	ErrInvalidDateRange                     = errors.New("invalid date range")
	ErrNoWorkingDays                        = errors.New("no working days")
)

type RecordLeaveInput struct {
	EmployeeUID  string
	LeaveTypeUID string
	StartDate    time.Time
	EndDate      time.Time
	RecordedBy   *int64
	Notes        *string
}

type RecordLeaveOutput struct {
	Records []*domain.LeaveRecord
}

type RecordLeaveUseCase struct {
	db               ports.DB
	employeeRepo     ports.EmployeeRepository
	leaveTypeRepo    ports.LeaveTypeRepository
	leaveBalanceRepo ports.LeaveBalanceRepository
	leaveRecordRepo  ports.LeaveRecordRepository
	balanceTxRepo    ports.LeaveBalanceTransactionRepository
	workingDaysCalc  *WorkingDaysCalculator
	leaveSync        ports.LeaveSyncPort
}

func NewRecordLeaveUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	leaveTypeRepo ports.LeaveTypeRepository,
	leaveBalanceRepo ports.LeaveBalanceRepository,
	leaveRecordRepo ports.LeaveRecordRepository,
	balanceTxRepo ports.LeaveBalanceTransactionRepository,
	workingDaysCalc *WorkingDaysCalculator,
	leaveSync ports.LeaveSyncPort,
) *RecordLeaveUseCase {
	return &RecordLeaveUseCase{
		db:               db,
		employeeRepo:     employeeRepo,
		leaveTypeRepo:    leaveTypeRepo,
		leaveBalanceRepo: leaveBalanceRepo,
		leaveRecordRepo:  leaveRecordRepo,
		balanceTxRepo:    balanceTxRepo,
		workingDaysCalc:  workingDaysCalc,
		leaveSync:        leaveSync,
	}
}

func (uc *RecordLeaveUseCase) Execute(ctx context.Context, input RecordLeaveInput) (*RecordLeaveOutput, error) {
	if input.EndDate.Before(input.StartDate) {
		return nil, ErrInvalidDateRange
	}

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	employee, err := uc.employeeRepo.GetByUID(ctx, tx, input.EmployeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	leaveType, err := uc.leaveTypeRepo.GetByUID(ctx, tx, input.LeaveTypeUID)
	if err != nil {
		return nil, err
	}
	if leaveType == nil {
		return nil, ErrLeaveTypeNotFound
	}

	totalWorkingDays, err := uc.workingDaysCalc.CalculateWorkingDays(ctx, tx, input.StartDate, input.EndDate)
	if err != nil {
		return nil, err
	}

	if leaveType.MaxConsecutive != nil && totalWorkingDays > *leaveType.MaxConsecutive {
		return nil, ErrExceedsConsecutiveDays
	}

	if leaveType.RecordingDeadlineDays != nil {
		deadline := input.EndDate.AddDate(0, 0, *leaveType.RecordingDeadlineDays)
		if time.Now().After(deadline) {
			return nil, ErrRecordingDeadlinePassed
		}
	}

	periods := uc.splitByYear(input.StartDate, input.EndDate)

	var records []*domain.LeaveRecord
	for _, period := range periods {
		workingDays, err := uc.workingDaysCalc.CalculateWorkingDays(ctx, tx, period.Start, period.End)
		if err != nil {
			return nil, err
		}

		if workingDays == 0 {
			continue
		}

		balance, err := uc.getOrCreateBalance(ctx, tx, employee.ID, leaveType, period.Year)
		if err != nil {
			return nil, err
		}

		remaining := balance.TotalDays - balance.UsedDays
		if workingDays > remaining {
			return nil, ErrInsufficientBalance
		}

		record := domain.NewLeaveRecord(
			employee.ID,
			leaveType.ID,
			period.Start,
			period.End,
			workingDays,
			input.RecordedBy,
			input.Notes,
			nil, // leaveRequestUID - direct recordings don't have a leave request
		)

		if err := uc.leaveRecordRepo.Create(ctx, tx, record); err != nil {
			return nil, err
		}

		balance.UsedDays += workingDays
		if err := uc.leaveBalanceRepo.Update(ctx, tx, balance); err != nil {
			return nil, err
		}

		balanceTx := domain.NewLeaveBalanceTransaction(
			balance.ID,
			domain.TransactionTypeDeduct,
			workingDays,
			&record.ID,
			input.RecordedBy,
			input.Notes,
		)
		if err := uc.balanceTxRepo.Create(ctx, tx, balanceTx); err != nil {
			return nil, err
		}

		if err := uc.leaveSync.SyncLeaveRecord(ctx, record, employee, leaveType); err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &RecordLeaveOutput{Records: records}, nil
}

type datePeriod struct {
	Start time.Time
	End   time.Time
	Year  int
}

func (uc *RecordLeaveUseCase) splitByYear(start, end time.Time) []datePeriod {
	var periods []datePeriod

	current := start
	for current.Year() < end.Year() {
		yearEnd := time.Date(current.Year(), 12, 31, 0, 0, 0, 0, current.Location())
		periods = append(periods, datePeriod{
			Start: current,
			End:   yearEnd,
			Year:  current.Year(),
		})
		current = time.Date(current.Year()+1, 1, 1, 0, 0, 0, 0, current.Location())
	}

	periods = append(periods, datePeriod{
		Start: current,
		End:   end,
		Year:  current.Year(),
	})

	return periods
}

func (uc *RecordLeaveUseCase) getOrCreateBalance(ctx context.Context, q ports.Querier, employeeID int64, leaveType *domain.LeaveType, year int) (*domain.LeaveBalance, error) {
	balance, err := uc.leaveBalanceRepo.GetByEmployeeAndTypeAndYear(ctx, q, employeeID, leaveType.ID, year)
	if err != nil {
		return nil, err
	}

	if balance != nil {
		return balance, nil
	}

	balance = domain.NewLeaveBalance(employeeID, leaveType.ID, year, leaveType.DefaultBalance)
	if err := uc.leaveBalanceRepo.Create(ctx, q, balance); err != nil {
		return nil, err
	}

	initialTx := domain.NewLeaveBalanceTransaction(
		balance.ID,
		domain.TransactionTypeInitial,
		leaveType.DefaultBalance,
		nil,
		nil,
		nil,
	)
	if err := uc.balanceTxRepo.Create(ctx, q, initialTx); err != nil {
		return nil, err
	}

	return balance, nil
}
