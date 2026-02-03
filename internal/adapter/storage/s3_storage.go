package storage

import (
	"context"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/plastinin/docrecognizer/internal/config"
)

// S3Storage реализация файлового хранилища на базе S3/MinIO
type S3Storage struct {
	client *minio.Client
	bucket string
}

// NewS3Storage создаёт новый экземпляр S3Storage
func NewS3Storage(ctx context.Context, cfg config.S3Config) (*S3Storage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	// Проверяем/создаём bucket
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &S3Storage{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// Upload загружает файл в S3 и возвращает ключ
func (s *S3Storage) Upload(ctx context.Context, fileName string, contentType string, reader io.Reader, size int64) (string, error) {
	// Генерируем уникальный ключ: year/month/day/uuid/filename
	now := time.Now()
	fileKey := path.Join(
		now.Format("2006"),
		now.Format("01"),
		now.Format("02"),
		uuid.New().String(),
		fileName,
	)

	_, err := s.client.PutObject(ctx, s.bucket, fileKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return fileKey, nil
}

// Download скачивает файл из S3
func (s *S3Storage) Download(ctx context.Context, fileKey string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, fileKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	// Проверяем, что объект существует
	_, err = obj.Stat()
	if err != nil {
		obj.Close()
		return nil, fmt.Errorf("failed to stat object: %w", err)
	}

	return obj, nil
}

// Delete удаляет файл из S3
func (s *S3Storage) Delete(ctx context.Context, fileKey string) error {
	err := s.client.RemoveObject(ctx, s.bucket, fileKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// GetURL возвращает presigned URL для доступа к файлу
func (s *S3Storage) GetURL(ctx context.Context, fileKey string) (string, error) {
	// URL действителен 1 час
	expiry := time.Hour

	url, err := s.client.PresignedGetObject(ctx, s.bucket, fileKey, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}