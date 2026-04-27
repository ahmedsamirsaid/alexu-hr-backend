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
	Date              time.Time
	EmployeeUID       string
	EmployeeName      string
	DepartmentUID     *string
	CheckIn           *time.Time
	CheckInLogUID     *string
	CheckOut          *time.Time
	CheckOutLogUID    *string
	CheckInDevice     *string
	CheckInDeviceUID  *string
	CheckOutDevice    *string
	CheckOutDeviceUID *string
	HasEditHistory    bool
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
	GetByUID(ctx context.Context, q Querier, uid string) (*domain.AttendanceRecord, error)
	GetByUIDs(ctx context.Context, q Querier, uids []string) ([]*domain.AttendanceRecord, error)
	Update(ctx context.Context, q Querier, record *domain.AttendanceRecord) error
	ListByDate(ctx context.Context, q Querier, date time.Time, employeeUID *string) ([]*domain.AttendanceRecord, error)
	ListByDateRange(ctx context.Context, q Querier, startDate, endDate time.Time, employeeUID *string) ([]*domain.AttendanceRecord, error)
	ListByDepartmentUID(ctx context.Context, q Querier, departmentUID string, filter DepartmentAttendanceLogsFilter, params ListParams) ([]*AttendanceRecordWithEmployee, error)
	CountByDepartmentUID(ctx context.Context, q Querier, departmentUID string, filter DepartmentAttendanceLogsFilter) (int, error)
	ListByEmployeeUID(ctx context.Context, q Querier, employeeUID string, filter DepartmentAttendanceLogsFilter, params ListParams) ([]*AttendanceRecordWithEmployee, error)
	CountByEmployeeUID(ctx context.Context, q Querier, employeeUID string, filter DepartmentAttendanceLogsFilter) (int, error)
	ListDailyByDepartmentUID(ctx context.Context, q Querier, departmentUID string, filter DepartmentAttendanceLogsFilter, params ListParams) ([]*DailyAttendanceGroup, error)
	CountDailyByDepartmentUID(ctx context.Context, q Querier, departmentUID string, filter DepartmentAttendanceLogsFilter) (int, error)
	ListDailyByEmployeeUID(ctx context.Context, q Querier, employeeUID string, filter DepartmentAttendanceLogsFilter, params ListParams) ([]*DailyAttendanceGroup, error)
	CountDailyByEmployeeUID(ctx context.Context, q Querier, employeeUID string, filter DepartmentAttendanceLogsFilter) (int, error)
	ListDaily(ctx context.Context, q Querier, filter DepartmentAttendanceLogsFilter, params ListParams) ([]*DailyAttendanceGroup, error)
	ResolveEmployeeUIDByDeviceUserID(ctx context.Context, q Querier, deviceUserID string) (*string, error)
}
