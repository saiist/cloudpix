package storage

import (
	"bytes"
	"cloudpix/internal/domain/model"
	"cloudpix/internal/domain/repository"
	"context"
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3を使用したストレージリポジトリの実装
type S3StorageRepository struct {
	s3Client *s3.S3
}

// 新しいS3ストレージリポジトリを作成
func NewS3StorageRepository(s3Client *s3.S3) repository.StorageRepository {
	return &S3StorageRepository{
		s3Client: s3Client,
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
	data, err := ioutil.ReadAll(resp.Body)
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
