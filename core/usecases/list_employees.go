package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

// ListEmployeesInput defines the filters for listing employees.
type ListEmployeesInput struct {
	Status                *domain.EmployeeStatus
	HireDateFrom          *time.Time
	HireDateTo            *time.Time
	Role                  *string // Filter by role name
	ManagedDepartmentUIDs []string
}

// EmployeeUserInfo contains the linked user account information.
type EmployeeUserInfo struct {
	UserUID  string
	Phone    string
	IsActive bool
	Roles    []string
}

// EmployeeListItem represents an employee in the list response.
type EmployeeListItem struct {
	UID                                   string
	Name                                  string
	Mobile                                string
	GovernmentID                          string
	UniversityID                          string
	Email                                 *string
	HireDate                              time.Time
	Status                                domain.EmployeeStatus
	Type                                  domain.EmployeeType
	SubType                               domain.EmployeeSubType
	TelephoneNumber                       *string
	PersonWithSpecialNeeds                bool
	DateOfBirth                           *time.Time
	Gender                                *string
	Religion                              *string
	MaritalStatus                         *string
	Address                               *string
	PlaceOfBirth                          *string
	PlaceOfResidence                      *string
	PoliceStation                         *string
	IDCardValidUntil                      *time.Time
	AcademicLevel                         *string
	EducationalQualification              *string
	UniversityName                        *string
	Faculty                               *string
	Specialization                        *string
	YearObtained                          *int
	ActualAppointmentReappointmentDate    *time.Time
	AppointmentDecisionDate               *time.Time
	AppointmentDecisionNumber             *string
	AppointmentType                       *string
	DepartmentName                        *string
	Grade                                 *string
	WorkEntity                            *string
	EmployeeFileNumber                    *string
	InsuranceNumber                       *string
	EmploymentStatus                      *string
	SolidarityFund                        bool
	SubscriptionDate                      *time.Time
	MilitaryStatus                        *string
	MedicalCadre                          *string
	PersonalPhotoURL                      *string
	NationalCardImageURL                  *string
	QualificationCertificateImageURL      *string
	CVURL                                 *string
	MemberNumber                          *string
	InsuranceCode                         *string
	JobGroup                              *string
	QualitativeGroup                      *string
	JobTitleAtLevel                       *string
	JobTitleBeforePlacement               *string
	FinancialGrade                        *string
	PreviousFinancialGrade                *string
	JobLevel                              *string
	DecisionDate                          *time.Time
	DecisionNumber                        *int64
	GradeGrantDate                        *time.Time
	NatureOfAppointment                   *string
	Notes                                 *string
	Reappointment                         bool
	DecisionFileURL                       *string
	AppointmentSeniorityOrGradeWithdrawal bool
	DepartmentUID                         *string // Department UID (nil if not assigned)
	ShiftUID                              *string
	User                                  *EmployeeUserInfo // Linked user account info (nil if no user linked)
}

// ListEmployeesOutput contains the list of employees.
type ListEmployeesOutput struct {
	Employees []EmployeeListItem
}

// ListEmployeesUseCase handles listing employees.
type ListEmployeesUseCase struct {
	db           ports.DB
	employeeRepo ports.EmployeeRepository
	userRepo     ports.UserRepository
	roleRepo     ports.RoleRepository
}

// NewListEmployeesUseCase creates a new list employees use case.
func NewListEmployeesUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	userRepo ports.UserRepository,
	roleRepo ports.RoleRepository,
) *ListEmployeesUseCase {
	return &ListEmployeesUseCase{
		db:           db,
		employeeRepo: employeeRepo,
		userRepo:     userRepo,
		roleRepo:     roleRepo,
	}
}

// Execute lists employees with optional filters.
func (uc *ListEmployeesUseCase) Execute(ctx context.Context, input ListEmployeesInput) (*ListEmployeesOutput, error) {
	filter := &ports.EmployeeListFilter{
		Status:       input.Status,
		HireDateFrom: input.HireDateFrom,
		HireDateTo:   input.HireDateTo,
	}

	employees, err := uc.employeeRepo.List(ctx, uc.db, filter)
	if err != nil {
		return nil, err
	}

	allowedDepartments := make(map[string]struct{}, len(input.ManagedDepartmentUIDs))
	for _, departmentUID := range input.ManagedDepartmentUIDs {
		allowedDepartments[departmentUID] = struct{}{}
	}
	limitByDepartment := len(allowedDepartments) > 0

	items := make([]EmployeeListItem, 0, len(employees))
	for _, emp := range employees {
		if limitByDepartment {
			if emp.DepartmentUID == nil {
				continue
			}
			if _, ok := allowedDepartments[*emp.DepartmentUID]; !ok {
				continue
			}
		}

		item := EmployeeListItem{
			UID:                                   emp.UID,
			Name:                                  emp.Name,
			Mobile:                                emp.Mobile,
			GovernmentID:                          emp.GovernmentID,
			UniversityID:                          emp.UniversityID,
			Email:                                 emp.Email,
			HireDate:                              emp.HireDate,
			Status:                                emp.Status,
			Type:                                  emp.Type,
			SubType:                               emp.SubType,
			TelephoneNumber:                       emp.TelephoneNumber,
			PersonWithSpecialNeeds:                emp.PersonWithSpecialNeeds,
			DateOfBirth:                           emp.DateOfBirth,
			Gender:                                emp.Gender,
			Religion:                              emp.Religion,
			MaritalStatus:                         emp.MaritalStatus,
			Address:                               emp.Address,
			PlaceOfBirth:                          emp.PlaceOfBirth,
			PlaceOfResidence:                      emp.PlaceOfResidence,
			PoliceStation:                         emp.PoliceStation,
			IDCardValidUntil:                      emp.IDCardValidUntil,
			AcademicLevel:                         emp.AcademicLevel,
			EducationalQualification:              emp.EducationalQualification,
			UniversityName:                        emp.UniversityName,
			Faculty:                               emp.Faculty,
			Specialization:                        emp.Specialization,
			YearObtained:                          emp.YearObtained,
			ActualAppointmentReappointmentDate:    emp.ActualAppointmentReappointmentDate,
			AppointmentDecisionDate:               emp.AppointmentDecisionDate,
			AppointmentDecisionNumber:             emp.AppointmentDecisionNumber,
			AppointmentType:                       emp.AppointmentType,
			DepartmentName:                        emp.DepartmentName,
			Grade:                                 emp.Grade,
			WorkEntity:                            emp.WorkEntity,
			EmployeeFileNumber:                    emp.EmployeeFileNumber,
			InsuranceNumber:                       emp.InsuranceNumber,
			EmploymentStatus:                      emp.EmploymentStatus,
			SolidarityFund:                        emp.SolidarityFund,
			SubscriptionDate:                      emp.SubscriptionDate,
			MilitaryStatus:                        emp.MilitaryStatus,
			MedicalCadre:                          emp.MedicalCadre,
			PersonalPhotoURL:                      emp.PersonalPhotoURL,
			NationalCardImageURL:                  emp.NationalCardImageURL,
			QualificationCertificateImageURL:      emp.QualificationCertificateImageURL,
			CVURL:                                 emp.CVURL,
			MemberNumber:                          emp.MemberNumber,
			InsuranceCode:                         emp.InsuranceCode,
			JobGroup:                              emp.JobGroup,
			QualitativeGroup:                      emp.QualitativeGroup,
			JobTitleAtLevel:                       emp.JobTitleAtLevel,
			JobTitleBeforePlacement:               emp.JobTitleBeforePlacement,
			FinancialGrade:                        emp.FinancialGrade,
			PreviousFinancialGrade:                emp.PreviousFinancialGrade,
			JobLevel:                              emp.JobLevel,
			DecisionDate:                          emp.DecisionDate,
			DecisionNumber:                        emp.DecisionNumber,
			GradeGrantDate:                        emp.GradeGrantDate,
			NatureOfAppointment:                   emp.NatureOfAppointment,
			Notes:                                 emp.Notes,
			Reappointment:                         emp.Reappointment,
			DecisionFileURL:                       emp.DecisionFileURL,
			AppointmentSeniorityOrGradeWithdrawal: emp.AppointmentSeniorityOrGradeWithdrawal,
			DepartmentUID:                         emp.DepartmentUID,
			ShiftUID:                              emp.ShiftUID,
		}

		// Fetch linked user details
		user, err := uc.userRepo.GetByEmployeeUID(ctx, uc.db, emp.UID)
		if err != nil {
			return nil, err
		}
		if user != nil {
			// Fetch user roles
			roles, err := uc.roleRepo.GetRolesForUser(ctx, uc.db, user.ID)
			if err != nil {
				return nil, err
			}
			roleNames := make([]string, 0, len(roles))
			for _, role := range roles {
				roleNames = append(roleNames, role.Name)
			}

			item.User = &EmployeeUserInfo{
				UserUID:  user.UID,
				Phone:    user.Phone,
				IsActive: user.IsActive,
				Roles:    roleNames,
			}
		}

		// Apply role filter if specified
		if input.Role != nil {
			if item.User == nil {
				continue // Skip employees without linked user when filtering by role
			}
			hasRole := false
			for _, r := range item.User.Roles {
				if r == *input.Role {
					hasRole = true
					break
				}
			}
			if !hasRole {
				continue // Skip employees without the specified role
			}
		}

		items = append(items, item)
	}

	return &ListEmployeesOutput{Employees: items}, nil
}
