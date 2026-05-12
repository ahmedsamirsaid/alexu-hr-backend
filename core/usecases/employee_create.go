package usecases

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

var (
	ErrEmployeeMobileAlreadyExists   = errors.New("employee mobile already exists")
	ErrGovernmentIDAlreadyExists     = errors.New("government id already exists")
	ErrUniversityIDAlreadyExists     = errors.New("university id already exists")
	ErrEmployeeDocumentTypeInvalid   = errors.New("employee document type is invalid")
	ErrEmployeeDocumentDuplicate     = errors.New("employee document type is duplicated")
	ErrInvalidEmployeeClassification = errors.New("invalid employee classification")
)

const (
	employeeDocumentTypePersonalPhoto                 = "personal_photo"
	employeeDocumentTypeNationalCardImage             = "national_card_image"
	employeeDocumentTypeQualificationCertificateImage = "qualification_certificate_image"
	employeeDocumentTypeCV                            = "cv"
	employeeDocumentTypeDecisionFile                  = "decision_file"
)

type CreateEmployeeInput struct {
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
	AppointmentSeniorityOrGradeWithdrawal bool
	DepartmentUID                         *string
	ShiftUID                              *string
	Documents                             []CreateEmployeeDocumentInput
}

type CreateEmployeeDocumentInput struct {
	DocumentType string
	FileName     string
	ContentType  string
}

type CreateEmployeeDocumentOutput struct {
	DocumentType string `json:"documentType"`
	FileName     string `json:"fileName"`
	URL          string `json:"url"`
	Method       string `json:"method"`
	Bucket       string `json:"bucket"`
	ObjectKey    string `json:"objectKey"`
	StoredURL    string `json:"storedUrl"`
}

type CreateEmployeeOutput struct {
	Employee  *domain.Employee
	UserUID   *string
	Documents []CreateEmployeeDocumentOutput
}

type CreateEmployeeUseCase struct {
	db                  ports.DB
	employeeRepo        ports.EmployeeRepository
	userRepo            ports.UserRepository
	roleRepo            ports.RoleRepository
	documentUploadURLUC *GenerateDocumentUploadURLUseCase
	auditor             audit.Auditor
}

func NewCreateEmployeeUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	userRepo ports.UserRepository,
	roleRepo ports.RoleRepository,
	documentUploadURLUC *GenerateDocumentUploadURLUseCase,
	auditor audit.Auditor,
) *CreateEmployeeUseCase {
	return &CreateEmployeeUseCase{
		db:                  db,
		employeeRepo:        employeeRepo,
		userRepo:            userRepo,
		roleRepo:            roleRepo,
		documentUploadURLUC: documentUploadURLUC,
		auditor:             auditor,
	}
}

func (uc *CreateEmployeeUseCase) Execute(ctx context.Context, input CreateEmployeeInput) (*CreateEmployeeOutput, error) {
	input.UniversityID = strings.TrimSpace(input.UniversityID)
	if input.UniversityID == "" {
		generatedUniversityID, err := uc.generateUniversityID(ctx)
		if err != nil {
			return nil, err
		}
		input.UniversityID = generatedUniversityID
	}

	if err := uc.validateUniqueness(ctx, input); err != nil {
		return nil, err
	}

	employee := domain.NewEmployee(input.Name, input.Mobile, input.GovernmentID, input.UniversityID, input.HireDate)
	employee.Email = trimStringPtr(input.Email)
	if input.Status != "" {
		employee.Status = input.Status
	}
	if input.Type != "" {
		employee.Type = input.Type
	}
	if input.SubType != "" {
		employee.SubType = input.SubType
	}
	employee.TelephoneNumber = trimStringPtr(input.TelephoneNumber)
	employee.PersonWithSpecialNeeds = input.PersonWithSpecialNeeds
	employee.DateOfBirth = input.DateOfBirth
	employee.Gender = trimStringPtr(input.Gender)
	employee.Religion = trimStringPtr(input.Religion)
	employee.MaritalStatus = trimStringPtr(input.MaritalStatus)
	employee.Address = trimStringPtr(input.Address)
	employee.PlaceOfBirth = trimStringPtr(input.PlaceOfBirth)
	employee.PlaceOfResidence = trimStringPtr(input.PlaceOfResidence)
	employee.PoliceStation = trimStringPtr(input.PoliceStation)
	employee.IDCardValidUntil = input.IDCardValidUntil
	employee.AcademicLevel = trimStringPtr(input.AcademicLevel)
	employee.EducationalQualification = trimStringPtr(input.EducationalQualification)
	employee.UniversityName = trimStringPtr(input.UniversityName)
	employee.Faculty = trimStringPtr(input.Faculty)
	employee.Specialization = trimStringPtr(input.Specialization)
	employee.YearObtained = input.YearObtained
	employee.ActualAppointmentReappointmentDate = input.ActualAppointmentReappointmentDate
	employee.AppointmentDecisionDate = input.AppointmentDecisionDate
	employee.AppointmentDecisionNumber = trimStringPtr(input.AppointmentDecisionNumber)
	employee.AppointmentType = trimStringPtr(input.AppointmentType)
	employee.DepartmentName = trimStringPtr(input.DepartmentName)
	employee.Grade = input.Grade
	employee.WorkEntity = trimStringPtr(input.WorkEntity)
	employee.EmployeeFileNumber = trimStringPtr(input.EmployeeFileNumber)
	employee.InsuranceNumber = trimStringPtr(input.InsuranceNumber)
	employee.EmploymentStatus = trimStringPtr(input.EmploymentStatus)
	employee.SolidarityFund = input.SolidarityFund
	employee.SubscriptionDate = input.SubscriptionDate
	employee.MilitaryStatus = trimStringPtr(input.MilitaryStatus)
	employee.MedicalCadre = trimStringPtr(input.MedicalCadre)
	employee.MemberNumber = trimStringPtr(input.MemberNumber)
	employee.InsuranceCode = trimStringPtr(input.InsuranceCode)
	employee.JobGroup = trimStringPtr(input.JobGroup)
	employee.QualitativeGroup = trimStringPtr(input.QualitativeGroup)
	employee.JobTitleAtLevel = trimStringPtr(input.JobTitleAtLevel)
	employee.JobTitleBeforePlacement = trimStringPtr(input.JobTitleBeforePlacement)
	employee.FinancialGrade = trimStringPtr(input.FinancialGrade)
	employee.PreviousFinancialGrade = trimStringPtr(input.PreviousFinancialGrade)
	employee.JobLevel = trimStringPtr(input.JobLevel)
	employee.DecisionDate = input.DecisionDate
	employee.DecisionNumber = input.DecisionNumber
	employee.GradeGrantDate = input.GradeGrantDate
	employee.NatureOfAppointment = trimStringPtr(input.NatureOfAppointment)
	employee.Notes = trimStringPtr(input.Notes)
	employee.Reappointment = input.Reappointment
	employee.AppointmentSeniorityOrGradeWithdrawal = input.AppointmentSeniorityOrGradeWithdrawal
	employee.DepartmentUID = input.DepartmentUID
	employee.ShiftUID = input.ShiftUID
	employee.ApplyClassificationDefaults()

	if !domain.IsValidEmployeeSubTypeForType(employee.Type, employee.SubType) {
		return nil, fmt.Errorf("%w: subtype %q for type %q", ErrInvalidEmployeeClassification, employee.SubType, employee.Type)
	}

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	documents, err := uc.prepareEmployeeDocuments(ctx, employee.UID, input.Documents, employee)
	if err != nil {
		return nil, err
	}

	if err := uc.employeeRepo.Create(ctx, tx, employee); err != nil {
		return nil, err
	}

	user := domain.NewUser(input.Mobile)
	user.EmployeeUID = &employee.UID
	if err := uc.userRepo.Create(ctx, tx, user); err != nil {
		if errors.Is(err, ErrPhoneAlreadyExists) {
			return nil, err
		}
		return nil, err
	}

	employeeRole, err := uc.roleRepo.GetByName(ctx, tx, "Employee")
	if err != nil {
		return nil, err
	}
	if employeeRole == nil {
		return nil, fmt.Errorf("Employee role not found - run migrations to create it")
	}
	if err := uc.roleRepo.AssignRoleToUser(ctx, tx, user.ID, employeeRole.ID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	defer uc.auditor.From(ctx).
		Did(audit.ActionCreate).
		On(audit.EntityEmployee, employee.UID).
		WithMeta("name", employee.Name).
		WithMeta("mobile", employee.Mobile).
		WithMeta("government_id", employee.GovernmentID).
		WithMeta("university_id", employee.UniversityID).
		WithMeta("documents_count", len(documents)).
		Save(ctx)

	return &CreateEmployeeOutput{
		Employee:  employee,
		UserUID:   &user.UID,
		Documents: documents,
	}, nil
}

func (uc *CreateEmployeeUseCase) validateUniqueness(ctx context.Context, input CreateEmployeeInput) error {
	existingEmployee, err := uc.employeeRepo.ExistingMobiles(ctx, uc.db, []string{input.Mobile})
	if err != nil {
		return err
	}
	if len(existingEmployee) > 0 {
		return ErrEmployeeMobileAlreadyExists
	}

	existingGov, err := uc.employeeRepo.ExistingGovernmentIDs(ctx, uc.db, []string{input.GovernmentID})
	if err != nil {
		return err
	}
	if len(existingGov) > 0 {
		return ErrGovernmentIDAlreadyExists
	}

	existingUni, err := uc.employeeRepo.ExistingUniversityIDs(ctx, uc.db, []string{input.UniversityID})
	if err != nil {
		return err
	}
	if len(existingUni) > 0 {
		return ErrUniversityIDAlreadyExists
	}

	existingUser, err := uc.userRepo.GetByPhone(ctx, uc.db, input.Mobile)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return ErrPhoneAlreadyExists
	}

	return nil
}

func (uc *CreateEmployeeUseCase) generateUniversityID(ctx context.Context) (string, error) {
	for {
		candidate := domain.GenerateUID("uni")
		existing, err := uc.employeeRepo.ExistingUniversityIDs(ctx, uc.db, []string{candidate})
		if err != nil {
			return "", err
		}
		if len(existing) == 0 {
			return candidate, nil
		}
	}
}

func (uc *CreateEmployeeUseCase) prepareEmployeeDocuments(
	ctx context.Context,
	userUID string,
	documents []CreateEmployeeDocumentInput,
	employee *domain.Employee,
) ([]CreateEmployeeDocumentOutput, error) {
	return prepareEmployeeDocumentUploads(ctx, uc.documentUploadURLUC, userUID, documents, employee)
}

func minioObjectURL(bucket, objectKey string) string {
	return "minio://" + bucket + "/" + objectKey
}

func trimStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
