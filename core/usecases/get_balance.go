package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type GetBalanceInput struct {
	EmployeeUID string
	Year        int
}

type BalanceInfo struct {
	LeaveTypeUID     string
	LeaveTypeCode    string
	LeaveTypeNameEN  string
	LeaveTypeNameAR  string
	Year             int
	InitialBalance   int
	UsedBalance      int
	RemainingBalance int
}

type GetBalanceOutput struct {
	Balances []BalanceInfo
}

type GetBalanceUseCase struct {
	db               ports.DB
	employeeRepo     ports.EmployeeRepository
	leaveTypeRepo    ports.LeaveTypeRepository
	leaveBalanceRepo ports.LeaveBalanceRepository
}

func NewGetBalanceUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	leaveTypeRepo ports.LeaveTypeRepository,
	leaveBalanceRepo ports.LeaveBalanceRepository,
) *GetBalanceUseCase {
	return &GetBalanceUseCase{
		db:               db,
		employeeRepo:     employeeRepo,
		leaveTypeRepo:    leaveTypeRepo,
		leaveBalanceRepo: leaveBalanceRepo,
	}
}

func (uc *GetBalanceUseCase) Execute(ctx context.Context, input GetBalanceInput) (*GetBalanceOutput, error) {
	employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, input.EmployeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	leaveTypes, err := uc.leaveTypeRepo.List(ctx, uc.db, true) // activeOnly=true for employee-facing balance
	if err != nil {
		return nil, err
	}

	leaveTypeMap := make(map[int64]*domain.LeaveType)
	for _, lt := range leaveTypes {
		leaveTypeMap[lt.ID] = lt
	}

	balances, err := uc.leaveBalanceRepo.ListByEmployeeAndYear(ctx, uc.db, employee.ID, input.Year)
	if err != nil {
		return nil, err
	}

	balanceMap := make(map[int64]*domain.LeaveBalance)
	for _, b := range balances {
		balanceMap[b.LeaveTypeID] = b
	}

	var result []BalanceInfo
	for _, lt := range leaveTypes {
		totalDays := entitledBalanceTotalDays(employee, lt, balanceAsOfDate(input.Year))
		info := BalanceInfo{
			LeaveTypeUID:    lt.UID,
			LeaveTypeCode:   lt.Code,
			LeaveTypeNameEN: lt.NameEN,
			LeaveTypeNameAR: lt.NameAR,
			Year:            input.Year,
			InitialBalance:  totalDays,
		}

		if balance, ok := balanceMap[lt.ID]; ok {
			info.UsedBalance = balance.UsedDays
			info.RemainingBalance = totalDays - balance.UsedDays
		} else {
			info.UsedBalance = 0
			info.RemainingBalance = totalDays
		}

		result = append(result, info)
	}

	return &GetBalanceOutput{Balances: result}, nil
}

func balanceAsOfDate(year int) time.Time {
	now := time.Now()
	if year == now.Year() {
		return now
	}
	return time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)
}
