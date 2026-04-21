package ports

import (
	"context"
	"time"
)

type PresignedUpload struct {
	URL       string
	Method    string
	Bucket    string
	ObjectKey string
	ExpiresAt time.Time
}

type PresignedDownload struct {
	URL       string
	Method    string
	Bucket    string
	ObjectKey string
	ExpiresAt time.Time
}

type ObjectInfo struct {
	Bucket      string
	ObjectKey   string
	Size        int64
	ETag        string
	ContentType string
	LastModified time.Time
}

type ObjectStorageService interface {
	EnsureBucket(ctx context.Context, bucket string) error

	PresignUpload(
		ctx context.Context,
		bucket string,
		objectKey string,
		expiry time.Duration,
	) (*PresignedUpload, error)

	PresignDownload(
		ctx context.Context,
		bucket string,
		objectKey string,
		expiry time.Duration,
		downloadFilename string,
	) (*PresignedDownload, error)

	StatObject(ctx context.Context, bucket string, objectKey string) (*ObjectInfo, error)
}