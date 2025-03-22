package s3

import (
	"bytes"
	"cloudpix/internal/domain/thumbnailmanagement/service"
	"cloudpix/internal/domain/thumbnailmanagement/valueobject"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3ThumbnailStorageService はS3を使ったストレージサービスの実装
type S3ThumbnailStorageService struct {
	s3Client  *s3.S3
	awsRegion string
}

// NewS3ThumbnailStorageService は新しいストレージサービスを作成します
func NewS3ThumbnailStorageService(s3Client *s3.S3, awsRegion string) service.StorageService {
	return &S3ThumbnailStorageService{
		s3Client:  s3Client,
		awsRegion: awsRegion,
	}
}

// FetchImage はストレージから画像を取得します
func (s *S3ThumbnailStorageService) FetchImage(ctx context.Context, bucket, key string) (valueobject.ImageData, error) {
	// S3からオブジェクトを取得
	resp, err := s.s3Client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return valueobject.ImageData{}, fmt.Errorf("failed to get S3 object: %w", err)
	}
	defer resp.Body.Close()

	// レスポンスボディを読み込む
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return valueobject.ImageData{}, fmt.Errorf("failed to read S3 object body: %w", err)
	}

	// コンテンツタイプを取得
	contentType := "application/octet-stream"
	if resp.ContentType != nil {
		contentType = *resp.ContentType
	}

	// 画像データを返す
	return valueobject.NewImageData(data, contentType), nil
}

// UploadThumbnail はサムネイルをアップロードします
func (s *S3ThumbnailStorageService) UploadThumbnail(ctx context.Context, bucket, key string, data valueobject.ImageData) error {
	// S3にアップロード
	_, err := s.s3Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
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

// GetObjectURL はオブジェクトの公開URLを生成します
func (s *S3ThumbnailStorageService) GetObjectURL(bucket, key string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, s.awsRegion, key)
}

// DeleteThumbnail はサムネイルを削除します
func (s *S3ThumbnailStorageService) DeleteThumbnail(ctx context.Context, bucket, key string) error {
	_, err := s.s3Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete thumbnail from S3: %w", err)
	}

	return nil
}
