package domain

type LeaveRequestDocument struct {
	LeaveRequestUID string
	FileName        string
	ObjectKey       string
}

func NewLeaveRequestDocument(leaveRequestUID, fileName, objectKey string) *LeaveRequestDocument {
	return &LeaveRequestDocument{
		LeaveRequestUID: leaveRequestUID,
		FileName:        fileName,
		ObjectKey:       objectKey,
	}
}
