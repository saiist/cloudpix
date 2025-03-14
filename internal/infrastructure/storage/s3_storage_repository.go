package storage

import (
	"bytes"
	"cloudpix/internal/domain/model"
	"cloudpix/internal/domain/repository"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3StorageRepository struct {
	s3Client  *s3.S3
	awsRegion string
}

func NewS3StorageRepository(s3Client *s3.S3, awsRegion string) repository.StorageRepository {
	return &S3StorageRepository{
		s3Client:  s3Client,
		awsRegion: awsRegion,
	}
}

// S3から画像を取得
func (r *S3StorageRepository) FetchImage(ctx context.Context, bucket, key string) (*model.ImageData, error) {
	// S3からオブジェクトを取得
	resp, err := r.s3Client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get S3 object: %w", err)
	}
	defer resp.Body.Close()

	// レスポンスボディを読み込む
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read S3 object body: %w", err)
	}

	// コンテンツタイプを取得
	contentType := "application/octet-stream"
	if resp.ContentType != nil {
		contentType = *resp.ContentType
	}

	return &model.ImageData{
		Data:        data,
		ContentType: contentType,
	}, nil
}

// サムネイルをS3にアップロード
func (r *S3StorageRepository) UploadThumbnail(ctx context.Context, bucket, key string, data *model.ImageData) error {
	// S3にアップロード
	_, err := r.s3Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data.Data),
		ContentType: aws.String(data.ContentType),
	})

	if err != nil {
		return fmt.Errorf("failed to upload thumbnail to S3: %w", err)
	}

	return nil
}

// Base64エンコードされた画像データをS3にアップロード
func (r *S3StorageRepository) UploadImage(ctx context.Context, bucket, key, contentType, base64Data string) (string, error) {
	// Base64デコード
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 data: %w", err)
	}

	// S3にアップロード
	_, err = r.s3Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(imageData),
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// ダウンロードURLを生成
	downloadURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, r.awsRegion, key)

	return downloadURL, nil
}

// S3へのアップロード用プレサインドURLを生成
func (r *S3StorageRepository) GeneratePresignedURL(ctx context.Context, bucket, key, contentType string, expiration time.Duration) (string, string, error) {
	// プレサインドURLリクエストを作成
	req, _ := r.s3Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	})

	// プレサインドURLを生成
	uploadURL, err := req.Presign(expiration)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	// ダウンロードURLを生成
	downloadURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, r.awsRegion, key)

	return uploadURL, downloadURL, nil
}
