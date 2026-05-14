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
	"github.com/banumusa/backend/core/i18n"
	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cfg := LoadConfig()

	var dbHost, dbPort string
	if cfg.PgBouncerEnabled {
		dbHost = cfg.PgBouncerHost
		dbPort = cfg.PgBouncerPort
	} else {
		dbHost = cfg.PostgresHost
		dbPort = cfg.PostgresPort
	}

	pgDB, err := db.NewPostgresDB(dbHost, dbPort, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB, cfg.PostgresSSLMode)
	if err != nil {
		slog.Error("main.main.connect_database", "error", err)
		os.Exit(1)
	}
	defer pgDB.Close()

	if err := runMigrations(pgDB.DB(), cfg.DBMigrationsPath); err != nil {
		slog.Error("main.main.run_migrations", "error", err)
		os.Exit(1)
	}

	// Run seed data if enabled
	if cfg.SeedDevData {
		if err := runSeedData(pgDB.DB()); err != nil {
			slog.Error("main.main.run_seed_data", "error", err)
			os.Exit(1)
		}
		slog.Info("main.main.seed_data_executed")
	}

	// Initialize I18n service
	i18nLoader := i18n.NewLoader(cfg.I18nLocalesPath)
	translations, err := i18nLoader.LoadAll()
	if err != nil {
		slog.Error("main.main.load_translations", "error", err)
		os.Exit(1)
	}

	i18nService := i18n.NewService(translations, cfg.I18nDefaultLocale, cfg.I18nSupportedLocales)
	slog.Info("main.main.i18n_initialized",
		"locales", i18nService.GetSupportedLocales(),
		"default", i18nService.GetDefaultLocale(),
		"translation_count", func() int {
			count := 0
			for _, localeTranslations := range translations {
				count += len(localeTranslations)
			}
			return count
		}(),
	)

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
	profileChangeRequestRepo := db.NewEmployeeProfileChangeRequestRepository()
	leaveRequestRepo := db.NewLeaveRequestRepository()
	permissionRequestRepo := db.NewPermissionRequestRepository()
	leaveRequestDocumentRepo := db.NewLeaveRequestDocumentRepository()

	deviceTokenRepo := db.NewDeviceTokenRepository()
	attendanceReminderRepo := db.NewAttendanceReminderRepository()

	attendanceRecordRepo := db.NewAttendanceRecordRepository()
	shiftRepo := db.NewShiftRepository()
	attendanceDeviceRepo := db.NewAttendanceDeviceRepository()
	penaltyRepo := db.NewPenaltyRepository()
	incentiveBonusRepo := db.NewIncentiveBonusRepository()
	annualReportRepo := db.NewAnnualReportRepository()
	auditLogRepo := db.NewAuditLogRepository()
	auditor := audit.NewAuditor(pgDB, auditLogRepo)
	var notificationService ports.NotificationService
	if cfg.FCMEnabled {
		fcmService, err := fcm.NewFCMNotificationService(context.Background(), fcm.Config{
			ServiceAccountJSONPath: cfg.FCMServiceAccountPath,
			DeviceTokenRepo:        deviceTokenRepo,
			UserRepo:               userRepo,
			I18nService:            i18nService,
			DB:                     pgDB,
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
		pgDB, employeeRepo, leaveTypeRepo, leaveBalanceRepo, leaveRecordRepo, balanceTxRepo, workingDaysCalc, leaveSync, auditor,
	)
	getBalanceUC := usecases.NewGetBalanceUseCase(pgDB, employeeRepo, leaveTypeRepo, leaveBalanceRepo)
	listLeaveRecordsUC := usecases.NewListLeaveRecordsUseCase(pgDB, employeeRepo, leaveTypeRepo, leaveRecordRepo)
	listAllLeaveRecordsUC := usecases.NewListAllLeaveRecordsUseCase(pgDB, leaveTypeRepo, leaveRecordRepo)

	getDashboardStatsUC := usecases.NewGetDashboardStatsUseCase(pgDB, employeeRepo, leaveRecordRepo, leaveRequestRepo, attendanceRecordRepo)

	setEmployeeManagerUC := usecases.NewSetEmployeeManagerUseCase(pgDB, employeeRepo, roleRepo, auditor)
	listManagerCandidatesUC := usecases.NewListManagerCandidatesUseCase(pgDB, employeeRepo, roleRepo)
	listEmployeesUC := usecases.NewListEmployeesUseCase(pgDB, employeeRepo, userRepo, roleRepo)
	importEmployeesUC := usecases.NewImportEmployeesUseCase(pgDB, employeeRepo, userRepo, roleRepo, auditor)
	exportEmployeesUC := usecases.NewExportEmployeesUseCase(pgDB, employeeRepo)
	exportEmployeesPDFUC := usecases.NewExportEmployeesPDFUseCase(pgDB, employeeRepo, cfg.FontPath)
	generateTemplateUC := usecases.NewGenerateImportTemplateUseCase()
	assignEmployeeDepartmentUC := usecases.NewAssignEmployeeDepartmentUseCase(pgDB, employeeRepo, departmentRepo, auditor)
	removeEmployeeDepartmentUC := usecases.NewRemoveEmployeeDepartmentUseCase(pgDB, employeeRepo, auditor)
	listShiftsUC := usecases.NewListShiftsUseCase(pgDB, shiftRepo)
	getShiftUC := usecases.NewGetShiftUseCase(pgDB, shiftRepo)
	createShiftUC := usecases.NewCreateShiftUseCase(pgDB, shiftRepo)
	updateShiftUC := usecases.NewUpdateShiftUseCase(pgDB, shiftRepo)

	jwtService := httpAdapter.NewJWTService(cfg.JWTSecret, cfg.AccessTokenMinutes)

	requestOTPUC := usecases.NewRequestOTPUseCase(pgDB, otpRepo, userRepo)
	verifyOTPUC := usecases.NewVerifyOTPUseCase(
		pgDB, userRepo, roleRepo, permissionRepo, employeeRepo, otpRepo, refreshTokenRepo,
		jwtService, cfg.DevOTPBypass, cfg.DevBypassOTP, cfg.RefreshTokenDays, auditor,
	)
	loginPasswordUC := usecases.NewLoginPasswordUseCase(
		pgDB, userRepo, roleRepo, permissionRepo, employeeRepo, refreshTokenRepo,
		jwtService, cfg.DevOTPBypass, cfg.DevBypassOTP, cfg.RefreshTokenDays, auditor,
	)
	refreshTokenUC := usecases.NewRefreshTokenUseCase(
		pgDB, userRepo, roleRepo, permissionRepo, refreshTokenRepo, jwtService, cfg.RefreshTokenDays,
	)
	logoutUC := usecases.NewLogoutUseCase(pgDB, refreshTokenRepo, userRepo, auditor)
	getCurrentUserUC := usecases.NewGetCurrentUserUseCase(pgDB, userRepo, roleRepo, permissionRepo, employeeRepo)

	listUsersUC := usecases.NewListUsersUseCase(pgDB, userRepo, roleRepo, employeeRepo)
	createUserUC := usecases.NewCreateUserUseCase(pgDB, userRepo, auditor)
	updateUserUC := usecases.NewUpdateUserUseCase(pgDB, userRepo, auditor)
	assignRoleUC := usecases.NewAssignRoleUseCase(pgDB, userRepo, roleRepo, departmentRepo, auditor)
	removeRoleUC := usecases.NewRemoveRoleUseCase(pgDB, userRepo, roleRepo, auditor)

	listRolesUC := usecases.NewListRolesUseCase(pgDB, roleRepo, permissionRepo)
	createRoleUC := usecases.NewCreateRoleUseCase(pgDB, roleRepo, auditor)
	setRoleScopeUC := usecases.NewSetRoleScopeUseCase(pgDB, roleRepo, auditor)
	setPermissionsUC := usecases.NewSetRolePermissionsUseCase(pgDB, roleRepo, permissionRepo, auditor)
	listPermissionsUC := usecases.NewListPermissionsUseCase(pgDB, permissionRepo)

	listApprovalFlowsUC := usecases.NewListApprovalFlowsUseCase(pgDB, approvalFlowRepo)
	getApprovalFlowUC := usecases.NewGetApprovalFlowUseCase(pgDB, approvalFlowRepo, approvalFlowStepRepo, roleRepo)

	createApprovalFlowUC := usecases.NewCreateApprovalFlowUseCase(pgDB, approvalFlowRepo, auditor)
	updateApprovalFlowUC := usecases.NewUpdateApprovalFlowUseCase(pgDB, approvalFlowRepo, auditor)
	listApprovalFlowStepsUC := usecases.NewListApprovalFlowStepsUseCase(pgDB, approvalFlowRepo, approvalFlowStepRepo)
	createApprovalFlowStepUC := usecases.NewCreateApprovalFlowStepUseCase(pgDB, approvalFlowRepo, approvalFlowStepRepo, roleRepo, auditor)
	updateApprovalFlowStepUC := usecases.NewUpdateApprovalFlowStepUseCase(pgDB, approvalFlowStepRepo, roleRepo, auditor)
	deleteApprovalFlowStepUC := usecases.NewDeleteApprovalFlowStepUseCase(pgDB, approvalFlowStepRepo, auditor)

	listDepartmentsUC := usecases.NewListDepartmentsUseCase(pgDB, departmentRepo)
	getDepartmentUC := usecases.NewGetDepartmentUseCase(pgDB, departmentRepo, roleRepo, employeeRepo)
	createDepartmentUC := usecases.NewCreateDepartmentUseCase(pgDB, departmentRepo, auditor)
	updateDepartmentUC := usecases.NewUpdateDepartmentUseCase(pgDB, departmentRepo, auditor)
	assignDepartmentManagerUC := usecases.NewAssignDepartmentManagerUseCase(pgDB, departmentRepo, userRepo, roleRepo, employeeRepo, auditor)
	removeDepartmentManagerUC := usecases.NewRemoveDepartmentManagerUseCase(pgDB, departmentRepo, roleRepo, userRepo, employeeRepo, auditor)

	getLeaveTypeDetailsUC := usecases.NewGetLeaveTypeDetailsUseCase(pgDB, leaveTypeRepo, approvalFlowRepo, approvalFlowStepRepo, roleRepo)
	listLeaveTypesUC := usecases.NewListLeaveTypesUseCase(pgDB, leaveTypeRepo)
	listSubLeaveTypesUC := usecases.NewListSubLeaveTypesUseCase(pgDB, leaveTypeRepo)
	updateLeaveTypeUC := usecases.NewUpdateLeaveTypeUseCase(pgDB, leaveTypeRepo, auditor)
	setLeaveTypeApprovalFlowUC := usecases.NewSetLeaveTypeApprovalFlowUseCase(pgDB, leaveTypeRepo, approvalFlowRepo, auditor)
	toggleLeaveTypeUC := usecases.NewToggleLeaveTypeUseCase(pgDB, leaveTypeRepo, auditor)
	listWeekendDaysUC := usecases.NewListWeekendDaysUseCase(pgDB, weekendRepo)
	listHolidaysUC := usecases.NewListHolidaysUseCase(pgDB, holidayDefinitionRepo)
	createManualHolidayUC := usecases.NewCreateManualHolidayUseCase(pgDB, holidayDefinitionRepo, weekendRepo, auditor)
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
	getEmployeeUC := usecases.NewGetEmployeeUseCase(pgDB, employeeRepo, generateDocumentDownloadURLUC, roleRepo)
	createEmployeeUC := usecases.NewCreateEmployeeUseCase(pgDB, employeeRepo, userRepo, roleRepo, generateDocumentUploadURLUC, auditor)
	updateOwnEmployeeProfileUC := usecases.NewUpdateOwnEmployeeProfileUseCase(pgDB, employeeRepo, userRepo, generateDocumentUploadURLUC, auditor)
	updateEmployeeProfileUC := usecases.NewUpdateEmployeeProfileUseCase(pgDB, employeeRepo, userRepo, roleRepo, generateDocumentUploadURLUC, auditor)
	listEmployeePenaltiesUC := usecases.NewListEmployeePenaltiesUseCase(pgDB, employeeRepo, penaltyRepo, generateDocumentDownloadURLUC)
	createEmployeePenaltyUC := usecases.NewCreateEmployeePenaltyUseCase(pgDB, employeeRepo, penaltyRepo, generateDocumentUploadURLUC)
	createEmployeePenaltyRemovalUC := usecases.NewCreateEmployeePenaltyRemovalUseCase(pgDB, penaltyRepo, generateDocumentUploadURLUC)
	updateEmployeePenaltiesUC := usecases.NewUpdateEmployeePenaltiesUseCase(pgDB, employeeRepo, penaltyRepo)
	listEmployeeIncentiveBonusesUC := usecases.NewListEmployeeIncentiveBonusesUseCase(pgDB, employeeRepo, incentiveBonusRepo, generateDocumentDownloadURLUC)
	createEmployeeIncentiveBonusUC := usecases.NewCreateEmployeeIncentiveBonusUseCase(pgDB, employeeRepo, incentiveBonusRepo, generateDocumentUploadURLUC)
	updateEmployeeIncentiveBonusesUC := usecases.NewUpdateEmployeeIncentiveBonusesUseCase(pgDB, employeeRepo, incentiveBonusRepo)
	listEmployeeAnnualReportsUC := usecases.NewListEmployeeAnnualReportsUseCase(pgDB, employeeRepo, annualReportRepo, generateDocumentDownloadURLUC)
	createEmployeeAnnualReportUC := usecases.NewCreateEmployeeAnnualReportUseCase(pgDB, employeeRepo, annualReportRepo, generateDocumentUploadURLUC)
	updateEmployeeAnnualReportsUC := usecases.NewUpdateEmployeeAnnualReportsUseCase(pgDB, employeeRepo, annualReportRepo)
	updateHolidayUC := usecases.NewUpdateHolidayUseCase(pgDB, holidayDefinitionRepo, weekendRepo, auditor)
	deleteHolidayUC := usecases.NewDeleteHolidayUseCase(pgDB, holidayDefinitionRepo, auditor)
	
	submitLeaveRequestUC := usecases.NewSubmitLeaveRequestUseCase(
		pgDB, userRepo, employeeRepo, leaveTypeRepo, leaveBalanceRepo, leaveRequestRepo, leaveRecordRepo,
		balanceTxRepo, approvalRequestRepo, approvalActionRepo, approvalFlowStepRepo, leaveRequestDocumentRepo, workingDaysCalc,
		generateDocumentUploadURLUC, notificationService, roleRepo, auditor,
	)
	cancelLeaveRequestUC := usecases.NewCancelLeaveRequestUseCase(
		pgDB, leaveRequestRepo, approvalRequestRepo, approvalActionRepo, auditor,
	)
	updateRejectedLeaveRequestUC := usecases.NewUpdateRejectedLeaveRequestUseCase(
		pgDB, employeeRepo, leaveTypeRepo, leaveRequestRepo, leaveRequestDocumentRepo, approvalRequestRepo, approvalActionRepo, workingDaysCalc, generateDocumentUploadURLUC, auditor,
	)
	listLeaveRequestsUC := usecases.NewListLeaveRequestsUseCase(pgDB, leaveRequestRepo, approvalRequestRepo, leaveTypeRepo)
	getLeaveRequestUC := usecases.NewGetLeaveRequestUseCase(
		pgDB, leaveRequestRepo, leaveRequestDocumentRepo, approvalRequestRepo, employeeRepo, leaveTypeRepo, generateDocumentDownloadURLUC,
	)

	listPendingApprovalsUC := usecases.NewListPendingApprovalsUseCase(
		pgDB, userRepo, employeeRepo, leaveTypeRepo, leaveRequestRepo, approvalRequestRepo, approvalFlowStepRepo, roleRepo,
	)
	approveRequestUC := usecases.NewApproveRequestUseCase(
		pgDB, employeeRepo, leaveTypeRepo, leaveBalanceRepo, leaveRecordRepo, balanceTxRepo,
		leaveRequestRepo, approvalRequestRepo, approvalActionRepo, approvalFlowStepRepo, roleRepo,
		notificationService, userRepo, auditor,
	)
	rejectRequestUC := usecases.NewRejectRequestUseCase(
		pgDB, employeeRepo, leaveRequestRepo, approvalRequestRepo, approvalActionRepo, approvalFlowStepRepo,
		roleRepo, notificationService, leaveTypeRepo, userRepo, auditor,
	)
	submitPermissionUC := usecases.NewSubmitPermissionRequestUseCase(
		pgDB, userRepo, employeeRepo, departmentRepo, shiftRepo,
		permissionRequestRepo, approvalRequestRepo, approvalActionRepo, approvalFlowStepRepo,
		weekendRepo, roleRepo, notificationService, auditor,
	)
	updatePermissionUC := usecases.NewUpdatePermissionRequestUseCase(
		pgDB, permissionRequestRepo, approvalRequestRepo,
		employeeRepo, departmentRepo, shiftRepo, weekendRepo, auditor,
	)
	cancelPermissionUC := usecases.NewCancelPermissionRequestUseCase(
		pgDB, permissionRequestRepo, approvalRequestRepo, approvalActionRepo,
		employeeRepo, departmentRepo, shiftRepo, userRepo, notificationService, auditor,
	)
	listPermissionRequestsUC := usecases.NewListPermissionRequestsUseCase(
		pgDB, permissionRequestRepo, approvalRequestRepo, employeeRepo,
	)
	getPermissionRequestUC := usecases.NewGetPermissionRequestUseCase(
		pgDB, permissionRequestRepo, approvalRequestRepo, approvalFlowStepRepo, employeeRepo, roleRepo,
	)
	listPendingPermissionApprovalsUC := usecases.NewListPendingPermissionApprovalsUseCase(
		pgDB, permissionRequestRepo, approvalRequestRepo, approvalFlowStepRepo, employeeRepo, roleRepo,
	)
	approvePermissionUC := usecases.NewApprovePermissionRequestUseCase(
		pgDB, permissionRequestRepo, approvalRequestRepo, approvalActionRepo, approvalFlowStepRepo,
		employeeRepo, departmentRepo, shiftRepo, roleRepo, userRepo, notificationService, auditor,
	)
	rejectPermissionUC := usecases.NewRejectPermissionRequestUseCase(
		pgDB, permissionRequestRepo, approvalRequestRepo, approvalActionRepo, approvalFlowStepRepo,
		employeeRepo, roleRepo, userRepo, notificationService, auditor,
	)
	autoRejectExpiredPermissionsUC := usecases.NewAutoRejectExpiredPermissionRequestsUseCase(
		pgDB, permissionRequestRepo, approvalRequestRepo, approvalActionRepo,
		employeeRepo, departmentRepo, shiftRepo, cfg.HolidaySyncTimezone,
	)
	permissionEligibilityUC := usecases.NewGetPermissionEligibilityUseCase(
		pgDB, employeeRepo, departmentRepo, shiftRepo, permissionRequestRepo, weekendRepo, cfg.HolidaySyncTimezone,
	)

	getApprovalHistoryUC := usecases.NewGetApprovalHistoryUseCase(
		pgDB, approvalRequestRepo, approvalActionRepo, employeeRepo,
	)
	submitEmployeeProfileChangeRequestUC := usecases.NewSubmitEmployeeProfileChangeRequestUseCase(
		pgDB, employeeRepo, roleRepo, approvalRequestRepo, approvalActionRepo, approvalFlowStepRepo, profileChangeRequestRepo, auditor,
	)
	getEmployeeProfileChangeRequestUC := usecases.NewGetEmployeeProfileChangeRequestUseCase(
		pgDB, roleRepo, profileChangeRequestRepo, approvalRequestRepo, employeeRepo,
	)
	listEmployeeProfileChangeRequestsUC := usecases.NewListEmployeeProfileChangeRequestsUseCase(
		pgDB, roleRepo, profileChangeRequestRepo, getEmployeeProfileChangeRequestUC,
	)
	listPendingEmployeeProfileChangeRequestsUC := usecases.NewListPendingEmployeeProfileChangeRequestsUseCase(
		pgDB, profileChangeRequestRepo, roleRepo, getEmployeeProfileChangeRequestUC,
	)
	approveEmployeeProfileChangeRequestUC := usecases.NewApproveEmployeeProfileChangeRequestUseCase(
		pgDB, employeeRepo, profileChangeRequestRepo, approvalRequestRepo, approvalActionRepo, approvalFlowStepRepo, roleRepo, auditor,
	)
	rejectEmployeeProfileChangeRequestUC := usecases.NewRejectEmployeeProfileChangeRequestUseCase(
		pgDB, profileChangeRequestRepo, approvalRequestRepo, approvalActionRepo, approvalFlowStepRepo, roleRepo, auditor,
	)

	// Device token use cases
	registerDeviceTokenUC := usecases.NewRegisterDeviceTokenUseCase(deviceTokenRepo, pgDB)
	unregisterDeviceTokenUC := usecases.NewUnregisterDeviceTokenUseCase(deviceTokenRepo, pgDB)
	notifyMissingCheckOutsUC := usecases.NewNotifyMissingCheckOutsUseCase(
		pgDB, attendanceRecordRepo, attendanceReminderRepo, employeeRepo, departmentRepo, shiftRepo, holidayDefinitionRepo, leaveRecordRepo, userRepo, notificationService,
	)
	sendTestPushUC := usecases.NewSendTestPushNotificationUseCase(pgDB, userRepo, notificationService)

	listDepartmentAttendanceLogsUC := usecases.NewListDepartmentAttendanceLogsUseCase(pgDB, departmentRepo, attendanceRecordRepo)
	listEmployeeAttendanceLogsUC := usecases.NewListEmployeeAttendanceLogsUseCase(pgDB, employeeRepo, attendanceRecordRepo)
	listDailyDepartmentAttendanceLogsUC := usecases.NewListDailyDepartmentAttendanceLogsUseCase(pgDB, departmentRepo, attendanceRecordRepo, employeeRepo, shiftRepo, leaveRecordRepo, weekendRepo, holidayDefinitionRepo, permissionRequestRepo)
	listDailyEmployeeAttendanceLogsUC := usecases.NewListDailyEmployeeAttendanceLogsUseCase(pgDB, employeeRepo, attendanceRecordRepo, departmentRepo, shiftRepo, leaveRecordRepo, weekendRepo, holidayDefinitionRepo, permissionRequestRepo)
	getDailyAttendanceSummaryUC := usecases.NewGetDailyAttendanceSummaryUseCase(pgDB, attendanceRecordRepo, employeeRepo, departmentRepo, shiftRepo)
	departmentAttendanceReportUC := usecases.NewGetDepartmentAttendanceReportUseCase(pgDB, departmentRepo, attendanceRecordRepo, employeeRepo, shiftRepo, weekendRepo, holidayDefinitionRepo)
	exportDepartmentAttendanceReportUC := usecases.NewExportDepartmentAttendanceReportUseCase(departmentAttendanceReportUC)

	// Attendance device use cases
	registerAttendanceDeviceUC := usecases.NewRegisterAttendanceDeviceUseCase(pgDB, attendanceDeviceRepo, auditor)
	listAttendanceDevicesUC := usecases.NewListAttendanceDevicesUseCase(pgDB, attendanceDeviceRepo)
	getAttendanceDeviceUC := usecases.NewGetAttendanceDeviceUseCase(pgDB, attendanceDeviceRepo)
	updateAttendanceDeviceUC := usecases.NewUpdateAttendanceDeviceUseCase(pgDB, attendanceDeviceRepo, auditor)
	deleteAttendanceDeviceUC := usecases.NewDeleteAttendanceDeviceUseCase(pgDB, attendanceDeviceRepo, auditor)
	activateAttendanceDeviceUC := usecases.NewActivateAttendanceDeviceUseCase(pgDB, attendanceDeviceRepo, auditor)
	attendanceDeviceStatsUC := usecases.NewGetAttendanceDeviceStatsUseCase(pgDB, attendanceDeviceRepo)
	checkAttendanceDeviceConnectionUC := usecases.NewCheckAttendanceDeviceConnectionUseCase(pgDB, attendanceDeviceRepo)
	checkAllAttendanceDevicesConnectionUC := usecases.NewCheckAllAttendanceDevicesConnectionUseCase(pgDB, attendanceDeviceRepo)
	createAttendanceLogUC := usecases.NewCreateAttendanceLogUseCase(pgDB, attendanceRecordRepo, employeeRepo, attendanceDeviceRepo, auditor)
	updateAttendanceLogUC := usecases.NewUpdateAttendanceLogUseCase(pgDB, attendanceRecordRepo, employeeRepo, attendanceDeviceRepo, auditor)
	importAttendanceLogsUC := usecases.NewImportAttendanceLogsUseCase(pgDB, attendanceRecordRepo, employeeRepo, attendanceDeviceRepo, auditor)
	getAttendanceLogHistoryUC := usecases.NewGetAttendanceLogHistoryUseCase(pgDB, attendanceRecordRepo, auditLogRepo, employeeRepo, userRepo)
	getMonthlyAttendanceStatsUC := usecases.NewGetMonthlyAttendanceStatsUseCase(
		pgDB,
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
	orgChartUC := usecases.NewOrgChartUseCase(pgDB, employeeRepo, departmentRepo, roleRepo)

	// Scheduler use cases
	autoRejectExpiredUC := usecases.NewAutoRejectExpiredRequestsUseCase(
		pgDB, leaveRequestRepo, approvalRequestRepo, approvalActionRepo, cfg.ExpiredLeaveGraceDays,
	)
	holidaySyncUC := usecases.NewSyncEgyptPublicHolidaysUseCase(
		pgDB,
		holidayDefinitionRepo,
		cfg.HolidaySyncEndpoint,
		cfg.HolidaySyncAPIKey,
		cfg.HolidaySyncCountry,
		cfg.HolidaySyncTimezone,
	)
	absenceSyncUC := usecases.NewSyncDailyAbsencesUseCase(
		pgDB,
		leaveRecordRepo,
		weekendRepo,
		holidayDefinitionRepo,
		cfg.HolidaySyncTimezone,
	)

	leaveHandler := httpAdapter.NewLeaveHandler(recordLeaveUC, getBalanceUC, listLeaveRecordsUC, listAllLeaveRecordsUC, i18nService)
	employeeHandler := httpAdapter.NewEmployeeHandler(
		createEmployeeUC,
		getEmployeeUC,
		listEmployeesUC,
		updateOwnEmployeeProfileUC,
		importEmployeesUC,
		exportEmployeesUC,
		exportEmployeesPDFUC,
		generateTemplateUC,
		assignEmployeeDepartmentUC,
		removeEmployeeDepartmentUC,
		listEmployeePenaltiesUC,
		createEmployeePenaltyUC,
		createEmployeePenaltyRemovalUC,
		updateEmployeePenaltiesUC,
		listEmployeeIncentiveBonusesUC,
		createEmployeeIncentiveBonusUC,
		updateEmployeeIncentiveBonusesUC,
		listEmployeeAnnualReportsUC,
		createEmployeeAnnualReportUC,
		updateEmployeeAnnualReportsUC,
		setEmployeeManagerUC,
		listManagerCandidatesUC,
	)
	authHandler := httpAdapter.NewAuthHandler(requestOTPUC, verifyOTPUC, loginPasswordUC, refreshTokenUC, logoutUC, getCurrentUserUC, i18nService)
	meHandler := httpAdapter.NewMeHandler(updateUserUC)

	userHandler := httpAdapter.NewUserHandler(listUsersUC, createUserUC, updateUserUC, assignRoleUC, removeRoleUC)
	roleHandler := httpAdapter.NewRoleHandler(listRolesUC, createRoleUC, setPermissionsUC, setRoleScopeUC, listPermissionsUC)
	dashboardHandler := httpAdapter.NewDashboardHandler(getDashboardStatsUC)
	approvalFlowHandler := httpAdapter.NewApprovalFlowHandler(
		listApprovalFlowsUC, getApprovalFlowUC, createApprovalFlowUC, updateApprovalFlowUC, listApprovalFlowStepsUC,
		createApprovalFlowStepUC, updateApprovalFlowStepUC, deleteApprovalFlowStepUC,
	)
	leaveRequestHandler := httpAdapter.NewLeaveRequestHandler(
		submitLeaveRequestUC, updateRejectedLeaveRequestUC, cancelLeaveRequestUC, listLeaveRequestsUC, getLeaveRequestUC,
		getEmployeeUC, listPendingApprovalsUC, approveRequestUC, rejectRequestUC, getApprovalHistoryUC, getCurrentUserUC, i18nService,
	)
	permissionRequestHandler := httpAdapter.NewPermissionRequestHandler(
		submitPermissionUC, updatePermissionUC, cancelPermissionUC,
		listPermissionRequestsUC, getPermissionRequestUC,
		listPendingPermissionApprovalsUC, approvePermissionUC, rejectPermissionUC,
		getApprovalHistoryUC, permissionEligibilityUC, getCurrentUserUC,
	)
	employeeProfileChangeHandler := httpAdapter.NewEmployeeProfileChangeHandler(
		submitEmployeeProfileChangeRequestUC,
		listEmployeeProfileChangeRequestsUC,
		getEmployeeProfileChangeRequestUC,
		listPendingEmployeeProfileChangeRequestsUC,
		approveEmployeeProfileChangeRequestUC,
		rejectEmployeeProfileChangeRequestUC,
		getApprovalHistoryUC,
		updateEmployeeProfileUC,
		getCurrentUserUC,
	)
	departmentHandler := httpAdapter.NewDepartmentHandler(
		listDepartmentsUC, getDepartmentUC, createDepartmentUC, updateDepartmentUC,
		assignDepartmentManagerUC, removeDepartmentManagerUC, i18nService, orgChartUC,
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
	leaveTypeHandler := httpAdapter.NewLeaveTypeHandler(getLeaveTypeDetailsUC, listLeaveTypesUC, listSubLeaveTypesUC, updateLeaveTypeUC, setLeaveTypeApprovalFlowUC, toggleLeaveTypeUC, i18nService)
	weekendHandler := httpAdapter.NewWeekendHandler(listWeekendDaysUC)
	holidayHandler := httpAdapter.NewHolidayHandler(pgDB, listHolidaysUC, listWeekendDaysUC, createManualHolidayUC, updateHolidayUC, deleteHolidayUC, employeeRepo)
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
    getAuditTrailUC := usecases.NewGetAuditTrailUseCase(pgDB, auditLogRepo, employeeRepo, i18nService)
	getActorAuditEventsUC := usecases.NewGetActorAuditEventsUseCase(pgDB, auditLogRepo, i18nService)
	listAllAuditLogsUC := usecases.NewListAllAuditLogsUseCase(auditLogRepo, pgDB, i18nService)
	listEnrichedAuditLogsUC := usecases.NewListEnrichedAuditLogsUseCase(
		auditLogRepo, employeeRepo, departmentRepo, leaveRequestRepo, leaveTypeRepo, attendanceRecordRepo, pgDB, i18nService,
	)
	getAuditFilterOptionsUC := usecases.NewGetAuditFilterOptionsUseCase(auditLogRepo, pgDB, i18nService)
	auditHandler := httpAdapter.NewAuditHandler(getAuditTrailUC, getActorAuditEventsUC, listAllAuditLogsUC, listEnrichedAuditLogsUC, getAuditFilterOptionsUC)
	documentHandler := httpAdapter.NewDocumentHandler(
		generateDocumentUploadURLUC,
		generateDocumentDownloadURLUC,
	)

	router := httpAdapter.NewRouter(httpAdapter.RouterConfig{
		LeaveHandler:                 leaveHandler,
		EmployeeHandler:              employeeHandler,
		AuthHandler:                  authHandler,
		UserHandler:                  userHandler,
		RoleHandler:                  roleHandler,
		DashboardHandler:             dashboardHandler,
		ApprovalFlowHandler:          approvalFlowHandler,
		LeaveRequestHandler:          leaveRequestHandler,
		PermissionRequestHandler:     permissionRequestHandler,
		MeHandler:                    meHandler,
		DepartmentHandler:            departmentHandler,
		DeviceTokenHandler:           deviceTokenHandler,
		AttendanceDeviceHandler:      attendanceDeviceHandler,
		LeaveTypeHandler:             leaveTypeHandler,
		WeekendHandler:               weekendHandler,
		HolidayHandler:               holidayHandler,
		ShiftHandler:                 shiftHandler,
		AttendanceHandler:            attendanceHandler,
		AuditHandler:                 auditHandler,
		DebugHandler:                 debugHandler,
		JWTService:                   jwtService,
		AuthEnabled:                  cfg.AuthEnabled,
		EmployeeProfileChangeHandler: employeeProfileChangeHandler,
		DocumentHandler:              documentHandler,
		I18nService:                  i18nService,
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

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
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

	if _, err := sqlDB.Exec(`UPDATE schema_migrations SET version = $1`, targetVersion); err != nil {
		return err
	}

	slog.Info("main.main.reconciled_migration_version", "from", currentVersion, "to", targetVersion)
	return nil
}

func readSchemaMigrationVersion(sqlDB *sql.DB) (version uint64, dirty bool, hasVersion bool, err error) {
	var count int
	if err = sqlDB.QueryRow(`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'schema_migrations'`).Scan(&count); err != nil {
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

func runSeedData(sqlDB *sql.DB) error {
	// Read the seed file
	seedPath := "./adapters/db/seed/seed_dev_consolidated.sql"
	seedSQL, err := os.ReadFile(seedPath)
	if err != nil {
		return fmt.Errorf("failed to read seed file: %w", err)
	}

	// Execute the seed SQL
	if _, err := sqlDB.Exec(string(seedSQL)); err != nil {
		return fmt.Errorf("failed to execute seed data: %w", err)
	}

	return nil
}
