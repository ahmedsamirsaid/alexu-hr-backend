package usecases

import (
	"context"
	"strings"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

// GetEmployeeInput defines the input for getting an employee.
type GetEmployeeInput struct {
	UID string
}

// GetEmployeeOutput contains the employee details.
type GetEmployeeOutput struct {
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
	Documents                             []EmployeeDocumentOutput
	AppointmentSeniorityOrGradeWithdrawal bool
	DepartmentUID                         *string
	ShiftUID                              *string
	ManagerUID                            *string
	ManagerName                           *string
	RoleName                              *string // primary non-employee role name; nil for ordinary employees
}

type EmployeeDocumentOutput struct {
	DocumentType string
	FileName     string
	URL          string
	Method       string
	Bucket       string
	ObjectKey    string
}

// GetEmployeeUseCase handles retrieving a single employee by UID.
type GetEmployeeUseCase struct {
	db                 ports.DB
	employeeRepo       ports.EmployeeRepository
	roleRepo           ports.RoleRepository
	documentDownloadUC *GenerateDocumentDownloadURLUseCase
}

// NewGetEmployeeUseCase creates a new get employee use case.
func NewGetEmployeeUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	documentDownloadUC *GenerateDocumentDownloadURLUseCase,
	roleRepo ports.RoleRepository,
) *GetEmployeeUseCase {
	return &GetEmployeeUseCase{
		db:                 db,
		employeeRepo:       employeeRepo,
		roleRepo:           roleRepo,
		documentDownloadUC: documentDownloadUC,
	}
}

// Execute retrieves an employee by UID.
func (uc *GetEmployeeUseCase) Execute(ctx context.Context, input GetEmployeeInput) (*GetEmployeeOutput, error) {
	employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, input.UID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	documents, err := uc.buildEmployeeDocuments(ctx, employee)
	if err != nil {
		return nil, err
	}

	output := &GetEmployeeOutput{
		UID:                                   employee.UID,
		Name:                                  employee.Name,
		Mobile:                                employee.Mobile,
		GovernmentID:                          employee.GovernmentID,
		UniversityID:                          employee.UniversityID,
		Email:                                 employee.Email,
		HireDate:                              employee.HireDate,
		Status:                                employee.Status,
		Type:                                  employee.Type,
		SubType:                               employee.SubType,
		TelephoneNumber:                       employee.TelephoneNumber,
		PersonWithSpecialNeeds:                employee.PersonWithSpecialNeeds,
		DateOfBirth:                           employee.DateOfBirth,
		Gender:                                employee.Gender,
		Religion:                              employee.Religion,
		MaritalStatus:                         employee.MaritalStatus,
		Address:                               employee.Address,
		PlaceOfBirth:                          employee.PlaceOfBirth,
		PlaceOfResidence:                      employee.PlaceOfResidence,
		PoliceStation:                         employee.PoliceStation,
		IDCardValidUntil:                      employee.IDCardValidUntil,
		AcademicLevel:                         employee.AcademicLevel,
		EducationalQualification:              employee.EducationalQualification,
		UniversityName:                        employee.UniversityName,
		Faculty:                               employee.Faculty,
		Specialization:                        employee.Specialization,
		YearObtained:                          employee.YearObtained,
		ActualAppointmentReappointmentDate:    employee.ActualAppointmentReappointmentDate,
		AppointmentDecisionDate:               employee.AppointmentDecisionDate,
		AppointmentDecisionNumber:             employee.AppointmentDecisionNumber,
		AppointmentType:                       employee.AppointmentType,
		DepartmentName:                        employee.DepartmentName,
		Grade:                                 employee.Grade,
		WorkEntity:                            employee.WorkEntity,
		EmployeeFileNumber:                    employee.EmployeeFileNumber,
		InsuranceNumber:                       employee.InsuranceNumber,
		EmploymentStatus:                      employee.EmploymentStatus,
		SolidarityFund:                        employee.SolidarityFund,
		SubscriptionDate:                      employee.SubscriptionDate,
		MilitaryStatus:                        employee.MilitaryStatus,
		MedicalCadre:                          employee.MedicalCadre,
		PersonalPhotoURL:                      employee.PersonalPhotoURL,
		NationalCardImageURL:                  employee.NationalCardImageURL,
		QualificationCertificateImageURL:      employee.QualificationCertificateImageURL,
		CVURL:                                 employee.CVURL,
		MemberNumber:                          employee.MemberNumber,
		InsuranceCode:                         employee.InsuranceCode,
		JobGroup:                              employee.JobGroup,
		QualitativeGroup:                      employee.QualitativeGroup,
		JobTitleAtLevel:                       employee.JobTitleAtLevel,
		JobTitleBeforePlacement:               employee.JobTitleBeforePlacement,
		FinancialGrade:                        employee.FinancialGrade,
		PreviousFinancialGrade:                employee.PreviousFinancialGrade,
		JobLevel:                              employee.JobLevel,
		DecisionDate:                          employee.DecisionDate,
		DecisionNumber:                        employee.DecisionNumber,
		GradeGrantDate:                        employee.GradeGrantDate,
		NatureOfAppointment:                   employee.NatureOfAppointment,
		Notes:                                 employee.Notes,
		Reappointment:                         employee.Reappointment,
		DecisionFileURL:                       employee.DecisionFileURL,
		Documents:                             documents,
		AppointmentSeniorityOrGradeWithdrawal: employee.AppointmentSeniorityOrGradeWithdrawal,
		DepartmentUID:                         employee.DepartmentUID,
		ShiftUID:                              employee.ShiftUID,
		ManagerUID:                            employee.ManagerUID,
	}

	if err := uc.enrichWithManagerAndRole(ctx, output); err != nil {
		return nil, err
	}

	return output, nil
}

func (uc *GetEmployeeUseCase) enrichWithManagerAndRole(ctx context.Context, output *GetEmployeeOutput) error {
	if uc.roleRepo == nil {
		return nil
	}

	uidLookup := []string{output.UID}
	if output.ManagerUID != nil {
		uidLookup = append(uidLookup, *output.ManagerUID)
	}

	roleNames, err := uc.roleRepo.GetRoleNamesByEmployeeUIDs(ctx, uc.db, uidLookup)
	if err != nil {
		return err
	}

	if name := roleNames[output.UID]; name != "" {
		output.RoleName = &name
	}

	if output.ManagerUID != nil {
		mgr, err := uc.employeeRepo.GetByUID(ctx, uc.db, *output.ManagerUID)
		if err != nil {
			return err
		}
		if mgr != nil {
			output.ManagerName = &mgr.Name
		}
	}

	return nil
}

func (uc *GetEmployeeUseCase) buildEmployeeDocuments(ctx context.Context, employee *domain.Employee) ([]EmployeeDocumentOutput, error) {
	documentRefs := []struct {
		documentType string
		fileName     string
		storedURL    *string
	}{
		{documentType: "personal_photo", fileName: "personal_photo", storedURL: employee.PersonalPhotoURL},
		{documentType: "national_card_image", fileName: "national_card_image", storedURL: employee.NationalCardImageURL},
		{documentType: "qualification_certificate_image", fileName: "qualification_certificate_image", storedURL: employee.QualificationCertificateImageURL},
		{documentType: "cv", fileName: "cv", storedURL: employee.CVURL},
		{documentType: "decision_file", fileName: "decision_file", storedURL: employee.DecisionFileURL},
	}

	documents := make([]EmployeeDocumentOutput, 0, len(documentRefs))
	for _, documentRef := range documentRefs {
		if documentRef.storedURL == nil || strings.TrimSpace(*documentRef.storedURL) == "" || uc.documentDownloadUC == nil {
			continue
		}

		objectKey, fileName, ok := parseEmployeeDocumentStoredURL(*documentRef.storedURL)
		if !ok {
			continue
		}
		if fileName == "" {
			fileName = documentRef.fileName
		}

		presigned, err := uc.documentDownloadUC.Execute(ctx, GenerateDocumentDownloadURLInput{
			ObjectKey:        objectKey,
			DownloadFileName: fileName,
		})
		if err != nil {
			return nil, err
		}

		documents = append(documents, EmployeeDocumentOutput{
			DocumentType: documentRef.documentType,
			FileName:     fileName,
			URL:          presigned.URL,
			Method:       presigned.Method,
			Bucket:       presigned.Bucket,
			ObjectKey:    presigned.ObjectKey,
		})
	}

	return documents, nil
}

func parseEmployeeDocumentStoredURL(value string) (objectKey, fileName string, ok bool) {
	const prefix = "minio://"
	if !strings.HasPrefix(value, prefix) {
		return "", "", false
	}

	trimmed := strings.TrimPrefix(value, prefix)
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
		return "", "", false
	}

	objectKey = parts[1]
	lastSlash := strings.LastIndex(objectKey, "/")
	if lastSlash >= 0 && lastSlash < len(objectKey)-1 {
		fileName = objectKey[lastSlash+1:]
	}

	return objectKey, fileName, true
}
