package usecases

import (
	"context"
	"strings"

	"github.com/banumusa/backend/core/domain"
)

func prepareEmployeeDocumentUploads(
	ctx context.Context,
	documentUploadURLUC *GenerateDocumentUploadURLUseCase,
	userUID string,
	documents []CreateEmployeeDocumentInput,
	employee *domain.Employee,
) ([]CreateEmployeeDocumentOutput, error) {
	if len(documents) == 0 {
		return nil, nil
	}

	seen := make(map[string]struct{}, len(documents))
	outputs := make([]CreateEmployeeDocumentOutput, 0, len(documents))
	for _, document := range documents {
		documentType := strings.TrimSpace(strings.ToLower(document.DocumentType))
		if !isValidEmployeeDocumentType(documentType) {
			return nil, ErrEmployeeDocumentTypeInvalid
		}
		if _, ok := seen[documentType]; ok {
			return nil, ErrEmployeeDocumentDuplicate
		}
		seen[documentType] = struct{}{}

		presigned, err := documentUploadURLUC.Execute(ctx, GenerateDocumentUploadURLInput{
			UserUID:     userUID,
			FileName:    document.FileName,
			ContentType: document.ContentType,
		})
		if err != nil {
			return nil, err
		}

		storedURL := minioObjectURL(presigned.Bucket, presigned.ObjectKey)
		assignEmployeeDocumentURL(employee, documentType, storedURL)
		outputs = append(outputs, CreateEmployeeDocumentOutput{
			DocumentType: documentType,
			FileName:     document.FileName,
			URL:          presigned.URL,
			Method:       presigned.Method,
			Bucket:       presigned.Bucket,
			ObjectKey:    presigned.ObjectKey,
			StoredURL:    storedURL,
		})
	}

	return outputs, nil
}

func isValidEmployeeDocumentType(documentType string) bool {
	switch documentType {
	case employeeDocumentTypePersonalPhoto,
		employeeDocumentTypeNationalCardImage,
		employeeDocumentTypeQualificationCertificateImage,
		employeeDocumentTypeCV,
		employeeDocumentTypeDecisionFile:
		return true
	default:
		return false
	}
}

func assignEmployeeDocumentURL(employee *domain.Employee, documentType, storedURL string) {
	switch documentType {
	case employeeDocumentTypePersonalPhoto:
		employee.PersonalPhotoURL = &storedURL
	case employeeDocumentTypeNationalCardImage:
		employee.NationalCardImageURL = &storedURL
	case employeeDocumentTypeQualificationCertificateImage:
		employee.QualificationCertificateImageURL = &storedURL
	case employeeDocumentTypeCV:
		employee.CVURL = &storedURL
	case employeeDocumentTypeDecisionFile:
		employee.DecisionFileURL = &storedURL
	}
}

func employeeDocumentFieldName(documentType string) string {
	switch strings.TrimSpace(strings.ToLower(documentType)) {
	case employeeDocumentTypePersonalPhoto:
		return "personalPhotoUrl"
	case employeeDocumentTypeNationalCardImage:
		return "nationalCardImageUrl"
	case employeeDocumentTypeQualificationCertificateImage:
		return "qualificationCertificateImageUrl"
	case employeeDocumentTypeCV:
		return "cvUrl"
	case employeeDocumentTypeDecisionFile:
		return "decisionFileUrl"
	default:
		return ""
	}
}

func employeeDocumentFieldValue(employee *domain.Employee, documentType string) *string {
	switch strings.TrimSpace(strings.ToLower(documentType)) {
	case employeeDocumentTypePersonalPhoto:
		return employee.PersonalPhotoURL
	case employeeDocumentTypeNationalCardImage:
		return employee.NationalCardImageURL
	case employeeDocumentTypeQualificationCertificateImage:
		return employee.QualificationCertificateImageURL
	case employeeDocumentTypeCV:
		return employee.CVURL
	case employeeDocumentTypeDecisionFile:
		return employee.DecisionFileURL
	default:
		return nil
	}
}
