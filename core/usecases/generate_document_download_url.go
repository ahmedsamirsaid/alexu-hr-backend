package usecases

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/banumusa/backend/core/ports"
)

var ErrInvalidObjectKey = errors.New("invalid object key")

type GenerateDocumentDownloadURLInput struct {
	ObjectKey        string
	DownloadFileName string
}

type GenerateDocumentDownloadURLOutput struct {
	URL       string    `json:"url"`
	Method    string    `json:"method"`
	Bucket    string    `json:"bucket"`
	ObjectKey string    `json:"objectKey"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type GenerateDocumentDownloadURLUseCase struct {
	storage       ports.ObjectStorageService
	bucket        string
	expiryMinutes int
}

func NewGenerateDocumentDownloadURLUseCase(
	storage ports.ObjectStorageService,
	bucket string,
	expiryMinutes int,
) *GenerateDocumentDownloadURLUseCase {
	return &GenerateDocumentDownloadURLUseCase{
		storage:       storage,
		bucket:        bucket,
		expiryMinutes: expiryMinutes,
	}
}

func (uc *GenerateDocumentDownloadURLUseCase) Execute(
	ctx context.Context,
	input GenerateDocumentDownloadURLInput,
) (*GenerateDocumentDownloadURLOutput, error) {
	if strings.TrimSpace(input.ObjectKey) == "" {
		return nil, ErrInvalidObjectKey
	}

	_, err := uc.storage.StatObject(ctx, uc.bucket, input.ObjectKey)
	if err != nil {
		return nil, err
	}

	out, err := uc.storage.PresignDownload(
		ctx,
		uc.bucket,
		input.ObjectKey,
		time.Duration(uc.expiryMinutes)*time.Minute,
		input.DownloadFileName,
	)
	if err != nil {
		return nil, err
	}

	return &GenerateDocumentDownloadURLOutput{
		URL:       out.URL,
		Method:    out.Method,
		Bucket:    out.Bucket,
		ObjectKey: out.ObjectKey,
		ExpiresAt: out.ExpiresAt,
	}, nil
}