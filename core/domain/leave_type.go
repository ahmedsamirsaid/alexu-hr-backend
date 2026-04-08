package domain

import "time"

type LeaveType struct {
	ID                    int64
	UID                   string
	Code                  string
	NameEN                string
	NameAR                string
	DefaultBalance        int
	MaxConsecutive        *int
	RecordingDeadlineDays *int
	AdvanceNoticeDays     *int
	IsActive              bool
	ApprovalFlowUID       *string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// RequiresApproval returns true if this leave type requires approval workflow
func (lt *LeaveType) RequiresApproval() bool {
	return lt.ApprovalFlowUID != nil
}
