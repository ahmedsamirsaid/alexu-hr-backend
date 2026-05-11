package domain

import "time"

type Penalty struct {
	ID                     int64
	EmployeeUID            string
	PenaltyType            string
	PenaltyReason          string
	PenaltyDecisionNumber  string
	PenaltyDecisionDate    time.Time
	PenaltyDecisionFileURL string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type PenaltyRemoval struct {
	ID                               int64
	PenaltyID                        int64
	PenaltyRemovalType               string
	PenaltyRemovalNumber             string
	PenaltyRemovalDate               time.Time
	Notes                            string
	PenaltyWithdrawalDecisionFileURL string
	CreatedAt                        time.Time
	UpdatedAt                        time.Time
}
