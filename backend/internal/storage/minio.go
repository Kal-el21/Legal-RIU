package storage

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"legal-riu-portal/internal/config"
)

type MinIOClient struct {
	client  *minio.Client
	bucket  string
	expires time.Duration
}

var Storage *MinIOClient

func InitMinIO(cfg *config.Config) *MinIOClient {
	client, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
		Secure: cfg.MinIO.UseSSL,
	})
	if err != nil {
		log.Fatalf("Failed to init MinIO client: %v", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.MinIO.Bucket)
	if err != nil {
		log.Fatalf("Failed to check bucket: %v", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.MinIO.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalf("Failed to create bucket: %v", err)
		}
		log.Printf("Bucket '%s' created successfully", cfg.MinIO.Bucket)
	}

	s := &MinIOClient{
		client:  client,
		bucket:  cfg.MinIO.Bucket,
		expires: time.Duration(cfg.MinIO.PresignExpiresMinutes) * time.Minute,
	}

	Storage = s
	log.Println("MinIO connected successfully")
	return s
}

// UploadFile uploads a file to MinIO and returns the stored path
func (m *MinIOClient) UploadFile(ctx context.Context, folder string, fileHeader *multipart.FileHeader) (string, string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	ext := filepath.Ext(fileHeader.Filename)
	objectName := fmt.Sprintf("%s/%s%s", folder, uuid.New().String(), ext)

	_, err = m.client.PutObject(ctx, m.bucket, objectName, file, fileHeader.Size, minio.PutObjectOptions{
		ContentType: fileHeader.Header.Get("Content-Type"),
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to upload file: %w", err)
	}

	return objectName, fileHeader.Filename, nil
}

// GetPresignedURL generates a temporary pre-signed URL for file access
func (m *MinIOClient) GetPresignedURL(ctx context.Context, objectPath string) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := m.client.PresignedGetObject(ctx, m.bucket, objectPath, m.expires, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return presignedURL.String(), nil
}

// DeleteFile removes a file from MinIO
func (m *MinIOClient) DeleteFile(ctx context.Context, objectPath string) error {
	err := m.client.RemoveObject(ctx, m.bucket, objectPath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
