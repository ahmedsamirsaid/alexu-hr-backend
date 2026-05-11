package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/usecases"
)

func TestGetEmployeeUseCase_Execute(t *testing.T) {
	email := "ahmed@test.com"
	employee := &domain.Employee{
		ID:           1,
		UID:          "emp_123",
		Name:         "أحمد محمد",
		Mobile:       "01012345678",
		GovernmentID: "28501011234567",
		UniversityID: "AU-12345-001",
		Email:        &email,
		HireDate:     time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC),
		Status:       domain.EmployeeStatusActive,
		Type:         domain.EmployeeTypePermanent,
		SubType:      domain.EmployeeSubTypeNormal,
	}

	tests := []struct {
		name        string
		input       usecases.GetEmployeeInput
		employee    *domain.Employee
		expectedErr error
	}{
		{
			name: "successfully retrieve employee",
			input: usecases.GetEmployeeInput{
				UID: "emp_123",
			},
			employee:    employee,
			expectedErr: nil,
		},
		{
			name: "employee not found",
			input: usecases.GetEmployeeInput{
				UID: "emp_notfound",
			},
			employee:    nil,
			expectedErr: usecases.ErrEmployeeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTx := &mockTx{}
			db := &mockDB{tx: mockTx}
			employeeRepo := &mockEmployeeRepo{employee: tt.employee}

			uc := usecases.NewGetEmployeeUseCase(db, employeeRepo, nil)

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

			if output.UID != tt.employee.UID {
				t.Errorf("expected UID %s, got %s", tt.employee.UID, output.UID)
			}
			if output.Name != tt.employee.Name {
				t.Errorf("expected Name %s, got %s", tt.employee.Name, output.Name)
			}
			if output.Mobile != tt.employee.Mobile {
				t.Errorf("expected Mobile %s, got %s", tt.employee.Mobile, output.Mobile)
			}
			if output.GovernmentID != tt.employee.GovernmentID {
				t.Errorf("expected GovernmentID %s, got %s", tt.employee.GovernmentID, output.GovernmentID)
			}
			if output.UniversityID != tt.employee.UniversityID {
				t.Errorf("expected UniversityID %s, got %s", tt.employee.UniversityID, output.UniversityID)
			}
			if *output.Email != *tt.employee.Email {
				t.Errorf("expected Email %s, got %s", *tt.employee.Email, *output.Email)
			}
			if !output.HireDate.Equal(tt.employee.HireDate) {
				t.Errorf("expected HireDate %v, got %v", tt.employee.HireDate, output.HireDate)
			}
			if output.Status != tt.employee.Status {
				t.Errorf("expected Status %s, got %s", tt.employee.Status, output.Status)
			}
			if output.Type != tt.employee.Type {
				t.Errorf("expected Type %s, got %s", tt.employee.Type, output.Type)
			}
			if output.SubType != tt.employee.SubType {
				t.Errorf("expected SubType %s, got %s", tt.employee.SubType, output.SubType)
			}
		})
	}
}
