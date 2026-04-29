package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/usecases"
)

func TestGetBalanceUseCase_Execute(t *testing.T) {
	employee := &domain.Employee{
		ID:       1,
		UID:      "emp_123",
		HireDate: time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	casualLeaveType := &domain.LeaveType{
		ID:             1,
		UID:            "lt_casual",
		Code:           "CASUAL",
		NameEN:         "Casual Leave",
		NameAR:         "إجازة عارضة",
		DefaultBalance: 7,
	}

	regularLeaveType := &domain.LeaveType{
		ID:             2,
		UID:            "lt_regular",
		Code:           "REGULAR",
		NameEN:         "Regular Leave",
		NameAR:         "إجازة اعتيادى",
		DefaultBalance: 21,
	}

	tests := []struct {
		name              string
		input             usecases.GetBalanceInput
		employee          *domain.Employee
		leaveType         *domain.LeaveType
		balance           *domain.LeaveBalance
		expectedErr       error
		expectedInitial   int
		expectedRemaining int
	}{
		{
			name: "existing balance",
			input: usecases.GetBalanceInput{
				EmployeeUID: "emp_123",
				Year:        2025,
			},
			employee:  employee,
			leaveType: casualLeaveType,
			balance: &domain.LeaveBalance{
				ID:          1,
				EmployeeID:  1,
				LeaveTypeID: 1,
				Year:        2025,
				TotalDays:   7,
				UsedDays:    3,
			},
			expectedErr:       nil,
			expectedInitial:   7,
			expectedRemaining: 4,
		},
		{
			name: "no balance yet - uses default",
			input: usecases.GetBalanceInput{
				EmployeeUID: "emp_123",
				Year:        2025,
			},
			employee:          employee,
			leaveType:         casualLeaveType,
			balance:           nil,
			expectedErr:       nil,
			expectedInitial:   7,
			expectedRemaining: 7, // Default balance
		},
		{
			name: "regular leave uses service based entitlement",
			input: usecases.GetBalanceInput{
				EmployeeUID: "emp_123",
				Year:        2025,
			},
			employee:          employee,
			leaveType:         regularLeaveType,
			balance:           nil,
			expectedErr:       nil,
			expectedInitial:   30,
			expectedRemaining: 30,
		},
		{
			name: "employee not found",
			input: usecases.GetBalanceInput{
				EmployeeUID: "emp_unknown",
				Year:        2025,
			},
			employee:    nil,
			leaveType:   casualLeaveType,
			expectedErr: usecases.ErrEmployeeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTx := &mockTx{}
			db := &mockDB{tx: mockTx}
			employeeRepo := &mockEmployeeRepo{employee: tt.employee}
			leaveTypeRepo := &mockLeaveTypeRepo{leaveType: tt.leaveType}
			leaveBalanceRepo := &mockLeaveBalanceRepo{balance: tt.balance}

			uc := usecases.NewGetBalanceUseCase(
				db,
				employeeRepo,
				leaveTypeRepo,
				leaveBalanceRepo,
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

			if len(output.Balances) == 0 {
				t.Fatal("expected at least one balance")
			}

			if output.Balances[0].InitialBalance != tt.expectedInitial {
				t.Errorf("expected initial balance %d, got %d", tt.expectedInitial, output.Balances[0].InitialBalance)
			}

			if output.Balances[0].RemainingBalance != tt.expectedRemaining {
				t.Errorf("expected remaining balance %d, got %d", tt.expectedRemaining, output.Balances[0].RemainingBalance)
			}
		})
	}
}
