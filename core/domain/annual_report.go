package domain

import "time"

type AnnualReport struct {
	ID             int64
	EmployeeUID    string
	ReportYear     string
	ReportGrade    string
	Notes          string
	ReportImageURL string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
