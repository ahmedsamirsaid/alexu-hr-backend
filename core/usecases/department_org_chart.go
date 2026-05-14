package usecases

import (
	"context"
	"sort"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type OrgChartNode struct {
	UID              string
	Name             string
	RoleName         *string
	DepartmentUID    *string
	DepartmentNameEN string
	DepartmentNameAR *string
	DirectReports    []*OrgChartNode
}

type OrgChartInput struct {
	DepartmentUIDs []string // nil = full tree; non-nil = department-scoped filter
}

type OrgChartOutput struct {
	Roots []*OrgChartNode
}

type OrgChartUseCase struct {
	db       ports.DB
	empRepo  ports.EmployeeRepository
	deptRepo ports.DepartmentRepository
	roleRepo ports.RoleRepository
}

func NewOrgChartUseCase(
	db ports.DB,
	empRepo ports.EmployeeRepository,
	deptRepo ports.DepartmentRepository,
	roleRepo ports.RoleRepository,
) *OrgChartUseCase {
	return &OrgChartUseCase{
		db:       db,
		empRepo:  empRepo,
		deptRepo: deptRepo,
		roleRepo: roleRepo,
	}
}

func (uc *OrgChartUseCase) Execute(ctx context.Context, input OrgChartInput) (*OrgChartOutput, error) {
	// 1. Fetch all active employees.
	activeStatus := domain.EmployeeStatusActive
	employees, err := uc.empRepo.List(ctx, uc.db, &ports.EmployeeListFilter{Status: &activeStatus})
	if err != nil {
		return nil, err
	}

	// 2. Fetch all departments → lookup map.
	departments, err := uc.deptRepo.List(ctx, uc.db, false)
	if err != nil {
		return nil, err
	}
	deptByUID := make(map[string]*domain.Department, len(departments))
	for _, d := range departments {
		deptByUID[d.UID] = d
	}

	// 3. Batch-fetch role names keyed by employee UID.
	employeeUIDs := make([]string, 0, len(employees))
	for _, emp := range employees {
		employeeUIDs = append(employeeUIDs, emp.UID)
	}
	roleNameByEmpUID, err := uc.roleRepo.GetRoleNamesByEmployeeUIDs(ctx, uc.db, employeeUIDs)
	if err != nil {
		return nil, err
	}

	// 4. Build the node map.
	nodeByUID := make(map[string]*OrgChartNode, len(employees))
	for _, emp := range employees {
		node := &OrgChartNode{
			UID:           emp.UID,
			Name:          emp.Name,
			DepartmentUID: emp.DepartmentUID,
			DirectReports: []*OrgChartNode{},
		}
		if name, ok := roleNameByEmpUID[emp.UID]; ok {
			node.RoleName = &name
		}
		if emp.DepartmentUID != nil {
			if dept, ok := deptByUID[*emp.DepartmentUID]; ok {
				node.DepartmentNameEN = dept.NameEN
				node.DepartmentNameAR = dept.NameAR
			}
		}
		nodeByUID[emp.UID] = node
	}

	// 5. Wire parent → children; collect roots.
	var roots []*OrgChartNode
	for _, emp := range employees {
		node := nodeByUID[emp.UID]
		if emp.ManagerUID == nil {
			roots = append(roots, node)
			continue
		}
		if parent, ok := nodeByUID[*emp.ManagerUID]; ok {
			parent.DirectReports = append(parent.DirectReports, node)
		} else {
			roots = append(roots, node)
		}
	}

	// 6. Sort alphabetically at every level.
	sortTree(roots)

	// 7. Apply department filter if present.
	if input.DepartmentUIDs != nil {
		filter := make(map[string]struct{}, len(input.DepartmentUIDs))
		for _, uid := range input.DepartmentUIDs {
			filter[uid] = struct{}{}
		}
		pruned := make([]*OrgChartNode, 0, len(roots))
		for _, root := range roots {
			if n := pruneNode(root, filter); n != nil {
				pruned = append(pruned, n)
			}
		}
		roots = pruned
	}

	return &OrgChartOutput{Roots: roots}, nil
}

func pruneNode(node *OrgChartNode, filter map[string]struct{}) *OrgChartNode {
	inFilter := false
	if node.DepartmentUID != nil {
		_, inFilter = filter[*node.DepartmentUID]
	}

	prunedReports := make([]*OrgChartNode, 0, len(node.DirectReports))
	for _, child := range node.DirectReports {
		if pruned := pruneNode(child, filter); pruned != nil {
			prunedReports = append(prunedReports, pruned)
		}
	}

	if !inFilter && len(prunedReports) == 0 {
		return nil
	}

	return &OrgChartNode{
		UID:              node.UID,
		Name:             node.Name,
		RoleName:         node.RoleName,
		DepartmentUID:    node.DepartmentUID,
		DepartmentNameEN: node.DepartmentNameEN,
		DepartmentNameAR: node.DepartmentNameAR,
		DirectReports:    prunedReports,
	}
}

func sortTree(nodes []*OrgChartNode) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})
	for _, n := range nodes {
		sortTree(n.DirectReports)
	}
}
