package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"legal-riu-portal/internal/config"
)

type MinIOClient struct {
	client       *minio.Client
	publicClient *minio.Client
	bucket       string
	expires      time.Duration
}

const maxUploadSizeBytes = 100 * 1024 * 1024

var allowedUploadExtensions = map[string]bool{
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
	".ppt":  true,
	".pptx": true,
	".jpg":  true,
	".jpeg": true,
	".png":  true,
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

	// Create public client for generating presigned URLs
	var publicClient *minio.Client
	publicEndpoint := cfg.MinIO.PublicEndpoint
	if publicEndpoint != "" && publicEndpoint != cfg.MinIO.Endpoint {
		// endpoint is used as-is (MinIO SDK expects host:port format, no scheme)
		publicClient, err = minio.New(publicEndpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
			Secure: cfg.MinIO.UseSSL,
		})
		if err != nil {
			log.Printf("Warning: failed to init public MinIO client, falling back to default endpoint: %v", err)
			publicClient = client
		}
	} else {
		publicClient = client
	}

	s := &MinIOClient{
		client:       client,
		publicClient: publicClient,
		bucket:       cfg.MinIO.Bucket,
		expires:      time.Duration(cfg.MinIO.PresignExpiresMinutes) * time.Minute,
	}

	Storage = s
	log.Println("MinIO connected successfully")
	return s
}

// UploadFile uploads a file to MinIO and returns the stored path
func (m *MinIOClient) UploadFile(ctx context.Context, folder string, fileHeader *multipart.FileHeader, customName ...string) (string, string, error) {
	if err := validateUpload(fileHeader); err != nil {
		return "", "", err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	ext := filepath.Ext(fileHeader.Filename)
	var objectName string
	if len(customName) > 0 && customName[0] != "" {
		sanitized := sanitizeFilename(customName[0])
		timestamp := time.Now().Format("20060102-150405")
		objectName = fmt.Sprintf("%s/%s-%s-%s%s", folder, sanitized, timestamp, uuid.New().String()[:8], ext)
	} else {
		objectName = fmt.Sprintf("%s/%s%s", folder, uuid.New().String(), ext)
	}

	_, err = m.client.PutObject(ctx, m.bucket, objectName, file, fileHeader.Size, minio.PutObjectOptions{
		ContentType: fileHeader.Header.Get("Content-Type"),
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to upload file: %w", err)
	}

	return objectName, fileHeader.Filename, nil
}

func (m *MinIOClient) UploadBytes(ctx context.Context, objectName string, data []byte, contentType string) error {
	if len(data) == 0 {
		return errors.New("file kosong tidak dapat diupload")
	}
	_, err := m.client.PutObject(ctx, m.bucket, objectName, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("failed to upload generated file: %w", err)
	}
	return nil
}

func sanitizeFilename(name string) string {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	base = strings.ReplaceAll(base, " ", "_")
	base = strings.ReplaceAll(base, ".", "_")
	var result strings.Builder
	for _, r := range base {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			result.WriteRune(r)
		}
	}
	sanitized := result.String()
	if sanitized == "" {
		sanitized = "file"
	}
	return sanitized
}

func validateUpload(fileHeader *multipart.FileHeader) error {
	if fileHeader == nil {
		return errors.New("file tidak ditemukan")
	}
	if fileHeader.Size <= 0 {
		return errors.New("file kosong tidak dapat diupload")
	}
	if fileHeader.Size > maxUploadSizeBytes {
		return fmt.Errorf("ukuran file maksimal %d MB", maxUploadSizeBytes/(1024*1024))
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedUploadExtensions[ext] {
		return errors.New("tipe file tidak diizinkan")
	}

	return nil
}

// GetPresignedURL generates a temporary pre-signed URL for file access
func (m *MinIOClient) GetPresignedURL(ctx context.Context, objectPath string) (string, error) {
	reqParams := make(url.Values)
	client := m.client
	if m.publicClient != nil {
		client = m.publicClient
	}
	presignedURL, err := client.PresignedGetObject(ctx, m.bucket, objectPath, m.expires, reqParams)
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

// GetFileObject retrieves a file object from MinIO for proxy download
func (m *MinIOClient) GetFileObject(ctx context.Context, objectPath string) (*minio.Object, error) {
	return m.client.GetObject(ctx, m.bucket, objectPath, minio.GetObjectOptions{})
}

// GetFileContentType returns the stored content type of an object (from MinIO metadata).
func (m *MinIOClient) GetFileContentType(ctx context.Context, objectPath string) string {
	info, err := m.client.StatObject(ctx, m.bucket, objectPath, minio.StatObjectOptions{})
	if err != nil {
		return ""
	}
	return info.ContentType
}
