package usecases

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

const permissionDefaultFlowUID = "apf_permission_default"

type SubmitPermissionRequestInput struct {
	UserUID        string
	EmployeeUID    string
	Type           domain.PermissionType
	PermissionDate time.Time
	StartTime      *string
	EndTime        *string
	Reason         *string
}

type SubmitPermissionRequestOutput struct {
	PermissionRequest *domain.PermissionRequest
	ApprovalRequest   *domain.ApprovalRequest
}

type SubmitPermissionRequestUseCase struct {
	db                   ports.DB
	userRepo             ports.UserRepository
	employeeRepo         ports.EmployeeRepository
	deptRepo             ports.DepartmentRepository
	shiftRepo            ports.ShiftRepository
	permissionRepo       ports.PermissionRequestRepository
	approvalRequestRepo  ports.ApprovalRequestRepository
	approvalActionRepo   ports.ApprovalActionRepository
	approvalFlowStepRepo ports.ApprovalFlowStepRepository
	weekendRepo          ports.WeekendConfigRepository
	roleRepo             ports.RoleRepository
	notificationService  ports.NotificationService
	auditor              audit.Auditor
	now                  func() time.Time
}

func NewSubmitPermissionRequestUseCase(
	db ports.DB,
	userRepo ports.UserRepository,
	employeeRepo ports.EmployeeRepository,
	deptRepo ports.DepartmentRepository,
	shiftRepo ports.ShiftRepository,
	permissionRepo ports.PermissionRequestRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	approvalActionRepo ports.ApprovalActionRepository,
	approvalFlowStepRepo ports.ApprovalFlowStepRepository,
	weekendRepo ports.WeekendConfigRepository,
	roleRepo ports.RoleRepository,
	notificationService ports.NotificationService,
	auditor audit.Auditor,
) *SubmitPermissionRequestUseCase {
	return &SubmitPermissionRequestUseCase{
		db:                   db,
		userRepo:             userRepo,
		employeeRepo:         employeeRepo,
		deptRepo:             deptRepo,
		shiftRepo:            shiftRepo,
		permissionRepo:       permissionRepo,
		approvalRequestRepo:  approvalRequestRepo,
		approvalActionRepo:   approvalActionRepo,
		approvalFlowStepRepo: approvalFlowStepRepo,
		weekendRepo:          weekendRepo,
		roleRepo:             roleRepo,
		notificationService:  notificationService,
		auditor:              auditor,
		now:                  time.Now,
	}
}

func (uc *SubmitPermissionRequestUseCase) Execute(ctx context.Context, input SubmitPermissionRequestInput) (*SubmitPermissionRequestOutput, error) {
	if !input.Type.IsValid() {
		return nil, ErrPermissionTypeInvalid
	}

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	employee, err := uc.employeeRepo.GetByUID(ctx, tx, input.EmployeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	shift, err := resolveEffectiveShift(ctx, uc.db, uc.employeeRepo, uc.deptRepo, uc.shiftRepo, employee.UID, employee.DepartmentUID)
	if err != nil {
		return nil, err
	}

	permissionDate := normalizeDateOnly(input.PermissionDate)
	now := uc.now()

	// 1. Type-specific deadline
	if err := validatePermissionSubmissionDeadline(input.Type, permissionDate, now); err != nil {
		return nil, err
	}

	// 2. Resolve the [start, end] window once at submission time. For morning / personal types
	// the resolution uses the employee's effective shift; for official / health_insurance the
	// caller-supplied times are validated against the shift bounds.
	startTime, endTime, err := validatePermissionWindow(input.Type, input.StartTime, input.EndTime, shift, permissionDate)
	if err != nil {
		return nil, err
	}

	// 2b. Reject permissions whose entire window has already passed — there's nothing to excuse.
	if err := validatePermissionWindowNotPast(permissionDate, endTime, now); err != nil {
		return nil, err
	}

	// 3. Same-day overlap with any pending/approved permission for this employee
	overlap, err := uc.permissionRepo.HasOverlappingOnDate(ctx, tx, employee.UID, permissionDate, startTime, endTime, nil)
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, ErrPermissionOverlap
	}

	// 4. Weekly cap
	if err := uc.checkWeeklyCap(ctx, tx, employee.UID, input.Type, permissionDate, nil); err != nil {
		return nil, err
	}

	// 5. Approval flow existence + first step resolution (depends on employee's department and role, so must be done in transaction)
	steps, err := uc.approvalFlowStepRepo.ListByFlow(ctx, tx, permissionDefaultFlowUID)
	if err != nil {
		return nil, err
	}
	if len(steps) == 0 {
		return nil, ErrApprovalFlowHasNoSteps
	}

	requesterUser, err := uc.userRepo.GetByUID(ctx, tx, input.UserUID)
	if err != nil {
		return nil, err
	}
	if requesterUser == nil {
		return nil, ErrUserNotFound
	}

	currentStep, err := resolvePermissionFirstStep(ctx, tx, uc.roleRepo, steps, employee, requesterUser.UID)
	if err != nil {
		return nil, err
	}
	maxStep := currentStep

	approvalRequest := domain.NewApprovalRequestWithState(
		permissionDefaultFlowUID,
		employee.UID,
		currentStep,
		maxStep,
		domain.ApprovalRequestStatusPending,
	)
	if err := uc.approvalRequestRepo.Create(ctx, tx, approvalRequest); err != nil {
		return nil, err
	}

	permissionRequest := domain.NewPermissionRequest(
		employee.UID,
		input.Type,
		permissionDate,
		startTime,
		endTime,
		input.Reason,
		approvalRequest.UID,
	)
	if err := uc.permissionRepo.Create(ctx, tx, permissionRequest); err != nil {
		return nil, err
	}

	submitAction := domain.NewApprovalAction(
		approvalRequest.UID,
		domain.ApprovalActionTypeSubmit,
		nil,
		employee.UID,
		nil,
	)
	if err := uc.approvalActionRepo.Create(ctx, tx, submitAction); err != nil {
		return nil, err
	}

	// Audit log for permission request submission
	actionParams := map[string]interface{}{
		"Employee":       employee.Name,
		"PermissionType": permissionRequest.Type.NameEN(),
		"Date":           permissionRequest.PermissionDate.Format("Jan 2, 2006"),
	}
	defer uc.auditor.Actor(employee.UID).
		Did(audit.ActionSubmit).
		On("permission_request", permissionRequest.UID).
		WithMeta("action_key", "audit.sentence.submit_permission").
		WithMeta("action_params", actionParams).
		WithMeta("permission_type", string(permissionRequest.Type)).
		WithMeta("permission_date", permissionRequest.PermissionDate.Format("2006-01-02")).
		WithMeta("start_time", permissionRequest.StartTime).
		WithMeta("end_time", permissionRequest.EndTime).
		WithMeta("new_status", "pending").
		WithMeta("approval_request_uid", approvalRequest.UID).
		Save(ctx)

	// Capture data for notification before commit
	step, err := uc.approvalFlowStepRepo.GetByFlowAndStep(ctx, tx, permissionDefaultFlowUID, currentStep)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	if step != nil {
		go uc.notifyApprovers(step.RoleUID, employee, permissionRequest)
	}

	return &SubmitPermissionRequestOutput{
		PermissionRequest: permissionRequest,
		ApprovalRequest:   approvalRequest,
	}, nil
}

// checkWeeklyCap enforces:
//   - health_insurance: at most 1 in week
//   - morning + personal: at most 1 of either in week (shared cap)
//   - official: unlimited
//
// Counts all `pending|approved` rows on permission_date within the configured working week.
func (uc *SubmitPermissionRequestUseCase) checkWeeklyCap(ctx context.Context, q ports.Querier, employeeUID string, t domain.PermissionType, date time.Time, excludeUID *string) error {
	if t == domain.PermissionTypeOfficial {
		return nil
	}

	weekendDays, err := uc.weekendRepo.GetWeekendDays(ctx, q)
	if err != nil {
		return err
	}
	weekStart, weekEnd := WeekBoundsForDate(weekendDays, date)

	statuses := []domain.ApprovalRequestStatus{domain.ApprovalRequestStatusPending, domain.ApprovalRequestStatusApproved}

	var types []domain.PermissionType
	switch t {
	case domain.PermissionTypeHealthInsurance:
		types = []domain.PermissionType{domain.PermissionTypeHealthInsurance}
	case domain.PermissionTypeMorning, domain.PermissionTypePersonal:
		types = []domain.PermissionType{domain.PermissionTypeMorning, domain.PermissionTypePersonal}
	default:
		return nil
	}

	count, err := uc.permissionRepo.CountInWeek(ctx, q, employeeUID, types, weekStart, weekEnd, statuses, excludeUID)
	if err != nil {
		return err
	}
	if count >= 1 {
		return ErrPermissionWeeklyCapExceeded
	}
	return nil
}

func validatePermissionSubmissionDeadline(t domain.PermissionType, permissionDate, now time.Time) error {
	day := normalizeDateOnly(permissionDate)
	today := normalizeDateOnly(now)

	switch t {
	case domain.PermissionTypeMorning:
		// Must be submitted at least 1 calendar day before — i.e. permission_date > today.
		if !day.After(today) {
			return ErrPermissionDeadlineMissed
		}
	case domain.PermissionTypePersonal:
		// Must be same day, before 13:00 local time.
		if !day.Equal(today) {
			return ErrPermissionDeadlineMissed
		}
		cutoff := time.Date(today.Year(), today.Month(), today.Day(), 13, 0, 0, 0, time.Local)
		if !now.Before(cutoff) {
			return ErrPermissionDeadlineMissed
		}
	case domain.PermissionTypeOfficial:
		if !day.Equal(today) {
			return ErrPermissionDeadlineMissed
		}
	case domain.PermissionTypeHealthInsurance:
		tomorrow := today.AddDate(0, 0, 1)
		if !(day.Equal(today) || day.Equal(tomorrow)) {
			return ErrPermissionDeadlineMissed
		}
	default:
		return ErrPermissionTypeInvalid
	}
	return nil
}

// validatePermissionWindowNotPast returns an error if the permission's effective end time is in the past relative to `now`.
func validatePermissionWindowNotPast(permissionDate time.Time, endTimeHHMM string, now time.Time) error {
	end, err := buildTimeOnDate(permissionDate, endTimeHHMM)
	if err != nil {
		return err
	}
	if !end.After(now) {
		return ErrPermissionAlreadyExpired
	}
	return nil
}

// validatePermissionWindow returns the [start, end] window for the permission, validating that it is well-formed and within shift bounds if applicable. For types without explicit windows, the effective window is resolved via the domain helper and returned in HH:MM format.
func validatePermissionWindow(t domain.PermissionType, start, end *string, shift *domain.Shift, permissionDate time.Time) (string, string, error) {
	if !t.HasExplicitTimeWindow() {
		// Resolve via the domain helper so the same logic drives reads + writes.
		req := &domain.PermissionRequest{Type: t, PermissionDate: permissionDate}
		startT, endT, err := req.EffectiveWindow(shift.StartTime, shift.EndTime)
		if err != nil {
			return "", "", err
		}
		return startT.Format("15:04"), endT.Format("15:04"), nil
	}

	if start == nil || end == nil || strings.TrimSpace(*start) == "" || strings.TrimSpace(*end) == "" {
		return "", "", ErrPermissionWindowRequired
	}
	startStr := strings.TrimSpace(*start)
	endStr := strings.TrimSpace(*end)

	startT, err := time.Parse("15:04", startStr)
	if err != nil {
		return "", "", ErrPermissionWindowInvalid
	}
	endT, err := time.Parse("15:04", endStr)
	if err != nil {
		return "", "", ErrPermissionWindowInvalid
	}
	if !endT.After(startT) {
		return "", "", ErrPermissionWindowInvalid
	}

	shiftStart, err := time.Parse("15:04", shift.StartTime)
	if err != nil {
		return "", "", err
	}
	shiftEnd, err := time.Parse("15:04", shift.EndTime)
	if err != nil {
		return "", "", err
	}

	if startT.Before(shiftStart) || endT.After(shiftEnd) {
		return "", "", ErrPermissionWindowOutsideShift
	}

	return startStr, endStr, nil
}

func (uc *SubmitPermissionRequestUseCase) notifyApprovers(roleUID string, employee *domain.Employee, request *domain.PermissionRequest) {
	if uc.notificationService == nil {
		return
	}
	var deptUID *string
	if employee.DepartmentUID != nil {
		v := *employee.DepartmentUID
		deptUID = &v
	}
	approvers, err := uc.roleRepo.GetUsersByRoleAndDepartment(context.Background(), uc.db, roleUID, deptUID)
	if err != nil {
		slog.Error("submit_permission_request.notifyApprovers.get_approvers", "error", err)
		return
	}
	if len(approvers) == 0 {
		return
	}
	userUIDs := make([]string, len(approvers))
	for i, u := range approvers {
		userUIDs[i] = u.UID
	}
	params := map[string]interface{}{
		"EmployeeName":   employee.Name,
		"PermissionType": ports.LocalizableString{Ar: request.Type.NameAR(), En: request.Type.NameEN()},
	}
	data := ports.NotificationData{
		"type":       "pending_permission",
		"requestUid": request.UID,
	}
	if _, err := uc.notificationService.SendToUsers(userUIDs, "notification.pending_permission.title", "notification.pending_permission.body", params, data); err != nil {
		slog.Error("submit_permission_request.notifyApprovers.send", "error", err)
	}
}
