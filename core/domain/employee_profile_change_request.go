package domain

import "time"

type EmployeeProfileChangeRequest struct {
	ID                        int64
	UID                       string
	EmployeeUID               string
	ApprovalRequestUID        string
	SubmittedByEmployeeUID    string
	CurrentFinancialGrade     *string
	CurrentIDCardValidUntil   *time.Time
	CurrentMaritalStatus      *string
	RequestedFinancialGrade   *string
	RequestedIDCardValidUntil *time.Time
	RequestedMaritalStatus    *string
	Comments                  *string
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

func NewEmployeeProfileChangeRequest(employeeUID, approvalRequestUID, submittedByEmployeeUID string) *EmployeeProfileChangeRequest {
	return &EmployeeProfileChangeRequest{
		UID:                    GenerateUID("epc"),
		EmployeeUID:            employeeUID,
		ApprovalRequestUID:     approvalRequestUID,
		SubmittedByEmployeeUID: submittedByEmployeeUID,
	}
}
