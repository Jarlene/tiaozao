package storage

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"path"
	"time"

	"flea-market/internal/config"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	productImagesBucket = "flea-products"
	presignedURLExpiry  = 24 * time.Hour
)

type MinIOClient struct {
	client *minio.Client
	cfg    *config.MinIOConfig
}

func NewMinIOClient(cfg *config.MinIOConfig) *MinIOClient {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		log.Fatalf("Failed to create MinIO client: %v", err)
	}

	mc := &MinIOClient{client: client, cfg: cfg}

	if err := mc.ensureBucket(cfg.Bucket); err != nil {
		log.Fatalf("Failed to create MinIO bucket %s: %v", cfg.Bucket, err)
	}
	if err := mc.ensureBucket(productImagesBucket); err != nil {
		log.Fatalf("Failed to create MinIO bucket %s: %v", productImagesBucket, err)
	}

	return mc
}

func (m *MinIOClient) Client() *minio.Client {
	return m.client
}

func (m *MinIOClient) AvatarBucket() string {
	return m.cfg.Bucket
}

func (m *MinIOClient) ProductBucket() string {
	return productImagesBucket
}

// UploadFile 上传文件到 MinIO，返回 object key
func (m *MinIOClient) UploadFile(bucket string, file multipart.File, header *multipart.FileHeader) (string, error) {
	ctx := context.Background()

	// 生成唯一 object key: user_id/uuid.ext
	ext := path.Ext(header.Filename)
	objectKey := fmt.Sprintf("%s/%s%s", time.Now().Format("2006/01/02"), uuid.New().String(), ext)

	_, err := m.client.PutObject(ctx, bucket, objectKey, file, header.Size, minio.PutObjectOptions{
		ContentType: header.Header.Get("Content-Type"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to MinIO: %w", err)
	}

	return objectKey, nil
}

// DeleteFile 从 MinIO 删除文件
func (m *MinIOClient) DeleteFile(bucket string, objectKey string) error {
	ctx := context.Background()
	return m.client.RemoveObject(ctx, bucket, objectKey, minio.RemoveObjectOptions{})
}

// GetFileURL 获取文件的公开访问 URL
func (m *MinIOClient) GetFileURL(bucket string, objectKey string) string {
	if m.cfg.UseSSL {
		return fmt.Sprintf("https://%s/%s/%s", m.cfg.Endpoint, bucket, objectKey)
	}
	return fmt.Sprintf("http://%s/%s/%s", m.cfg.Endpoint, bucket, objectKey)
}

// PresignedGetURL 生成预签名 GET URL（临时访问）
func (m *MinIOClient) PresignedGetURL(bucket string, objectKey string, expires time.Duration) (string, error) {
	ctx := context.Background()
	url, err := m.client.PresignedGetObject(ctx, bucket, objectKey, expires, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (m *MinIOClient) ensureBucket(bucket string) error {
	ctx := context.Background()
	exists, err := m.client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if !exists {
		return m.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	}
	return nil
}
