package domain

import "time"

type IncentiveBonus struct {
	ID               int64
	EmployeeUID      string
	BonusDate        time.Time
	DecisionNumber   string
	DecisionDate     time.Time
	DecisionImageURL string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
