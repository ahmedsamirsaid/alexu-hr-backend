package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type UpdatePermissionRequestInput struct {
	PermissionRequestUID string
	ActorEmployeeUID     string
	Type                 domain.PermissionType
	PermissionDate       time.Time
	StartTime            *string
	EndTime              *string
	Reason               *string
}

type UpdatePermissionRequestUseCase struct {
	db                  ports.DB
	permissionRepo      ports.PermissionRequestRepository
	approvalRequestRepo ports.ApprovalRequestRepository
	employeeRepo        ports.EmployeeRepository
	deptRepo            ports.DepartmentRepository
	shiftRepo           ports.ShiftRepository
	weekendRepo         ports.WeekendConfigRepository
	auditor             audit.Auditor
	now                 func() time.Time
}

func NewUpdatePermissionRequestUseCase(
	db ports.DB,
	permissionRepo ports.PermissionRequestRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	employeeRepo ports.EmployeeRepository,
	deptRepo ports.DepartmentRepository,
	shiftRepo ports.ShiftRepository,
	weekendRepo ports.WeekendConfigRepository,
	auditor audit.Auditor,
) *UpdatePermissionRequestUseCase {
	return &UpdatePermissionRequestUseCase{
		db:                  db,
		permissionRepo:      permissionRepo,
		approvalRequestRepo: approvalRequestRepo,
		employeeRepo:        employeeRepo,
		deptRepo:            deptRepo,
		shiftRepo:           shiftRepo,
		weekendRepo:         weekendRepo,
		auditor:             auditor,
		now:                 time.Now,
	}
}

func (uc *UpdatePermissionRequestUseCase) Execute(ctx context.Context, input UpdatePermissionRequestInput) (*domain.PermissionRequest, error) {
	if !input.Type.IsValid() {
		return nil, ErrPermissionTypeInvalid
	}

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	request, err := uc.permissionRepo.GetByUID(ctx, tx, input.PermissionRequestUID)
	if err != nil {
		return nil, err
	}
	if request == nil {
		return nil, ErrPermissionRequestNotFound
	}
	if request.EmployeeUID != input.ActorEmployeeUID {
		return nil, ErrNotRequestOwner
	}

	approvalRequest, err := uc.approvalRequestRepo.GetByUID(ctx, tx, request.ApprovalRequestUID)
	if err != nil {
		return nil, err
	}
	if approvalRequest == nil {
		return nil, ErrApprovalRequestNotFound
	}
	if !approvalRequest.IsPending() {
		return nil, ErrPermissionNotEditable
	}

	employee, err := uc.employeeRepo.GetByUID(ctx, tx, request.EmployeeUID)
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

	if err := validatePermissionSubmissionDeadline(input.Type, permissionDate, now); err != nil {
		return nil, err
	}
	startTime, endTime, err := validatePermissionWindow(input.Type, input.StartTime, input.EndTime, shift, permissionDate)
	if err != nil {
		return nil, err
	}
	if err := validatePermissionWindowNotPast(permissionDate, endTime, now); err != nil {
		return nil, err
	}

	overlap, err := uc.permissionRepo.HasOverlappingOnDate(ctx, tx, employee.UID, permissionDate, startTime, endTime, &request.UID)
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, ErrPermissionOverlap
	}

	if err := checkWeeklyCapForUpdate(ctx, tx, uc.permissionRepo, uc.weekendRepo, employee.UID, input.Type, permissionDate, &request.UID); err != nil {
		return nil, err
	}

	// Capture old values for audit
	oldType := request.Type
	oldDate := request.PermissionDate

	request.Type = input.Type
	request.PermissionDate = permissionDate
	request.StartTime = startTime
	request.EndTime = endTime
	request.Reason = input.Reason
	if err := uc.permissionRepo.Update(ctx, tx, request); err != nil {
		return nil, err
	}

	// Audit log for permission request update
	actionParams := map[string]interface{}{
		"Actor": employee.Name,
	}
	auditBuilder := uc.auditor.Actor(employee.UID).
		Did(audit.ActionUpdate).
		On("permission_request", request.UID).
		WithMeta("action_key", "audit.sentence.update_permission").
		WithMeta("action_params", actionParams)
	
	if oldType != request.Type {
		auditBuilder.WithMeta("old_permission_type", string(oldType)).
			WithMeta("new_permission_type", string(request.Type))
	}
	if !oldDate.Equal(request.PermissionDate) {
		auditBuilder.WithMeta("old_permission_date", oldDate.Format("2006-01-02")).
			WithMeta("new_permission_date", request.PermissionDate.Format("2006-01-02"))
	}
	defer auditBuilder.Save(ctx)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return request, nil
}

func checkWeeklyCapForUpdate(ctx context.Context, q ports.Querier, permRepo ports.PermissionRequestRepository, weekendRepo ports.WeekendConfigRepository, employeeUID string, t domain.PermissionType, date time.Time, excludeUID *string) error {
	if t == domain.PermissionTypeOfficial {
		return nil
	}
	weekendDays, err := weekendRepo.GetWeekendDays(ctx, q)
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

	count, err := permRepo.CountInWeek(ctx, q, employeeUID, types, weekStart, weekEnd, statuses, excludeUID)
	if err != nil {
		return err
	}
	if count >= 1 {
		return ErrPermissionWeeklyCapExceeded
	}
	return nil
}
