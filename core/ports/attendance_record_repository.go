package ports

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
)

type AttendanceRecordWithEmployee struct {
	Record        *domain.AttendanceRecord
	EmployeeName  string
	DepartmentUID *string
	DeviceName    string
}

type DailyAttendanceGroup struct {
	Date           time.Time
	EmployeeUID    string
	EmployeeName   string
	DepartmentUID  *string
	CheckIn        *time.Time
	CheckOut       *time.Time
	CheckInDevice  *string
	CheckOutDevice *string
}

type DepartmentAttendanceLogsFilter struct {
	EmployeeUID      *string
	EmployeeName     *string
	EmployeeNameMode *string
	DeviceUID        *string
	PunchType        *domain.AttendancePunchType
	StartDate        *time.Time
	EndDate          *time.Time
}

type AttendanceRecordRepository interface {
	Create(ctx context.Context, q Querier, record *domain.AttendanceRecord) (bool, error)
	ListByDate(ctx context.Context, q Querier, date time.Time, employeeUID *string) ([]*domain.AttendanceRecord, error)
	ListByDepartmentUID(ctx context.Context, q Querier, departmentUID string, filter DepartmentAttendanceLogsFilter, params ListParams) ([]*AttendanceRecordWithEmployee, error)
	CountByDepartmentUID(ctx context.Context, q Querier, departmentUID string, filter DepartmentAttendanceLogsFilter) (int, error)
	ListByEmployeeUID(ctx context.Context, q Querier, employeeUID string, filter DepartmentAttendanceLogsFilter, params ListParams) ([]*AttendanceRecordWithEmployee, error)
	CountByEmployeeUID(ctx context.Context, q Querier, employeeUID string, filter DepartmentAttendanceLogsFilter) (int, error)
	ListDailyByDepartmentUID(ctx context.Context, q Querier, departmentUID string, filter DepartmentAttendanceLogsFilter, params ListParams) ([]*DailyAttendanceGroup, error)
	CountDailyByDepartmentUID(ctx context.Context, q Querier, departmentUID string, filter DepartmentAttendanceLogsFilter) (int, error)
	ListDailyByEmployeeUID(ctx context.Context, q Querier, employeeUID string, filter DepartmentAttendanceLogsFilter, params ListParams) ([]*DailyAttendanceGroup, error)
	CountDailyByEmployeeUID(ctx context.Context, q Querier, employeeUID string, filter DepartmentAttendanceLogsFilter) (int, error)
	ResolveEmployeeUIDByDeviceUserID(ctx context.Context, q Querier, deviceUserID string) (*string, error)
}
