package usecases

import (
	"context"
	"strings"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type AssignRoleInput struct {
	UserUID       string
	RoleUID       string
	DepartmentUID *string
}

type AssignRoleUseCase struct {
	db       ports.DB
	userRepo ports.UserRepository
	roleRepo ports.RoleRepository
	deptRepo ports.DepartmentRepository
	auditor  audit.Auditor
}

func NewAssignRoleUseCase(
	db ports.DB,
	userRepo ports.UserRepository,
	roleRepo ports.RoleRepository,
	deptRepo ports.DepartmentRepository,
	auditor audit.Auditor,
) *AssignRoleUseCase {
	return &AssignRoleUseCase{
		db:       db,
		userRepo: userRepo,
		roleRepo: roleRepo,
		deptRepo: deptRepo,
		auditor:  auditor,
	}
}

func (uc *AssignRoleUseCase) Execute(ctx context.Context, input AssignRoleInput) error {
	user, err := uc.userRepo.GetByUID(ctx, uc.db, input.UserUID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	role, err := uc.roleRepo.GetByUID(ctx, uc.db, input.RoleUID)
	if err != nil {
		return err
	}
	if role == nil {
		return ErrRoleNotFound
	}

	normalizedScopeType := domain.NormalizeRoleScopeType(role.ScopeType)
	if normalizedScopeType == "" {
		return ErrInvalidRoleScopeType
	}

	departmentUIDInput := ""
	if input.DepartmentUID != nil {
		departmentUIDInput = strings.TrimSpace(*input.DepartmentUID)
	}

	switch normalizedScopeType {
	case domain.RoleScopeDepartment:
		if departmentUIDInput == "" {
			return ErrRoleScopeRequired
		}
	case domain.RoleScopeGlobal, domain.RoleScopeSelf:
		if departmentUIDInput != "" {
			return ErrRoleScopeConflict
		}
	}

	// Audit log after successful role assignment
	auditBuilder := uc.auditor.From(ctx).Did(audit.ActionAssign).On(audit.EntityUser, input.UserUID).
		WithMeta("role_uid", role.UID).
		WithMeta("role_name", role.Name)

	if departmentUIDInput == "" {
		defer auditBuilder.Save(ctx)
		return uc.roleRepo.AssignRoleToUser(ctx, uc.db, user.ID, role.ID)
	}

	department, err := uc.deptRepo.GetByUID(ctx, uc.db, departmentUIDInput)
	if err != nil {
		return err
	}
	if department == nil {
		return ErrDepartmentNotFound
	}

	departmentUID := department.UID
	auditBuilder.WithMeta("department_uid", departmentUID)
	defer auditBuilder.Save(ctx)
	
	return uc.roleRepo.AssignRoleToUserWithDepartment(ctx, uc.db, user.ID, role.ID, &departmentUID)
}

type RemoveRoleInput struct {
	UserUID string
	RoleUID string
}

type RemoveRoleUseCase struct {
	db       ports.DB
	userRepo ports.UserRepository
	roleRepo ports.RoleRepository
	auditor  audit.Auditor
}

func NewRemoveRoleUseCase(
	db ports.DB,
	userRepo ports.UserRepository,
	roleRepo ports.RoleRepository,
	auditor audit.Auditor,
) *RemoveRoleUseCase {
	return &RemoveRoleUseCase{
		db:       db,
		userRepo: userRepo,
		roleRepo: roleRepo,
		auditor:  auditor,
	}
}

func (uc *RemoveRoleUseCase) Execute(ctx context.Context, input RemoveRoleInput) error {
	user, err := uc.userRepo.GetByUID(ctx, uc.db, input.UserUID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	role, err := uc.roleRepo.GetByUID(ctx, uc.db, input.RoleUID)
	if err != nil {
		return err
	}
	if role == nil {
		return ErrRoleNotFound
	}

	// Audit log after successful role removal
	defer uc.auditor.From(ctx).Did(audit.ActionUnassign).On(audit.EntityUser, input.UserUID).
		WithMeta("role_uid", role.UID).
		WithMeta("role_name", role.Name).
		Save(ctx)

	return uc.roleRepo.RemoveRoleFromUser(ctx, uc.db, user.ID, role.ID)
}
