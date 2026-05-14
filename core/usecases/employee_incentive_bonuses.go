package usecases

import (
	"context"
	"strings"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type EmployeeIncentiveBonusInput struct {
	BonusDate        time.Time
	DecisionNumber   string
	DecisionDate     time.Time
	DecisionImageURL string
}

type EmployeeIncentiveBonusDecisionImageInput struct {
	FileName    string
	ContentType string
}

type EmployeeIncentiveBonusDecisionImageOutput struct {
	FileName  string    `json:"fileName"`
	URL       string    `json:"url"`
	Method    string    `json:"method"`
	Bucket    string    `json:"bucket"`
	ObjectKey string    `json:"objectKey"`
	StoredURL string    `json:"storedUrl"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type EmployeeIncentiveBonusOutput struct {
	ID               int64
	BonusDate        time.Time
	DecisionNumber   string
	DecisionDate     time.Time
	DecisionImageURL string
	DecisionImage    *EmployeeIncentiveBonusDecisionImageOutput
}

type ListEmployeeIncentiveBonusesInput struct {
	EmployeeUID string
}

type ListEmployeeIncentiveBonusesOutput struct {
	Bonuses []EmployeeIncentiveBonusOutput
}

type ListEmployeeIncentiveBonusesUseCase struct {
	db                 ports.DB
	employeeRepo       ports.EmployeeRepository
	bonusRepo          ports.IncentiveBonusRepository
	documentDownloadUC *GenerateDocumentDownloadURLUseCase
}

func NewListEmployeeIncentiveBonusesUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	bonusRepo ports.IncentiveBonusRepository,
	documentDownloadUC *GenerateDocumentDownloadURLUseCase,
) *ListEmployeeIncentiveBonusesUseCase {
	return &ListEmployeeIncentiveBonusesUseCase{
		db:                 db,
		employeeRepo:       employeeRepo,
		bonusRepo:          bonusRepo,
		documentDownloadUC: documentDownloadUC,
	}
}

func (uc *ListEmployeeIncentiveBonusesUseCase) Execute(ctx context.Context, input ListEmployeeIncentiveBonusesInput) (*ListEmployeeIncentiveBonusesOutput, error) {
	employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, input.EmployeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	bonuses, err := uc.bonusRepo.ListByEmployeeUID(ctx, uc.db, input.EmployeeUID)
	if err != nil {
		return nil, err
	}

	output := &ListEmployeeIncentiveBonusesOutput{Bonuses: make([]EmployeeIncentiveBonusOutput, 0, len(bonuses))}
	for _, bonus := range bonuses {
		decisionImage, err := uc.buildDecisionImage(ctx, bonus)
		if err != nil {
			return nil, err
		}
		output.Bonuses = append(output.Bonuses, mapEmployeeIncentiveBonusOutput(bonus, decisionImage))
	}

	return output, nil
}

type CreateEmployeeIncentiveBonusInput struct {
	EmployeeUID   string
	UserUID       string
	Bonus         EmployeeIncentiveBonusInput
	DecisionImage *EmployeeIncentiveBonusDecisionImageInput
}

type CreateEmployeeIncentiveBonusOutput struct {
	Bonus         EmployeeIncentiveBonusOutput
	DecisionImage *EmployeeIncentiveBonusDecisionImageOutput
}

type CreateEmployeeIncentiveBonusUseCase struct {
	db                  ports.DB
	employeeRepo        ports.EmployeeRepository
	bonusRepo           ports.IncentiveBonusRepository
	documentUploadURLUC *GenerateDocumentUploadURLUseCase
}

func NewCreateEmployeeIncentiveBonusUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	bonusRepo ports.IncentiveBonusRepository,
	documentUploadURLUC *GenerateDocumentUploadURLUseCase,
) *CreateEmployeeIncentiveBonusUseCase {
	return &CreateEmployeeIncentiveBonusUseCase{
		db:                  db,
		employeeRepo:        employeeRepo,
		bonusRepo:           bonusRepo,
		documentUploadURLUC: documentUploadURLUC,
	}
}

func (uc *CreateEmployeeIncentiveBonusUseCase) Execute(ctx context.Context, input CreateEmployeeIncentiveBonusInput) (*CreateEmployeeIncentiveBonusOutput, error) {
	employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, input.EmployeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	var decisionImage *EmployeeIncentiveBonusDecisionImageOutput
	if input.DecisionImage != nil {
		presigned, err := uc.documentUploadURLUC.Execute(ctx, GenerateDocumentUploadURLInput{
			UserUID:     input.UserUID,
			FileName:    input.DecisionImage.FileName,
			ContentType: input.DecisionImage.ContentType,
		})
		if err != nil {
			return nil, err
		}

		storedURL := minioObjectURL(presigned.Bucket, presigned.ObjectKey)
		input.Bonus.DecisionImageURL = storedURL
		decisionImage = &EmployeeIncentiveBonusDecisionImageOutput{
			FileName:  input.DecisionImage.FileName,
			URL:       presigned.URL,
			Method:    presigned.Method,
			Bucket:    presigned.Bucket,
			ObjectKey: presigned.ObjectKey,
			StoredURL: storedURL,
			ExpiresAt: presigned.ExpiresAt,
		}
	}

	bonus := buildIncentiveBonusDomain(input.EmployeeUID, input.Bonus)
	if err := uc.bonusRepo.Create(ctx, uc.db, bonus); err != nil {
		return nil, err
	}

	return &CreateEmployeeIncentiveBonusOutput{
		Bonus:         mapEmployeeIncentiveBonusOutput(bonus, decisionImage),
		DecisionImage: decisionImage,
	}, nil
}

type UpdateEmployeeIncentiveBonusesInput struct {
	EmployeeUID string
	Bonuses     []EmployeeIncentiveBonusInput
}

type UpdateEmployeeIncentiveBonusesOutput struct {
	Bonuses []EmployeeIncentiveBonusOutput
}

type UpdateEmployeeIncentiveBonusesUseCase struct {
	db           ports.DB
	employeeRepo ports.EmployeeRepository
	bonusRepo    ports.IncentiveBonusRepository
}

func NewUpdateEmployeeIncentiveBonusesUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	bonusRepo ports.IncentiveBonusRepository,
) *UpdateEmployeeIncentiveBonusesUseCase {
	return &UpdateEmployeeIncentiveBonusesUseCase{
		db:           db,
		employeeRepo: employeeRepo,
		bonusRepo:    bonusRepo,
	}
}

func (uc *UpdateEmployeeIncentiveBonusesUseCase) Execute(ctx context.Context, input UpdateEmployeeIncentiveBonusesInput) (*UpdateEmployeeIncentiveBonusesOutput, error) {
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

	if err := uc.bonusRepo.DeleteByEmployeeUID(ctx, tx, input.EmployeeUID); err != nil {
		return nil, err
	}

	output := &UpdateEmployeeIncentiveBonusesOutput{Bonuses: make([]EmployeeIncentiveBonusOutput, 0, len(input.Bonuses))}
	for _, bonusInput := range input.Bonuses {
		bonus := buildIncentiveBonusDomain(input.EmployeeUID, bonusInput)
		if err := uc.bonusRepo.Create(ctx, tx, bonus); err != nil {
			return nil, err
		}
		output.Bonuses = append(output.Bonuses, mapEmployeeIncentiveBonusOutput(bonus, nil))
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return output, nil
}

func buildIncentiveBonusDomain(employeeUID string, input EmployeeIncentiveBonusInput) *domain.IncentiveBonus {
	return &domain.IncentiveBonus{
		EmployeeUID:      employeeUID,
		BonusDate:        input.BonusDate,
		DecisionNumber:   strings.TrimSpace(input.DecisionNumber),
		DecisionDate:     input.DecisionDate,
		DecisionImageURL: strings.TrimSpace(input.DecisionImageURL),
	}
}

func mapEmployeeIncentiveBonusOutput(bonus *domain.IncentiveBonus, decisionImage *EmployeeIncentiveBonusDecisionImageOutput) EmployeeIncentiveBonusOutput {
	return EmployeeIncentiveBonusOutput{
		ID:               bonus.ID,
		BonusDate:        bonus.BonusDate,
		DecisionNumber:   bonus.DecisionNumber,
		DecisionDate:     bonus.DecisionDate,
		DecisionImageURL: bonus.DecisionImageURL,
		DecisionImage:    decisionImage,
	}
}

func (uc *ListEmployeeIncentiveBonusesUseCase) buildDecisionImage(ctx context.Context, bonus *domain.IncentiveBonus) (*EmployeeIncentiveBonusDecisionImageOutput, error) {
	if strings.TrimSpace(bonus.DecisionImageURL) == "" || uc.documentDownloadUC == nil {
		return nil, nil
	}

	objectKey, fileName, ok := parseEmployeeDocumentStoredURL(bonus.DecisionImageURL)
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

	return &EmployeeIncentiveBonusDecisionImageOutput{
		FileName:  fileName,
		URL:       presigned.URL,
		Method:    presigned.Method,
		Bucket:    presigned.Bucket,
		ObjectKey: presigned.ObjectKey,
		StoredURL: bonus.DecisionImageURL,
		ExpiresAt: presigned.ExpiresAt,
	}, nil
}
