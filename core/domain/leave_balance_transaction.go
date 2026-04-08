package domain

import "time"

type TransactionType string

const (
	TransactionTypeInitial    TransactionType = "INITIAL"
	TransactionTypeDeduct     TransactionType = "DEDUCT"
	TransactionTypeRefund     TransactionType = "REFUND"
	TransactionTypeAdjustment TransactionType = "ADJUSTMENT"
	TransactionTypeCarryOver  TransactionType = "CARRY_OVER"
)

type LeaveBalanceTransaction struct {
	ID              int64
	UID             string
	BalanceID       int64
	TransactionType TransactionType
	Days            int
	LeaveRecordID   *int64
	Notes           *string
	CreatedBy       *int64
	CreatedAt       time.Time
}

func NewLeaveBalanceTransaction(balanceID int64, txType TransactionType, days int, leaveRecordID, createdBy *int64, notes *string) *LeaveBalanceTransaction {
	return &LeaveBalanceTransaction{
		UID:             GenerateUID("lbtx"),
		BalanceID:       balanceID,
		TransactionType: txType,
		Days:            days,
		LeaveRecordID:   leaveRecordID,
		CreatedBy:       createdBy,
		Notes:           notes,
	}
}
