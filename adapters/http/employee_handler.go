package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/usecases"
)

const maxUploadSize = 20 * 1024 * 1024 // 20MB

// EmployeeHandler handles employee HTTP requests.
type EmployeeHandler struct {
	createUC                 *usecases.CreateEmployeeUseCase
	getUC                    *usecases.GetEmployeeUseCase
	listUC                   *usecases.ListEmployeesUseCase
	updateOwnProfileUC       *usecases.UpdateOwnEmployeeProfileUseCase
	importUC                 *usecases.ImportEmployeesUseCase
	exportUC                 *usecases.ExportEmployeesUseCase
	exportPDFUC              *usecases.ExportEmployeesPDFUseCase
	templateUC               *usecases.GenerateImportTemplateUseCase
	assignDeptUC             *usecases.AssignEmployeeDepartmentUseCase
	removeDeptUC             *usecases.RemoveEmployeeDepartmentUseCase
	listPenaltiesUC          *usecases.ListEmployeePenaltiesUseCase
	createPenaltyUC          *usecases.CreateEmployeePenaltyUseCase
	createPenaltyRemovalUC   *usecases.CreateEmployeePenaltyRemovalUseCase
	updatePenaltiesUC        *usecases.UpdateEmployeePenaltiesUseCase
	listIncentiveBonusesUC   *usecases.ListEmployeeIncentiveBonusesUseCase
	createIncentiveBonusUC   *usecases.CreateEmployeeIncentiveBonusUseCase
	updateIncentiveBonusesUC *usecases.UpdateEmployeeIncentiveBonusesUseCase
	listAnnualReportsUC      *usecases.ListEmployeeAnnualReportsUseCase
	createAnnualReportUC     *usecases.CreateEmployeeAnnualReportUseCase
	updateAnnualReportsUC    *usecases.UpdateEmployeeAnnualReportsUseCase
}

// NewEmployeeHandler creates a new employee handler.
func NewEmployeeHandler(
	createUC *usecases.CreateEmployeeUseCase,
	getUC *usecases.GetEmployeeUseCase,
	listUC *usecases.ListEmployeesUseCase,
	updateOwnProfileUC *usecases.UpdateOwnEmployeeProfileUseCase,
	importUC *usecases.ImportEmployeesUseCase,
	exportUC *usecases.ExportEmployeesUseCase,
	exportPDFUC *usecases.ExportEmployeesPDFUseCase,
	templateUC *usecases.GenerateImportTemplateUseCase,
	assignDeptUC *usecases.AssignEmployeeDepartmentUseCase,
	removeDeptUC *usecases.RemoveEmployeeDepartmentUseCase,
	listPenaltiesUC *usecases.ListEmployeePenaltiesUseCase,
	createPenaltyUC *usecases.CreateEmployeePenaltyUseCase,
	createPenaltyRemovalUC *usecases.CreateEmployeePenaltyRemovalUseCase,
	updatePenaltiesUC *usecases.UpdateEmployeePenaltiesUseCase,
	listIncentiveBonusesUC *usecases.ListEmployeeIncentiveBonusesUseCase,
	createIncentiveBonusUC *usecases.CreateEmployeeIncentiveBonusUseCase,
	updateIncentiveBonusesUC *usecases.UpdateEmployeeIncentiveBonusesUseCase,
	listAnnualReportsUC *usecases.ListEmployeeAnnualReportsUseCase,
	createAnnualReportUC *usecases.CreateEmployeeAnnualReportUseCase,
	updateAnnualReportsUC *usecases.UpdateEmployeeAnnualReportsUseCase,
) *EmployeeHandler {
	return &EmployeeHandler{
		createUC:                 createUC,
		getUC:                    getUC,
		listUC:                   listUC,
		updateOwnProfileUC:       updateOwnProfileUC,
		importUC:                 importUC,
		exportUC:                 exportUC,
		exportPDFUC:              exportPDFUC,
		templateUC:               templateUC,
		assignDeptUC:             assignDeptUC,
		removeDeptUC:             removeDeptUC,
		listPenaltiesUC:          listPenaltiesUC,
		createPenaltyUC:          createPenaltyUC,
		createPenaltyRemovalUC:   createPenaltyRemovalUC,
		updatePenaltiesUC:        updatePenaltiesUC,
		listIncentiveBonusesUC:   listIncentiveBonusesUC,
		createIncentiveBonusUC:   createIncentiveBonusUC,
		updateIncentiveBonusesUC: updateIncentiveBonusesUC,
		listAnnualReportsUC:      listAnnualReportsUC,
		createAnnualReportUC:     createAnnualReportUC,
		updateAnnualReportsUC:    updateAnnualReportsUC,
	}
}

type CreateEmployeeRequest struct {
	Name                                  string                       `json:"name"`
	Mobile                                string                       `json:"mobile"`
	GovernmentID                          string                       `json:"governmentId"`
	UniversityID                          string                       `json:"universityId"`
	Email                                 *string                      `json:"email,omitempty"`
	HireDate                              string                       `json:"hireDate"`
	Status                                *string                      `json:"status,omitempty"`
	Type                                  *string                      `json:"type,omitempty"`
	SubType                               *string                      `json:"subType,omitempty"`
	TelephoneNumber                       *string                      `json:"telephoneNumber,omitempty"`
	PersonWithSpecialNeeds                bool                         `json:"personWithSpecialNeeds"`
	DateOfBirth                           *string                      `json:"dateOfBirth,omitempty"`
	Gender                                *string                      `json:"gender,omitempty"`
	Religion                              *string                      `json:"religion,omitempty"`
	MaritalStatus                         *string                      `json:"maritalStatus,omitempty"`
	Address                               *string                      `json:"address,omitempty"`
	PlaceOfBirth                          *string                      `json:"placeOfBirth,omitempty"`
	PlaceOfResidence                      *string                      `json:"placeOfResidence,omitempty"`
	PoliceStation                         *string                      `json:"policeStation,omitempty"`
	IDCardValidUntil                      *string                      `json:"idCardValidUntil,omitempty"`
	AcademicLevel                         *string                      `json:"academicLevel,omitempty"`
	EducationalQualification              *string                      `json:"educationalQualification,omitempty"`
	UniversityName                        *string                      `json:"universityName,omitempty"`
	Faculty                               *string                      `json:"faculty,omitempty"`
	Specialization                        *string                      `json:"specialization,omitempty"`
	YearObtained                          *FlexibleInt                 `json:"yearObtained,omitempty"`
	ActualAppointmentReappointmentDate    *string                      `json:"actualAppointmentReappointmentDate,omitempty"`
	AppointmentDecisionDate               *string                      `json:"appointmentDecisionDate,omitempty"`
	AppointmentDecisionNumber             *string                      `json:"appointmentDecisionNumber,omitempty"`
	AppointmentType                       *string                      `json:"appointmentType,omitempty"`
	DepartmentName                        *string                      `json:"departmentName,omitempty"`
	Grade                                 *string                      `json:"grade,omitempty"`
	WorkEntity                            *string                      `json:"workEntity,omitempty"`
	EmployeeFileNumber                    *string                      `json:"employeeFileNumber,omitempty"`
	InsuranceNumber                       *string                      `json:"insuranceNumber,omitempty"`
	EmploymentStatus                      *string                      `json:"employmentStatus,omitempty"`
	SolidarityFund                        bool                         `json:"solidarityFund"`
	SubscriptionDate                      *string                      `json:"subscriptionDate,omitempty"`
	MilitaryStatus                        *string                      `json:"militaryStatus,omitempty"`
	MedicalCadre                          *string                      `json:"medicalCadre,omitempty"`
	MemberNumber                          *string                      `json:"memberNumber,omitempty"`
	InsuranceCode                         *string                      `json:"insuranceCode,omitempty"`
	JobGroup                              *string                      `json:"jobGroup,omitempty"`
	QualitativeGroup                      *string                      `json:"qualitativeGroup,omitempty"`
	JobTitleAtLevel                       *string                      `json:"jobTitleAtLevel,omitempty"`
	JobTitleBeforePlacement               *string                      `json:"jobTitleBeforePlacement,omitempty"`
	FinancialGrade                        *string                      `json:"financialGrade,omitempty"`
	PreviousFinancialGrade                *string                      `json:"previousFinancialGrade,omitempty"`
	JobLevel                              *string                      `json:"jobLevel,omitempty"`
	DecisionDate                          *string                      `json:"decisionDate,omitempty"`
	DecisionNumber                        *int64                       `json:"decisionNumber,omitempty"`
	GradeGrantDate                        *string                      `json:"gradeGrantDate,omitempty"`
	NatureOfAppointment                   *string                      `json:"natureOfAppointment,omitempty"`
	Notes                                 *string                      `json:"notes,omitempty"`
	Reappointment                         bool                         `json:"reappointment"`
	AppointmentSeniorityOrGradeWithdrawal bool                         `json:"appointmentSeniorityOrGradeWithdrawal"`
	DepartmentUID                         *string                      `json:"departmentUid,omitempty"`
	ShiftUID                              *string                      `json:"shiftUid,omitempty"`
	Documents                             []CreateEmployeeDocumentItem `json:"documents,omitempty"`
}

type CreateEmployeeDocumentItem struct {
	DocumentType string `json:"documentType"`
	FileName     string `json:"fileName"`
	ContentType  string `json:"contentType"`
}

type CreateEmployeeResponse struct {
	UID          string                           `json:"uid"`
	UniversityID string                           `json:"universityId"`
	UserUID      *string                          `json:"userUid,omitempty"`
	Documents    []CreateEmployeeDocumentResponse `json:"documents,omitempty"`
}

type CreateEmployeeDocumentResponse struct {
	DocumentType string `json:"documentType"`
	FileName     string `json:"fileName"`
	URL          string `json:"url"`
	Method       string `json:"method"`
	Bucket       string `json:"bucket"`
	ObjectKey    string `json:"objectKey"`
	StoredURL    string `json:"storedUrl"`
}

type FlexibleInt int

func (v *FlexibleInt) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" {
		return nil
	}

	if strings.HasPrefix(trimmed, `"`) {
		var raw string
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return nil
		}
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return err
		}
		*v = FlexibleInt(parsed)
		return nil
	}

	var parsed int
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	*v = FlexibleInt(parsed)
	return nil
}

type EmployeeDocumentDownloadResponse struct {
	DocumentType string `json:"documentType"`
	FileName     string `json:"fileName"`
	URL          string `json:"url"`
	Method       string `json:"method"`
	Bucket       string `json:"bucket"`
	ObjectKey    string `json:"objectKey"`
}

type EmployeePenaltyRequest struct {
	PenaltyType            string                           `json:"penaltyType"`
	PenaltyReason          string                           `json:"penaltyReason"`
	PenaltyDecisionNumber  string                           `json:"penaltyDecisionNumber"`
	PenaltyDecisionDate    string                           `json:"penaltyDecisionDate"`
	PenaltyDecisionFile    *EmployeePenaltyDecisionFileItem `json:"penaltyDecisionFile,omitempty"`
	PenaltyDecisionFileURL string                           `json:"penaltyDecisionFileUrl,omitempty"`
}

type EmployeePenaltyDecisionFileItem struct {
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
}

type UpdateEmployeePenaltiesRequest struct {
	Penalties []EmployeePenaltyRequest `json:"penalties"`
}

type EmployeePenaltyResponse struct {
	ID                     int64                                      `json:"id"`
	PenaltyType            string                                     `json:"penaltyType"`
	PenaltyReason          string                                     `json:"penaltyReason"`
	PenaltyDecisionNumber  string                                     `json:"penaltyDecisionNumber"`
	PenaltyDecisionDate    string                                     `json:"penaltyDecisionDate"`
	PenaltyDecisionFileURL string                                     `json:"penaltyDecisionFileUrl"`
	PenaltyDecisionFile    *EmployeePenaltyDecisionFileUploadResponse `json:"penaltyDecisionFile,omitempty"`
}

type EmployeePenaltyDecisionFileUploadResponse struct {
	FileName  string `json:"fileName"`
	URL       string `json:"url"`
	Method    string `json:"method"`
	Bucket    string `json:"bucket"`
	ObjectKey string `json:"objectKey"`
	StoredURL string `json:"storedUrl"`
	ExpiresAt string `json:"expiresAt"`
}

type CreateEmployeePenaltyResponse struct {
	ID                     int64                                      `json:"id"`
	PenaltyType            string                                     `json:"penaltyType"`
	PenaltyReason          string                                     `json:"penaltyReason"`
	PenaltyDecisionNumber  string                                     `json:"penaltyDecisionNumber"`
	PenaltyDecisionDate    string                                     `json:"penaltyDecisionDate"`
	PenaltyDecisionFileURL string                                     `json:"penaltyDecisionFileUrl"`
	PenaltyDecisionFile    *EmployeePenaltyDecisionFileUploadResponse `json:"penaltyDecisionFile,omitempty"`
}

type EmployeePenaltiesResponse struct {
	Penalties []EmployeePenaltyResponse `json:"penalties"`
}

type EmployeePenaltyRemovalRequest struct {
	PenaltyRemovalType            string                           `json:"penaltyRemovalType"`
	PenaltyRemovalNumber          string                           `json:"penaltyRemovalNumber"`
	PenaltyRemovalDate            string                           `json:"penaltyRemovalDate"`
	Notes                         string                           `json:"notes"`
	PenaltyWithdrawalDecisionFile *EmployeePenaltyDecisionFileItem `json:"penaltyWithdrawalDecisionFile"`
}

type EmployeePenaltyRemovalResponse struct {
	ID                               int64                                      `json:"id"`
	PenaltyID                        int64                                      `json:"penaltyId"`
	PenaltyRemovalType               string                                     `json:"penaltyRemovalType"`
	PenaltyRemovalNumber             string                                     `json:"penaltyRemovalNumber"`
	PenaltyRemovalDate               string                                     `json:"penaltyRemovalDate"`
	Notes                            string                                     `json:"notes"`
	PenaltyWithdrawalDecisionFileURL string                                     `json:"penaltyWithdrawalDecisionFileUrl"`
	PenaltyWithdrawalDecisionFile    *EmployeePenaltyDecisionFileUploadResponse `json:"penaltyWithdrawalDecisionFile,omitempty"`
}

type EmployeeAnnualReportRequest struct {
	ReportYear     string                         `json:"reportYear"`
	ReportGrade    string                         `json:"reportGrade"`
	Notes          string                         `json:"notes"`
	ReportImage    *EmployeeAnnualReportImageItem `json:"reportImage,omitempty"`
	ReportImageURL string                         `json:"reportImageUrl,omitempty"`
}

type EmployeeAnnualReportImageItem struct {
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
}

type UpdateEmployeeAnnualReportsRequest struct {
	Reports []EmployeeAnnualReportRequest `json:"reports"`
}

type EmployeeAnnualReportImageResponse struct {
	FileName  string `json:"fileName"`
	URL       string `json:"url"`
	Method    string `json:"method"`
	Bucket    string `json:"bucket"`
	ObjectKey string `json:"objectKey"`
	StoredURL string `json:"storedUrl"`
	ExpiresAt string `json:"expiresAt"`
}

type EmployeeAnnualReportResponse struct {
	ReportYear     string                             `json:"reportYear"`
	ReportGrade    string                             `json:"reportGrade"`
	Notes          string                             `json:"notes"`
	ReportImageURL string                             `json:"reportImageUrl"`
	ReportImage    *EmployeeAnnualReportImageResponse `json:"reportImage,omitempty"`
}

type CreateEmployeeAnnualReportResponse struct {
	ReportYear     string                             `json:"reportYear"`
	ReportGrade    string                             `json:"reportGrade"`
	Notes          string                             `json:"notes"`
	ReportImageURL string                             `json:"reportImageUrl"`
	ReportImage    *EmployeeAnnualReportImageResponse `json:"reportImage,omitempty"`
}

type EmployeeAnnualReportsResponse struct {
	Reports []EmployeeAnnualReportResponse `json:"reports"`
}

type EmployeeIncentiveBonusRequest struct {
	BonusDate        string                           `json:"bonusDate"`
	DecisionNumber   string                           `json:"decisionNumber"`
	DecisionDate     string                           `json:"decisionDate"`
	DecisionImage    *EmployeePenaltyDecisionFileItem `json:"decisionImage,omitempty"`
	DecisionImageURL string                           `json:"decisionImageUrl,omitempty"`
}

type UpdateEmployeeIncentiveBonusesRequest struct {
	Bonuses []EmployeeIncentiveBonusRequest `json:"bonuses"`
}

type EmployeeIncentiveBonusDecisionImageResponse struct {
	FileName  string `json:"fileName"`
	URL       string `json:"url"`
	Method    string `json:"method"`
	Bucket    string `json:"bucket"`
	ObjectKey string `json:"objectKey"`
	StoredURL string `json:"storedUrl"`
	ExpiresAt string `json:"expiresAt"`
}

type EmployeeIncentiveBonusResponse struct {
	ID               int64                                        `json:"id"`
	BonusDate        string                                       `json:"bonusDate"`
	DecisionNumber   string                                       `json:"decisionNumber"`
	DecisionDate     string                                       `json:"decisionDate"`
	DecisionImageURL string                                       `json:"decisionImageUrl"`
	DecisionImage    *EmployeeIncentiveBonusDecisionImageResponse `json:"decisionImage,omitempty"`
}

type CreateEmployeeIncentiveBonusResponse struct {
	ID               int64                                        `json:"id"`
	BonusDate        string                                       `json:"bonusDate"`
	DecisionNumber   string                                       `json:"decisionNumber"`
	DecisionDate     string                                       `json:"decisionDate"`
	DecisionImageURL string                                       `json:"decisionImageUrl"`
	DecisionImage    *EmployeeIncentiveBonusDecisionImageResponse `json:"decisionImage,omitempty"`
}

type EmployeeIncentiveBonusesResponse struct {
	Bonuses []EmployeeIncentiveBonusResponse `json:"bonuses"`
}

// GetEmployeeResponse represents the get employee API response.
type GetEmployeeResponse struct {
	UID                                   string                             `json:"uid"`
	Name                                  string                             `json:"name"`
	Mobile                                string                             `json:"mobile"`
	GovernmentID                          string                             `json:"governmentId"`
	UniversityID                          string                             `json:"universityId"`
	Email                                 *string                            `json:"email,omitempty"`
	HireDate                              string                             `json:"hireDate"`
	Status                                string                             `json:"status"`
	Type                                  string                             `json:"type"`
	SubType                               string                             `json:"subType"`
	TelephoneNumber                       *string                            `json:"telephoneNumber,omitempty"`
	PersonWithSpecialNeeds                bool                               `json:"personWithSpecialNeeds"`
	DateOfBirth                           *string                            `json:"dateOfBirth,omitempty"`
	Gender                                *string                            `json:"gender,omitempty"`
	Religion                              *string                            `json:"religion,omitempty"`
	MaritalStatus                         *string                            `json:"maritalStatus,omitempty"`
	Address                               *string                            `json:"address,omitempty"`
	PlaceOfBirth                          *string                            `json:"placeOfBirth,omitempty"`
	PlaceOfResidence                      *string                            `json:"placeOfResidence,omitempty"`
	PoliceStation                         *string                            `json:"policeStation,omitempty"`
	IDCardValidUntil                      *string                            `json:"idCardValidUntil,omitempty"`
	AcademicLevel                         *string                            `json:"academicLevel,omitempty"`
	EducationalQualification              *string                            `json:"educationalQualification,omitempty"`
	UniversityName                        *string                            `json:"universityName,omitempty"`
	Faculty                               *string                            `json:"faculty,omitempty"`
	Specialization                        *string                            `json:"specialization,omitempty"`
	YearObtained                          *int                               `json:"yearObtained,omitempty"`
	ActualAppointmentReappointmentDate    *string                            `json:"actualAppointmentReappointmentDate,omitempty"`
	AppointmentDecisionDate               *string                            `json:"appointmentDecisionDate,omitempty"`
	AppointmentDecisionNumber             *string                            `json:"appointmentDecisionNumber,omitempty"`
	AppointmentType                       *string                            `json:"appointmentType,omitempty"`
	DepartmentName                        *string                            `json:"departmentName,omitempty"`
	Grade                                 *string                            `json:"grade,omitempty"`
	WorkEntity                            *string                            `json:"workEntity,omitempty"`
	EmployeeFileNumber                    *string                            `json:"employeeFileNumber,omitempty"`
	InsuranceNumber                       *string                            `json:"insuranceNumber,omitempty"`
	EmploymentStatus                      *string                            `json:"employmentStatus,omitempty"`
	SolidarityFund                        bool                               `json:"solidarityFund"`
	SubscriptionDate                      *string                            `json:"subscriptionDate,omitempty"`
	MilitaryStatus                        *string                            `json:"militaryStatus,omitempty"`
	MedicalCadre                          *string                            `json:"medicalCadre,omitempty"`
	PersonalPhotoURL                      *string                            `json:"personalPhotoUrl,omitempty"`
	NationalCardImageURL                  *string                            `json:"nationalCardImageUrl,omitempty"`
	QualificationCertificateImageURL      *string                            `json:"qualificationCertificateImageUrl,omitempty"`
	CVURL                                 *string                            `json:"cvUrl,omitempty"`
	MemberNumber                          *string                            `json:"memberNumber,omitempty"`
	InsuranceCode                         *string                            `json:"insuranceCode,omitempty"`
	JobGroup                              *string                            `json:"jobGroup,omitempty"`
	QualitativeGroup                      *string                            `json:"qualitativeGroup,omitempty"`
	JobTitleAtLevel                       *string                            `json:"jobTitleAtLevel,omitempty"`
	JobTitleBeforePlacement               *string                            `json:"jobTitleBeforePlacement,omitempty"`
	FinancialGrade                        *string                            `json:"financialGrade,omitempty"`
	PreviousFinancialGrade                *string                            `json:"previousFinancialGrade,omitempty"`
	JobLevel                              *string                            `json:"jobLevel,omitempty"`
	DecisionDate                          *string                            `json:"decisionDate,omitempty"`
	DecisionNumber                        *int64                             `json:"decisionNumber,omitempty"`
	GradeGrantDate                        *string                            `json:"gradeGrantDate,omitempty"`
	NatureOfAppointment                   *string                            `json:"natureOfAppointment,omitempty"`
	Notes                                 *string                            `json:"notes,omitempty"`
	Reappointment                         bool                               `json:"reappointment"`
	DecisionFileURL                       *string                            `json:"decisionFileUrl,omitempty"`
	Documents                             []EmployeeDocumentDownloadResponse `json:"documents,omitempty"`
	AppointmentSeniorityOrGradeWithdrawal bool                               `json:"appointmentSeniorityOrGradeWithdrawal"`
	DepartmentUID                         *string                            `json:"departmentUid,omitempty"`
	ShiftUID                              *string                            `json:"shiftUid,omitempty"`
}

type UpdateOwnEmployeeProfileRequest struct {
	Name            string  `json:"name"`
	Mobile          string  `json:"mobile"`
	TelephoneNumber *string `json:"telephoneNumber,omitempty"`
	Email           *string `json:"email,omitempty"`
}

type UpdateOwnEmployeeProfileResponse struct {
	UID             string  `json:"uid"`
	Name            string  `json:"name"`
	Mobile          string  `json:"mobile"`
	TelephoneNumber *string `json:"telephoneNumber,omitempty"`
	Email           *string `json:"email,omitempty"`
}

// EmployeeListResponse represents the list employees API response.
type EmployeeListResponse struct {
	Employees []EmployeeListItemResponse `json:"employees"`
}

// EmployeeUserInfoResponse represents linked user account info in the response.
type EmployeeUserInfoResponse struct {
	UserUID  string   `json:"userUid"`
	Phone    string   `json:"phone"`
	IsActive bool     `json:"isActive"`
	Roles    []string `json:"roles"`
}

// EmployeeListItemResponse represents an employee in the list response.
type EmployeeListItemResponse struct {
	UID                                   string                    `json:"uid"`
	Name                                  string                    `json:"name"`
	Mobile                                string                    `json:"mobile"`
	GovernmentID                          string                    `json:"governmentId"`
	UniversityID                          string                    `json:"universityId"`
	Email                                 *string                   `json:"email,omitempty"`
	HireDate                              string                    `json:"hireDate"`
	Status                                string                    `json:"status"`
	Type                                  string                    `json:"type"`
	SubType                               string                    `json:"subType"`
	TelephoneNumber                       *string                   `json:"telephoneNumber,omitempty"`
	PersonWithSpecialNeeds                bool                      `json:"personWithSpecialNeeds"`
	DateOfBirth                           *string                   `json:"dateOfBirth,omitempty"`
	Gender                                *string                   `json:"gender,omitempty"`
	Religion                              *string                   `json:"religion,omitempty"`
	MaritalStatus                         *string                   `json:"maritalStatus,omitempty"`
	Address                               *string                   `json:"address,omitempty"`
	PlaceOfBirth                          *string                   `json:"placeOfBirth,omitempty"`
	PlaceOfResidence                      *string                   `json:"placeOfResidence,omitempty"`
	PoliceStation                         *string                   `json:"policeStation,omitempty"`
	IDCardValidUntil                      *string                   `json:"idCardValidUntil,omitempty"`
	AcademicLevel                         *string                   `json:"academicLevel,omitempty"`
	EducationalQualification              *string                   `json:"educationalQualification,omitempty"`
	UniversityName                        *string                   `json:"universityName,omitempty"`
	Faculty                               *string                   `json:"faculty,omitempty"`
	Specialization                        *string                   `json:"specialization,omitempty"`
	YearObtained                          *int                      `json:"yearObtained,omitempty"`
	ActualAppointmentReappointmentDate    *string                   `json:"actualAppointmentReappointmentDate,omitempty"`
	AppointmentDecisionDate               *string                   `json:"appointmentDecisionDate,omitempty"`
	AppointmentDecisionNumber             *string                   `json:"appointmentDecisionNumber,omitempty"`
	AppointmentType                       *string                   `json:"appointmentType,omitempty"`
	DepartmentName                        *string                   `json:"departmentName,omitempty"`
	Grade                                 *string                   `json:"grade,omitempty"`
	WorkEntity                            *string                   `json:"workEntity,omitempty"`
	EmployeeFileNumber                    *string                   `json:"employeeFileNumber,omitempty"`
	InsuranceNumber                       *string                   `json:"insuranceNumber,omitempty"`
	EmploymentStatus                      *string                   `json:"employmentStatus,omitempty"`
	SolidarityFund                        bool                      `json:"solidarityFund"`
	SubscriptionDate                      *string                   `json:"subscriptionDate,omitempty"`
	MilitaryStatus                        *string                   `json:"militaryStatus,omitempty"`
	MedicalCadre                          *string                   `json:"medicalCadre,omitempty"`
	PersonalPhotoURL                      *string                   `json:"personalPhotoUrl,omitempty"`
	NationalCardImageURL                  *string                   `json:"nationalCardImageUrl,omitempty"`
	QualificationCertificateImageURL      *string                   `json:"qualificationCertificateImageUrl,omitempty"`
	CVURL                                 *string                   `json:"cvUrl,omitempty"`
	MemberNumber                          *string                   `json:"memberNumber,omitempty"`
	InsuranceCode                         *string                   `json:"insuranceCode,omitempty"`
	JobGroup                              *string                   `json:"jobGroup,omitempty"`
	QualitativeGroup                      *string                   `json:"qualitativeGroup,omitempty"`
	JobTitleAtLevel                       *string                   `json:"jobTitleAtLevel,omitempty"`
	JobTitleBeforePlacement               *string                   `json:"jobTitleBeforePlacement,omitempty"`
	FinancialGrade                        *string                   `json:"financialGrade,omitempty"`
	PreviousFinancialGrade                *string                   `json:"previousFinancialGrade,omitempty"`
	JobLevel                              *string                   `json:"jobLevel,omitempty"`
	DecisionDate                          *string                   `json:"decisionDate,omitempty"`
	DecisionNumber                        *int64                    `json:"decisionNumber,omitempty"`
	GradeGrantDate                        *string                   `json:"gradeGrantDate,omitempty"`
	NatureOfAppointment                   *string                   `json:"natureOfAppointment,omitempty"`
	Notes                                 *string                   `json:"notes,omitempty"`
	Reappointment                         bool                      `json:"reappointment"`
	DecisionFileURL                       *string                   `json:"decisionFileUrl,omitempty"`
	AppointmentSeniorityOrGradeWithdrawal bool                      `json:"appointmentSeniorityOrGradeWithdrawal"`
	DepartmentUID                         *string                   `json:"departmentUid,omitempty"`
	ShiftUID                              *string                   `json:"shiftUid,omitempty"`
	User                                  *EmployeeUserInfoResponse `json:"user,omitempty"`
}

// Create handles POST /api/v1/employees
func (h *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("employee_handler.Create.decode_request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Mobile == "" {
		writeError(w, http.StatusBadRequest, "mobile is required")
		return
	}
	if req.GovernmentID == "" {
		writeError(w, http.StatusBadRequest, "governmentId is required")
		return
	}
	if req.HireDate == "" {
		writeError(w, http.StatusBadRequest, "hireDate is required")
		return
	}

	hireDate, err := time.Parse("2006-01-02", req.HireDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid hireDate format, expected YYYY-MM-DD")
		return
	}

	input := usecases.CreateEmployeeInput{
		Name:                                  req.Name,
		Mobile:                                req.Mobile,
		GovernmentID:                          req.GovernmentID,
		UniversityID:                          req.UniversityID,
		Email:                                 req.Email,
		HireDate:                              hireDate,
		TelephoneNumber:                       req.TelephoneNumber,
		PersonWithSpecialNeeds:                req.PersonWithSpecialNeeds,
		Gender:                                req.Gender,
		Religion:                              req.Religion,
		MaritalStatus:                         req.MaritalStatus,
		Address:                               req.Address,
		PlaceOfBirth:                          req.PlaceOfBirth,
		PlaceOfResidence:                      req.PlaceOfResidence,
		PoliceStation:                         req.PoliceStation,
		AcademicLevel:                         req.AcademicLevel,
		EducationalQualification:              req.EducationalQualification,
		UniversityName:                        req.UniversityName,
		Faculty:                               req.Faculty,
		Specialization:                        req.Specialization,
		ActualAppointmentReappointmentDate:    nil,
		AppointmentDecisionDate:               nil,
		AppointmentDecisionNumber:             req.AppointmentDecisionNumber,
		AppointmentType:                       req.AppointmentType,
		DepartmentName:                        req.DepartmentName,
		Grade:                                 req.Grade,
		WorkEntity:                            req.WorkEntity,
		EmployeeFileNumber:                    req.EmployeeFileNumber,
		InsuranceNumber:                       req.InsuranceNumber,
		EmploymentStatus:                      req.EmploymentStatus,
		SolidarityFund:                        req.SolidarityFund,
		MilitaryStatus:                        req.MilitaryStatus,
		MedicalCadre:                          req.MedicalCadre,
		MemberNumber:                          req.MemberNumber,
		InsuranceCode:                         req.InsuranceCode,
		JobGroup:                              req.JobGroup,
		QualitativeGroup:                      req.QualitativeGroup,
		JobTitleAtLevel:                       req.JobTitleAtLevel,
		JobTitleBeforePlacement:               req.JobTitleBeforePlacement,
		FinancialGrade:                        req.FinancialGrade,
		PreviousFinancialGrade:                req.PreviousFinancialGrade,
		JobLevel:                              req.JobLevel,
		DecisionNumber:                        req.DecisionNumber,
		NatureOfAppointment:                   req.NatureOfAppointment,
		Notes:                                 req.Notes,
		Reappointment:                         req.Reappointment,
		AppointmentSeniorityOrGradeWithdrawal: req.AppointmentSeniorityOrGradeWithdrawal,
		DepartmentUID:                         req.DepartmentUID,
		ShiftUID:                              req.ShiftUID,
		Documents:                             make([]usecases.CreateEmployeeDocumentInput, 0, len(req.Documents)),
	}
	if req.YearObtained != nil {
		yearObtained := int(*req.YearObtained)
		input.YearObtained = &yearObtained
	}

	if req.Status != nil && *req.Status != "" {
		status := domain.EmployeeStatus(*req.Status)
		switch status {
		case domain.EmployeeStatusActive, domain.EmployeeStatusInactive, domain.EmployeeStatusTerminated:
			input.Status = status
		default:
			writeError(w, http.StatusBadRequest, "invalid status value")
			return
		}
	}

	if req.Type != nil && *req.Type != "" {
		employeeType := domain.NormalizeEmployeeType(*req.Type)
		if employeeType == "" {
			writeError(w, http.StatusBadRequest, "invalid type value")
			return
		}
		input.Type = employeeType
	}

	if req.SubType != nil && *req.SubType != "" {
		subType := domain.NormalizeEmployeeSubType(*req.SubType)
		if subType == "" {
			writeError(w, http.StatusBadRequest, "invalid subType value")
			return
		}
		input.SubType = subType
	}

	if input.Type != "" && input.SubType != "" && !domain.IsValidEmployeeSubTypeForType(input.Type, input.SubType) {
		writeError(w, http.StatusBadRequest, "subType does not match type")
		return
	}

	input.DateOfBirth, err = parseOptionalDate(req.DateOfBirth)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid dateOfBirth format, expected YYYY-MM-DD")
		return
	}
	input.IDCardValidUntil, err = parseOptionalDate(req.IDCardValidUntil)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid idCardValidUntil format, expected YYYY-MM-DD")
		return
	}
	input.SubscriptionDate, err = parseOptionalDate(req.SubscriptionDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid subscriptionDate format, expected YYYY-MM-DD")
		return
	}
	input.ActualAppointmentReappointmentDate, err = parseOptionalDate(req.ActualAppointmentReappointmentDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid actualAppointmentReappointmentDate format, expected YYYY-MM-DD")
		return
	}
	input.AppointmentDecisionDate, err = parseOptionalDate(req.AppointmentDecisionDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid appointmentDecisionDate format, expected YYYY-MM-DD")
		return
	}
	input.DecisionDate, err = parseOptionalDate(req.DecisionDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid decisionDate format, expected YYYY-MM-DD")
		return
	}
	input.GradeGrantDate, err = parseOptionalDate(req.GradeGrantDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid gradeGrantDate format, expected YYYY-MM-DD")
		return
	}

	for _, document := range req.Documents {
		input.Documents = append(input.Documents, usecases.CreateEmployeeDocumentInput{
			DocumentType: document.DocumentType,
			FileName:     document.FileName,
			ContentType:  document.ContentType,
		})
	}

	output, err := h.createUC.Execute(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrEmployeeMobileAlreadyExists):
			writeError(w, http.StatusConflict, err.Error())
		case errors.Is(err, usecases.ErrGovernmentIDAlreadyExists):
			writeError(w, http.StatusConflict, err.Error())
		case errors.Is(err, usecases.ErrUniversityIDAlreadyExists):
			writeError(w, http.StatusConflict, err.Error())
		case errors.Is(err, usecases.ErrPhoneAlreadyExists):
			writeError(w, http.StatusConflict, err.Error())
		case errors.Is(err, usecases.ErrEmployeeDocumentTypeInvalid):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, usecases.ErrEmployeeDocumentDuplicate):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, usecases.ErrInvalidEmployeeClassification):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, usecases.ErrInvalidFilename), errors.Is(err, usecases.ErrInvalidContentType):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			slog.Error("employee_handler.Create.execute_usecase", "error", err)
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusCreated, CreateEmployeeResponse{
		UID:          output.Employee.UID,
		UniversityID: output.Employee.UniversityID,
		UserUID:      output.UserUID,
		Documents:    toCreateEmployeeDocumentResponses(output.Documents),
	})
}

// GetEmployee handles GET /api/v1/employees/{uid}
func (h *EmployeeHandler) GetEmployee(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	input := usecases.GetEmployeeInput{UID: uid}
	output, err := h.getUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrEmployeeNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("employee_handler.GetEmployee.execute_usecase", "error", err)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	if isSelfScopedClaims(claims) {
		if claims.EmployeeUID == nil || *claims.EmployeeUID != output.UID {
			writeJSONError(w, http.StatusForbidden, "permission_denied", "Access to this employee is not permitted")
			return
		}
	}

	if isScopedDepartmentClaims(claims) {
		if output.DepartmentUID == nil || !claims.HasDepartmentAccess(*output.DepartmentUID) {
			writeJSONError(w, http.StatusForbidden, "permission_denied", "Access to this employee is not permitted")
			return
		}
	}

	writeJSON(w, http.StatusOK, GetEmployeeResponse{
		UID:                                   output.UID,
		Name:                                  output.Name,
		Mobile:                                output.Mobile,
		GovernmentID:                          output.GovernmentID,
		UniversityID:                          output.UniversityID,
		Email:                                 output.Email,
		HireDate:                              output.HireDate.Format("2006-01-02"),
		Status:                                string(output.Status),
		Type:                                  string(output.Type),
		SubType:                               string(output.SubType),
		TelephoneNumber:                       output.TelephoneNumber,
		PersonWithSpecialNeeds:                output.PersonWithSpecialNeeds,
		DateOfBirth:                           formatDatePtr(output.DateOfBirth),
		Gender:                                output.Gender,
		Religion:                              output.Religion,
		MaritalStatus:                         output.MaritalStatus,
		Address:                               output.Address,
		PlaceOfBirth:                          output.PlaceOfBirth,
		PlaceOfResidence:                      output.PlaceOfResidence,
		PoliceStation:                         output.PoliceStation,
		IDCardValidUntil:                      formatDatePtr(output.IDCardValidUntil),
		AcademicLevel:                         output.AcademicLevel,
		EducationalQualification:              output.EducationalQualification,
		UniversityName:                        output.UniversityName,
		Faculty:                               output.Faculty,
		Specialization:                        output.Specialization,
		YearObtained:                          output.YearObtained,
		ActualAppointmentReappointmentDate:    formatDatePtr(output.ActualAppointmentReappointmentDate),
		AppointmentDecisionDate:               formatDatePtr(output.AppointmentDecisionDate),
		AppointmentDecisionNumber:             output.AppointmentDecisionNumber,
		AppointmentType:                       output.AppointmentType,
		DepartmentName:                        output.DepartmentName,
		Grade:                                 output.Grade,
		WorkEntity:                            output.WorkEntity,
		EmployeeFileNumber:                    output.EmployeeFileNumber,
		InsuranceNumber:                       output.InsuranceNumber,
		EmploymentStatus:                      output.EmploymentStatus,
		SolidarityFund:                        output.SolidarityFund,
		SubscriptionDate:                      formatDatePtr(output.SubscriptionDate),
		MilitaryStatus:                        output.MilitaryStatus,
		MedicalCadre:                          output.MedicalCadre,
		PersonalPhotoURL:                      output.PersonalPhotoURL,
		NationalCardImageURL:                  output.NationalCardImageURL,
		QualificationCertificateImageURL:      output.QualificationCertificateImageURL,
		CVURL:                                 output.CVURL,
		MemberNumber:                          output.MemberNumber,
		InsuranceCode:                         output.InsuranceCode,
		JobGroup:                              output.JobGroup,
		QualitativeGroup:                      output.QualitativeGroup,
		JobTitleAtLevel:                       output.JobTitleAtLevel,
		JobTitleBeforePlacement:               output.JobTitleBeforePlacement,
		FinancialGrade:                        output.FinancialGrade,
		PreviousFinancialGrade:                output.PreviousFinancialGrade,
		JobLevel:                              output.JobLevel,
		DecisionDate:                          formatDatePtr(output.DecisionDate),
		DecisionNumber:                        output.DecisionNumber,
		GradeGrantDate:                        formatDatePtr(output.GradeGrantDate),
		NatureOfAppointment:                   output.NatureOfAppointment,
		Notes:                                 output.Notes,
		Reappointment:                         output.Reappointment,
		DecisionFileURL:                       output.DecisionFileURL,
		Documents:                             toEmployeeDocumentDownloadResponses(output.Documents),
		AppointmentSeniorityOrGradeWithdrawal: output.AppointmentSeniorityOrGradeWithdrawal,
		DepartmentUID:                         output.DepartmentUID,
		ShiftUID:                              output.ShiftUID,
	})
}

// ListEmployees handles GET /api/v1/employees
func (h *EmployeeHandler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	allowFullEmployeeList := claims != nil && claims.HasRole("hr_staff")
	input, err := parseListFilters(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if !allowFullEmployeeList && isScopedDepartmentClaims(claims) {
		input.ManagedDepartmentUIDs = append([]string(nil), claims.ManagedDepartmentUIDs...)
	}

	output, err := h.listUC.Execute(r.Context(), input)
	if err != nil {
		slog.Error("employee_handler.ListEmployees.execute_usecase", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !allowFullEmployeeList && isSelfScopedClaims(claims) {
		if claims.EmployeeUID == nil {
			writeJSON(w, http.StatusOK, EmployeeListResponse{Employees: []EmployeeListItemResponse{}})
			return
		}

		filtered := output.Employees[:0]
		for _, emp := range output.Employees {
			if emp.UID == *claims.EmployeeUID {
				filtered = append(filtered, emp)
			}
		}
		output.Employees = filtered
	}

	employees := make([]EmployeeListItemResponse, 0, len(output.Employees))
	for _, emp := range output.Employees {
		item := EmployeeListItemResponse{
			UID:                                   emp.UID,
			Name:                                  emp.Name,
			Mobile:                                emp.Mobile,
			GovernmentID:                          emp.GovernmentID,
			UniversityID:                          emp.UniversityID,
			Email:                                 emp.Email,
			HireDate:                              emp.HireDate.Format("2006-01-02"),
			Status:                                string(emp.Status),
			Type:                                  string(emp.Type),
			SubType:                               string(emp.SubType),
			TelephoneNumber:                       emp.TelephoneNumber,
			PersonWithSpecialNeeds:                emp.PersonWithSpecialNeeds,
			DateOfBirth:                           formatDatePtr(emp.DateOfBirth),
			Gender:                                emp.Gender,
			Religion:                              emp.Religion,
			MaritalStatus:                         emp.MaritalStatus,
			Address:                               emp.Address,
			PlaceOfBirth:                          emp.PlaceOfBirth,
			PlaceOfResidence:                      emp.PlaceOfResidence,
			PoliceStation:                         emp.PoliceStation,
			IDCardValidUntil:                      formatDatePtr(emp.IDCardValidUntil),
			AcademicLevel:                         emp.AcademicLevel,
			EducationalQualification:              emp.EducationalQualification,
			UniversityName:                        emp.UniversityName,
			Faculty:                               emp.Faculty,
			Specialization:                        emp.Specialization,
			YearObtained:                          emp.YearObtained,
			ActualAppointmentReappointmentDate:    formatDatePtr(emp.ActualAppointmentReappointmentDate),
			AppointmentDecisionDate:               formatDatePtr(emp.AppointmentDecisionDate),
			AppointmentDecisionNumber:             emp.AppointmentDecisionNumber,
			AppointmentType:                       emp.AppointmentType,
			DepartmentName:                        emp.DepartmentName,
			Grade:                                 emp.Grade,
			WorkEntity:                            emp.WorkEntity,
			EmployeeFileNumber:                    emp.EmployeeFileNumber,
			InsuranceNumber:                       emp.InsuranceNumber,
			EmploymentStatus:                      emp.EmploymentStatus,
			SolidarityFund:                        emp.SolidarityFund,
			SubscriptionDate:                      formatDatePtr(emp.SubscriptionDate),
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
			DecisionDate:                          formatDatePtr(emp.DecisionDate),
			DecisionNumber:                        emp.DecisionNumber,
			GradeGrantDate:                        formatDatePtr(emp.GradeGrantDate),
			NatureOfAppointment:                   emp.NatureOfAppointment,
			Notes:                                 emp.Notes,
			Reappointment:                         emp.Reappointment,
			DecisionFileURL:                       emp.DecisionFileURL,
			AppointmentSeniorityOrGradeWithdrawal: emp.AppointmentSeniorityOrGradeWithdrawal,
			DepartmentUID:                         emp.DepartmentUID,
			ShiftUID:                              emp.ShiftUID,
		}
		if emp.User != nil {
			item.User = &EmployeeUserInfoResponse{
				UserUID:  emp.User.UserUID,
				Phone:    emp.User.Phone,
				IsActive: emp.User.IsActive,
				Roles:    emp.User.Roles,
			}
		}
		employees = append(employees, item)
	}

	writeJSON(w, http.StatusOK, EmployeeListResponse{Employees: employees})
}

// UpdateOwnProfile handles PUT /api/v1/employees/me/profile
func (h *EmployeeHandler) UpdateOwnProfile(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}
	if claims.EmployeeUID == nil || *claims.EmployeeUID == "" {
		writeJSONError(w, http.StatusBadRequest, "no_employee_linked", "No employee profile linked to user")
		return
	}

	var req UpdateOwnEmployeeProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	output, err := h.updateOwnProfileUC.Execute(r.Context(), usecases.UpdateOwnEmployeeProfileInput{
		UserID:          claims.UserID,
		Name:            req.Name,
		Mobile:          req.Mobile,
		TelephoneNumber: req.TelephoneNumber,
		Email:           req.Email,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrNoEmployeeLinked):
			writeJSONError(w, http.StatusBadRequest, "no_employee_linked", "No employee profile linked to user")
		case errors.Is(err, usecases.ErrEmployeeNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, usecases.ErrEmployeeMobileAlreadyExists):
			writeError(w, http.StatusConflict, err.Error())
		case errors.Is(err, usecases.ErrPhoneAlreadyExists):
			writeError(w, http.StatusConflict, err.Error())
		case err.Error() == "name is required", err.Error() == "mobile is required":
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			slog.Error("employee_handler.UpdateOwnProfile.execute_usecase", "error", err)
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, UpdateOwnEmployeeProfileResponse{
		UID:             output.UID,
		Name:            output.Name,
		Mobile:          output.Mobile,
		TelephoneNumber: output.TelephoneNumber,
		Email:           output.Email,
	})
}

// ImportEmployees handles POST /api/v1/employees/import
func (h *EmployeeHandler) ImportEmployees(w http.ResponseWriter, r *http.Request) {
	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	// Parse multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		if err.Error() == "http: request body too large" {
			writeError(w, http.StatusRequestEntityTooLarge, "file exceeds maximum size of 20MB")
			return
		}
		writeError(w, http.StatusBadRequest, "failed to parse form: "+err.Error())
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	// Validate file extension
	if !isXLSXFile(header.Filename) {
		writeError(w, http.StatusBadRequest, "only .xlsx files are supported")
		return
	}

	input := usecases.ImportEmployeesInput{
		File:     file,
		FileSize: header.Size,
	}

	output, err := h.importUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecases.ErrInvalidFileType):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrFileTooLarge):
			statusCode = http.StatusRequestEntityTooLarge
		case errors.Is(err, usecases.ErrEmptyFile):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrInvalidHeaders):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrImportValidation):
			// Return validation errors with 400 status but include the output
			writeJSON(w, http.StatusBadRequest, output)
			return
		default:
			slog.Error("employee_handler.ImportEmployees.execute_usecase", "error", err)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, output)
}

// ExportEmployees handles GET /api/v1/employees/export
func (h *EmployeeHandler) ExportEmployees(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if isScopedDepartmentClaims(claims) {
		writeJSONError(w, http.StatusForbidden, "permission_denied", "Export is not permitted for scoped department users")
		return
	}

	input, err := parseExportFilters(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	output, err := h.exportUC.Execute(r.Context(), input)
	if err != nil {
		slog.Error("employee_handler.ExportEmployees.execute_usecase", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeFileResponse(w, output.Data, output.Filename, output.ContentType)
}

// ExportEmployeesPDF handles GET /api/v1/employees/export/pdf
func (h *EmployeeHandler) ExportEmployeesPDF(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if isScopedDepartmentClaims(claims) {
		writeJSONError(w, http.StatusForbidden, "permission_denied", "Export is not permitted for scoped department users")
		return
	}

	input, err := parseExportFilters(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	output, err := h.exportPDFUC.Execute(r.Context(), input)
	if err != nil {
		slog.Error("employee_handler.ExportEmployeesPDF.execute_usecase", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeFileResponse(w, output.Data, output.Filename, output.ContentType)
}

// DownloadImportTemplate handles GET /api/v1/employees/import/template
func (h *EmployeeHandler) DownloadImportTemplate(w http.ResponseWriter, r *http.Request) {
	output, err := h.templateUC.Execute()
	if err != nil {
		slog.Error("employee_handler.DownloadImportTemplate.execute_usecase", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeFileResponse(w, output.Data, output.Filename, output.ContentType)
}

// AssignDepartmentRequest represents the request body for assigning a department.
type AssignDepartmentRequest struct {
	DepartmentUID string  `json:"departmentUid"`
	ShiftUID      *string `json:"shiftUid,omitempty"`
}

// AssignDepartment handles PUT /api/v1/employees/{uid}/department
func (h *EmployeeHandler) AssignDepartment(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	var req AssignDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("employee_handler.AssignDepartment.decode_request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DepartmentUID == "" {
		writeError(w, http.StatusBadRequest, "departmentUid is required")
		return
	}

	input := usecases.AssignEmployeeDepartmentInput{
		EmployeeUID:   uid,
		DepartmentUID: req.DepartmentUID,
		ShiftUID:      req.ShiftUID,
	}
	err := h.assignDeptUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecases.ErrEmployeeNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, usecases.ErrDepartmentNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, usecases.ErrDepartmentInactive):
			statusCode = http.StatusBadRequest
		default:
			slog.Error("employee_handler.AssignDepartment.execute_usecase", "error", err)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveDepartment handles DELETE /api/v1/employees/{uid}/department
func (h *EmployeeHandler) RemoveDepartment(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	input := usecases.RemoveEmployeeDepartmentInput{
		EmployeeUID: uid,
	}
	err := h.removeDeptUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrEmployeeNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("employee_handler.RemoveDepartment.execute_usecase", "error", err)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListPenalties handles GET /api/v1/employees/{employeeUid}/penalties
func (h *EmployeeHandler) ListPenalties(w http.ResponseWriter, r *http.Request) {
	employeeUID := r.PathValue("employeeUid")
	if employeeUID == "" {
		writeError(w, http.StatusBadRequest, "employeeUid is required")
		return
	}

	output, err := h.listPenaltiesUC.Execute(r.Context(), usecases.ListEmployeePenaltiesInput{EmployeeUID: employeeUID})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrEmployeeNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("employee_handler.ListPenalties.execute_usecase", "error", err, "employee_uid", employeeUID)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, EmployeePenaltiesResponse{Penalties: toEmployeePenaltyResponses(output.Penalties)})
}

// CreatePenalty handles POST /api/v1/employees/{employeeUid}/penalties
func (h *EmployeeHandler) CreatePenalty(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}

	employeeUID := r.PathValue("employeeUid")
	if employeeUID == "" {
		writeError(w, http.StatusBadRequest, "employeeUid is required")
		return
	}

	var req EmployeePenaltyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("employee_handler.CreatePenalty.decode_request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	penaltyInput, err := parseEmployeePenaltyRequest(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	output, err := h.createPenaltyUC.Execute(r.Context(), usecases.CreateEmployeePenaltyInput{
		EmployeeUID:  employeeUID,
		UserUID:      claims.UserUID,
		Penalty:      penaltyInput,
		DecisionFile: toEmployeePenaltyDecisionFileInput(req.PenaltyDecisionFile),
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrEmployeeNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("employee_handler.CreatePenalty.execute_usecase", "error", err, "employee_uid", employeeUID)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toCreateEmployeePenaltyResponse(output))
}

// CreatePenaltyRemoval handles POST /api/v1/penalties/{penaltyId}/removals
func (h *EmployeeHandler) CreatePenaltyRemoval(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}

	penaltyIDText := r.PathValue("penaltyId")
	if strings.TrimSpace(penaltyIDText) == "" {
		writeError(w, http.StatusBadRequest, "penaltyId is required")
		return
	}

	penaltyID, err := strconv.ParseInt(penaltyIDText, 10, 64)
	if err != nil || penaltyID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid penaltyId")
		return
	}

	var req EmployeePenaltyRemovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("employee_handler.CreatePenaltyRemoval.decode_request", "error", err, "penalty_id", penaltyID)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	removalInput, err := parseEmployeePenaltyRemovalRequest(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	output, err := h.createPenaltyRemovalUC.Execute(r.Context(), usecases.CreateEmployeePenaltyRemovalInput{
		PenaltyID:    penaltyID,
		UserUID:      claims.UserUID,
		Removal:      removalInput,
		DecisionFile: toEmployeePenaltyRemovalDecisionFileInput(req.PenaltyWithdrawalDecisionFile),
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecases.ErrPenaltyNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, usecases.ErrPenaltyAlreadyRemoved):
			statusCode = http.StatusConflict
		default:
			slog.Error("employee_handler.CreatePenaltyRemoval.execute_usecase", "error", err, "penalty_id", penaltyID)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toEmployeePenaltyRemovalResponse(output))
}

// UpdatePenalties handles PUT /api/v1/employees/{employeeUid}/penalties
func (h *EmployeeHandler) UpdatePenalties(w http.ResponseWriter, r *http.Request) {
	employeeUID := r.PathValue("employeeUid")
	if employeeUID == "" {
		writeError(w, http.StatusBadRequest, "employeeUid is required")
		return
	}

	var req UpdateEmployeePenaltiesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("employee_handler.UpdatePenalties.decode_request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	penalties := make([]usecases.EmployeePenaltyInput, 0, len(req.Penalties))
	for _, item := range req.Penalties {
		penaltyInput, err := parseEmployeePenaltyRequest(item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		penalties = append(penalties, penaltyInput)
	}

	output, err := h.updatePenaltiesUC.Execute(r.Context(), usecases.UpdateEmployeePenaltiesInput{
		EmployeeUID: employeeUID,
		Penalties:   penalties,
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrEmployeeNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("employee_handler.UpdatePenalties.execute_usecase", "error", err, "employee_uid", employeeUID)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, EmployeePenaltiesResponse{Penalties: toEmployeePenaltyResponses(output.Penalties)})
}

// ListIncentiveBonuses handles GET /api/v1/employees/{employeeUid}/incentive-bonuses
func (h *EmployeeHandler) ListIncentiveBonuses(w http.ResponseWriter, r *http.Request) {
	employeeUID := r.PathValue("employeeUid")
	if employeeUID == "" {
		writeError(w, http.StatusBadRequest, "employeeUid is required")
		return
	}

	output, err := h.listIncentiveBonusesUC.Execute(r.Context(), usecases.ListEmployeeIncentiveBonusesInput{EmployeeUID: employeeUID})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrEmployeeNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("employee_handler.ListIncentiveBonuses.execute_usecase", "error", err, "employee_uid", employeeUID)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, EmployeeIncentiveBonusesResponse{Bonuses: toEmployeeIncentiveBonusResponses(output.Bonuses)})
}

// CreateIncentiveBonus handles POST /api/v1/employees/{employeeUid}/incentive-bonuses
func (h *EmployeeHandler) CreateIncentiveBonus(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}

	employeeUID := r.PathValue("employeeUid")
	if employeeUID == "" {
		writeError(w, http.StatusBadRequest, "employeeUid is required")
		return
	}

	var req EmployeeIncentiveBonusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("employee_handler.CreateIncentiveBonus.decode_request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	bonusInput, err := parseEmployeeIncentiveBonusRequest(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	output, err := h.createIncentiveBonusUC.Execute(r.Context(), usecases.CreateEmployeeIncentiveBonusInput{
		EmployeeUID:   employeeUID,
		UserUID:       claims.UserUID,
		Bonus:         bonusInput,
		DecisionImage: toEmployeeIncentiveBonusDecisionImageInput(req.DecisionImage),
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrEmployeeNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("employee_handler.CreateIncentiveBonus.execute_usecase", "error", err, "employee_uid", employeeUID)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toCreateEmployeeIncentiveBonusResponse(output))
}

// UpdateIncentiveBonuses handles PUT /api/v1/employees/{employeeUid}/incentive-bonuses
func (h *EmployeeHandler) UpdateIncentiveBonuses(w http.ResponseWriter, r *http.Request) {
	employeeUID := r.PathValue("employeeUid")
	if employeeUID == "" {
		writeError(w, http.StatusBadRequest, "employeeUid is required")
		return
	}

	var req UpdateEmployeeIncentiveBonusesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("employee_handler.UpdateIncentiveBonuses.decode_request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	bonuses := make([]usecases.EmployeeIncentiveBonusInput, 0, len(req.Bonuses))
	for _, item := range req.Bonuses {
		bonusInput, err := parseEmployeeIncentiveBonusRequest(item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		bonuses = append(bonuses, bonusInput)
	}

	output, err := h.updateIncentiveBonusesUC.Execute(r.Context(), usecases.UpdateEmployeeIncentiveBonusesInput{
		EmployeeUID: employeeUID,
		Bonuses:     bonuses,
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrEmployeeNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("employee_handler.UpdateIncentiveBonuses.execute_usecase", "error", err, "employee_uid", employeeUID)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, EmployeeIncentiveBonusesResponse{Bonuses: toEmployeeIncentiveBonusResponses(output.Bonuses)})
}

// ListAnnualReports handles GET /api/v1/employees/{employeeUid}/annual-reports
func (h *EmployeeHandler) ListAnnualReports(w http.ResponseWriter, r *http.Request) {
	employeeUID := r.PathValue("employeeUid")
	if employeeUID == "" {
		writeError(w, http.StatusBadRequest, "employeeUid is required")
		return
	}

	output, err := h.listAnnualReportsUC.Execute(r.Context(), usecases.ListEmployeeAnnualReportsInput{EmployeeUID: employeeUID})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrEmployeeNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("employee_handler.ListAnnualReports.execute_usecase", "error", err, "employee_uid", employeeUID)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, EmployeeAnnualReportsResponse{Reports: toEmployeeAnnualReportResponses(output.Reports)})
}

// CreateAnnualReport handles POST /api/v1/employees/{employeeUid}/annual-reports
func (h *EmployeeHandler) CreateAnnualReport(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}

	employeeUID := r.PathValue("employeeUid")
	if employeeUID == "" {
		writeError(w, http.StatusBadRequest, "employeeUid is required")
		return
	}

	var req EmployeeAnnualReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("employee_handler.CreateAnnualReport.decode_request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	reportInput, err := parseEmployeeAnnualReportRequest(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	output, err := h.createAnnualReportUC.Execute(r.Context(), usecases.CreateEmployeeAnnualReportInput{
		EmployeeUID: employeeUID,
		UserUID:     claims.UserUID,
		Report:      reportInput,
		ReportImage: toEmployeeAnnualReportImageInput(req.ReportImage),
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrEmployeeNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("employee_handler.CreateAnnualReport.execute_usecase", "error", err, "employee_uid", employeeUID)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toCreateEmployeeAnnualReportResponse(output))
}

// UpdateAnnualReports handles PUT /api/v1/employees/{employeeUid}/annual-reports
func (h *EmployeeHandler) UpdateAnnualReports(w http.ResponseWriter, r *http.Request) {
	employeeUID := r.PathValue("employeeUid")
	if employeeUID == "" {
		writeError(w, http.StatusBadRequest, "employeeUid is required")
		return
	}

	var req UpdateEmployeeAnnualReportsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("employee_handler.UpdateAnnualReports.decode_request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	reports := make([]usecases.EmployeeAnnualReportInput, 0, len(req.Reports))
	for _, item := range req.Reports {
		reportInput, err := parseEmployeeAnnualReportRequest(item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		reports = append(reports, reportInput)
	}

	output, err := h.updateAnnualReportsUC.Execute(r.Context(), usecases.UpdateEmployeeAnnualReportsInput{
		EmployeeUID: employeeUID,
		Reports:     reports,
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrEmployeeNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("employee_handler.UpdateAnnualReports.execute_usecase", "error", err, "employee_uid", employeeUID)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, EmployeeAnnualReportsResponse{Reports: toEmployeeAnnualReportResponses(output.Reports)})
}

func parseListFilters(r *http.Request) (usecases.ListEmployeesInput, error) {
	var input usecases.ListEmployeesInput

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := domain.EmployeeStatus(statusStr)
		switch status {
		case domain.EmployeeStatusActive, domain.EmployeeStatusInactive, domain.EmployeeStatusTerminated:
			input.Status = &status
		default:
			return input, errors.New("invalid status value")
		}
	}

	if dateStr := r.URL.Query().Get("hire_date_from"); dateStr != "" {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return input, errors.New("invalid hire_date_from format, expected YYYY-MM-DD")
		}
		input.HireDateFrom = &date
	}

	if dateStr := r.URL.Query().Get("hire_date_to"); dateStr != "" {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return input, errors.New("invalid hire_date_to format, expected YYYY-MM-DD")
		}
		input.HireDateTo = &date
	}

	if role := r.URL.Query().Get("role"); role != "" {
		input.Role = &role
	}

	return input, nil
}

func parseExportFilters(r *http.Request) (usecases.ExportEmployeesInput, error) {
	var input usecases.ExportEmployeesInput

	// Parse status filter
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := domain.EmployeeStatus(statusStr)
		switch status {
		case domain.EmployeeStatusActive, domain.EmployeeStatusInactive, domain.EmployeeStatusTerminated:
			input.Status = &status
		default:
			return input, errors.New("invalid status value")
		}
	}

	// Parse hire_date_from filter
	if dateStr := r.URL.Query().Get("hire_date_from"); dateStr != "" {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return input, errors.New("invalid hire_date_from format, expected YYYY-MM-DD")
		}
		input.HireDateFrom = &date
	}

	// Parse hire_date_to filter
	if dateStr := r.URL.Query().Get("hire_date_to"); dateStr != "" {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return input, errors.New("invalid hire_date_to format, expected YYYY-MM-DD")
		}
		input.HireDateTo = &date
	}

	return input, nil
}

func isXLSXFile(filename string) bool {
	if len(filename) < 5 {
		return false
	}
	ext := filename[len(filename)-5:]
	return ext == ".xlsx"
}

func isScopedDepartmentClaims(claims *JWTClaims) bool {
	return claims != nil && !claims.HasPermission("*") && claims.IsDepartmentScope()
}

func isSelfScopedClaims(claims *JWTClaims) bool {
	return claims != nil && !claims.HasPermission("*") && claims.IsSelfScope()
}

func writeFileResponse(w http.ResponseWriter, data []byte, filename, contentType string) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func formatDatePtr(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format("2006-01-02")
	return &formatted
}

func parseOptionalDate(value *string) (*time.Time, error) {
	if value == nil || *value == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", *value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseEmployeePenaltyRequest(req EmployeePenaltyRequest) (usecases.EmployeePenaltyInput, error) {
	if strings.TrimSpace(req.PenaltyType) == "" {
		return usecases.EmployeePenaltyInput{}, errors.New("penaltyType is required")
	}
	if strings.TrimSpace(req.PenaltyReason) == "" {
		return usecases.EmployeePenaltyInput{}, errors.New("penaltyReason is required")
	}
	if strings.TrimSpace(req.PenaltyDecisionNumber) == "" {
		return usecases.EmployeePenaltyInput{}, errors.New("penaltyDecisionNumber is required")
	}
	if strings.TrimSpace(req.PenaltyDecisionDate) == "" {
		return usecases.EmployeePenaltyInput{}, errors.New("penaltyDecisionDate is required")
	}
	hasFileUpload := req.PenaltyDecisionFile != nil
	hasStoredURL := strings.TrimSpace(req.PenaltyDecisionFileURL) != ""
	if !hasFileUpload && !hasStoredURL {
		return usecases.EmployeePenaltyInput{}, errors.New("penaltyDecisionFile or penaltyDecisionFileUrl is required")
	}
	if hasFileUpload {
		if strings.TrimSpace(req.PenaltyDecisionFile.FileName) == "" {
			return usecases.EmployeePenaltyInput{}, errors.New("penaltyDecisionFile.fileName is required")
		}
		if strings.TrimSpace(req.PenaltyDecisionFile.ContentType) == "" {
			return usecases.EmployeePenaltyInput{}, errors.New("penaltyDecisionFile.contentType is required")
		}
	}

	decisionDate, err := time.Parse("2006-01-02", req.PenaltyDecisionDate)
	if err != nil {
		return usecases.EmployeePenaltyInput{}, errors.New("invalid penaltyDecisionDate format, expected YYYY-MM-DD")
	}

	return usecases.EmployeePenaltyInput{
		PenaltyType:            req.PenaltyType,
		PenaltyReason:          req.PenaltyReason,
		PenaltyDecisionNumber:  req.PenaltyDecisionNumber,
		PenaltyDecisionDate:    decisionDate,
		PenaltyDecisionFileURL: req.PenaltyDecisionFileURL,
	}, nil
}

func parseEmployeePenaltyRemovalRequest(req EmployeePenaltyRemovalRequest) (usecases.EmployeePenaltyRemovalInput, error) {
	if strings.TrimSpace(req.PenaltyRemovalType) == "" {
		return usecases.EmployeePenaltyRemovalInput{}, errors.New("penaltyRemovalType is required")
	}
	if strings.TrimSpace(req.PenaltyRemovalNumber) == "" {
		return usecases.EmployeePenaltyRemovalInput{}, errors.New("penaltyRemovalNumber is required")
	}
	if strings.TrimSpace(req.PenaltyRemovalDate) == "" {
		return usecases.EmployeePenaltyRemovalInput{}, errors.New("penaltyRemovalDate is required")
	}
	if strings.TrimSpace(req.Notes) == "" {
		return usecases.EmployeePenaltyRemovalInput{}, errors.New("notes is required")
	}
	if req.PenaltyWithdrawalDecisionFile == nil {
		return usecases.EmployeePenaltyRemovalInput{}, errors.New("penaltyWithdrawalDecisionFile is required")
	}
	if strings.TrimSpace(req.PenaltyWithdrawalDecisionFile.FileName) == "" {
		return usecases.EmployeePenaltyRemovalInput{}, errors.New("penaltyWithdrawalDecisionFile.fileName is required")
	}
	if strings.TrimSpace(req.PenaltyWithdrawalDecisionFile.ContentType) == "" {
		return usecases.EmployeePenaltyRemovalInput{}, errors.New("penaltyWithdrawalDecisionFile.contentType is required")
	}

	removalDate, err := time.Parse("2006-01-02", req.PenaltyRemovalDate)
	if err != nil {
		return usecases.EmployeePenaltyRemovalInput{}, errors.New("invalid penaltyRemovalDate format, expected YYYY-MM-DD")
	}

	return usecases.EmployeePenaltyRemovalInput{
		PenaltyRemovalType:   req.PenaltyRemovalType,
		PenaltyRemovalNumber: req.PenaltyRemovalNumber,
		PenaltyRemovalDate:   removalDate,
		Notes:                req.Notes,
	}, nil
}

func parseEmployeeAnnualReportRequest(req EmployeeAnnualReportRequest) (usecases.EmployeeAnnualReportInput, error) {
	if strings.TrimSpace(req.ReportYear) == "" {
		return usecases.EmployeeAnnualReportInput{}, errors.New("reportYear is required")
	}
	if strings.TrimSpace(req.ReportGrade) == "" {
		return usecases.EmployeeAnnualReportInput{}, errors.New("reportGrade is required")
	}
	if strings.TrimSpace(req.Notes) == "" {
		return usecases.EmployeeAnnualReportInput{}, errors.New("notes is required")
	}
	hasFileUpload := req.ReportImage != nil
	hasStoredURL := strings.TrimSpace(req.ReportImageURL) != ""
	if !hasFileUpload && !hasStoredURL {
		return usecases.EmployeeAnnualReportInput{}, errors.New("reportImage or reportImageUrl is required")
	}
	if hasFileUpload {
		if strings.TrimSpace(req.ReportImage.FileName) == "" {
			return usecases.EmployeeAnnualReportInput{}, errors.New("reportImage.fileName is required")
		}
		if strings.TrimSpace(req.ReportImage.ContentType) == "" {
			return usecases.EmployeeAnnualReportInput{}, errors.New("reportImage.contentType is required")
		}
	}

	return usecases.EmployeeAnnualReportInput{
		ReportYear:     req.ReportYear,
		ReportGrade:    req.ReportGrade,
		Notes:          req.Notes,
		ReportImageURL: req.ReportImageURL,
	}, nil
}

func parseEmployeeIncentiveBonusRequest(req EmployeeIncentiveBonusRequest) (usecases.EmployeeIncentiveBonusInput, error) {
	if strings.TrimSpace(req.BonusDate) == "" {
		return usecases.EmployeeIncentiveBonusInput{}, errors.New("bonusDate is required")
	}
	if strings.TrimSpace(req.DecisionNumber) == "" {
		return usecases.EmployeeIncentiveBonusInput{}, errors.New("decisionNumber is required")
	}
	if strings.TrimSpace(req.DecisionDate) == "" {
		return usecases.EmployeeIncentiveBonusInput{}, errors.New("decisionDate is required")
	}
	hasFileUpload := req.DecisionImage != nil
	hasStoredURL := strings.TrimSpace(req.DecisionImageURL) != ""
	if !hasFileUpload && !hasStoredURL {
		return usecases.EmployeeIncentiveBonusInput{}, errors.New("decisionImage or decisionImageUrl is required")
	}
	if hasFileUpload {
		if strings.TrimSpace(req.DecisionImage.FileName) == "" {
			return usecases.EmployeeIncentiveBonusInput{}, errors.New("decisionImage.fileName is required")
		}
		if strings.TrimSpace(req.DecisionImage.ContentType) == "" {
			return usecases.EmployeeIncentiveBonusInput{}, errors.New("decisionImage.contentType is required")
		}
	}

	bonusDate, err := time.Parse("2006-01-02", req.BonusDate)
	if err != nil {
		return usecases.EmployeeIncentiveBonusInput{}, errors.New("invalid bonusDate format, expected YYYY-MM-DD")
	}
	decisionDate, err := time.Parse("2006-01-02", req.DecisionDate)
	if err != nil {
		return usecases.EmployeeIncentiveBonusInput{}, errors.New("invalid decisionDate format, expected YYYY-MM-DD")
	}

	return usecases.EmployeeIncentiveBonusInput{
		BonusDate:        bonusDate,
		DecisionNumber:   req.DecisionNumber,
		DecisionDate:     decisionDate,
		DecisionImageURL: req.DecisionImageURL,
	}, nil
}

func toEmployeePenaltyDecisionFileInput(file *EmployeePenaltyDecisionFileItem) *usecases.EmployeePenaltyDecisionFileInput {
	if file == nil {
		return nil
	}
	return &usecases.EmployeePenaltyDecisionFileInput{
		FileName:    file.FileName,
		ContentType: file.ContentType,
	}
}

func toEmployeePenaltyRemovalDecisionFileInput(file *EmployeePenaltyDecisionFileItem) *usecases.EmployeePenaltyRemovalDecisionFileInput {
	if file == nil {
		return nil
	}
	return &usecases.EmployeePenaltyRemovalDecisionFileInput{
		FileName:    file.FileName,
		ContentType: file.ContentType,
	}
}

func toEmployeeAnnualReportImageInput(file *EmployeeAnnualReportImageItem) *usecases.EmployeeAnnualReportImageInput {
	if file == nil {
		return nil
	}
	return &usecases.EmployeeAnnualReportImageInput{
		FileName:    file.FileName,
		ContentType: file.ContentType,
	}
}

func toEmployeeIncentiveBonusDecisionImageInput(file *EmployeePenaltyDecisionFileItem) *usecases.EmployeeIncentiveBonusDecisionImageInput {
	if file == nil {
		return nil
	}
	return &usecases.EmployeeIncentiveBonusDecisionImageInput{
		FileName:    file.FileName,
		ContentType: file.ContentType,
	}
}

func toCreateEmployeeDocumentResponses(documents []usecases.CreateEmployeeDocumentOutput) []CreateEmployeeDocumentResponse {
	if len(documents) == 0 {
		return nil
	}

	response := make([]CreateEmployeeDocumentResponse, 0, len(documents))
	for _, document := range documents {
		response = append(response, CreateEmployeeDocumentResponse{
			DocumentType: document.DocumentType,
			FileName:     document.FileName,
			URL:          document.URL,
			Method:       document.Method,
			Bucket:       document.Bucket,
			ObjectKey:    document.ObjectKey,
			StoredURL:    document.StoredURL,
		})
	}

	return response
}

func toEmployeeDocumentDownloadResponses(documents []usecases.EmployeeDocumentOutput) []EmployeeDocumentDownloadResponse {
	if len(documents) == 0 {
		return nil
	}

	response := make([]EmployeeDocumentDownloadResponse, 0, len(documents))
	for _, document := range documents {
		response = append(response, EmployeeDocumentDownloadResponse{
			DocumentType: document.DocumentType,
			FileName:     document.FileName,
			URL:          document.URL,
			Method:       document.Method,
			Bucket:       document.Bucket,
			ObjectKey:    document.ObjectKey,
		})
	}

	return response
}

func toEmployeePenaltyResponses(penalties []usecases.EmployeePenaltyOutput) []EmployeePenaltyResponse {
	response := make([]EmployeePenaltyResponse, 0, len(penalties))
	for _, penalty := range penalties {
		response = append(response, toEmployeePenaltyResponse(penalty))
	}
	return response
}

func toEmployeePenaltyResponse(penalty usecases.EmployeePenaltyOutput) EmployeePenaltyResponse {
	response := EmployeePenaltyResponse{
		ID:                     penalty.ID,
		PenaltyType:            penalty.PenaltyType,
		PenaltyReason:          penalty.PenaltyReason,
		PenaltyDecisionNumber:  penalty.PenaltyDecisionNumber,
		PenaltyDecisionDate:    penalty.PenaltyDecisionDate.Format("2006-01-02"),
		PenaltyDecisionFileURL: penalty.PenaltyDecisionFileURL,
	}

	if penalty.PenaltyDecisionFile != nil {
		response.PenaltyDecisionFile = &EmployeePenaltyDecisionFileUploadResponse{
			FileName:  penalty.PenaltyDecisionFile.FileName,
			URL:       penalty.PenaltyDecisionFile.URL,
			Method:    penalty.PenaltyDecisionFile.Method,
			Bucket:    penalty.PenaltyDecisionFile.Bucket,
			ObjectKey: penalty.PenaltyDecisionFile.ObjectKey,
			StoredURL: penalty.PenaltyDecisionFile.StoredURL,
			ExpiresAt: penalty.PenaltyDecisionFile.ExpiresAt.Format(time.RFC3339),
		}
	}

	return response
}

func toCreateEmployeePenaltyResponse(output *usecases.CreateEmployeePenaltyOutput) CreateEmployeePenaltyResponse {
	response := CreateEmployeePenaltyResponse{
		ID:                     output.Penalty.ID,
		PenaltyType:            output.Penalty.PenaltyType,
		PenaltyReason:          output.Penalty.PenaltyReason,
		PenaltyDecisionNumber:  output.Penalty.PenaltyDecisionNumber,
		PenaltyDecisionDate:    output.Penalty.PenaltyDecisionDate.Format("2006-01-02"),
		PenaltyDecisionFileURL: output.Penalty.PenaltyDecisionFileURL,
	}

	if output.DecisionFile != nil {
		response.PenaltyDecisionFile = &EmployeePenaltyDecisionFileUploadResponse{
			FileName:  output.DecisionFile.FileName,
			URL:       output.DecisionFile.URL,
			Method:    output.DecisionFile.Method,
			Bucket:    output.DecisionFile.Bucket,
			ObjectKey: output.DecisionFile.ObjectKey,
			StoredURL: output.DecisionFile.StoredURL,
			ExpiresAt: output.DecisionFile.ExpiresAt.Format(time.RFC3339),
		}
	}

	return response
}

func toEmployeePenaltyRemovalResponse(output *usecases.CreateEmployeePenaltyRemovalOutput) EmployeePenaltyRemovalResponse {
	response := EmployeePenaltyRemovalResponse{
		ID:                               output.Removal.ID,
		PenaltyID:                        output.Removal.PenaltyID,
		PenaltyRemovalType:               output.Removal.PenaltyRemovalType,
		PenaltyRemovalNumber:             output.Removal.PenaltyRemovalNumber,
		PenaltyRemovalDate:               output.Removal.PenaltyRemovalDate.Format("2006-01-02"),
		Notes:                            output.Removal.Notes,
		PenaltyWithdrawalDecisionFileURL: output.Removal.PenaltyWithdrawalDecisionFileURL,
	}

	if output.DecisionFile != nil {
		response.PenaltyWithdrawalDecisionFile = &EmployeePenaltyDecisionFileUploadResponse{
			FileName:  output.DecisionFile.FileName,
			URL:       output.DecisionFile.URL,
			Method:    output.DecisionFile.Method,
			Bucket:    output.DecisionFile.Bucket,
			ObjectKey: output.DecisionFile.ObjectKey,
			StoredURL: output.DecisionFile.StoredURL,
			ExpiresAt: output.DecisionFile.ExpiresAt.Format(time.RFC3339),
		}
	}

	return response
}

func toEmployeeIncentiveBonusResponses(bonuses []usecases.EmployeeIncentiveBonusOutput) []EmployeeIncentiveBonusResponse {
	response := make([]EmployeeIncentiveBonusResponse, 0, len(bonuses))
	for _, bonus := range bonuses {
		response = append(response, toEmployeeIncentiveBonusResponse(bonus))
	}
	return response
}

func toEmployeeIncentiveBonusResponse(bonus usecases.EmployeeIncentiveBonusOutput) EmployeeIncentiveBonusResponse {
	response := EmployeeIncentiveBonusResponse{
		ID:               bonus.ID,
		BonusDate:        bonus.BonusDate.Format("2006-01-02"),
		DecisionNumber:   bonus.DecisionNumber,
		DecisionDate:     bonus.DecisionDate.Format("2006-01-02"),
		DecisionImageURL: bonus.DecisionImageURL,
	}
	if bonus.DecisionImage != nil {
		response.DecisionImage = &EmployeeIncentiveBonusDecisionImageResponse{
			FileName:  bonus.DecisionImage.FileName,
			URL:       bonus.DecisionImage.URL,
			Method:    bonus.DecisionImage.Method,
			Bucket:    bonus.DecisionImage.Bucket,
			ObjectKey: bonus.DecisionImage.ObjectKey,
			StoredURL: bonus.DecisionImage.StoredURL,
			ExpiresAt: bonus.DecisionImage.ExpiresAt.Format(time.RFC3339),
		}
	}
	return response
}

func toCreateEmployeeIncentiveBonusResponse(output *usecases.CreateEmployeeIncentiveBonusOutput) CreateEmployeeIncentiveBonusResponse {
	response := CreateEmployeeIncentiveBonusResponse{
		ID:               output.Bonus.ID,
		BonusDate:        output.Bonus.BonusDate.Format("2006-01-02"),
		DecisionNumber:   output.Bonus.DecisionNumber,
		DecisionDate:     output.Bonus.DecisionDate.Format("2006-01-02"),
		DecisionImageURL: output.Bonus.DecisionImageURL,
	}
	if output.DecisionImage != nil {
		response.DecisionImage = &EmployeeIncentiveBonusDecisionImageResponse{
			FileName:  output.DecisionImage.FileName,
			URL:       output.DecisionImage.URL,
			Method:    output.DecisionImage.Method,
			Bucket:    output.DecisionImage.Bucket,
			ObjectKey: output.DecisionImage.ObjectKey,
			StoredURL: output.DecisionImage.StoredURL,
			ExpiresAt: output.DecisionImage.ExpiresAt.Format(time.RFC3339),
		}
	}
	return response
}

func toEmployeeAnnualReportResponses(reports []usecases.EmployeeAnnualReportOutput) []EmployeeAnnualReportResponse {
	response := make([]EmployeeAnnualReportResponse, 0, len(reports))
	for _, report := range reports {
		response = append(response, toEmployeeAnnualReportResponse(report))
	}
	return response
}

func toEmployeeAnnualReportResponse(report usecases.EmployeeAnnualReportOutput) EmployeeAnnualReportResponse {
	response := EmployeeAnnualReportResponse{
		ReportYear:     report.ReportYear,
		ReportGrade:    report.ReportGrade,
		Notes:          report.Notes,
		ReportImageURL: report.ReportImageURL,
	}
	if report.ReportImage != nil {
		response.ReportImage = &EmployeeAnnualReportImageResponse{
			FileName:  report.ReportImage.FileName,
			URL:       report.ReportImage.URL,
			Method:    report.ReportImage.Method,
			Bucket:    report.ReportImage.Bucket,
			ObjectKey: report.ReportImage.ObjectKey,
			StoredURL: report.ReportImage.StoredURL,
			ExpiresAt: report.ReportImage.ExpiresAt.Format(time.RFC3339),
		}
	}
	return response
}

func toCreateEmployeeAnnualReportResponse(output *usecases.CreateEmployeeAnnualReportOutput) CreateEmployeeAnnualReportResponse {
	response := CreateEmployeeAnnualReportResponse{
		ReportYear:     output.Report.ReportYear,
		ReportGrade:    output.Report.ReportGrade,
		Notes:          output.Report.Notes,
		ReportImageURL: output.Report.ReportImageURL,
	}
	if output.ReportImage != nil {
		response.ReportImage = &EmployeeAnnualReportImageResponse{
			FileName:  output.ReportImage.FileName,
			URL:       output.ReportImage.URL,
			Method:    output.ReportImage.Method,
			Bucket:    output.ReportImage.Bucket,
			ObjectKey: output.ReportImage.ObjectKey,
			StoredURL: output.ReportImage.StoredURL,
			ExpiresAt: output.ReportImage.ExpiresAt.Format(time.RFC3339),
		}
	}
	return response
}
