package s3

import (
	"bytes"
	"cloudpix/internal/domain/imagemanagement/service"
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3StorageService はS3を使ったストレージサービスの実装
type S3StorageService struct {
	s3Client  *s3.S3
	awsRegion string
}

// NewS3StorageService は新しいS3ストレージサービスを作成します
func NewS3StorageService(s3Client *s3.S3, awsRegion string) service.StorageService {
	return &S3StorageService{
		s3Client:  s3Client,
		awsRegion: awsRegion,
	}
}

// StoreImage はBase64エンコードされた画像データをS3に保存します
func (s *S3StorageService) StoreImage(ctx context.Context, bucket, key, contentType, base64Data string) (string, error) {
	// Base64デコード
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 data: %w", err)
	}

	// S3にアップロード
	_, err = s.s3Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(imageData),
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// ダウンロードURLを生成
	downloadURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, s.awsRegion, key)

	return downloadURL, nil
}

// GenerateImageURL は画像アップロード用のプレサインドURLを生成します
func (s *S3StorageService) GenerateImageURL(ctx context.Context, bucket, key, contentType string, expiration time.Duration) (string, string, error) {
	// プレサインドURLリクエストを作成
	req, _ := s.s3Client.PutObjectRequest(&s3.PutObjectInput{
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
	downloadURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, s.awsRegion, key)

	return uploadURL, downloadURL, nil
}

// DeleteImage はS3から画像を削除します
func (s *S3StorageService) DeleteImage(ctx context.Context, bucket, key string) error {
	_, err := s.s3Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object from S3: %w", err)
	}

	return nil
}
