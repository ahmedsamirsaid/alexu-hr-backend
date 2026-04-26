package http

import (
	"net/http"
)

type RouterConfig struct {
	LeaveHandler            *LeaveHandler
	EmployeeHandler         *EmployeeHandler
	ShiftHandler            *ShiftHandler
	AuthHandler             *AuthHandler
	UserHandler             *UserHandler
	RoleHandler             *RoleHandler
	DashboardHandler        *DashboardHandler
	ApprovalFlowHandler     *ApprovalFlowHandler
	LeaveRequestHandler     *LeaveRequestHandler
	DepartmentHandler       *DepartmentHandler
	DeviceTokenHandler      *DeviceTokenHandler
	AttendanceDeviceHandler *AttendanceDeviceHandler
	LeaveTypeHandler        *LeaveTypeHandler
	WeekendHandler          *WeekendHandler
	AttendanceHandler       *AttendanceHandler
	HolidayHandler          *HolidayHandler
	DebugHandler            *DebugHandler
	JWTService              *JWTService
	AuthEnabled             bool
	DocumentHandler         *DocumentHandler
}

func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/auth/otp/request", cfg.AuthHandler.RequestOTP)
	mux.HandleFunc("POST /api/v1/auth/otp/verify", cfg.AuthHandler.VerifyOTP)
	mux.HandleFunc("POST /api/v1/auth/login", cfg.AuthHandler.LoginPassword)
	mux.HandleFunc("POST /api/v1/auth/refresh", cfg.AuthHandler.RefreshToken)

	protectedMux := http.NewServeMux()

	protectedMux.Handle("GET /api/v1/dashboard/stats", RequirePermission("employees:read")(http.HandlerFunc(cfg.DashboardHandler.GetStats)))
	protectedMux.HandleFunc("POST /api/v1/auth/logout", cfg.AuthHandler.Logout)
	protectedMux.HandleFunc("GET /api/v1/auth/me", cfg.AuthHandler.GetCurrentUser)

	protectedMux.Handle("POST /api/v1/leave", RequirePermission("leave:record")(http.HandlerFunc(cfg.LeaveHandler.RecordLeave)))
	protectedMux.Handle("GET /api/v1/leave-records", RequirePermission("leave:record")(http.HandlerFunc(cfg.LeaveHandler.ListAllLeaveRecords)))
	protectedMux.Handle("GET /api/v1/employees/{employeeUid}/balance", RequirePermission("leave:read")(http.HandlerFunc(cfg.LeaveHandler.GetBalance)))
	protectedMux.Handle("GET /api/v1/employees/{employeeUid}/leave-records", RequirePermission("leave:read")(http.HandlerFunc(cfg.LeaveHandler.ListLeaveRecords)))
	protectedMux.Handle("POST /api/v1/documents/presign-upload", RequirePermission("documents:write")(http.HandlerFunc(cfg.DocumentHandler.GenerateUploadURL)))
	protectedMux.Handle("POST /api/v1/documents/presign-download", RequirePermission("documents:read")(http.HandlerFunc(cfg.DocumentHandler.GenerateDownloadURL)))

	protectedMux.Handle("GET /api/v1/employees/{uid}", RequirePermission("employees:read")(http.HandlerFunc(cfg.EmployeeHandler.GetEmployee)))
	protectedMux.Handle("GET /api/v1/employees", RequirePermission("employees:read")(http.HandlerFunc(cfg.EmployeeHandler.ListEmployees)))
	protectedMux.Handle("POST /api/v1/employees/import", RequirePermission("employees:import")(http.HandlerFunc(cfg.EmployeeHandler.ImportEmployees)))
	protectedMux.Handle("GET /api/v1/employees/export", RequirePermission("employees:export")(http.HandlerFunc(cfg.EmployeeHandler.ExportEmployees)))
	protectedMux.Handle("GET /api/v1/employees/export/pdf", RequirePermission("employees:export")(http.HandlerFunc(cfg.EmployeeHandler.ExportEmployeesPDF)))
	protectedMux.Handle("GET /api/v1/employees/import/template", RequirePermission("employees:import")(http.HandlerFunc(cfg.EmployeeHandler.DownloadImportTemplate)))
	protectedMux.Handle("PUT /api/v1/employees/{uid}/department", RequirePermission("employees:write")(http.HandlerFunc(cfg.EmployeeHandler.AssignDepartment)))
	protectedMux.Handle("DELETE /api/v1/employees/{uid}/department", RequirePermission("employees:write")(http.HandlerFunc(cfg.EmployeeHandler.RemoveDepartment)))

	protectedMux.Handle("GET /api/v1/users", RequirePermission("users:read")(http.HandlerFunc(cfg.UserHandler.ListUsers)))
	protectedMux.Handle("POST /api/v1/users", RequirePermission("users:write")(http.HandlerFunc(cfg.UserHandler.CreateUser)))
	protectedMux.Handle("PUT /api/v1/users/{uid}", RequirePermission("users:write")(http.HandlerFunc(cfg.UserHandler.UpdateUser)))
	protectedMux.Handle("POST /api/v1/users/{uid}/roles", RequirePermission("users:write")(http.HandlerFunc(cfg.UserHandler.AssignRole)))
	protectedMux.Handle("DELETE /api/v1/users/{uid}/roles/{roleUid}", RequirePermission("users:write")(http.HandlerFunc(cfg.UserHandler.RemoveRole)))

	protectedMux.Handle("GET /api/v1/roles", RequirePermission("roles:read")(http.HandlerFunc(cfg.RoleHandler.ListRoles)))
	protectedMux.Handle("POST /api/v1/roles", RequirePermission("roles:write")(http.HandlerFunc(cfg.RoleHandler.CreateRole)))
	protectedMux.Handle("PATCH /api/v1/roles/{uid}/scope", RequirePermission("roles:write")(http.HandlerFunc(cfg.RoleHandler.SetScope)))
	protectedMux.Handle("POST /api/v1/roles/{uid}/permissions", RequirePermission("roles:write")(http.HandlerFunc(cfg.RoleHandler.SetPermissions)))
	protectedMux.Handle("GET /api/v1/permissions", RequirePermission("roles:read")(http.HandlerFunc(cfg.RoleHandler.ListPermissions)))

	protectedMux.Handle("GET /api/v1/admin/approval-flows", RequirePermission("approval-flows:read")(http.HandlerFunc(cfg.ApprovalFlowHandler.ListApprovalFlows)))
	protectedMux.Handle("GET /api/v1/admin/approval-flows/{uid}", RequirePermission("approval-flows:read")(http.HandlerFunc(cfg.ApprovalFlowHandler.GetApprovalFlow)))
	protectedMux.Handle("POST /api/v1/admin/approval-flows", RequirePermission("approval-flows:write")(http.HandlerFunc(cfg.ApprovalFlowHandler.CreateApprovalFlow)))
	protectedMux.Handle("PATCH /api/v1/admin/approval-flows/{uid}", RequirePermission("approval-flows:write")(http.HandlerFunc(cfg.ApprovalFlowHandler.UpdateApprovalFlow)))
	protectedMux.Handle("GET /api/v1/admin/approval-flows/{uid}/steps", RequirePermission("approval-flows:read")(http.HandlerFunc(cfg.ApprovalFlowHandler.ListApprovalFlowSteps)))
	protectedMux.Handle("POST /api/v1/admin/approval-flows/{uid}/steps", RequirePermission("approval-flows:write")(http.HandlerFunc(cfg.ApprovalFlowHandler.CreateApprovalFlowStep)))
	protectedMux.Handle("PATCH /api/v1/admin/approval-flows/{uid}/steps/{stepUid}", RequirePermission("approval-flows:write")(http.HandlerFunc(cfg.ApprovalFlowHandler.UpdateApprovalFlowStep)))
	protectedMux.Handle("DELETE /api/v1/admin/approval-flows/{uid}/steps/{stepUid}", RequirePermission("approval-flows:write")(http.HandlerFunc(cfg.ApprovalFlowHandler.DeleteApprovalFlowStep)))

	protectedMux.Handle("GET /api/v1/admin/shifts", RequirePermission("shift:read")(http.HandlerFunc(cfg.ShiftHandler.List)))
	protectedMux.Handle("GET /api/v1/admin/shifts/{uid}", RequirePermission("shift:read")(http.HandlerFunc(cfg.ShiftHandler.Get)))
	protectedMux.Handle("POST /api/v1/admin/shifts", RequirePermission("shift:write")(http.HandlerFunc(cfg.ShiftHandler.Create)))
	protectedMux.Handle("PATCH /api/v1/admin/shifts/{uid}", RequirePermission("shift:write")(http.HandlerFunc(cfg.ShiftHandler.Update)))

	protectedMux.Handle("GET /api/v1/admin/departments", RequirePermission("departments:read")(http.HandlerFunc(cfg.DepartmentHandler.ListDepartments)))
	protectedMux.Handle("GET /api/v1/admin/departments/{uid}", RequirePermission("departments:read")(http.HandlerFunc(cfg.DepartmentHandler.GetDepartment)))
	protectedMux.Handle("POST /api/v1/admin/departments", RequirePermission("departments:write")(http.HandlerFunc(cfg.DepartmentHandler.CreateDepartment)))
	protectedMux.Handle("PATCH /api/v1/admin/departments/{uid}", RequirePermission("departments:write")(http.HandlerFunc(cfg.DepartmentHandler.UpdateDepartment)))
	protectedMux.Handle("POST /api/v1/admin/departments/{uid}/manager", RequirePermission("departments:write")(http.HandlerFunc(cfg.DepartmentHandler.AssignManager)))
	protectedMux.Handle("DELETE /api/v1/admin/departments/{uid}/manager", RequirePermission("departments:write")(http.HandlerFunc(cfg.DepartmentHandler.RemoveManager)))

	protectedMux.Handle("GET /api/v1/admin/leave-types", RequirePermission("leave-types:read")(http.HandlerFunc(cfg.LeaveTypeHandler.ListLeaveTypes)))
	protectedMux.Handle("PATCH /api/v1/admin/leave-types/{uid}/active", RequirePermission("leave-types:write")(http.HandlerFunc(cfg.LeaveTypeHandler.ToggleLeaveType)))
	protectedMux.Handle("GET /api/v1/leave-types/{uid}/sub-leave-types", RequirePermission("leave:request")(http.HandlerFunc(cfg.LeaveTypeHandler.ListSubLeaveTypes)))
	protectedMux.Handle("GET /api/v1/admin/weekend-config", RequirePermission("departments:read")(http.HandlerFunc(cfg.WeekendHandler.ListWeekendDays)))
	protectedMux.Handle("GET /api/v1/admin/holidays", RequirePermission("holidays:read")(http.HandlerFunc(cfg.HolidayHandler.ListHolidays)))
	protectedMux.Handle("POST /api/v1/admin/holidays", RequirePermission("holidays:write")(http.HandlerFunc(cfg.HolidayHandler.CreateHoliday)))
	protectedMux.Handle("PATCH /api/v1/admin/holidays/{uid}", RequirePermission("holidays:write")(http.HandlerFunc(cfg.HolidayHandler.UpdateHoliday)))
	protectedMux.Handle("DELETE /api/v1/admin/holidays/{uid}", RequirePermission("holidays:write")(http.HandlerFunc(cfg.HolidayHandler.DeleteHoliday)))

	protectedMux.Handle("POST /api/v1/leave-requests", RequirePermission("leave:request")(http.HandlerFunc(cfg.LeaveRequestHandler.SubmitLeaveRequest)))
	protectedMux.Handle("GET /api/v1/leave-requests", RequirePermission("leave:request")(http.HandlerFunc(cfg.LeaveRequestHandler.ListLeaveRequests)))
	protectedMux.Handle("GET /api/v1/leave-requests/{uid}", RequirePermission("leave:request")(http.HandlerFunc(cfg.LeaveRequestHandler.GetLeaveRequest)))
	protectedMux.Handle("PATCH /api/v1/leave-requests/{uid}", RequirePermission("leave:request")(http.HandlerFunc(cfg.LeaveRequestHandler.UpdateRejectedLeaveRequest)))
	protectedMux.Handle("POST /api/v1/leave-requests/{uid}/cancel", RequirePermission("leave:request")(http.HandlerFunc(cfg.LeaveRequestHandler.CancelLeaveRequest)))

	protectedMux.Handle("GET /api/v1/approvals/pending", RequirePermission("leave:approve")(http.HandlerFunc(cfg.LeaveRequestHandler.ListPendingApprovals)))
	protectedMux.Handle("POST /api/v1/approvals/{uid}/approve", RequirePermission("leave:approve")(http.HandlerFunc(cfg.LeaveRequestHandler.ApproveRequest)))
	protectedMux.Handle("POST /api/v1/approvals/{uid}/reject", RequirePermission("leave:approve")(http.HandlerFunc(cfg.LeaveRequestHandler.RejectRequest)))
	protectedMux.Handle("GET /api/v1/approvals/{uid}/history", RequirePermission("leave:approve")(http.HandlerFunc(cfg.LeaveRequestHandler.GetApprovalHistory)))

	protectedMux.HandleFunc("POST /api/v1/device-tokens", cfg.DeviceTokenHandler.Register)
	protectedMux.HandleFunc("DELETE /api/v1/device-tokens/{uid}", cfg.DeviceTokenHandler.Unregister)

	protectedMux.Handle("GET /api/v1/attendance/departments/{departmentUid}/logs", RequirePermission("attendance:read")(http.HandlerFunc(cfg.AttendanceHandler.ListDepartmentLogs)))
	protectedMux.Handle("GET /api/v1/attendance/departments/{departmentUid}/daily-logs", RequirePermission("attendance:read")(http.HandlerFunc(cfg.AttendanceHandler.ListDailyDepartmentLogs)))
	protectedMux.Handle("GET /api/v1/attendance/employees/{employeeUid}/logs", RequirePermission("attendance:read")(http.HandlerFunc(cfg.AttendanceHandler.ListEmployeeLogs)))
	protectedMux.Handle("GET /api/v1/attendance/employees/{employeeUid}/daily-logs", RequirePermission("attendance:read")(http.HandlerFunc(cfg.AttendanceHandler.ListDailyEmployeeLogs)))
	protectedMux.Handle("POST /api/v1/attendance/logs", RequirePermission("attendance:write")(http.HandlerFunc(cfg.AttendanceHandler.CreateLog)))
	protectedMux.Handle("PATCH /api/v1/attendance/logs/{uid}", RequirePermission("attendance:write")(http.HandlerFunc(cfg.AttendanceHandler.UpdateLog)))
	protectedMux.Handle("GET /api/v1/attendance/logs/{uid}/history", RequirePermission("attendance:read")(http.HandlerFunc(cfg.AttendanceHandler.GetLogHistory)))
	protectedMux.Handle("GET /api/v1/attendance/summary/daily", RequirePermission("attendance:read")(http.HandlerFunc(cfg.AttendanceHandler.GetDailySummary)))

	protectedMux.Handle("GET /api/v1/admin/attendance-devices/stats", RequirePermission("attendance-devices:read")(http.HandlerFunc(cfg.AttendanceDeviceHandler.Stats)))
	protectedMux.Handle("POST /api/v1/admin/debug/push-test", RequirePermission("users:write")(http.HandlerFunc(cfg.DebugHandler.SendTestPush)))
	protectedMux.Handle("GET /api/v1/admin/attendance-devices", RequirePermission("attendance-devices:read")(http.HandlerFunc(cfg.AttendanceDeviceHandler.List)))
	protectedMux.Handle("GET /api/v1/admin/attendance-devices/{uid}", RequirePermission("attendance-devices:read")(http.HandlerFunc(cfg.AttendanceDeviceHandler.Get)))
	protectedMux.Handle("POST /api/v1/admin/attendance-devices", RequirePermission("attendance-devices:write")(http.HandlerFunc(cfg.AttendanceDeviceHandler.Register)))
	protectedMux.Handle("PATCH /api/v1/admin/attendance-devices/{uid}", RequirePermission("attendance-devices:write")(http.HandlerFunc(cfg.AttendanceDeviceHandler.Update)))
	protectedMux.Handle("POST /api/v1/admin/attendance-devices/check-all", RequirePermission("attendance-devices:read")(http.HandlerFunc(cfg.AttendanceDeviceHandler.CheckAllConnections)))
	protectedMux.Handle("POST /api/v1/admin/attendance-devices/{uid}/check-connection", RequirePermission("attendance-devices:read")(http.HandlerFunc(cfg.AttendanceDeviceHandler.CheckConnection)))
	protectedMux.Handle("DELETE /api/v1/admin/attendance-devices/{uid}", RequirePermission("attendance-devices:write")(http.HandlerFunc(cfg.AttendanceDeviceHandler.Delete)))
	protectedMux.Handle("PATCH /api/v1/admin/attendance-devices/{uid}/activate", RequirePermission("attendance-devices:write")(http.HandlerFunc(cfg.AttendanceDeviceHandler.Activate)))
	protectedMux.Handle("GET /api/v1/attendance/employees/{employeeUid}/logs/stats/monthly", RequirePermission("attendance:read")(http.HandlerFunc(cfg.AttendanceHandler.GetMonthlyStats)))
	authMiddleware := AuthMiddleware(cfg.JWTService, cfg.AuthEnabled)
	mux.Handle("/api/v1/", authMiddleware(protectedMux))

	RegisterSwaggerRoutes(mux)

	return CORSMiddleware(mux)
}
