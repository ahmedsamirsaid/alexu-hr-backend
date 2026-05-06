package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/banumusa/backend/adapters/db"
	httpAdapter "github.com/banumusa/backend/adapters/http"
	"github.com/banumusa/backend/adapters/legacy"
	"github.com/banumusa/backend/adapters/notifications/fcm"
	"github.com/banumusa/backend/adapters/notifications/noop"
	"github.com/banumusa/backend/adapters/scheduler"
	minioAdapter "github.com/banumusa/backend/adapters/storage/minio"
	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cfg := LoadConfig()

	sqliteDB, err := db.NewSQLiteDB(cfg.DBPath)
	if err != nil {
		slog.Error("main.main.connect_database", "error", err)
		os.Exit(1)
	}
	defer sqliteDB.Close()

	if err := runMigrations(sqliteDB.DB(), cfg.DBMigrationsPath); err != nil {
		slog.Error("main.main.run_migrations", "error", err)
		os.Exit(1)
	}

	minioService, err := minioAdapter.NewService(minioAdapter.Config{
		Endpoint:  cfg.MinIOEndpoint,
		AccessKey: cfg.MinIOAccessKey,
		SecretKey: cfg.MinIOSecretKey,
		UseSSL:    cfg.MinIOUseSSL,
	})
	if err != nil {
		slog.Error("main.main.init_minio", "error", err)
		os.Exit(1)
	}

	if cfg.MinIOAutoCreateBucket {
		if err := minioService.EnsureBucket(context.Background(), cfg.MinIODocumentsBucket); err != nil {
			slog.Error("main.main.ensure_documents_bucket", "error", err)
			os.Exit(1)
		}
	}

	employeeRepo := db.NewEmployeeRepository()
	leaveTypeRepo := db.NewLeaveTypeRepository()
	leaveBalanceRepo := db.NewLeaveBalanceRepository()
	leaveRecordRepo := db.NewLeaveRecordRepository()
	balanceTxRepo := db.NewLeaveBalanceTransactionRepository()
	weekendRepo := db.NewWeekendConfigRepository()
	holidayDefinitionRepo := db.NewHolidayDefinitionRepository()

	userRepo := db.NewUserRepository()
	roleRepo := db.NewRoleRepository()
	permissionRepo := db.NewPermissionRepository()
	otpRepo := db.NewOTPRepository()
	refreshTokenRepo := db.NewRefreshTokenRepository()

	departmentRepo := db.NewDepartmentRepository()

	approvalFlowRepo := db.NewApprovalFlowRepository()
	approvalFlowStepRepo := db.NewApprovalFlowStepRepository()
	approvalRequestRepo := db.NewApprovalRequestRepository()
	approvalActionRepo := db.NewApprovalActionRepository()
	leaveRequestRepo := db.NewLeaveRequestRepository()
	permissionRequestRepo := db.NewPermissionRequestRepository()
	leaveRequestDocumentRepo := db.NewLeaveRequestDocumentRepository()

	deviceTokenRepo := db.NewDeviceTokenRepository()
	attendanceReminderRepo := db.NewAttendanceReminderRepository()

	attendanceRecordRepo := db.NewAttendanceRecordRepository()
	shiftRepo := db.NewShiftRepository()
	attendanceDeviceRepo := db.NewAttendanceDeviceRepository()
	auditLogRepo := db.NewAuditLogRepository()
	auditor := audit.NewAuditor(sqliteDB, auditLogRepo)
	var notificationService ports.NotificationService
	if cfg.FCMEnabled {
		fcmService, err := fcm.NewFCMNotificationService(context.Background(), fcm.Config{
			ServiceAccountJSONPath: cfg.FCMServiceAccountPath,
			DeviceTokenRepo:        deviceTokenRepo,
			DB:                     sqliteDB,
		})
		if err != nil {
			slog.Error("main.main.init_fcm", "error", err)
			os.Exit(1)
		}
		notificationService = fcmService
		slog.Info("main.main.fcm_enabled")
	} else {
		notificationService = noop.NewNoopNotificationService()
		slog.Info("main.main.fcm_disabled")
	}

	leaveSync := legacy.NewNoopLeaveSyncAdapter()
	workingDaysCalc := usecases.NewWorkingDaysCalculator(weekendRepo, holidayDefinitionRepo)

	recordLeaveUC := usecases.NewRecordLeaveUseCase(
		sqliteDB, employeeRepo, leaveTypeRepo, leaveBalanceRepo, leaveRecordRepo, balanceTxRepo, workingDaysCalc, leaveSync, auditor,
	)
	getBalanceUC := usecases.NewGetBalanceUseCase(sqliteDB, employeeRepo, leaveTypeRepo, leaveBalanceRepo)
	listLeaveRecordsUC := usecases.NewListLeaveRecordsUseCase(sqliteDB, employeeRepo, leaveTypeRepo, leaveRecordRepo)
	listAllLeaveRecordsUC := usecases.NewListAllLeaveRecordsUseCase(sqliteDB, leaveTypeRepo, leaveRecordRepo)

	getDashboardStatsUC := usecases.NewGetDashboardStatsUseCase(sqliteDB, employeeRepo, leaveRecordRepo, leaveRequestRepo, attendanceRecordRepo)

	getEmployeeUC := usecases.NewGetEmployeeUseCase(sqliteDB, employeeRepo)
	listEmployeesUC := usecases.NewListEmployeesUseCase(sqliteDB, employeeRepo, userRepo, roleRepo)
	importEmployeesUC := usecases.NewImportEmployeesUseCase(sqliteDB, employeeRepo, userRepo, roleRepo, auditor)
	exportEmployeesUC := usecases.NewExportEmployeesUseCase(sqliteDB, employeeRepo)
	exportEmployeesPDFUC := usecases.NewExportEmployeesPDFUseCase(sqliteDB, employeeRepo, cfg.FontPath)
	generateTemplateUC := usecases.NewGenerateImportTemplateUseCase()
	assignEmployeeDepartmentUC := usecases.NewAssignEmployeeDepartmentUseCase(sqliteDB, employeeRepo, departmentRepo, auditor)
	removeEmployeeDepartmentUC := usecases.NewRemoveEmployeeDepartmentUseCase(sqliteDB, employeeRepo, auditor)
	listShiftsUC := usecases.NewListShiftsUseCase(sqliteDB, shiftRepo)
	getShiftUC := usecases.NewGetShiftUseCase(sqliteDB, shiftRepo)
	createShiftUC := usecases.NewCreateShiftUseCase(sqliteDB, shiftRepo)
	updateShiftUC := usecases.NewUpdateShiftUseCase(sqliteDB, shiftRepo)

	jwtService := httpAdapter.NewJWTService(cfg.JWTSecret, cfg.AccessTokenMinutes)

	requestOTPUC := usecases.NewRequestOTPUseCase(sqliteDB, otpRepo, userRepo)
	verifyOTPUC := usecases.NewVerifyOTPUseCase(
		sqliteDB, userRepo, roleRepo, permissionRepo, employeeRepo, otpRepo, refreshTokenRepo,
		jwtService, cfg.DevOTPBypass, cfg.DevBypassOTP, cfg.RefreshTokenDays,
	)
	loginPasswordUC := usecases.NewLoginPasswordUseCase(
		sqliteDB, userRepo, roleRepo, permissionRepo, employeeRepo, refreshTokenRepo,
		jwtService, cfg.DevOTPBypass, cfg.DevBypassOTP, cfg.RefreshTokenDays,
	)
	refreshTokenUC := usecases.NewRefreshTokenUseCase(
		sqliteDB, userRepo, roleRepo, permissionRepo, refreshTokenRepo, jwtService, cfg.RefreshTokenDays,
	)
	logoutUC := usecases.NewLogoutUseCase(sqliteDB, refreshTokenRepo)
	getCurrentUserUC := usecases.NewGetCurrentUserUseCase(sqliteDB, userRepo, roleRepo, permissionRepo, employeeRepo)

	listUsersUC := usecases.NewListUsersUseCase(sqliteDB, userRepo, roleRepo, employeeRepo)
	createUserUC := usecases.NewCreateUserUseCase(sqliteDB, userRepo, auditor)
	updateUserUC := usecases.NewUpdateUserUseCase(sqliteDB, userRepo, auditor)
	assignRoleUC := usecases.NewAssignRoleUseCase(sqliteDB, userRepo, roleRepo, departmentRepo, auditor)
	removeRoleUC := usecases.NewRemoveRoleUseCase(sqliteDB, userRepo, roleRepo, auditor)

	listRolesUC := usecases.NewListRolesUseCase(sqliteDB, roleRepo, permissionRepo)
	createRoleUC := usecases.NewCreateRoleUseCase(sqliteDB, roleRepo, auditor)
	setRoleScopeUC := usecases.NewSetRoleScopeUseCase(sqliteDB, roleRepo, auditor)
	setPermissionsUC := usecases.NewSetRolePermissionsUseCase(sqliteDB, roleRepo, permissionRepo, auditor)
	listPermissionsUC := usecases.NewListPermissionsUseCase(sqliteDB, permissionRepo)

	listApprovalFlowsUC := usecases.NewListApprovalFlowsUseCase(sqliteDB, approvalFlowRepo)
	getApprovalFlowUC := usecases.NewGetApprovalFlowUseCase(sqliteDB, approvalFlowRepo, approvalFlowStepRepo, roleRepo)

	createApprovalFlowUC := usecases.NewCreateApprovalFlowUseCase(sqliteDB, approvalFlowRepo, auditor)
	updateApprovalFlowUC := usecases.NewUpdateApprovalFlowUseCase(sqliteDB, approvalFlowRepo, auditor)
	listApprovalFlowStepsUC := usecases.NewListApprovalFlowStepsUseCase(sqliteDB, approvalFlowRepo, approvalFlowStepRepo)
	createApprovalFlowStepUC := usecases.NewCreateApprovalFlowStepUseCase(sqliteDB, approvalFlowRepo, approvalFlowStepRepo, roleRepo, auditor)
	updateApprovalFlowStepUC := usecases.NewUpdateApprovalFlowStepUseCase(sqliteDB, approvalFlowStepRepo, roleRepo, auditor)
	deleteApprovalFlowStepUC := usecases.NewDeleteApprovalFlowStepUseCase(sqliteDB, approvalFlowStepRepo, auditor)

	listDepartmentsUC := usecases.NewListDepartmentsUseCase(sqliteDB, departmentRepo)
	getDepartmentUC := usecases.NewGetDepartmentUseCase(sqliteDB, departmentRepo, roleRepo, employeeRepo)
	createDepartmentUC := usecases.NewCreateDepartmentUseCase(sqliteDB, departmentRepo, auditor)
	updateDepartmentUC := usecases.NewUpdateDepartmentUseCase(sqliteDB, departmentRepo, auditor)
	assignDepartmentManagerUC := usecases.NewAssignDepartmentManagerUseCase(sqliteDB, departmentRepo, userRepo, roleRepo, employeeRepo, auditor)
	removeDepartmentManagerUC := usecases.NewRemoveDepartmentManagerUseCase(sqliteDB, departmentRepo, roleRepo, userRepo, employeeRepo, auditor)

	getLeaveTypeDetailsUC := usecases.NewGetLeaveTypeDetailsUseCase(sqliteDB, leaveTypeRepo, approvalFlowRepo, approvalFlowStepRepo, roleRepo)
	listLeaveTypesUC := usecases.NewListLeaveTypesUseCase(sqliteDB, leaveTypeRepo)
	listSubLeaveTypesUC := usecases.NewListSubLeaveTypesUseCase(sqliteDB, leaveTypeRepo)
	updateLeaveTypeUC := usecases.NewUpdateLeaveTypeUseCase(sqliteDB, leaveTypeRepo, auditor)
	setLeaveTypeApprovalFlowUC := usecases.NewSetLeaveTypeApprovalFlowUseCase(sqliteDB, leaveTypeRepo, approvalFlowRepo, auditor)
	toggleLeaveTypeUC := usecases.NewToggleLeaveTypeUseCase(sqliteDB, leaveTypeRepo, auditor)
	listWeekendDaysUC := usecases.NewListWeekendDaysUseCase(sqliteDB, weekendRepo)
	listHolidaysUC := usecases.NewListHolidaysUseCase(sqliteDB, holidayDefinitionRepo)
	createManualHolidayUC := usecases.NewCreateManualHolidayUseCase(sqliteDB, holidayDefinitionRepo, weekendRepo, auditor)
	generateDocumentUploadURLUC := usecases.NewGenerateDocumentUploadURLUseCase(
		minioService,
		cfg.MinIODocumentsBucket,
		cfg.MinIOUploadExpiryMinutes,
	)
	generateDocumentDownloadURLUC := usecases.NewGenerateDocumentDownloadURLUseCase(
		minioService,
		cfg.MinIODocumentsBucket,
		cfg.MinIODownloadExpiryMinutes,
	)
	updateHolidayUC := usecases.NewUpdateHolidayUseCase(sqliteDB, holidayDefinitionRepo, weekendRepo, auditor)
	deleteHolidayUC := usecases.NewDeleteHolidayUseCase(sqliteDB, holidayDefinitionRepo, auditor)

	submitLeaveRequestUC := usecases.NewSubmitLeaveRequestUseCase(
		sqliteDB, userRepo, employeeRepo, leaveTypeRepo, leaveBalanceRepo, leaveRequestRepo, leaveRecordRepo,
		balanceTxRepo, approvalRequestRepo, approvalActionRepo, approvalFlowStepRepo, leaveRequestDocumentRepo, workingDaysCalc,
		generateDocumentUploadURLUC, notificationService, roleRepo, auditor,
	)
	cancelLeaveRequestUC := usecases.NewCancelLeaveRequestUseCase(
		sqliteDB, leaveRequestRepo, approvalRequestRepo, approvalActionRepo, auditor,
	)
	updateRejectedLeaveRequestUC := usecases.NewUpdateRejectedLeaveRequestUseCase(
		sqliteDB, employeeRepo, leaveTypeRepo, leaveRequestRepo, leaveRequestDocumentRepo, approvalRequestRepo, approvalActionRepo, workingDaysCalc, generateDocumentUploadURLUC, auditor,
	)
	listLeaveRequestsUC := usecases.NewListLeaveRequestsUseCase(sqliteDB, leaveRequestRepo, approvalRequestRepo, leaveTypeRepo)
	getLeaveRequestUC := usecases.NewGetLeaveRequestUseCase(
		sqliteDB, leaveRequestRepo, leaveRequestDocumentRepo, approvalRequestRepo, employeeRepo, leaveTypeRepo, generateDocumentDownloadURLUC,
	)

	listPendingApprovalsUC := usecases.NewListPendingApprovalsUseCase(
		sqliteDB, userRepo, employeeRepo, leaveTypeRepo, leaveRequestRepo, approvalRequestRepo, approvalFlowStepRepo, roleRepo,
	)
	approveRequestUC := usecases.NewApproveRequestUseCase(
		sqliteDB, employeeRepo, leaveTypeRepo, leaveBalanceRepo, leaveRecordRepo, balanceTxRepo,
		leaveRequestRepo, approvalRequestRepo, approvalActionRepo, approvalFlowStepRepo, roleRepo,
		notificationService, userRepo, auditor,
	)
	rejectRequestUC := usecases.NewRejectRequestUseCase(
		sqliteDB, employeeRepo, leaveRequestRepo, approvalRequestRepo, approvalActionRepo, approvalFlowStepRepo,
		roleRepo, notificationService, leaveTypeRepo, userRepo, auditor,
	)
	submitPermissionUC := usecases.NewSubmitPermissionRequestUseCase(
		sqliteDB, userRepo, employeeRepo, departmentRepo, shiftRepo,
		permissionRequestRepo, approvalRequestRepo, approvalActionRepo, approvalFlowStepRepo,
		weekendRepo, roleRepo, notificationService,
	)
	updatePermissionUC := usecases.NewUpdatePermissionRequestUseCase(
		sqliteDB, permissionRequestRepo, approvalRequestRepo,
		employeeRepo, departmentRepo, shiftRepo, weekendRepo,
	)
	cancelPermissionUC := usecases.NewCancelPermissionRequestUseCase(
		sqliteDB, permissionRequestRepo, approvalRequestRepo, approvalActionRepo,
		employeeRepo, departmentRepo, shiftRepo, userRepo, notificationService,
	)
	listPermissionRequestsUC := usecases.NewListPermissionRequestsUseCase(
		sqliteDB, permissionRequestRepo, approvalRequestRepo, employeeRepo,
	)
	getPermissionRequestUC := usecases.NewGetPermissionRequestUseCase(
		sqliteDB, permissionRequestRepo, approvalRequestRepo, approvalFlowStepRepo, employeeRepo, roleRepo,
	)
	listPendingPermissionApprovalsUC := usecases.NewListPendingPermissionApprovalsUseCase(
		sqliteDB, permissionRequestRepo, approvalRequestRepo, approvalFlowStepRepo, employeeRepo, roleRepo,
	)
	approvePermissionUC := usecases.NewApprovePermissionRequestUseCase(
		sqliteDB, permissionRequestRepo, approvalRequestRepo, approvalActionRepo, approvalFlowStepRepo,
		employeeRepo, departmentRepo, shiftRepo, roleRepo, userRepo, notificationService,
	)
	rejectPermissionUC := usecases.NewRejectPermissionRequestUseCase(
		sqliteDB, permissionRequestRepo, approvalRequestRepo, approvalActionRepo, approvalFlowStepRepo,
		employeeRepo, roleRepo, userRepo, notificationService,
	)
	autoRejectExpiredPermissionsUC := usecases.NewAutoRejectExpiredPermissionRequestsUseCase(
		sqliteDB, permissionRequestRepo, approvalRequestRepo, approvalActionRepo,
		employeeRepo, departmentRepo, shiftRepo, cfg.HolidaySyncTimezone,
	)
	permissionEligibilityUC := usecases.NewGetPermissionEligibilityUseCase(
		sqliteDB, employeeRepo, departmentRepo, shiftRepo, permissionRequestRepo, weekendRepo, cfg.HolidaySyncTimezone,
	)

	getApprovalHistoryUC := usecases.NewGetApprovalHistoryUseCase(
		sqliteDB, approvalRequestRepo, approvalActionRepo, employeeRepo,
	)

	registerDeviceTokenUC := usecases.NewRegisterDeviceTokenUseCase(deviceTokenRepo, sqliteDB)
	unregisterDeviceTokenUC := usecases.NewUnregisterDeviceTokenUseCase(deviceTokenRepo, sqliteDB)
	notifyMissingCheckOutsUC := usecases.NewNotifyMissingCheckOutsUseCase(
		sqliteDB, attendanceRecordRepo, attendanceReminderRepo, employeeRepo, departmentRepo, shiftRepo, holidayDefinitionRepo, leaveRecordRepo, userRepo, notificationService,
	)
	sendTestPushUC := usecases.NewSendTestPushNotificationUseCase(sqliteDB, userRepo, notificationService)

	listDepartmentAttendanceLogsUC := usecases.NewListDepartmentAttendanceLogsUseCase(sqliteDB, departmentRepo, attendanceRecordRepo)
	listEmployeeAttendanceLogsUC := usecases.NewListEmployeeAttendanceLogsUseCase(sqliteDB, employeeRepo, attendanceRecordRepo)
	listDailyDepartmentAttendanceLogsUC := usecases.NewListDailyDepartmentAttendanceLogsUseCase(sqliteDB, departmentRepo, attendanceRecordRepo, employeeRepo, shiftRepo, leaveRecordRepo, weekendRepo, holidayDefinitionRepo, permissionRequestRepo)
	listDailyEmployeeAttendanceLogsUC := usecases.NewListDailyEmployeeAttendanceLogsUseCase(sqliteDB, employeeRepo, attendanceRecordRepo, departmentRepo, shiftRepo, leaveRecordRepo, weekendRepo, holidayDefinitionRepo, permissionRequestRepo)
	getDailyAttendanceSummaryUC := usecases.NewGetDailyAttendanceSummaryUseCase(sqliteDB, attendanceRecordRepo, employeeRepo, departmentRepo, shiftRepo)
	departmentAttendanceReportUC := usecases.NewGetDepartmentAttendanceReportUseCase(sqliteDB, departmentRepo, attendanceRecordRepo, employeeRepo, shiftRepo, weekendRepo, holidayDefinitionRepo)
	exportDepartmentAttendanceReportUC := usecases.NewExportDepartmentAttendanceReportUseCase(departmentAttendanceReportUC)

	registerAttendanceDeviceUC := usecases.NewRegisterAttendanceDeviceUseCase(sqliteDB, attendanceDeviceRepo, auditor)
	listAttendanceDevicesUC := usecases.NewListAttendanceDevicesUseCase(sqliteDB, attendanceDeviceRepo)
	getAttendanceDeviceUC := usecases.NewGetAttendanceDeviceUseCase(sqliteDB, attendanceDeviceRepo)
	updateAttendanceDeviceUC := usecases.NewUpdateAttendanceDeviceUseCase(sqliteDB, attendanceDeviceRepo, auditor)
	deleteAttendanceDeviceUC := usecases.NewDeleteAttendanceDeviceUseCase(sqliteDB, attendanceDeviceRepo, auditor)
	activateAttendanceDeviceUC := usecases.NewActivateAttendanceDeviceUseCase(sqliteDB, attendanceDeviceRepo, auditor)
	attendanceDeviceStatsUC := usecases.NewGetAttendanceDeviceStatsUseCase(sqliteDB, attendanceDeviceRepo)
	checkAttendanceDeviceConnectionUC := usecases.NewCheckAttendanceDeviceConnectionUseCase(sqliteDB, attendanceDeviceRepo)
	checkAllAttendanceDevicesConnectionUC := usecases.NewCheckAllAttendanceDevicesConnectionUseCase(sqliteDB, attendanceDeviceRepo)
	createAttendanceLogUC := usecases.NewCreateAttendanceLogUseCase(sqliteDB, attendanceRecordRepo, employeeRepo, attendanceDeviceRepo, auditor)
	updateAttendanceLogUC := usecases.NewUpdateAttendanceLogUseCase(sqliteDB, attendanceRecordRepo, employeeRepo, attendanceDeviceRepo, auditor)
	importAttendanceLogsUC := usecases.NewImportAttendanceLogsUseCase(sqliteDB, attendanceRecordRepo, employeeRepo, attendanceDeviceRepo, auditor)
	getAttendanceLogHistoryUC := usecases.NewGetAttendanceLogHistoryUseCase(sqliteDB, attendanceRecordRepo, auditLogRepo, employeeRepo, userRepo)
	
	getMonthlyAttendanceStatsUC := usecases.NewGetMonthlyAttendanceStatsUseCase(
		sqliteDB,
		attendanceRecordRepo,
		usecases.MonthlyAttendanceStatsDependencies{
			EmployeeRepo: employeeRepo,
			DeptRepo:     departmentRepo,
			ShiftRepo:    shiftRepo,
			WeekendRepo:  weekendRepo,
			HolidayRepo:  holidayDefinitionRepo,
		},
	)
	exportEmployeeAttendanceReportUC := usecases.NewExportEmployeeAttendanceReportUseCase(getMonthlyAttendanceStatsUC)

	autoRejectExpiredUC := usecases.NewAutoRejectExpiredRequestsUseCase(
		sqliteDB, leaveRequestRepo, approvalRequestRepo, approvalActionRepo, cfg.ExpiredLeaveGraceDays,
	)
	holidaySyncUC := usecases.NewSyncEgyptPublicHolidaysUseCase(
		sqliteDB,
		holidayDefinitionRepo,
		cfg.HolidaySyncEndpoint,
		cfg.HolidaySyncAPIKey,
		cfg.HolidaySyncCountry,
		cfg.HolidaySyncTimezone,
	)
	absenceSyncUC := usecases.NewSyncDailyAbsencesUseCase(
		sqliteDB,
		leaveRecordRepo,
		weekendRepo,
		holidayDefinitionRepo,
		cfg.HolidaySyncTimezone,
	)

	leaveHandler := httpAdapter.NewLeaveHandler(recordLeaveUC, getBalanceUC, listLeaveRecordsUC, listAllLeaveRecordsUC)
	employeeHandler := httpAdapter.NewEmployeeHandler(getEmployeeUC, listEmployeesUC, importEmployeesUC, exportEmployeesUC, exportEmployeesPDFUC, generateTemplateUC, assignEmployeeDepartmentUC, removeEmployeeDepartmentUC)
	authHandler := httpAdapter.NewAuthHandler(requestOTPUC, verifyOTPUC, loginPasswordUC, refreshTokenUC, logoutUC, getCurrentUserUC)
	userHandler := httpAdapter.NewUserHandler(listUsersUC, createUserUC, updateUserUC, assignRoleUC, removeRoleUC)
	roleHandler := httpAdapter.NewRoleHandler(listRolesUC, createRoleUC, setPermissionsUC, setRoleScopeUC, listPermissionsUC)
	dashboardHandler := httpAdapter.NewDashboardHandler(getDashboardStatsUC)
	approvalFlowHandler := httpAdapter.NewApprovalFlowHandler(
		listApprovalFlowsUC, getApprovalFlowUC, createApprovalFlowUC, updateApprovalFlowUC, listApprovalFlowStepsUC,
		createApprovalFlowStepUC, updateApprovalFlowStepUC, deleteApprovalFlowStepUC,
	)
	leaveRequestHandler := httpAdapter.NewLeaveRequestHandler(
		submitLeaveRequestUC, updateRejectedLeaveRequestUC, cancelLeaveRequestUC, listLeaveRequestsUC, getLeaveRequestUC,
		getEmployeeUC, listPendingApprovalsUC, approveRequestUC, rejectRequestUC, getApprovalHistoryUC, getCurrentUserUC,
	)
	permissionRequestHandler := httpAdapter.NewPermissionRequestHandler(
		submitPermissionUC, updatePermissionUC, cancelPermissionUC,
		listPermissionRequestsUC, getPermissionRequestUC,
		listPendingPermissionApprovalsUC, approvePermissionUC, rejectPermissionUC,
		getApprovalHistoryUC, permissionEligibilityUC, getCurrentUserUC,
	)
	departmentHandler := httpAdapter.NewDepartmentHandler(
		listDepartmentsUC, getDepartmentUC, createDepartmentUC, updateDepartmentUC,
		assignDepartmentManagerUC, removeDepartmentManagerUC,
	)
	deviceTokenHandler := httpAdapter.NewDeviceTokenHandler(registerDeviceTokenUC, unregisterDeviceTokenUC)
	attendanceDeviceHandler := httpAdapter.NewAttendanceDeviceHandler(
		registerAttendanceDeviceUC,
		listAttendanceDevicesUC,
		getAttendanceDeviceUC,
		updateAttendanceDeviceUC,
		deleteAttendanceDeviceUC,
		activateAttendanceDeviceUC,
		attendanceDeviceStatsUC,
		checkAttendanceDeviceConnectionUC,
		checkAllAttendanceDevicesConnectionUC,
	)
	leaveTypeHandler := httpAdapter.NewLeaveTypeHandler(getLeaveTypeDetailsUC, listLeaveTypesUC, listSubLeaveTypesUC, updateLeaveTypeUC, setLeaveTypeApprovalFlowUC, toggleLeaveTypeUC)
	weekendHandler := httpAdapter.NewWeekendHandler(listWeekendDaysUC)
	holidayHandler := httpAdapter.NewHolidayHandler(sqliteDB, listHolidaysUC, listWeekendDaysUC, createManualHolidayUC, updateHolidayUC, deleteHolidayUC, employeeRepo)
	shiftHandler := httpAdapter.NewShiftHandler(listShiftsUC, getShiftUC, createShiftUC, updateShiftUC)
	debugHandler := httpAdapter.NewDebugHandler(sendTestPushUC)
	attendanceHandler := httpAdapter.NewAttendanceHandler(
		listDepartmentAttendanceLogsUC,
		listEmployeeAttendanceLogsUC,
		listDailyDepartmentAttendanceLogsUC,
		listDailyEmployeeAttendanceLogsUC,
		createAttendanceLogUC,
		updateAttendanceLogUC,
		importAttendanceLogsUC,
		getAttendanceLogHistoryUC,
		getMonthlyAttendanceStatsUC,
		getDailyAttendanceSummaryUC,
		httpAdapter.AttendanceReportDependencies{
			GetDepartmentReportUC:    departmentAttendanceReportUC,
			ExportDepartmentReportUC: exportDepartmentAttendanceReportUC,
			ExportEmployeeReportUC:   exportEmployeeAttendanceReportUC,
		},
	)

	getAuditTrailUC := usecases.NewGetAuditTrailUseCase(sqliteDB, auditLogRepo, employeeRepo)
	getActorAuditEventsUC := usecases.NewGetActorAuditEventsUseCase(sqliteDB, auditLogRepo)
	listAllAuditLogsUC := usecases.NewListAllAuditLogsUseCase(auditLogRepo, sqliteDB)
	listEnrichedAuditLogsUC := usecases.NewListEnrichedAuditLogsUseCase(
		auditLogRepo, employeeRepo, departmentRepo, leaveRequestRepo, leaveTypeRepo, attendanceRecordRepo, sqliteDB,
	)
	getAuditFilterOptionsUC := usecases.NewGetAuditFilterOptionsUseCase(auditLogRepo, sqliteDB)
	auditHandler := httpAdapter.NewAuditHandler(getAuditTrailUC, getActorAuditEventsUC, listAllAuditLogsUC, listEnrichedAuditLogsUC, getAuditFilterOptionsUC)


	documentHandler := httpAdapter.NewDocumentHandler(
		generateDocumentUploadURLUC,
		generateDocumentDownloadURLUC,
	)

	router := httpAdapter.NewRouter(httpAdapter.RouterConfig{
		LeaveHandler:            leaveHandler,
		EmployeeHandler:         employeeHandler,
		AuthHandler:             authHandler,
		UserHandler:             userHandler,
		RoleHandler:             roleHandler,
		DashboardHandler:        dashboardHandler,
		ApprovalFlowHandler:     approvalFlowHandler,
		LeaveRequestHandler:     leaveRequestHandler,
		DepartmentHandler:       departmentHandler,
		DeviceTokenHandler:      deviceTokenHandler,
		AttendanceDeviceHandler: attendanceDeviceHandler,
		LeaveTypeHandler:        leaveTypeHandler,
		WeekendHandler:          weekendHandler,
		HolidayHandler:          holidayHandler,
		ShiftHandler:            shiftHandler,
		AttendanceHandler:       attendanceHandler,
		AuditHandler:            auditHandler,
		DebugHandler:            debugHandler,
		JWTService:              jwtService,
		AuthEnabled:             cfg.AuthEnabled,
		DocumentHandler:         documentHandler,
		PermissionRequestHandler: permissionRequestHandler,
	})

	var sched *scheduler.Scheduler
	if cfg.SchedulerEnabled {
		interval := time.Duration(cfg.SchedulerIntervalMinutes) * time.Minute
		if interval <= 0 {
			interval = time.Duration(cfg.SchedulerIntervalHours) * time.Hour
		}
		sched = scheduler.New(autoRejectExpiredUC, autoRejectExpiredPermissionsUC, holidaySyncUC, absenceSyncUC, notifyMissingCheckOutsUC, interval, cfg.HolidaySyncTimezone)
		sched.Start(context.Background())
		slog.Info("main.main.scheduler_enabled", "interval", interval, "grace_days", cfg.ExpiredLeaveGraceDays)
	} else {
		slog.Info("main.main.scheduler_disabled")
	}

	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", cfg.Port),
		Handler: router,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("main.main.server_listening", "addr", server.Addr, "auth_enabled", cfg.AuthEnabled)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("main.main.server_error", "error", err)
			os.Exit(1)
		}
	}()

	<-stop
	slog.Info("main.main.shutdown_initiated")

	if sched != nil {
		sched.Stop()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("main.main.server_shutdown_error", "error", err)
	}

	slog.Info("main.main.shutdown_complete")
}

func runMigrations(sqlDB *sql.DB, migrationsPath string) error {
	if err := reconcileMigrationVersion(sqlDB, migrationsPath); err != nil {
		return fmt.Errorf("failed to reconcile migration version: %w", err)
	}

	driver, err := sqlite.WithInstance(sqlDB, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"sqlite",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func reconcileMigrationVersion(sqlDB *sql.DB, migrationsPath string) error {
	currentVersion, dirty, hasVersion, err := readSchemaMigrationVersion(sqlDB)
	if err != nil || !hasVersion || dirty {
		return err
	}

	versions, err := listMigrationVersions(migrationsPath)
	if err != nil {
		return err
	}
	if len(versions) == 0 {
		return nil
	}

	if _, ok := versions[currentVersion]; ok {
		return nil
	}

	targetVersion, ok := legacyOrdinalToVersion(currentVersion, versions)
	if !ok {
		return nil
	}

	if _, err := sqlDB.Exec(`UPDATE schema_migrations SET version = ?`, targetVersion); err != nil {
		return err
	}

	slog.Info("main.main.reconciled_migration_version", "from", currentVersion, "to", targetVersion)
	return nil
}

func readSchemaMigrationVersion(sqlDB *sql.DB) (version uint64, dirty bool, hasVersion bool, err error) {
	var count int
	if err = sqlDB.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'schema_migrations'`).Scan(&count); err != nil {
		return 0, false, false, err
	}
	if count == 0 {
		return 0, false, false, nil
	}

	err = sqlDB.QueryRow(`SELECT version, dirty FROM schema_migrations LIMIT 1`).Scan(&version, &dirty)
	if err == sql.ErrNoRows {
		return 0, false, false, nil
	}
	if err != nil {
		return 0, false, false, err
	}

	return version, dirty, true, nil
}

func listMigrationVersions(migrationsPath string) (map[uint64]struct{}, error) {
	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return nil, err
	}

	versions := make(map[uint64]struct{})
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if filepath.Ext(name) != ".sql" || (!strings.HasSuffix(name, ".up.sql") && !strings.HasSuffix(name, ".down.sql")) {
			continue
		}

		prefix, _, found := strings.Cut(name, "_")
		if !found {
			continue
		}

		version, err := strconv.ParseUint(prefix, 10, 64)
		if err != nil {
			continue
		}

		versions[version] = struct{}{}
	}

	return versions, nil
}

func legacyOrdinalToVersion(legacyVersion uint64, versions map[uint64]struct{}) (uint64, bool) {
	ordered := make([]uint64, 0, len(versions))
	for version := range versions {
		ordered = append(ordered, version)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })

	if legacyVersion == 0 || legacyVersion > uint64(len(ordered)) {
		return 0, false
	}

	return ordered[legacyVersion-1], true
}
