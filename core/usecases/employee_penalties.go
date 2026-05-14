package usecases

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

var ErrPenaltyNotFound = errors.New("penalty not found")
var ErrPenaltyAlreadyRemoved = errors.New("penalty already removed")

type EmployeePenaltyInput struct {
	PenaltyType            string
	PenaltyReason          string
	PenaltyDecisionNumber  string
	PenaltyDecisionDate    time.Time
	PenaltyDecisionFileURL string
}

type EmployeePenaltyDecisionFileInput struct {
	FileName    string
	ContentType string
}

type EmployeePenaltyOutput struct {
	ID                     int64
	PenaltyType            string
	PenaltyReason          string
	PenaltyDecisionNumber  string
	PenaltyDecisionDate    time.Time
	PenaltyDecisionFileURL string
	PenaltyDecisionFile    *EmployeePenaltyDecisionFileOutput
}

type EmployeePenaltyDecisionFileOutput struct {
	FileName  string    `json:"fileName"`
	URL       string    `json:"url"`
	Method    string    `json:"method"`
	Bucket    string    `json:"bucket"`
	ObjectKey string    `json:"objectKey"`
	StoredURL string    `json:"storedUrl"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type ListEmployeePenaltiesInput struct {
	EmployeeUID string
}

type ListEmployeePenaltiesOutput struct {
	Penalties []EmployeePenaltyOutput
}

type ListEmployeePenaltiesUseCase struct {
	db                 ports.DB
	employeeRepo       ports.EmployeeRepository
	penaltyRepo        ports.PenaltyRepository
	documentDownloadUC *GenerateDocumentDownloadURLUseCase
}

func NewListEmployeePenaltiesUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	penaltyRepo ports.PenaltyRepository,
	documentDownloadUC *GenerateDocumentDownloadURLUseCase,
) *ListEmployeePenaltiesUseCase {
	return &ListEmployeePenaltiesUseCase{
		db:                 db,
		employeeRepo:       employeeRepo,
		penaltyRepo:        penaltyRepo,
		documentDownloadUC: documentDownloadUC,
	}
}

func (uc *ListEmployeePenaltiesUseCase) Execute(ctx context.Context, input ListEmployeePenaltiesInput) (*ListEmployeePenaltiesOutput, error) {
	employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, input.EmployeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	penalties, err := uc.penaltyRepo.ListByEmployeeUID(ctx, uc.db, input.EmployeeUID)
	if err != nil {
		return nil, err
	}

	output := &ListEmployeePenaltiesOutput{Penalties: make([]EmployeePenaltyOutput, 0, len(penalties))}
	for _, penalty := range penalties {
		decisionFile, err := uc.buildPenaltyDecisionFile(ctx, penalty)
		if err != nil {
			return nil, err
		}
		output.Penalties = append(output.Penalties, mapEmployeePenaltyOutput(penalty, decisionFile))
	}

	return output, nil
}

type CreateEmployeePenaltyInput struct {
	EmployeeUID  string
	UserUID      string
	Penalty      EmployeePenaltyInput
	DecisionFile *EmployeePenaltyDecisionFileInput
}

type CreateEmployeePenaltyOutput struct {
	Penalty      EmployeePenaltyOutput
	DecisionFile *EmployeePenaltyDecisionFileOutput
}

type EmployeePenaltyRemovalInput struct {
	PenaltyRemovalType               string
	PenaltyRemovalNumber             string
	PenaltyRemovalDate               time.Time
	Notes                            string
	PenaltyWithdrawalDecisionFileURL string
}

type EmployeePenaltyRemovalDecisionFileInput struct {
	FileName    string
	ContentType string
}

type EmployeePenaltyRemovalOutput struct {
	ID                               int64
	PenaltyID                        int64
	PenaltyRemovalType               string
	PenaltyRemovalNumber             string
	PenaltyRemovalDate               time.Time
	Notes                            string
	PenaltyWithdrawalDecisionFileURL string
	PenaltyWithdrawalDecisionFile    *EmployeePenaltyDecisionFileOutput
}

type CreateEmployeePenaltyRemovalInput struct {
	PenaltyID    int64
	UserUID      string
	Removal      EmployeePenaltyRemovalInput
	DecisionFile *EmployeePenaltyRemovalDecisionFileInput
}

type CreateEmployeePenaltyRemovalOutput struct {
	Removal      EmployeePenaltyRemovalOutput
	DecisionFile *EmployeePenaltyDecisionFileOutput
}

type CreateEmployeePenaltyUseCase struct {
	db                  ports.DB
	employeeRepo        ports.EmployeeRepository
	penaltyRepo         ports.PenaltyRepository
	documentUploadURLUC *GenerateDocumentUploadURLUseCase
}

func NewCreateEmployeePenaltyUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	penaltyRepo ports.PenaltyRepository,
	documentUploadURLUC *GenerateDocumentUploadURLUseCase,
) *CreateEmployeePenaltyUseCase {
	return &CreateEmployeePenaltyUseCase{
		db:                  db,
		employeeRepo:        employeeRepo,
		penaltyRepo:         penaltyRepo,
		documentUploadURLUC: documentUploadURLUC,
	}
}

func (uc *CreateEmployeePenaltyUseCase) Execute(ctx context.Context, input CreateEmployeePenaltyInput) (*CreateEmployeePenaltyOutput, error) {
	employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, input.EmployeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	var decisionFile *EmployeePenaltyDecisionFileOutput
	if input.DecisionFile != nil {
		presigned, err := uc.documentUploadURLUC.Execute(ctx, GenerateDocumentUploadURLInput{
			UserUID:     input.UserUID,
			FileName:    input.DecisionFile.FileName,
			ContentType: input.DecisionFile.ContentType,
		})
		if err != nil {
			return nil, err
		}

		storedURL := minioObjectURL(presigned.Bucket, presigned.ObjectKey)
		input.Penalty.PenaltyDecisionFileURL = storedURL
		decisionFile = &EmployeePenaltyDecisionFileOutput{
			FileName:  input.DecisionFile.FileName,
			URL:       presigned.URL,
			Method:    presigned.Method,
			Bucket:    presigned.Bucket,
			ObjectKey: presigned.ObjectKey,
			StoredURL: storedURL,
			ExpiresAt: presigned.ExpiresAt,
		}
	}

	penalty := buildPenaltyDomain(input.EmployeeUID, input.Penalty)
	if err := uc.penaltyRepo.Create(ctx, uc.db, penalty); err != nil {
		return nil, err
	}

	return &CreateEmployeePenaltyOutput{
		Penalty:      mapEmployeePenaltyOutput(penalty, decisionFile),
		DecisionFile: decisionFile,
	}, nil
}

type UpdateEmployeePenaltiesInput struct {
	EmployeeUID string
	Penalties   []EmployeePenaltyInput
}

type UpdateEmployeePenaltiesOutput struct {
	Penalties []EmployeePenaltyOutput
}

type UpdateEmployeePenaltiesUseCase struct {
	db           ports.DB
	employeeRepo ports.EmployeeRepository
	penaltyRepo  ports.PenaltyRepository
}

func NewUpdateEmployeePenaltiesUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	penaltyRepo ports.PenaltyRepository,
) *UpdateEmployeePenaltiesUseCase {
	return &UpdateEmployeePenaltiesUseCase{
		db:           db,
		employeeRepo: employeeRepo,
		penaltyRepo:  penaltyRepo,
	}
}

func (uc *UpdateEmployeePenaltiesUseCase) Execute(ctx context.Context, input UpdateEmployeePenaltiesInput) (*UpdateEmployeePenaltiesOutput, error) {
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

	if err := uc.penaltyRepo.DeleteByEmployeeUID(ctx, tx, input.EmployeeUID); err != nil {
		return nil, err
	}

	output := &UpdateEmployeePenaltiesOutput{Penalties: make([]EmployeePenaltyOutput, 0, len(input.Penalties))}
	for _, penaltyInput := range input.Penalties {
		penalty := buildPenaltyDomain(input.EmployeeUID, penaltyInput)
		if err := uc.penaltyRepo.Create(ctx, tx, penalty); err != nil {
			return nil, err
		}
		output.Penalties = append(output.Penalties, mapEmployeePenaltyOutput(penalty, nil))
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return output, nil
}

type CreateEmployeePenaltyRemovalUseCase struct {
	db                  ports.DB
	penaltyRepo         ports.PenaltyRepository
	documentUploadURLUC *GenerateDocumentUploadURLUseCase
}

func NewCreateEmployeePenaltyRemovalUseCase(
	db ports.DB,
	penaltyRepo ports.PenaltyRepository,
	documentUploadURLUC *GenerateDocumentUploadURLUseCase,
) *CreateEmployeePenaltyRemovalUseCase {
	return &CreateEmployeePenaltyRemovalUseCase{
		db:                  db,
		penaltyRepo:         penaltyRepo,
		documentUploadURLUC: documentUploadURLUC,
	}
}

func (uc *CreateEmployeePenaltyRemovalUseCase) Execute(ctx context.Context, input CreateEmployeePenaltyRemovalInput) (*CreateEmployeePenaltyRemovalOutput, error) {
	penalty, err := uc.penaltyRepo.GetByID(ctx, uc.db, input.PenaltyID)
	if err != nil {
		return nil, err
	}
	if penalty == nil {
		return nil, ErrPenaltyNotFound
	}

	isRemoved, err := uc.penaltyRepo.IsRemoved(ctx, uc.db, input.PenaltyID)
	if err != nil {
		return nil, err
	}
	if isRemoved {
		return nil, ErrPenaltyAlreadyRemoved
	}

	var decisionFile *EmployeePenaltyDecisionFileOutput
	if input.DecisionFile != nil {
		presigned, err := uc.documentUploadURLUC.Execute(ctx, GenerateDocumentUploadURLInput{
			UserUID:     input.UserUID,
			FileName:    input.DecisionFile.FileName,
			ContentType: input.DecisionFile.ContentType,
		})
		if err != nil {
			return nil, err
		}

		storedURL := minioObjectURL(presigned.Bucket, presigned.ObjectKey)
		input.Removal.PenaltyWithdrawalDecisionFileURL = storedURL
		decisionFile = &EmployeePenaltyDecisionFileOutput{
			FileName:  input.DecisionFile.FileName,
			URL:       presigned.URL,
			Method:    presigned.Method,
			Bucket:    presigned.Bucket,
			ObjectKey: presigned.ObjectKey,
			StoredURL: storedURL,
			ExpiresAt: presigned.ExpiresAt,
		}
	}

	removal := buildPenaltyRemovalDomain(input.PenaltyID, input.Removal)
	if err := uc.penaltyRepo.CreateRemoval(ctx, uc.db, removal); err != nil {
		return nil, err
	}

	return &CreateEmployeePenaltyRemovalOutput{
		Removal:      mapEmployeePenaltyRemovalOutput(removal, decisionFile),
		DecisionFile: decisionFile,
	}, nil
}

func buildPenaltyDomain(employeeUID string, input EmployeePenaltyInput) *domain.Penalty {
	return &domain.Penalty{
		EmployeeUID:            employeeUID,
		PenaltyType:            strings.TrimSpace(input.PenaltyType),
		PenaltyReason:          strings.TrimSpace(input.PenaltyReason),
		PenaltyDecisionNumber:  strings.TrimSpace(input.PenaltyDecisionNumber),
		PenaltyDecisionDate:    input.PenaltyDecisionDate,
		PenaltyDecisionFileURL: strings.TrimSpace(input.PenaltyDecisionFileURL),
	}
}

func buildPenaltyRemovalDomain(penaltyID int64, input EmployeePenaltyRemovalInput) *domain.PenaltyRemoval {
	return &domain.PenaltyRemoval{
		PenaltyID:                        penaltyID,
		PenaltyRemovalType:               strings.TrimSpace(input.PenaltyRemovalType),
		PenaltyRemovalNumber:             strings.TrimSpace(input.PenaltyRemovalNumber),
		PenaltyRemovalDate:               input.PenaltyRemovalDate,
		Notes:                            strings.TrimSpace(input.Notes),
		PenaltyWithdrawalDecisionFileURL: strings.TrimSpace(input.PenaltyWithdrawalDecisionFileURL),
	}
}

func mapEmployeePenaltyOutput(penalty *domain.Penalty, decisionFile *EmployeePenaltyDecisionFileOutput) EmployeePenaltyOutput {
	return EmployeePenaltyOutput{
		ID:                     penalty.ID,
		PenaltyType:            penalty.PenaltyType,
		PenaltyReason:          penalty.PenaltyReason,
		PenaltyDecisionNumber:  penalty.PenaltyDecisionNumber,
		PenaltyDecisionDate:    penalty.PenaltyDecisionDate,
		PenaltyDecisionFileURL: penalty.PenaltyDecisionFileURL,
		PenaltyDecisionFile:    decisionFile,
	}
}

func mapEmployeePenaltyRemovalOutput(removal *domain.PenaltyRemoval, decisionFile *EmployeePenaltyDecisionFileOutput) EmployeePenaltyRemovalOutput {
	return EmployeePenaltyRemovalOutput{
		ID:                               removal.ID,
		PenaltyID:                        removal.PenaltyID,
		PenaltyRemovalType:               removal.PenaltyRemovalType,
		PenaltyRemovalNumber:             removal.PenaltyRemovalNumber,
		PenaltyRemovalDate:               removal.PenaltyRemovalDate,
		Notes:                            removal.Notes,
		PenaltyWithdrawalDecisionFileURL: removal.PenaltyWithdrawalDecisionFileURL,
		PenaltyWithdrawalDecisionFile:    decisionFile,
	}
}

func (uc *ListEmployeePenaltiesUseCase) buildPenaltyDecisionFile(ctx context.Context, penalty *domain.Penalty) (*EmployeePenaltyDecisionFileOutput, error) {
	if strings.TrimSpace(penalty.PenaltyDecisionFileURL) == "" || uc.documentDownloadUC == nil {
		return nil, nil
	}

	objectKey, fileName, ok := parseEmployeeDocumentStoredURL(penalty.PenaltyDecisionFileURL)
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

	return &EmployeePenaltyDecisionFileOutput{
		FileName:  fileName,
		URL:       presigned.URL,
		Method:    presigned.Method,
		Bucket:    presigned.Bucket,
		ObjectKey: presigned.ObjectKey,
		StoredURL: penalty.PenaltyDecisionFileURL,
		ExpiresAt: presigned.ExpiresAt,
	}, nil
}
