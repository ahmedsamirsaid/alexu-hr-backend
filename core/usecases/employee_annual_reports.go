package usecases

import (
	"context"
	"strings"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type EmployeeAnnualReportInput struct {
	ReportYear     string
	ReportGrade    string
	Notes          string
	ReportImageURL string
}

type EmployeeAnnualReportImageInput struct {
	FileName    string
	ContentType string
}

type EmployeeAnnualReportImageOutput struct {
	FileName  string    `json:"fileName"`
	URL       string    `json:"url"`
	Method    string    `json:"method"`
	Bucket    string    `json:"bucket"`
	ObjectKey string    `json:"objectKey"`
	StoredURL string    `json:"storedUrl"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type EmployeeAnnualReportOutput struct {
	ReportYear     string
	ReportGrade    string
	Notes          string
	ReportImageURL string
	ReportImage    *EmployeeAnnualReportImageOutput
}

type ListEmployeeAnnualReportsInput struct {
	EmployeeUID string
}

type ListEmployeeAnnualReportsOutput struct {
	Reports []EmployeeAnnualReportOutput
}

type ListEmployeeAnnualReportsUseCase struct {
	db                 ports.DB
	employeeRepo       ports.EmployeeRepository
	reportRepo         ports.AnnualReportRepository
	documentDownloadUC *GenerateDocumentDownloadURLUseCase
}

func NewListEmployeeAnnualReportsUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	reportRepo ports.AnnualReportRepository,
	documentDownloadUC *GenerateDocumentDownloadURLUseCase,
) *ListEmployeeAnnualReportsUseCase {
	return &ListEmployeeAnnualReportsUseCase{
		db:                 db,
		employeeRepo:       employeeRepo,
		reportRepo:         reportRepo,
		documentDownloadUC: documentDownloadUC,
	}
}

func (uc *ListEmployeeAnnualReportsUseCase) Execute(ctx context.Context, input ListEmployeeAnnualReportsInput) (*ListEmployeeAnnualReportsOutput, error) {
	employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, input.EmployeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	reports, err := uc.reportRepo.ListByEmployeeUID(ctx, uc.db, input.EmployeeUID)
	if err != nil {
		return nil, err
	}

	output := &ListEmployeeAnnualReportsOutput{Reports: make([]EmployeeAnnualReportOutput, 0, len(reports))}
	for _, report := range reports {
		image, err := uc.buildReportImage(ctx, report)
		if err != nil {
			return nil, err
		}
		output.Reports = append(output.Reports, mapEmployeeAnnualReportOutput(report, image))
	}
	return output, nil
}

type CreateEmployeeAnnualReportInput struct {
	EmployeeUID string
	UserUID     string
	Report      EmployeeAnnualReportInput
	ReportImage *EmployeeAnnualReportImageInput
}

type CreateEmployeeAnnualReportOutput struct {
	Report      EmployeeAnnualReportOutput
	ReportImage *EmployeeAnnualReportImageOutput
}

type CreateEmployeeAnnualReportUseCase struct {
	db                  ports.DB
	employeeRepo        ports.EmployeeRepository
	reportRepo          ports.AnnualReportRepository
	documentUploadURLUC *GenerateDocumentUploadURLUseCase
}

func NewCreateEmployeeAnnualReportUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	reportRepo ports.AnnualReportRepository,
	documentUploadURLUC *GenerateDocumentUploadURLUseCase,
) *CreateEmployeeAnnualReportUseCase {
	return &CreateEmployeeAnnualReportUseCase{
		db:                  db,
		employeeRepo:        employeeRepo,
		reportRepo:          reportRepo,
		documentUploadURLUC: documentUploadURLUC,
	}
}

func (uc *CreateEmployeeAnnualReportUseCase) Execute(ctx context.Context, input CreateEmployeeAnnualReportInput) (*CreateEmployeeAnnualReportOutput, error) {
	employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, input.EmployeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	var reportImage *EmployeeAnnualReportImageOutput
	if input.ReportImage != nil {
		presigned, err := uc.documentUploadURLUC.Execute(ctx, GenerateDocumentUploadURLInput{
			UserUID:     input.UserUID,
			FileName:    input.ReportImage.FileName,
			ContentType: input.ReportImage.ContentType,
		})
		if err != nil {
			return nil, err
		}
		storedURL := minioObjectURL(presigned.Bucket, presigned.ObjectKey)
		input.Report.ReportImageURL = storedURL
		reportImage = &EmployeeAnnualReportImageOutput{
			FileName:  input.ReportImage.FileName,
			URL:       presigned.URL,
			Method:    presigned.Method,
			Bucket:    presigned.Bucket,
			ObjectKey: presigned.ObjectKey,
			StoredURL: storedURL,
			ExpiresAt: presigned.ExpiresAt,
		}
	}

	report := buildAnnualReportDomain(input.EmployeeUID, input.Report)
	if err := uc.reportRepo.Create(ctx, uc.db, report); err != nil {
		return nil, err
	}

	return &CreateEmployeeAnnualReportOutput{
		Report:      mapEmployeeAnnualReportOutput(report, reportImage),
		ReportImage: reportImage,
	}, nil
}

type UpdateEmployeeAnnualReportsInput struct {
	EmployeeUID string
	Reports     []EmployeeAnnualReportInput
}

type UpdateEmployeeAnnualReportsOutput struct {
	Reports []EmployeeAnnualReportOutput
}

type UpdateEmployeeAnnualReportsUseCase struct {
	db           ports.DB
	employeeRepo ports.EmployeeRepository
	reportRepo   ports.AnnualReportRepository
}

func NewUpdateEmployeeAnnualReportsUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	reportRepo ports.AnnualReportRepository,
) *UpdateEmployeeAnnualReportsUseCase {
	return &UpdateEmployeeAnnualReportsUseCase{
		db:           db,
		employeeRepo: employeeRepo,
		reportRepo:   reportRepo,
	}
}

func (uc *UpdateEmployeeAnnualReportsUseCase) Execute(ctx context.Context, input UpdateEmployeeAnnualReportsInput) (*UpdateEmployeeAnnualReportsOutput, error) {
	employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, input.EmployeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := uc.reportRepo.DeleteByEmployeeUID(ctx, tx, input.EmployeeUID); err != nil {
		return nil, err
	}

	output := &UpdateEmployeeAnnualReportsOutput{Reports: make([]EmployeeAnnualReportOutput, 0, len(input.Reports))}
	for _, reportInput := range input.Reports {
		report := buildAnnualReportDomain(input.EmployeeUID, reportInput)
		if err := uc.reportRepo.Create(ctx, tx, report); err != nil {
			return nil, err
		}
		output.Reports = append(output.Reports, mapEmployeeAnnualReportOutput(report, nil))
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return output, nil
}

func buildAnnualReportDomain(employeeUID string, input EmployeeAnnualReportInput) *domain.AnnualReport {
	return &domain.AnnualReport{
		EmployeeUID:    employeeUID,
		ReportYear:     strings.TrimSpace(input.ReportYear),
		ReportGrade:    strings.TrimSpace(input.ReportGrade),
		Notes:          strings.TrimSpace(input.Notes),
		ReportImageURL: strings.TrimSpace(input.ReportImageURL),
	}
}

func mapEmployeeAnnualReportOutput(report *domain.AnnualReport, image *EmployeeAnnualReportImageOutput) EmployeeAnnualReportOutput {
	return EmployeeAnnualReportOutput{
		ReportYear:     report.ReportYear,
		ReportGrade:    report.ReportGrade,
		Notes:          report.Notes,
		ReportImageURL: report.ReportImageURL,
		ReportImage:    image,
	}
}

func (uc *ListEmployeeAnnualReportsUseCase) buildReportImage(ctx context.Context, report *domain.AnnualReport) (*EmployeeAnnualReportImageOutput, error) {
	if strings.TrimSpace(report.ReportImageURL) == "" || uc.documentDownloadUC == nil {
		return nil, nil
	}
	objectKey, fileName, ok := parseEmployeeDocumentStoredURL(report.ReportImageURL)
	if !ok {
		return nil, nil
	}
	presigned, err := uc.documentDownloadUC.Execute(ctx, GenerateDocumentDownloadURLInput{
		ObjectKey:        objectKey,
		DownloadFileName: fileName,
	})
	if err != nil {
		return nil, err
	}
	return &EmployeeAnnualReportImageOutput{
		FileName:  fileName,
		URL:       presigned.URL,
		Method:    presigned.Method,
		Bucket:    presigned.Bucket,
		ObjectKey: presigned.ObjectKey,
		StoredURL: report.ReportImageURL,
		ExpiresAt: presigned.ExpiresAt,
	}, nil
}
