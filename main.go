package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/banumusa/backend/adapters/db"
	httpAdapter "github.com/banumusa/backend/adapters/http"
	"github.com/banumusa/backend/adapters/legacy"
	"github.com/banumusa/backend/adapters/notifications/fcm"
	"github.com/banumusa/backend/adapters/notifications/noop"
	"github.com/banumusa/backend/adapters/scheduler"
	minioAdapter "github.com/banumusa/backend/adapters/storage/minio"
	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	leaveRequestDocumentRepo := db.NewLeaveRequestDocumentRepository()

	deviceTokenRepo := db.NewDeviceTokenRepository()
	attendanceReminderRepo := db.NewAttendanceReminderRepository()

	attendanceRecordRepo := db.NewAttendanceRecordRepository()
	shiftRepo := db.NewShiftRepository()
	attendanceDeviceRepo := db.NewAttendanceDeviceRepository()

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
		sqliteDB, employeeRepo, leaveTypeRepo, leaveBalanceRepo, leaveRecordRepo, balanceTxRepo, workingDaysCalc, leaveSync,
	)
	getBalanceUC := usecases.NewGetBalanceUseCase(sqliteDB, employeeRepo, leaveTypeRepo, leaveBalanceRepo)
	listLeaveRecordsUC := usecases.NewListLeaveRecordsUseCase(sqliteDB, employeeRepo, leaveTypeRepo, leaveRecordRepo)
	listAllLeaveRecordsUC := usecases.NewListAllLeaveRecordsUseCase(sqliteDB, leaveTypeRepo, leaveRecordRepo)

	getDashboardStatsUC := usecases.NewGetDashboardStatsUseCase(sqliteDB, employeeRepo, leaveRecordRepo, leaveRequestRepo, attendanceRecordRepo)

	getEmployeeUC := usecases.NewGetEmployeeUseCase(sqliteDB, employeeRepo)
	listEmployeesUC := usecases.NewListEmployeesUseCase(sqliteDB, employeeRepo, userRepo, roleRepo)
	importEmployeesUC := usecases.NewImportEmployeesUseCase(sqliteDB, employeeRepo, userRepo, roleRepo)
	exportEmployeesUC := usecases.NewExportEmployeesUseCase(sqliteDB, employeeRepo)
	exportEmployeesPDFUC := usecases.NewExportEmployeesPDFUseCase(sqliteDB, employeeRepo, cfg.FontPath)
	generateTemplateUC := usecases.NewGenerateImportTemplateUseCase()
	assignEmployeeDepartmentUC := usecases.NewAssignEmployeeDepartmentUseCase(sqliteDB, employeeRepo, departmentRepo)
	removeEmployeeDepartmentUC := usecases.NewRemoveEmployeeDepartmentUseCase(sqliteDB, employeeRepo)
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
		sqliteDB, userRepo, roleRepo, refreshTokenRepo, jwtService, cfg.RefreshTokenDays,
	)
	logoutUC := usecases.NewLogoutUseCase(sqliteDB, refreshTokenRepo)
	getCurrentUserUC := usecases.NewGetCurrentUserUseCase(sqliteDB, userRepo, roleRepo, permissionRepo, employeeRepo)

	listUsersUC := usecases.NewListUsersUseCase(sqliteDB, userRepo, roleRepo, employeeRepo)
	createUserUC := usecases.NewCreateUserUseCase(sqliteDB, userRepo)
	updateUserUC := usecases.NewUpdateUserUseCase(sqliteDB, userRepo)
	assignRoleUC := usecases.NewAssignRoleUseCase(sqliteDB, userRepo, roleRepo)
	removeRoleUC := usecases.NewRemoveRoleUseCase(sqliteDB, userRepo, roleRepo)

	listRolesUC := usecases.NewListRolesUseCase(sqliteDB, roleRepo, permissionRepo)
	createRoleUC := usecases.NewCreateRoleUseCase(sqliteDB, roleRepo)
	setPermissionsUC := usecases.NewSetRolePermissionsUseCase(sqliteDB, roleRepo, permissionRepo)
	listPermissionsUC := usecases.NewListPermissionsUseCase(sqliteDB, permissionRepo)

	listApprovalFlowsUC := usecases.NewListApprovalFlowsUseCase(sqliteDB, approvalFlowRepo)
	getApprovalFlowUC := usecases.NewGetApprovalFlowUseCase(sqliteDB, approvalFlowRepo, approvalFlowStepRepo, roleRepo)
	createApprovalFlowUC := usecases.NewCreateApprovalFlowUseCase(sqliteDB, approvalFlowRepo)
	updateApprovalFlowUC := usecases.NewUpdateApprovalFlowUseCase(sqliteDB, approvalFlowRepo)
	listApprovalFlowStepsUC := usecases.NewListApprovalFlowStepsUseCase(sqliteDB, approvalFlowRepo, approvalFlowStepRepo)
	createApprovalFlowStepUC := usecases.NewCreateApprovalFlowStepUseCase(sqliteDB, approvalFlowRepo, approvalFlowStepRepo, roleRepo)
	updateApprovalFlowStepUC := usecases.NewUpdateApprovalFlowStepUseCase(sqliteDB, approvalFlowStepRepo, roleRepo)
	deleteApprovalFlowStepUC := usecases.NewDeleteApprovalFlowStepUseCase(sqliteDB, approvalFlowStepRepo)

	listDepartmentsUC := usecases.NewListDepartmentsUseCase(sqliteDB, departmentRepo)
	getDepartmentUC := usecases.NewGetDepartmentUseCase(sqliteDB, departmentRepo, roleRepo, employeeRepo)
	createDepartmentUC := usecases.NewCreateDepartmentUseCase(sqliteDB, departmentRepo)
	updateDepartmentUC := usecases.NewUpdateDepartmentUseCase(sqliteDB, departmentRepo)
	assignDepartmentManagerUC := usecases.NewAssignDepartmentManagerUseCase(sqliteDB, departmentRepo, userRepo, roleRepo)
	removeDepartmentManagerUC := usecases.NewRemoveDepartmentManagerUseCase(sqliteDB, departmentRepo, roleRepo)

	listLeaveTypesUC := usecases.NewListLeaveTypesUseCase(sqliteDB, leaveTypeRepo)
	listSubLeaveTypesUC := usecases.NewListSubLeaveTypesUseCase(sqliteDB, leaveTypeRepo)
	toggleLeaveTypeUC := usecases.NewToggleLeaveTypeUseCase(sqliteDB, leaveTypeRepo)
	listWeekendDaysUC := usecases.NewListWeekendDaysUseCase(sqliteDB, weekendRepo)
	listHolidaysUC := usecases.NewListHolidaysUseCase(sqliteDB, holidayDefinitionRepo)
	createManualHolidayUC := usecases.NewCreateManualHolidayUseCase(sqliteDB, holidayDefinitionRepo, weekendRepo)
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

	submitLeaveRequestUC := usecases.NewSubmitLeaveRequestUseCase(
		sqliteDB, employeeRepo, leaveTypeRepo, leaveBalanceRepo, leaveRequestRepo, leaveRecordRepo,
		balanceTxRepo, approvalRequestRepo, approvalActionRepo, approvalFlowStepRepo, leaveRequestDocumentRepo, workingDaysCalc, generateDocumentUploadURLUC,
		notificationService, roleRepo,
	)
	cancelLeaveRequestUC := usecases.NewCancelLeaveRequestUseCase(
		sqliteDB, leaveRequestRepo, approvalRequestRepo, approvalActionRepo,
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
		notificationService, userRepo,
	)
	rejectRequestUC := usecases.NewRejectRequestUseCase(
		sqliteDB, employeeRepo, leaveRequestRepo, approvalRequestRepo, approvalActionRepo, approvalFlowStepRepo,
		roleRepo, notificationService, leaveTypeRepo, userRepo,
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
	listDailyDepartmentAttendanceLogsUC := usecases.NewListDailyDepartmentAttendanceLogsUseCase(sqliteDB, departmentRepo, attendanceRecordRepo, employeeRepo, shiftRepo, leaveRecordRepo, weekendRepo, holidayDefinitionRepo)
	listDailyEmployeeAttendanceLogsUC := usecases.NewListDailyEmployeeAttendanceLogsUseCase(sqliteDB, employeeRepo, attendanceRecordRepo, departmentRepo, shiftRepo, leaveRecordRepo, weekendRepo, holidayDefinitionRepo)
	getDailyAttendanceSummaryUC := usecases.NewGetDailyAttendanceSummaryUseCase(sqliteDB, attendanceRecordRepo, employeeRepo, departmentRepo, shiftRepo)

	registerAttendanceDeviceUC := usecases.NewRegisterAttendanceDeviceUseCase(sqliteDB, attendanceDeviceRepo)
	listAttendanceDevicesUC := usecases.NewListAttendanceDevicesUseCase(sqliteDB, attendanceDeviceRepo)
	getAttendanceDeviceUC := usecases.NewGetAttendanceDeviceUseCase(sqliteDB, attendanceDeviceRepo)
	updateAttendanceDeviceUC := usecases.NewUpdateAttendanceDeviceUseCase(sqliteDB, attendanceDeviceRepo)
	deleteAttendanceDeviceUC := usecases.NewDeleteAttendanceDeviceUseCase(sqliteDB, attendanceDeviceRepo)
	activateAttendanceDeviceUC := usecases.NewActivateAttendanceDeviceUseCase(sqliteDB, attendanceDeviceRepo)
	attendanceDeviceStatsUC := usecases.NewGetAttendanceDeviceStatsUseCase(sqliteDB, attendanceDeviceRepo)
	checkAttendanceDeviceConnectionUC := usecases.NewCheckAttendanceDeviceConnectionUseCase(sqliteDB, attendanceDeviceRepo)
	checkAllAttendanceDevicesConnectionUC := usecases.NewCheckAllAttendanceDevicesConnectionUseCase(sqliteDB, attendanceDeviceRepo)
	createAttendanceLogUC := usecases.NewCreateAttendanceLogUseCase(sqliteDB, attendanceRecordRepo, employeeRepo, attendanceDeviceRepo)
	updateAttendanceLogUC := usecases.NewUpdateAttendanceLogUseCase(sqliteDB, attendanceRecordRepo, employeeRepo, attendanceDeviceRepo)
	getMonthlyAttendanceStatsUC := usecases.NewGetMonthlyAttendanceStatsUseCase(sqliteDB, attendanceRecordRepo)

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
	roleHandler := httpAdapter.NewRoleHandler(listRolesUC, createRoleUC, setPermissionsUC, listPermissionsUC)
	dashboardHandler := httpAdapter.NewDashboardHandler(getDashboardStatsUC)
	approvalFlowHandler := httpAdapter.NewApprovalFlowHandler(
		listApprovalFlowsUC, getApprovalFlowUC, createApprovalFlowUC, updateApprovalFlowUC, listApprovalFlowStepsUC,
		createApprovalFlowStepUC, updateApprovalFlowStepUC, deleteApprovalFlowStepUC,
	)
	leaveRequestHandler := httpAdapter.NewLeaveRequestHandler(
		submitLeaveRequestUC, cancelLeaveRequestUC, listLeaveRequestsUC, getLeaveRequestUC,
		listPendingApprovalsUC, approveRequestUC, rejectRequestUC, getApprovalHistoryUC, getCurrentUserUC,
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
	leaveTypeHandler := httpAdapter.NewLeaveTypeHandler(listLeaveTypesUC, listSubLeaveTypesUC, toggleLeaveTypeUC)
	weekendHandler := httpAdapter.NewWeekendHandler(listWeekendDaysUC)
	holidayHandler := httpAdapter.NewHolidayHandler(listHolidaysUC, listWeekendDaysUC, createManualHolidayUC)
	shiftHandler := httpAdapter.NewShiftHandler(listShiftsUC, getShiftUC, createShiftUC, updateShiftUC)
	debugHandler := httpAdapter.NewDebugHandler(sendTestPushUC)
	attendanceHandler := httpAdapter.NewAttendanceHandler(
		listDepartmentAttendanceLogsUC,
		listEmployeeAttendanceLogsUC,
		listDailyDepartmentAttendanceLogsUC,
		listDailyEmployeeAttendanceLogsUC,
		createAttendanceLogUC,
		updateAttendanceLogUC,
		getMonthlyAttendanceStatsUC,
		getDailyAttendanceSummaryUC,
	)
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
		DebugHandler:            debugHandler,
		JWTService:              jwtService,
		AuthEnabled:             cfg.AuthEnabled,
		DocumentHandler:         documentHandler,
	})

	var sched *scheduler.Scheduler
	if cfg.SchedulerEnabled {
		interval := time.Duration(cfg.SchedulerIntervalMinutes) * time.Minute
		if interval <= 0 {
			interval = time.Duration(cfg.SchedulerIntervalHours) * time.Hour
		}
		sched = scheduler.New(autoRejectExpiredUC, holidaySyncUC, absenceSyncUC, notifyMissingCheckOutsUC, interval, cfg.HolidaySyncTimezone)
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
