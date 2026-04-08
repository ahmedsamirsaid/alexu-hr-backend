package domain

import "time"

type ApprovalFlow struct {
	ID          int64
	UID         string
	Code        string
	NameEN      string
	NameAR      *string
	Description *string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewApprovalFlow(code, nameEN string, nameAR, description *string) *ApprovalFlow {
	return &ApprovalFlow{
		UID:         GenerateUID("apf"),
		Code:        code,
		NameEN:      nameEN,
		NameAR:      nameAR,
		Description: description,
		IsActive:    true,
	}
}
