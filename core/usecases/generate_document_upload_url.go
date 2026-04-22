package usecases

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/banumusa/backend/core/ports"
)

var (
	ErrInvalidFilename    = errors.New("invalid filename")
	ErrInvalidContentType = errors.New("invalid content type")
)

type GenerateDocumentUploadURLInput struct {
	UserUID     string
	FileName    string
	ContentType string
}

type GenerateDocumentUploadURLOutput struct {
	URL       string    `json:"url"`
	Method    string    `json:"method"`
	Bucket    string    `json:"bucket"`
	ObjectKey string    `json:"objectKey"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type GenerateDocumentUploadURLUseCase struct {
	storage       ports.ObjectStorageService
	bucket        string
	expiryMinutes int
}

func NewGenerateDocumentUploadURLUseCase(
	storage ports.ObjectStorageService,
	bucket string,
	expiryMinutes int,
) *GenerateDocumentUploadURLUseCase {
	return &GenerateDocumentUploadURLUseCase{
		storage:       storage,
		bucket:        bucket,
		expiryMinutes: expiryMinutes,
	}
}

func (uc *GenerateDocumentUploadURLUseCase) Execute(
	ctx context.Context,
	input GenerateDocumentUploadURLInput,
) (*GenerateDocumentUploadURLOutput, error) {
	if strings.TrimSpace(input.FileName) == "" {
		return nil, ErrInvalidFilename
	}

	if !isAllowedContentType(input.ContentType) {
		return nil, ErrInvalidContentType
	}

	ext := strings.ToLower(filepath.Ext(input.FileName))
	objectKey := fmt.Sprintf(
		"documents/%s/%s/%s%s",
		input.UserUID,
		time.Now().UTC().Format("2006/01"),
		uuid.NewString(),
		ext,
	)

	out, err := uc.storage.PresignUpload(
		ctx,
		uc.bucket,
		objectKey,
		time.Duration(uc.expiryMinutes)*time.Minute,
	)
	if err != nil {
		return nil, err
	}

	return &GenerateDocumentUploadURLOutput{
		URL:       out.URL,
		Method:    out.Method,
		Bucket:    out.Bucket,
		ObjectKey: out.ObjectKey,
		ExpiresAt: out.ExpiresAt,
	}, nil
}

func isAllowedContentType(contentType string) bool {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case
		"application/pdf",
		"image/jpeg",
		"image/png",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return true
	default:
		return false
	}
}