package domain

import "time"

type Department struct {
	ID              int64
	UID             string
	Code            string
	NameEN          string
	NameAR          *string
	IsActive        bool
	DefaultShiftUID *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewDepartment(code, nameEN string, nameAR *string) *Department {
	return &Department{
		UID:      GenerateUID("dep"),
		Code:     code,
		NameEN:   nameEN,
		NameAR:   nameAR,
		IsActive: true,
	}
}
