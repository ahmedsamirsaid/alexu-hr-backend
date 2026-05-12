package usecases

import (
	"context"
	"sort"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type ManagerCandidate struct {
	EmployeeUID string
	Name        string
	RoleName    string
}

type ListManagerCandidatesUseCase struct {
	db       ports.DB
	empRepo  ports.EmployeeRepository
	roleRepo ports.RoleRepository
}

func NewListManagerCandidatesUseCase(
	db ports.DB,
	empRepo ports.EmployeeRepository,
	roleRepo ports.RoleRepository,
) *ListManagerCandidatesUseCase {
	return &ListManagerCandidatesUseCase{db: db, empRepo: empRepo, roleRepo: roleRepo}
}

// Execute returns all employees that can serve as managers for leadership employees.
// Valid managers hold one of: role_university_president, role_vice_president, role_dean.
func (uc *ListManagerCandidatesUseCase) Execute(ctx context.Context) ([]*ManagerCandidate, error) {
	const (
		presidentUID = "role_university_president"
		vpUID        = "role_vice_president"
		deanUID      = "role_dean"
	)

	seen := map[string]bool{}
	var empUIDs []string

	for _, roleUID := range []string{presidentUID, vpUID, deanUID} {
		users, err := uc.roleRepo.GetUsersByRoleAndDepartment(ctx, uc.db, roleUID, nil)
		if err != nil {
			return nil, err
		}
		for _, u := range users {
			if u.EmployeeUID != nil && !seen[*u.EmployeeUID] {
				seen[*u.EmployeeUID] = true
				empUIDs = append(empUIDs, *u.EmployeeUID)
			}
		}
	}

	if len(empUIDs) == 0 {
		return []*ManagerCandidate{}, nil
	}

	employees, err := uc.empRepo.GetByUIDs(ctx, uc.db, empUIDs)
	if err != nil {
		return nil, err
	}

	roleNames, err := uc.roleRepo.GetRoleNamesByEmployeeUIDs(ctx, uc.db, empUIDs)
	if err != nil {
		return nil, err
	}

	candidates := make([]*ManagerCandidate, 0, len(employees))
	for _, emp := range employees {
		if emp == nil || emp.Status != domain.EmployeeStatusActive {
			continue
		}
		candidates = append(candidates, &ManagerCandidate{
			EmployeeUID: emp.UID,
			Name:        emp.Name,
			RoleName:    roleNames[emp.UID],
		})
	}

	// Sort by role priority then name for consistent ordering
	rolePriority := map[string]int{
		"University President": 1,
		"Vice President":       2,
		"Dean":                 3,
	}
	sort.Slice(candidates, func(i, j int) bool {
		pi := rolePriority[candidates[i].RoleName]
		pj := rolePriority[candidates[j].RoleName]
		if pi != pj {
			return pi < pj
		}
		return candidates[i].Name < candidates[j].Name
	})

	return candidates, nil
}
