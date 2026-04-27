package usecases

import (
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
)

func TestEntitledBalanceTotalDays(t *testing.T) {
	asOf := time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		employee  *domain.Employee
		leaveType *domain.LeaveType
		want      int
	}{
		{
			name: "casual is always seven days",
			employee: &domain.Employee{
				HireDate: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			leaveType: &domain.LeaveType{Code: leaveTypeCodeCasual, DefaultBalance: 3},
			want:      7,
		},
		{
			name: "regular leave under ten years of service",
			employee: &domain.Employee{
				HireDate: time.Date(2018, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			leaveType: &domain.LeaveType{Code: leaveTypeCodeRegular, DefaultBalance: 21},
			want:      21,
		},
		{
			name: "regular leave more than ten years of service and under fifty",
			employee: &domain.Employee{
				HireDate: time.Date(2010, 4, 1, 0, 0, 0, 0, time.UTC),
			},
			leaveType: &domain.LeaveType{Code: leaveTypeCodeRegular, DefaultBalance: 21},
			want:      30,
		},
		{
			name: "regular leave more than fifty years since hire",
			employee: &domain.Employee{
				HireDate: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			leaveType: &domain.LeaveType{Code: leaveTypeCodeRegular, DefaultBalance: 21},
			want:      45,
		},
		{
			name: "other leave types keep default balance",
			employee: &domain.Employee{
				HireDate: time.Date(2010, 4, 1, 0, 0, 0, 0, time.UTC),
			},
			leaveType: &domain.LeaveType{Code: "SICK", DefaultBalance: 15},
			want:      15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := entitledBalanceTotalDays(tt.employee, tt.leaveType, asOf)
			if got != tt.want {
				t.Fatalf("entitledBalanceTotalDays() = %d, want %d", got, tt.want)
			}
		})
	}
}
