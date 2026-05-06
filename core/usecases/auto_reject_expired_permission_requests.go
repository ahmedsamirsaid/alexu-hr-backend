package usecases

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

const autoRejectActorUID = "system"

type AutoRejectExpiredPermissionRequestsOutput struct {
	RejectedCount int
}

type AutoRejectExpiredPermissionRequestsUseCase struct {
	db                  ports.DB
	permissionRepo      ports.PermissionRequestRepository
	approvalRequestRepo ports.ApprovalRequestRepository
	approvalActionRepo  ports.ApprovalActionRepository
	employeeRepo        ports.EmployeeRepository
	deptRepo            ports.DepartmentRepository
	shiftRepo           ports.ShiftRepository
	location            *time.Location
}

func NewAutoRejectExpiredPermissionRequestsUseCase(
	db ports.DB,
	permissionRepo ports.PermissionRequestRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	approvalActionRepo ports.ApprovalActionRepository,
	employeeRepo ports.EmployeeRepository,
	deptRepo ports.DepartmentRepository,
	shiftRepo ports.ShiftRepository,
	timezone string,
) *AutoRejectExpiredPermissionRequestsUseCase {
	loc, err := time.LoadLocation(strings.TrimSpace(timezone))
	if err != nil {
		loc = time.FixedZone("Africa/Cairo", 2*60*60)
	}
	return &AutoRejectExpiredPermissionRequestsUseCase{
		db:                  db,
		permissionRepo:      permissionRepo,
		approvalRequestRepo: approvalRequestRepo,
		approvalActionRepo:  approvalActionRepo,
		employeeRepo:        employeeRepo,
		deptRepo:            deptRepo,
		shiftRepo:           shiftRepo,
		location:            loc,
	}
}

func (uc *AutoRejectExpiredPermissionRequestsUseCase) Execute(ctx context.Context) (*AutoRejectExpiredPermissionRequestsOutput, error) {
	now := time.Now().In(uc.location)
	candidates, err := uc.permissionRepo.FindExpiredPending(ctx, uc.db, now)
	if err != nil {
		return nil, err
	}

	rejected := 0
	for _, p := range candidates {
		expired, err := uc.isExpired(ctx, p, now)
		if err != nil {
			slog.Error("auto_reject_expired_permission_requests.is_expired", "error", err, "uid", p.UID)
			continue
		}
		if !expired {
			continue
		}
		if err := uc.rejectOne(ctx, p); err != nil {
			slog.Error("auto_reject_expired_permission_requests.reject", "error", err, "uid", p.UID)
			continue
		}
		rejected++
	}
	return &AutoRejectExpiredPermissionRequestsOutput{RejectedCount: rejected}, nil
}

func (uc *AutoRejectExpiredPermissionRequestsUseCase) isExpired(ctx context.Context, p *domain.PermissionRequest, now time.Time) (bool, error) {
	day := time.Date(p.PermissionDate.Year(), p.PermissionDate.Month(), p.PermissionDate.Day(), 0, 0, 0, 0, uc.location)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, uc.location)

	switch p.Type {
	case domain.PermissionTypeMorning:
		// Expires at 10:00 of the permission date.
		cutoff := time.Date(day.Year(), day.Month(), day.Day(), 10, 0, 0, 0, uc.location)
		return now.After(cutoff), nil
	case domain.PermissionTypePersonal:
		// Expires at end of permission date (use shift end if known, else 23:59).
		end, err := uc.endOfShiftFor(ctx, p.EmployeeUID, day)
		if err != nil {
			return false, err
		}
		return now.After(end), nil
	case domain.PermissionTypeOfficial, domain.PermissionTypeHealthInsurance:
		// Expires at end of permission date.
		end, err := uc.endOfShiftFor(ctx, p.EmployeeUID, day)
		if err != nil {
			return false, err
		}
		// For health_insurance, also allow until end of permission_date+1 (since it can be submitted day-before).
		if p.Type == domain.PermissionTypeHealthInsurance {
			if today.Before(day) {
				return false, nil
			}
		}
		return now.After(end), nil
	}
	return false, nil
}

func (uc *AutoRejectExpiredPermissionRequestsUseCase) endOfShiftFor(ctx context.Context, employeeUID string, day time.Time) (time.Time, error) {
	employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, employeeUID)
	if err != nil || employee == nil {
		return time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 0, 0, uc.location), err
	}
	shift, err := resolveEffectiveShift(ctx, uc.db, uc.employeeRepo, uc.deptRepo, uc.shiftRepo, employeeUID, employee.DepartmentUID)
	if err != nil || shift == nil {
		return time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 0, 0, uc.location), err
	}
	end, err := time.Parse("15:04", shift.EndTime)
	if err != nil {
		return time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 0, 0, uc.location), nil
	}
	return time.Date(day.Year(), day.Month(), day.Day(), end.Hour(), end.Minute(), 0, 0, uc.location), nil
}

func (uc *AutoRejectExpiredPermissionRequestsUseCase) rejectOne(ctx context.Context, p *domain.PermissionRequest) error {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	approval, err := uc.approvalRequestRepo.GetByUID(ctx, tx, p.ApprovalRequestUID)
	if err != nil {
		return err
	}
	if approval == nil || !approval.IsPending() {
		return nil
	}

	approval.Reject()
	if err := uc.approvalRequestRepo.Update(ctx, tx, approval); err != nil {
		return err
	}

	current := approval.CurrentStep
	comments := "auto-rejected: deadline expired"
	action := domain.NewApprovalAction(
		approval.UID,
		domain.ApprovalActionTypeReject,
		&current,
		autoRejectActorUID,
		&comments,
	)
	if err := uc.approvalActionRepo.Create(ctx, tx, action); err != nil {
		return err
	}

	p.SetDecided()
	if err := uc.permissionRepo.Update(ctx, tx, p); err != nil {
		return err
	}

	return tx.Commit()
}
