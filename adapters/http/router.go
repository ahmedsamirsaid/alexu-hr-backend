package http

import (
	"net/http"
)

type RouterConfig struct {
	LeaveHandler         *LeaveHandler
	EmployeeHandler      *EmployeeHandler
	AuthHandler          *AuthHandler
	UserHandler          *UserHandler
	RoleHandler          *RoleHandler
	DashboardHandler     *DashboardHandler
	ApprovalFlowHandler  *ApprovalFlowHandler
	LeaveRequestHandler  *LeaveRequestHandler
	DepartmentHandler    *DepartmentHandler
	DeviceTokenHandler   *DeviceTokenHandler
	LeaveTypeHandler     *LeaveTypeHandler
	JWTService           *JWTService
	AuthEnabled          bool
}

func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	// Public auth routes (no authentication required)
	mux.HandleFunc("POST /api/v1/auth/otp/request", cfg.AuthHandler.RequestOTP)
	mux.HandleFunc("POST /api/v1/auth/otp/verify", cfg.AuthHandler.VerifyOTP)
	mux.HandleFunc("POST /api/v1/auth/login", cfg.AuthHandler.LoginPassword)
	mux.HandleFunc("POST /api/v1/auth/refresh", cfg.AuthHandler.RefreshToken)

	// Protected routes - wrap with auth middleware
	protectedMux := http.NewServeMux()

	// Dashboard routes
	protectedMux.HandleFunc("GET /api/v1/dashboard/stats", cfg.DashboardHandler.GetStats)

	// Auth routes requiring authentication
	protectedMux.HandleFunc("POST /api/v1/auth/logout", cfg.AuthHandler.Logout)
	protectedMux.HandleFunc("GET /api/v1/auth/me", cfg.AuthHandler.GetCurrentUser)

	// Leave management routes
	protectedMux.Handle("POST /api/v1/leave", RequirePermission("leave:record")(http.HandlerFunc(cfg.LeaveHandler.RecordLeave)))
	protectedMux.Handle("GET /api/v1/leave-records", RequirePermission("leave:read")(http.HandlerFunc(cfg.LeaveHandler.ListAllLeaveRecords)))
	protectedMux.Handle("GET /api/v1/employees/{employeeUid}/balance", RequirePermission("leave:read")(http.HandlerFunc(cfg.LeaveHandler.GetBalance)))
	protectedMux.Handle("GET /api/v1/employees/{employeeUid}/leave-records", RequirePermission("leave:read")(http.HandlerFunc(cfg.LeaveHandler.ListLeaveRecords)))

	// Employee routes
	protectedMux.Handle("GET /api/v1/employees/{uid}", RequirePermission("employees:read")(http.HandlerFunc(cfg.EmployeeHandler.GetEmployee)))
	protectedMux.Handle("GET /api/v1/employees", RequirePermission("employees:read")(http.HandlerFunc(cfg.EmployeeHandler.ListEmployees)))
	protectedMux.Handle("POST /api/v1/employees/import", RequirePermission("employees:import")(http.HandlerFunc(cfg.EmployeeHandler.ImportEmployees)))
	protectedMux.Handle("GET /api/v1/employees/export", RequirePermission("employees:export")(http.HandlerFunc(cfg.EmployeeHandler.ExportEmployees)))
	protectedMux.Handle("GET /api/v1/employees/export/pdf", RequirePermission("employees:export")(http.HandlerFunc(cfg.EmployeeHandler.ExportEmployeesPDF)))
	protectedMux.Handle("GET /api/v1/employees/import/template", RequirePermission("employees:import")(http.HandlerFunc(cfg.EmployeeHandler.DownloadImportTemplate)))
	protectedMux.Handle("PUT /api/v1/employees/{uid}/department", RequirePermission("employees:write")(http.HandlerFunc(cfg.EmployeeHandler.AssignDepartment)))
	protectedMux.Handle("DELETE /api/v1/employees/{uid}/department", RequirePermission("employees:write")(http.HandlerFunc(cfg.EmployeeHandler.RemoveDepartment)))

	// User management routes
	protectedMux.Handle("GET /api/v1/users", RequirePermission("users:read")(http.HandlerFunc(cfg.UserHandler.ListUsers)))
	protectedMux.Handle("POST /api/v1/users", RequirePermission("users:write")(http.HandlerFunc(cfg.UserHandler.CreateUser)))
	protectedMux.Handle("PUT /api/v1/users/{uid}", RequirePermission("users:write")(http.HandlerFunc(cfg.UserHandler.UpdateUser)))
	protectedMux.Handle("POST /api/v1/users/{uid}/roles", RequirePermission("users:write")(http.HandlerFunc(cfg.UserHandler.AssignRole)))
	protectedMux.Handle("DELETE /api/v1/users/{uid}/roles/{roleUid}", RequirePermission("users:write")(http.HandlerFunc(cfg.UserHandler.RemoveRole)))

	// Role management routes
	protectedMux.Handle("GET /api/v1/roles", RequirePermission("roles:read")(http.HandlerFunc(cfg.RoleHandler.ListRoles)))
	protectedMux.Handle("POST /api/v1/roles", RequirePermission("roles:write")(http.HandlerFunc(cfg.RoleHandler.CreateRole)))
	protectedMux.Handle("POST /api/v1/roles/{uid}/permissions", RequirePermission("roles:write")(http.HandlerFunc(cfg.RoleHandler.SetPermissions)))
	protectedMux.Handle("GET /api/v1/permissions", RequirePermission("roles:read")(http.HandlerFunc(cfg.RoleHandler.ListPermissions)))

	// Approval Flow management routes (admin)
	protectedMux.Handle("GET /api/v1/admin/approval-flows", RequirePermission("approval-flows:read")(http.HandlerFunc(cfg.ApprovalFlowHandler.ListApprovalFlows)))
	protectedMux.Handle("POST /api/v1/admin/approval-flows", RequirePermission("approval-flows:write")(http.HandlerFunc(cfg.ApprovalFlowHandler.CreateApprovalFlow)))
	protectedMux.Handle("PATCH /api/v1/admin/approval-flows/{uid}", RequirePermission("approval-flows:write")(http.HandlerFunc(cfg.ApprovalFlowHandler.UpdateApprovalFlow)))
	protectedMux.Handle("GET /api/v1/admin/approval-flows/{uid}/steps", RequirePermission("approval-flows:read")(http.HandlerFunc(cfg.ApprovalFlowHandler.ListApprovalFlowSteps)))
	protectedMux.Handle("POST /api/v1/admin/approval-flows/{uid}/steps", RequirePermission("approval-flows:write")(http.HandlerFunc(cfg.ApprovalFlowHandler.CreateApprovalFlowStep)))
	protectedMux.Handle("PATCH /api/v1/admin/approval-flows/{uid}/steps/{stepUid}", RequirePermission("approval-flows:write")(http.HandlerFunc(cfg.ApprovalFlowHandler.UpdateApprovalFlowStep)))
	protectedMux.Handle("DELETE /api/v1/admin/approval-flows/{uid}/steps/{stepUid}", RequirePermission("approval-flows:write")(http.HandlerFunc(cfg.ApprovalFlowHandler.DeleteApprovalFlowStep)))

	// Department management routes (admin)
	protectedMux.Handle("GET /api/v1/admin/departments", RequirePermission("departments:read")(http.HandlerFunc(cfg.DepartmentHandler.ListDepartments)))
	protectedMux.Handle("GET /api/v1/admin/departments/{uid}", RequirePermission("departments:read")(http.HandlerFunc(cfg.DepartmentHandler.GetDepartment)))
	protectedMux.Handle("POST /api/v1/admin/departments", RequirePermission("departments:write")(http.HandlerFunc(cfg.DepartmentHandler.CreateDepartment)))
	protectedMux.Handle("PATCH /api/v1/admin/departments/{uid}", RequirePermission("departments:write")(http.HandlerFunc(cfg.DepartmentHandler.UpdateDepartment)))
	protectedMux.Handle("POST /api/v1/admin/departments/{uid}/manager", RequirePermission("departments:write")(http.HandlerFunc(cfg.DepartmentHandler.AssignManager)))
	protectedMux.Handle("DELETE /api/v1/admin/departments/{uid}/manager", RequirePermission("departments:write")(http.HandlerFunc(cfg.DepartmentHandler.RemoveManager)))

	// Leave Type management routes (admin)
	protectedMux.Handle("GET /api/v1/admin/leave-types", RequirePermission("leave-types:read")(http.HandlerFunc(cfg.LeaveTypeHandler.ListLeaveTypes)))
	protectedMux.Handle("PATCH /api/v1/admin/leave-types/{uid}/active", RequirePermission("leave-types:write")(http.HandlerFunc(cfg.LeaveTypeHandler.ToggleLeaveType)))

	// Leave Request routes (for employees)
	protectedMux.HandleFunc("POST /api/v1/leave-requests", cfg.LeaveRequestHandler.SubmitLeaveRequest)
	protectedMux.HandleFunc("GET /api/v1/leave-requests", cfg.LeaveRequestHandler.ListLeaveRequests)
	protectedMux.HandleFunc("GET /api/v1/leave-requests/{uid}", cfg.LeaveRequestHandler.GetLeaveRequest)
	protectedMux.HandleFunc("POST /api/v1/leave-requests/{uid}/cancel", cfg.LeaveRequestHandler.CancelLeaveRequest)

	// Approval routes (for approvers)
	protectedMux.HandleFunc("GET /api/v1/approvals/pending", cfg.LeaveRequestHandler.ListPendingApprovals)
	protectedMux.HandleFunc("POST /api/v1/approvals/{uid}/approve", cfg.LeaveRequestHandler.ApproveRequest)
	protectedMux.HandleFunc("POST /api/v1/approvals/{uid}/reject", cfg.LeaveRequestHandler.RejectRequest)
	protectedMux.HandleFunc("GET /api/v1/approvals/{uid}/history", cfg.LeaveRequestHandler.GetApprovalHistory)

	// Device token routes (for push notifications)
	protectedMux.HandleFunc("POST /api/v1/device-tokens", cfg.DeviceTokenHandler.Register)
	protectedMux.HandleFunc("DELETE /api/v1/device-tokens/{uid}", cfg.DeviceTokenHandler.Unregister)

	// Apply auth middleware to protected routes
	authMiddleware := AuthMiddleware(cfg.JWTService, cfg.AuthEnabled)
	mux.Handle("/api/v1/", authMiddleware(protectedMux))

	RegisterSwaggerRoutes(mux)

	return CORSMiddleware(mux)
}
