package domain

import "time"

type EmployeeStatus string

const (
	EmployeeStatusActive     EmployeeStatus = "active"
	EmployeeStatusInactive   EmployeeStatus = "inactive"
	EmployeeStatusTerminated EmployeeStatus = "terminated"
)

type Employee struct {
	ID            int64
	UID           string
	Name          string
	Mobile        string
	GovernmentID  string
	UniversityID  string
	Email         *string
	HireDate      time.Time
	Status        EmployeeStatus
	DepartmentUID *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewEmployee(name, mobile, governmentID, universityID string, hireDate time.Time) *Employee {
	return &Employee{
		UID:          GenerateUID("emp"),
		Name:         name,
		Mobile:       mobile,
		GovernmentID: governmentID,
		UniversityID: universityID,
		HireDate:     hireDate,
		Status:       EmployeeStatusActive,
	}
}
