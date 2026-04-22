package minio

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/banumusa/backend/core/ports"
	minioSDK "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

type Service struct {
	client *minioSDK.Client
}

func NewService(cfg Config) (*Service, error) {
	client, err := minioSDK.New(cfg.Endpoint, &minioSDK.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}

	return &Service{client: client}, nil
}

func (s *Service) EnsureBucket(ctx context.Context, bucket string) error {
	exists, err := s.client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("check bucket exists: %w", err)
	}
	if exists {
		return nil
	}

	if err := s.client.MakeBucket(ctx, bucket, minioSDK.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create bucket: %w", err)
	}

	return nil
}

func (s *Service) PresignUpload(
	ctx context.Context,
	bucket string,
	objectKey string,
	expiry time.Duration,
) (*ports.PresignedUpload, error) {
	u, err := s.client.PresignedPutObject(ctx, bucket, objectKey, expiry)
	if err != nil {
		return nil, fmt.Errorf("presign upload: %w", err)
	}

	return &ports.PresignedUpload{
		URL:       u.String(),
		Method:    "PUT",
		Bucket:    bucket,
		ObjectKey: objectKey,
		ExpiresAt: time.Now().UTC().Add(expiry),
	}, nil
}

func (s *Service) PresignDownload(
	ctx context.Context,
	bucket string,
	objectKey string,
	expiry time.Duration,
	downloadFilename string,
) (*ports.PresignedDownload, error) {
	reqParams := make(url.Values)
	if downloadFilename != "" {
		reqParams.Set("response-content-disposition", fmt.Sprintf(`attachment; filename="%s"`, downloadFilename))
	}

	u, err := s.client.PresignedGetObject(ctx, bucket, objectKey, expiry, reqParams)
	if err != nil {
		return nil, fmt.Errorf("presign download: %w", err)
	}

	return &ports.PresignedDownload{
		URL:       u.String(),
		Method:    "GET",
		Bucket:    bucket,
		ObjectKey: objectKey,
		ExpiresAt: time.Now().UTC().Add(expiry),
	}, nil
}

func (s *Service) StatObject(ctx context.Context, bucket string, objectKey string) (*ports.ObjectInfo, error) {
	info, err := s.client.StatObject(ctx, bucket, objectKey, minioSDK.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("stat object: %w", err)
	}

	return &ports.ObjectInfo{
		Bucket:       bucket,
		ObjectKey:    objectKey,
		Size:         info.Size,
		ETag:         info.ETag,
		ContentType:  info.ContentType,
		LastModified: info.LastModified,
	}, nil
}