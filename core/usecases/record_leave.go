package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/banumusa/backend/core/audit"
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
	EmployeeUID     string
	LeaveTypeUID    string
	StartDate       time.Time
	EndDate         time.Time
	RecordedBy      *int64
	RecordedByUID   *string // Employee UID of the actor recording the leave (for audit)
	Notes           *string
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
	auditor          audit.Auditor
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
	auditor audit.Auditor,
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
		auditor:          auditor,
	}
}

func (uc *RecordLeaveUseCase) Execute(ctx context.Context, input RecordLeaveInput) (*RecordLeaveOutput, error) {
	if input.EndDate.Before(input.StartDate) {
		return nil, ErrInvalidDateRange
	}

	// Determine actor UID for audit logging
	// Use RecordedByUID if provided, otherwise try to extract from context
	actorUID := ""
	if input.RecordedByUID != nil {
		actorUID = *input.RecordedByUID
	} else {
		actorUID = audit.ActorFromContext(ctx)
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

	totalWorkingDays, err := calculateLeaveDays(ctx, tx, uc.workingDaysCalc, leaveType, input.StartDate, input.EndDate)
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
		workingDays, err := calculateLeaveDays(ctx, tx, uc.workingDaysCalc, leaveType, period.Start, period.End)
		if err != nil {
			return nil, err
		}

		if workingDays == 0 {
			continue
		}

		balance, created, err := uc.getOrCreateBalance(ctx, tx, employee, leaveType, period.Year, period.Start)
		if err != nil {
			return nil, err
		}
		if created {
			initialTx := domain.NewLeaveBalanceTransaction(
				balance.ID,
				domain.TransactionTypeInitial,
				balance.TotalDays,
				nil,
				nil,
				nil,
			)
			if err := uc.balanceTxRepo.Create(ctx, tx, initialTx); err != nil {
				return nil, err
			}
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

		// Capture old balance before deduction
		oldUsedDays := balance.UsedDays
		oldRemainingDays := remaining

		balance.UsedDays += workingDays
		if err := uc.leaveBalanceRepo.Update(ctx, tx, balance); err != nil {
			return nil, err
		}

		// Audit log for balance deduction (only if we have an actor)
		if actorUID != "" {
			// Resolve actor name for human-readable sentence
			actorName := actorUID
			if actor, err := uc.employeeRepo.GetByUID(ctx, tx, actorUID); err == nil && actor != nil {
				actorName = actor.Name
			}
			
			actionSentence := fmt.Sprintf(
				"%s recorded %s leave for %s — deducted %d days from balance (was %d used, now %d used)",
				actorName, leaveType.NameEN, employee.Name,
				workingDays, oldUsedDays, balance.UsedDays,
			)
			
			defer uc.auditor.Actor(actorUID).
				Did(audit.ActionDeductBalance).
				On(audit.EntityLeaveBalance, balance.UID).
				WithMeta("action", actionSentence).
				WithMeta("employee_uid", input.EmployeeUID).
				WithMeta("employee_name", employee.Name).
				WithMeta("leave_type_uid", input.LeaveTypeUID).
				WithMeta("leave_type_name", leaveType.NameEN).
				WithMeta("deduction_amount", workingDays).
				WithMeta("old_used_days", oldUsedDays).
				WithMeta("new_used_days", balance.UsedDays).
				WithMeta("old_remaining_days", oldRemainingDays).
				WithMeta("new_remaining_days", balance.TotalDays-balance.UsedDays).
				WithMeta("start_date", period.Start.Format("2006-01-02")).
				WithMeta("end_date", period.End.Format("2006-01-02")).
				WithMeta("year", period.Year).
				Save(ctx)
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

func (uc *RecordLeaveUseCase) getOrCreateBalance(ctx context.Context, q ports.Querier, employee *domain.Employee, leaveType *domain.LeaveType, year int, asOf time.Time) (*domain.LeaveBalance, bool, error) {
	return ensureLeaveBalance(ctx, q, uc.leaveBalanceRepo, employee, leaveType, year, asOf)
}
