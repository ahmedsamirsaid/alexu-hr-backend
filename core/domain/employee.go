package domain

import (
	"strings"
	"time"
)

type EmployeeStatus string
type EmployeeType string
type EmployeeSubType string

const (
	EmployeeStatusActive     EmployeeStatus = "active"
	EmployeeStatusInactive   EmployeeStatus = "inactive"
	EmployeeStatusTerminated EmployeeStatus = "terminated"

	EmployeeTypePermanent EmployeeType = "permanent"
	EmployeeTypeTemporary EmployeeType = "temporary"

	EmployeeSubTypeNormal                         EmployeeSubType = "normal"
	EmployeeSubTypeSpecialNeeds                   EmployeeSubType = "special_needs"
	EmployeeSubTypeSeparationTerminationForBudget EmployeeSubType = "separation_termination_for_budget"
	EmployeeSubTypeComprehensiveBonus             EmployeeSubType = "comprehensive_bonus"
	EmployeeSubTypeContractEmployees              EmployeeSubType = "contract_employees"
)

type Employee struct {
	ID                                    int64
	UID                                   string
	Name                                  string
	Mobile                                string
	GovernmentID                          string
	UniversityID                          string
	Email                                 *string
	HireDate                              time.Time
	Status                                EmployeeStatus
	Type                                  EmployeeType
	SubType                               EmployeeSubType
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
	DepartmentUID                         *string
	ShiftUID                              *string
	CreatedAt                             time.Time
	UpdatedAt                             time.Time
}

func NewEmployee(name, mobile, governmentID, universityID string, hireDate time.Time) *Employee {
	return &Employee{
		UID:          GenerateUID("emp"),
		Name:         name,
		Mobile:       mobile,
		GovernmentID: governmentID,
		UniversityID: universityID,
		HireDate:     hireDate,
		Status:       EmployeeStatusActive,
		Type:         EmployeeTypePermanent,
		SubType:      EmployeeSubTypeNormal,
	}
}

func (e *Employee) ApplyClassificationDefaults() {
	if e.Type == "" {
		e.Type = EmployeeTypePermanent
	}
	if e.SubType != "" {
		return
	}
	switch e.Type {
	case EmployeeTypeTemporary:
		e.SubType = EmployeeSubTypeContractEmployees
	default:
		e.SubType = EmployeeSubTypeNormal
	}
}

func NormalizeEmployeeType(value string) EmployeeType {
	switch normalizeEmployeeClassificationValue(value) {
	case "permanent", "permenant":
		return EmployeeTypePermanent
	case "temporary":
		return EmployeeTypeTemporary
	default:
		return ""
	}
}

func NormalizeEmployeeSubType(value string) EmployeeSubType {
	switch normalizeEmployeeClassificationValue(value) {
	case "normal":
		return EmployeeSubTypeNormal
	case "specialneeds", "special_needs":
		return EmployeeSubTypeSpecialNeeds
	case "separationterminationforbudget", "separation_termination_for_budget":
		return EmployeeSubTypeSeparationTerminationForBudget
	case "comprehensivebonus", "comprehensive_bonus", "omprehensivebonus", "omprehensive_bonus":
		return EmployeeSubTypeComprehensiveBonus
	case "contractemployee", "contractemployees", "contract_employee", "contract_employees":
		return EmployeeSubTypeContractEmployees
	default:
		return ""
	}
}

func IsValidEmployeeSubTypeForType(employeeType EmployeeType, subType EmployeeSubType) bool {
	switch employeeType {
	case EmployeeTypePermanent:
		return subType == EmployeeSubTypeNormal || subType == EmployeeSubTypeSpecialNeeds
	case EmployeeTypeTemporary:
		return subType == EmployeeSubTypeSeparationTerminationForBudget ||
			subType == EmployeeSubTypeComprehensiveBonus ||
			subType == EmployeeSubTypeContractEmployees
	default:
		return false
	}
}

func normalizeEmployeeClassificationValue(value string) string {
	replacer := strings.NewReplacer(" ", "", "-", "", "_", "", "/", "")
	return replacer.Replace(strings.ToLower(strings.TrimSpace(value)))
}
