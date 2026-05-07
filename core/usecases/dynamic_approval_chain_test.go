package usecases_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	dbadapter "github.com/banumusa/backend/adapters/db"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type notificationRecorder struct {
	sentToUserCount  int
	sentToUsersCount int
	lastUsers        []string
}

func (n *notificationRecorder) SendToUser(userUID, title, body string, data ports.NotificationData) (int, error) {
	n.sentToUserCount++
	return 1, nil
}

func (n *notificationRecorder) SendToUsers(userUIDs []string, title, body string, data ports.NotificationData) (int, error) {
	n.sentToUsersCount++
	n.lastUsers = append([]string(nil), userUIDs...)
	return len(userUIDs), nil
}

type approvalChainTestEnv struct {
	db                  *dbadapter.SQLiteDB
	submitUC            *usecases.SubmitLeaveRequestUseCase
	approveUC           *usecases.ApproveRequestUseCase
	listPendingUC       *usecases.ListPendingApprovalsUseCase
	notificationService *notificationRecorder
	roleRepo            *dbadapter.RoleRepository
	userRepo            *dbadapter.UserRepository
	approvalRequestRepo *dbadapter.ApprovalRequestRepository
	leaveRequestRepo    *dbadapter.LeaveRequestRepository
	leaveRecordRepo     *dbadapter.LeaveRecordRepository
	actionRepo          *dbadapter.ApprovalActionRepository
}

func newApprovalChainTestEnv(t *testing.T) *approvalChainTestEnv {
	t.Helper()

	db := newDynamicApprovalSQLiteDB(t)
	notificationService := &notificationRecorder{}

	userRepo := dbadapter.NewUserRepository()
	employeeRepo := dbadapter.NewEmployeeRepository()
	roleRepo := dbadapter.NewRoleRepository()
	leaveTypeRepo := dbadapter.NewLeaveTypeRepository()
	leaveBalanceRepo := dbadapter.NewLeaveBalanceRepository()
	leaveRequestRepo := dbadapter.NewLeaveRequestRepository()
	leaveRecordRepo := dbadapter.NewLeaveRecordRepository()
	balanceTxRepo := dbadapter.NewLeaveBalanceTransactionRepository()
	approvalRequestRepo := dbadapter.NewApprovalRequestRepository()
	actionRepo := dbadapter.NewApprovalActionRepository()
	stepRepo := dbadapter.NewApprovalFlowStepRepository()
	weekendRepo := dbadapter.NewWeekendConfigRepository()
	holidayRepo := dbadapter.NewHolidayDefinitionRepository()
	workingDaysCalc := usecases.NewWorkingDaysCalculator(weekendRepo, holidayRepo)

	submitUC := usecases.NewSubmitLeaveRequestUseCase(
		db,
		userRepo,
		employeeRepo,
		leaveTypeRepo,
		leaveBalanceRepo,
		leaveRequestRepo,
		leaveRecordRepo,
		balanceTxRepo,
		approvalRequestRepo,
		actionRepo,
		stepRepo,
		dbadapter.NewLeaveRequestDocumentRepository(),
		workingDaysCalc,
		nil,
		notificationService,
		roleRepo,
		nil,
	)

	approveUC := usecases.NewApproveRequestUseCase(
		db,
		employeeRepo,
		leaveTypeRepo,
		leaveBalanceRepo,
		leaveRecordRepo,
		balanceTxRepo,
		leaveRequestRepo,
		approvalRequestRepo,
		actionRepo,
		stepRepo,
		roleRepo,
		notificationService,
		userRepo,
		nil,
	)

	listPendingUC := usecases.NewListPendingApprovalsUseCase(
		db,
		userRepo,
		employeeRepo,
		leaveTypeRepo,
		leaveRequestRepo,
		approvalRequestRepo,
		stepRepo,
		roleRepo,
	)

	return &approvalChainTestEnv{
		db:                  db,
		submitUC:            submitUC,
		approveUC:           approveUC,
		listPendingUC:       listPendingUC,
		notificationService: notificationService,
		roleRepo:            roleRepo,
		userRepo:            userRepo,
		approvalRequestRepo: approvalRequestRepo,
		leaveRequestRepo:    leaveRequestRepo,
		leaveRecordRepo:     leaveRecordRepo,
		actionRepo:          actionRepo,
	}
}

func (e *approvalChainTestEnv) close() {
	if err := e.db.Close(); err != nil {
		panic(err)
	}
}

func newDynamicApprovalSQLiteDB(t *testing.T) *dbadapter.SQLiteDB {
	t.Helper()

	dsn := fmt.Sprintf("file:dynamic_approval_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := dbadapter.NewSQLiteDB(dsn)
	if err != nil {
		t.Fatalf("failed to create sqlite db: %v", err)
	}

	migrationsPath := findMigrationsDir(t)
	driver, err := sqlite.WithInstance(db.DB(), &sqlite.Config{})
	if err != nil {
		t.Fatalf("failed to create migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"sqlite",
		driver,
	)
	if err != nil {
		t.Fatalf("failed to create migrate instance: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("failed to run migrations: %v", err)
	}

	clearDynamicApprovalSeedData(t, db)
	return db
}

func clearDynamicApprovalSeedData(t *testing.T, db *dbadapter.SQLiteDB) {
	t.Helper()

	tables := []string{
		"leave_request_documents",
		"sub_leave_types",
		"leave_requests",
		"approval_actions",
		"approval_requests",
		"leave_balance_transactions",
		"leave_records",
		"leave_balances",
		"leave_types",
		"approval_flow_steps",
		"approval_flows",
		"user_roles",
		"role_permissions",
		"refresh_tokens",
		"otp_codes",
		"holiday_definitions",
		"weekend_config",
		"employees",
		"users",
		"roles",
		"permissions",
		"departments",
	}

	for _, table := range tables {
		if _, err := db.ExecContext(context.Background(), fmt.Sprintf("DELETE FROM %s", table)); err != nil {
			t.Fatalf("failed to clear table %s: %v", table, err)
		}
	}
}

func findMigrationsDir(t *testing.T) string {
	t.Helper()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	dir := cwd
	for range 6 {
		candidate := filepath.Join(dir, "adapters", "db", "migrations")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		dir = filepath.Dir(dir)
	}

	t.Fatal("could not find migrations directory")
	return ""
}

func mustExec(t *testing.T, db *dbadapter.SQLiteDB, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), query, args...); err != nil {
		t.Fatalf("failed to exec query: %v", err)
	}
}

func seedDepartment(t *testing.T, env *approvalChainTestEnv, uid, code string) {
	t.Helper()
	mustExec(t, env.db,
		`INSERT INTO departments (uid, code, name_en, name_ar, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, 1, datetime('now'), datetime('now'))`,
		uid, code, code, code,
	)
}

func seedRole(t *testing.T, env *approvalChainTestEnv, uid, name string) int64 {
	t.Helper()
	mustExec(t, env.db,
		`INSERT INTO roles (uid, name, description, is_system, created_at, updated_at) VALUES (?, ?, ?, 0, datetime('now'), datetime('now'))`,
		uid, name, name,
	)

	var id int64
	if err := env.db.QueryRowContext(context.Background(), `SELECT id FROM roles WHERE uid = ?`, uid).Scan(&id); err != nil {
		t.Fatalf("failed to load role id for %s: %v", uid, err)
	}
	return id
}

func seedEmployeeAndUser(t *testing.T, env *approvalChainTestEnv, employeeUID, userUID, phone, departmentUID string) (*domain.Employee, *domain.User) {
	t.Helper()

	employee := domain.NewEmployee(userUID, phone, "gov_"+employeeUID, "uni_"+employeeUID, time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC))
	employee.UID = employeeUID
	employee.DepartmentUID = &departmentUID

	if err := dbadapter.NewEmployeeRepository().Create(context.Background(), env.db, employee); err != nil {
		t.Fatalf("failed to seed employee %s: %v", employeeUID, err)
	}

	user := domain.NewUser(phone)
	user.UID = userUID
	user.EmployeeUID = &employeeUID
	if err := env.userRepo.Create(context.Background(), env.db, user); err != nil {
		t.Fatalf("failed to seed user %s: %v", userUID, err)
	}

	return employee, user
}

func assignRole(t *testing.T, env *approvalChainTestEnv, userID, roleID int64, departmentUID *string) {
	t.Helper()
	if err := env.roleRepo.AssignRoleToUserWithDepartment(context.Background(), env.db, userID, roleID, departmentUID); err != nil {
		t.Fatalf("failed to assign role: %v", err)
	}
}

func seedApprovalFlowWithSteps(t *testing.T, env *approvalChainTestEnv, flowUID string, rolesByStep []string) {
	t.Helper()
	mustExec(t, env.db,
		`INSERT INTO approval_flows (uid, code, name_en, name_ar, description, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 1, datetime('now'), datetime('now'))`,
		flowUID, flowUID, flowUID, flowUID, flowUID,
	)

	for idx, roleUID := range rolesByStep {
		mustExec(t, env.db,
			`INSERT INTO approval_flow_steps (uid, approval_flow_uid, step_order, role_uid, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))`,
			flowUID+"_step_"+roleUID, flowUID, idx+1, roleUID,
		)
	}
}

func seedLeaveType(t *testing.T, env *approvalChainTestEnv, uid, code, flowUID string) {
	t.Helper()
	mustExec(t, env.db,
		`INSERT INTO leave_types (uid, code, name_en, name_ar, default_balance, max_consecutive, recording_deadline_days, advance_notice_days, is_active, approval_flow_uid, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, NULL, NULL, NULL, 1, ?, datetime('now'), datetime('now'))`,
		uid, code, code, code, 30, flowUID,
	)
}

func TestDynamicApprovalChain_SubmitSkipsRequesterStepAndShowsOnlyHigherApprover(t *testing.T) {
	env := newApprovalChainTestEnv(t)
	defer env.close()

	seedDepartment(t, env, "dept_math", "MATH")
	step1RoleID := seedRole(t, env, "role_department_manager", "Department Manager")
	step2RoleID := seedRole(t, env, "role_dean", "Dean")
	step3RoleID := seedRole(t, env, "role_uhr", "University Human Resources")

	_, requesterUser := seedEmployeeAndUser(t, env, "emp_requester", "usr_requester", "+201111111111", "dept_math")
	_, approver1User := seedEmployeeAndUser(t, env, "emp_dean", "usr_dean", "+201111111112", "dept_math")
	_, approver2User := seedEmployeeAndUser(t, env, "emp_uhr", "usr_uhr", "+201111111113", "dept_math")

	dept := "dept_math"
	assignRole(t, env, requesterUser.ID, step1RoleID, &dept)
	assignRole(t, env, approver1User.ID, step2RoleID, &dept)
	assignRole(t, env, approver2User.ID, step3RoleID, &dept)

	seedApprovalFlowWithSteps(t, env, "apf_dynamic_three", []string{"role_department_manager", "role_dean", "role_uhr"})
	seedLeaveType(t, env, "ltype_dynamic_three", "ANNUAL_DYNAMIC_THREE", "apf_dynamic_three")

	output, err := env.submitUC.Execute(context.Background(), usecases.SubmitLeaveRequestInput{
		UserUID:      requesterUser.UID,
		EmployeeUID:  "emp_requester",
		LeaveTypeUID: "ltype_dynamic_three",
		StartDate:    time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("submit leave request failed: %v", err)
	}

	if output.ApprovalRequest.CurrentStep != 2 {
		t.Fatalf("expected current step 2, got %d", output.ApprovalRequest.CurrentStep)
	}
	if output.ApprovalRequest.MaxStep != 3 {
		t.Fatalf("expected max step 3, got %d", output.ApprovalRequest.MaxStep)
	}
	if output.ApprovalRequest.Status != domain.ApprovalRequestStatusPending {
		t.Fatalf("expected pending approval request, got %s", output.ApprovalRequest.Status)
	}

	pending, err := env.listPendingUC.Execute(context.Background(), approver1User.ID)
	if err != nil {
		t.Fatalf("list pending approvals failed: %v", err)
	}
	if len(pending.Items) != 1 {
		t.Fatalf("expected 1 pending item for step-1 approver, got %d", len(pending.Items))
	}
	if pending.Items[0].ApprovalRequest.UID != output.ApprovalRequest.UID {
		t.Fatalf("unexpected pending approval uid %s", pending.Items[0].ApprovalRequest.UID)
	}

	pendingTop, err := env.listPendingUC.Execute(context.Background(), approver2User.ID)
	if err != nil {
		t.Fatalf("list pending approvals for top approver failed: %v", err)
	}
	if len(pendingTop.Items) != 0 {
		t.Fatalf("expected 0 pending items for top approver at this stage, got %d", len(pendingTop.Items))
	}
}

func TestDynamicApprovalChain_SubmitTopRoleAutoApprovesAndCreatesLeaveRecord(t *testing.T) {
	env := newApprovalChainTestEnv(t)
	defer env.close()

	seedDepartment(t, env, "dept_science", "SCI")
	step1RoleID := seedRole(t, env, "role_dean", "Dean")
	_, requesterUser := seedEmployeeAndUser(t, env, "emp_top_requester", "usr_top_requester", "+201122222221", "dept_science")
	dept := "dept_science"
	assignRole(t, env, requesterUser.ID, step1RoleID, &dept)

	seedApprovalFlowWithSteps(t, env, "apf_top_role", []string{"role_dean"})
	seedLeaveType(t, env, "ltype_top_role", "ANNUAL_TOP_ROLE", "apf_top_role")

	output, err := env.submitUC.Execute(context.Background(), usecases.SubmitLeaveRequestInput{
		UserUID:      requesterUser.UID,
		EmployeeUID:  "emp_top_requester",
		LeaveTypeUID: "ltype_top_role",
		StartDate:    time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("submit leave request failed: %v", err)
	}

	if output.ApprovalRequest.Status != domain.ApprovalRequestStatusApproved {
		t.Fatalf("expected approved approval request, got %s", output.ApprovalRequest.Status)
	}
	if output.ApprovalRequest.CurrentStep != 0 || output.ApprovalRequest.MaxStep != 0 {
		t.Fatalf("expected current/max step 0/0, got %d/%d", output.ApprovalRequest.CurrentStep, output.ApprovalRequest.MaxStep)
	}
	if output.LeaveRecord == nil {
		t.Fatal("expected leave record to be created for top-role requester")
	}
	if output.LeaveRequest.DecidedAt == nil {
		t.Fatal("expected leave request decided_at to be set")
	}
	if env.notificationService.sentToUsersCount != 0 {
		t.Fatalf("expected no approver notifications, got %d", env.notificationService.sentToUsersCount)
	}

	actions, err := env.actionRepo.ListByRequest(context.Background(), env.db, output.ApprovalRequest.UID)
	if err != nil {
		t.Fatalf("list approval actions failed: %v", err)
	}
	if len(actions) != 1 || actions[0].Action != domain.ApprovalActionTypeSubmit {
		t.Fatalf("expected only one submit action, got %+v", actions)
	}
	if actions[0].ActorUID != "emp_top_requester" {
		t.Fatalf("expected submit action actor_uid emp_top_requester, got %s", actions[0].ActorUID)
	}
}

func TestDynamicApprovalChain_SubmitOnBehalfUsesSubmittingEmployeeAsActor(t *testing.T) {
	env := newApprovalChainTestEnv(t)
	defer env.close()

	seedDepartment(t, env, "dept_eng", "ENG")
	step1RoleID := seedRole(t, env, "role_department_manager", "Department Manager")
	step2RoleID := seedRole(t, env, "role_dean", "Dean")

	_, managerUser := seedEmployeeAndUser(t, env, "emp_delegate_manager", "usr_delegate_manager", "+201155555551", "dept_eng")
	seedEmployeeAndUser(t, env, "emp_delegate_requester", "usr_delegate_requester", "+201155555552", "dept_eng")
	_, approverUser := seedEmployeeAndUser(t, env, "emp_delegate_dean", "usr_delegate_dean", "+201155555553", "dept_eng")

	dept := "dept_eng"
	assignRole(t, env, managerUser.ID, step1RoleID, &dept)
	assignRole(t, env, approverUser.ID, step2RoleID, &dept)

	seedApprovalFlowWithSteps(t, env, "apf_delegate_submit", []string{"role_department_manager", "role_dean"})
	seedLeaveType(t, env, "ltype_delegate_submit", "ANNUAL_DELEGATE", "apf_delegate_submit")

	output, err := env.submitUC.Execute(context.Background(), usecases.SubmitLeaveRequestInput{
		UserUID:          managerUser.UID,
		EmployeeUID:      "emp_delegate_requester",
		ActorEmployeeUID: managerUser.EmployeeUID,
		LeaveTypeUID:     "ltype_delegate_submit",
		StartDate:        time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
		EndDate:          time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("submit leave request failed: %v", err)
	}

	actions, err := env.actionRepo.ListByRequest(context.Background(), env.db, output.ApprovalRequest.UID)
	if err != nil {
		t.Fatalf("list approval actions failed: %v", err)
	}
	if len(actions) != 1 || actions[0].Action != domain.ApprovalActionTypeSubmit {
		t.Fatalf("expected only one submit action, got %+v", actions)
	}
	if actions[0].ActorUID != "emp_delegate_manager" {
		t.Fatalf("expected submit action actor_uid emp_delegate_manager, got %s", actions[0].ActorUID)
	}
	if output.ApprovalRequest.CurrentStep != 1 {
		t.Fatalf("expected delegated submission for non-manager employee to stay at step 1, got %d", output.ApprovalRequest.CurrentStep)
	}
}

func TestDynamicApprovalChain_AdminSubmittingForDepartmentManagerSkipsManagerStep(t *testing.T) {
	env := newApprovalChainTestEnv(t)
	defer env.close()

	seedDepartment(t, env, "dept_law", "LAW")
	step1RoleID := seedRole(t, env, "role_department_manager", "Department Manager")
	step2RoleID := seedRole(t, env, "role_dean", "Dean")

	_, adminUser := seedEmployeeAndUser(t, env, "emp_delegate_admin", "usr_delegate_admin", "+201166666661", "dept_law")
	managerEmployee, managerUser := seedEmployeeAndUser(t, env, "emp_delegate_target_manager", "usr_delegate_target_manager", "+201166666662", "dept_law")
	_, deanUser := seedEmployeeAndUser(t, env, "emp_delegate_law_dean", "usr_delegate_law_dean", "+201166666663", "dept_law")

	dept := "dept_law"
	assignRole(t, env, managerUser.ID, step1RoleID, &dept)
	assignRole(t, env, deanUser.ID, step2RoleID, &dept)

	seedApprovalFlowWithSteps(t, env, "apf_admin_for_manager", []string{"role_department_manager", "role_dean"})
	seedLeaveType(t, env, "ltype_admin_for_manager", "ANNUAL_ADMIN_FOR_MANAGER", "apf_admin_for_manager")

	output, err := env.submitUC.Execute(context.Background(), usecases.SubmitLeaveRequestInput{
		UserUID:          adminUser.UID,
		EmployeeUID:      managerEmployee.UID,
		ActorEmployeeUID: adminUser.EmployeeUID,
		LeaveTypeUID:     "ltype_admin_for_manager",
		StartDate:        time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
		EndDate:          time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("submit leave request failed: %v", err)
	}

	if output.ApprovalRequest.CurrentStep != 2 {
		t.Fatalf("expected department manager step to be skipped, got current step %d", output.ApprovalRequest.CurrentStep)
	}

	pendingForDean, err := env.listPendingUC.Execute(context.Background(), deanUser.ID)
	if err != nil {
		t.Fatalf("list pending approvals for dean failed: %v", err)
	}
	if len(pendingForDean.Items) != 1 || pendingForDean.Items[0].ApprovalRequest.UID != output.ApprovalRequest.UID {
		t.Fatalf("expected dean to receive the delegated manager request, got %+v", pendingForDean.Items)
	}
}

func TestDynamicApprovalChain_ApproveUsesOnlyHigherSteps(t *testing.T) {
	env := newApprovalChainTestEnv(t)
	defer env.close()

	seedDepartment(t, env, "dept_eng", "ENG")
	step1RoleID := seedRole(t, env, "role_department_manager", "Department Manager")
	step2RoleID := seedRole(t, env, "role_dean", "Dean")
	step3RoleID := seedRole(t, env, "role_uhr", "University Human Resources")

	_, requesterUser := seedEmployeeAndUser(t, env, "emp_chain_requester", "usr_chain_requester", "+201133333331", "dept_eng")
	_, approver1User := seedEmployeeAndUser(t, env, "emp_chain_dean", "usr_chain_dean", "+201133333332", "dept_eng")
	_, approver2User := seedEmployeeAndUser(t, env, "emp_chain_hr", "usr_chain_hr", "+201133333333", "dept_eng")

	dept := "dept_eng"
	assignRole(t, env, requesterUser.ID, step1RoleID, &dept)
	assignRole(t, env, approver1User.ID, step2RoleID, &dept)
	assignRole(t, env, approver2User.ID, step3RoleID, &dept)

	seedApprovalFlowWithSteps(t, env, "apf_dynamic_three", []string{"role_department_manager", "role_dean", "role_uhr"})
	seedLeaveType(t, env, "ltype_dynamic_three", "ANNUAL_DYNAMIC_THREE", "apf_dynamic_three")

	submitOutput, err := env.submitUC.Execute(context.Background(), usecases.SubmitLeaveRequestInput{
		UserUID:      requesterUser.UID,
		EmployeeUID:  "emp_chain_requester",
		LeaveTypeUID: "ltype_dynamic_three",
		StartDate:    time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("submit leave request failed: %v", err)
	}
	if submitOutput.ApprovalRequest.CurrentStep != 2 || submitOutput.ApprovalRequest.MaxStep != 3 {
		t.Fatalf("expected current/max step 2/3, got %d/%d", submitOutput.ApprovalRequest.CurrentStep, submitOutput.ApprovalRequest.MaxStep)
	}

	firstApprove, err := env.approveUC.Execute(context.Background(), usecases.ApproveRequestInput{
		ApprovalRequestUID: submitOutput.ApprovalRequest.UID,
		ActorUserID:        approver1User.ID,
		ActorEmployeeUID:   "emp_chain_dean",
	})
	if err != nil {
		t.Fatalf("first approval failed: %v", err)
	}
	if firstApprove.ApprovalRequest.Status != domain.ApprovalRequestStatusPending || firstApprove.ApprovalRequest.CurrentStep != 3 {
		t.Fatalf("expected pending request at raw step 3, got status=%s step=%d", firstApprove.ApprovalRequest.Status, firstApprove.ApprovalRequest.CurrentStep)
	}
	if firstApprove.LeaveRecord != nil {
		t.Fatal("did not expect leave record after first approval")
	}

	pendingForStep2, err := env.listPendingUC.Execute(context.Background(), approver2User.ID)
	if err != nil {
		t.Fatalf("list pending approvals for step 2 approver failed: %v", err)
	}
	if len(pendingForStep2.Items) != 1 || pendingForStep2.Items[0].ApprovalRequest.UID != submitOutput.ApprovalRequest.UID {
		t.Fatalf("expected request to move to second approver, got %+v", pendingForStep2.Items)
	}

	secondApprove, err := env.approveUC.Execute(context.Background(), usecases.ApproveRequestInput{
		ApprovalRequestUID: submitOutput.ApprovalRequest.UID,
		ActorUserID:        approver2User.ID,
		ActorEmployeeUID:   "emp_chain_hr",
	})
	if err != nil {
		t.Fatalf("second approval failed: %v", err)
	}
	if secondApprove.ApprovalRequest.Status != domain.ApprovalRequestStatusApproved {
		t.Fatalf("expected approved status after second approval, got %s", secondApprove.ApprovalRequest.Status)
	}
	if secondApprove.LeaveRecord == nil {
		t.Fatal("expected leave record after final approval")
	}
}

func TestDynamicApprovalChain_DepartmentScopedRequesterRoleMustMatchOwnDepartment(t *testing.T) {
	env := newApprovalChainTestEnv(t)
	defer env.close()

	seedDepartment(t, env, "dept_a", "A")
	seedDepartment(t, env, "dept_b", "B")
	step1RoleID := seedRole(t, env, "role_department_manager", "Department Manager")
	step2RoleID := seedRole(t, env, "role_dean", "Dean")
	step3RoleID := seedRole(t, env, "role_uhr", "University Human Resources")

	_, requesterUser := seedEmployeeAndUser(t, env, "emp_scoped_requester", "usr_scoped_requester", "+201144444441", "dept_a")
	deptB := "dept_b"
	assignRole(t, env, requesterUser.ID, step1RoleID, &deptB)
	_, _ = step2RoleID, step3RoleID

	seedApprovalFlowWithSteps(t, env, "apf_scoped", []string{"role_department_manager", "role_dean", "role_uhr"})
	seedLeaveType(t, env, "ltype_scoped", "ANNUAL_SCOPED", "apf_scoped")

	output, err := env.submitUC.Execute(context.Background(), usecases.SubmitLeaveRequestInput{
		UserUID:      requesterUser.UID,
		EmployeeUID:  "emp_scoped_requester",
		LeaveTypeUID: "ltype_scoped",
		StartDate:    time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("submit leave request failed: %v", err)
	}

	if output.ApprovalRequest.CurrentStep != 1 || output.ApprovalRequest.MaxStep != 3 {
		t.Fatalf("expected full flow to remain when requester role scope does not match department, got %d/%d", output.ApprovalRequest.CurrentStep, output.ApprovalRequest.MaxStep)
	}
}
