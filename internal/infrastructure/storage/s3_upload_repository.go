package storage

import (
	"bytes"
	"cloudpix/internal/domain/model"
	"cloudpix/internal/domain/repository"
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3UploadRepository struct {
	s3Client          *s3.S3
	dynamoDBClient    *dynamodb.DynamoDB
	bucketName        string
	metadataTableName string
	awsRegion         string
}

func NewS3UploadRepository(s3Client *s3.S3, dynamoDBClient *dynamodb.DynamoDB, bucketName, metadataTableName, awsRegion string) repository.UploadRepository {
	return &S3UploadRepository{
		s3Client:          s3Client,
		dynamoDBClient:    dynamoDBClient,
		bucketName:        bucketName,
		metadataTableName: metadataTableName,
		awsRegion:         awsRegion,
	}
}

// UploadImage はBase64エンコードされた画像データをS3にアップロードします
func (r *S3UploadRepository) UploadImage(ctx context.Context, imageID string, request *model.UploadRequest) (string, error) {
	// オブジェクトキーを生成
	objectKey := fmt.Sprintf("uploads/%s-%s", imageID, request.FileName)

	// Base64デコード
	imageData, err := base64.StdEncoding.DecodeString(request.Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 data: %w", err)
	}

	// S3にアップロード
	_, err = r.s3Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(imageData),
		ContentType: aws.String(request.ContentType),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// ダウンロードURLを生成
	downloadURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", r.bucketName, r.awsRegion, objectKey)

	return downloadURL, nil
}

// GeneratePresignedURL はS3へのアップロード用プレサインドURLを生成します
func (r *S3UploadRepository) GeneratePresignedURL(ctx context.Context, imageID string, request *model.UploadRequest, expiration time.Duration) (string, string, error) {
	// オブジェクトキーを生成
	objectKey := fmt.Sprintf("uploads/%s-%s", imageID, request.FileName)

	// プレサインドURLリクエストを作成
	req, _ := r.s3Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(objectKey),
		ContentType: aws.String(request.ContentType),
	})

	// プレサインドURLを生成
	uploadURL, err := req.Presign(expiration)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	// ダウンロードURLを生成
	downloadURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", r.bucketName, r.awsRegion, objectKey)

	return uploadURL, downloadURL, nil
}

// SaveMetadata はアップロードされた画像のメタデータをDynamoDBに保存します
func (r *S3UploadRepository) SaveMetadata(ctx context.Context, metadata *model.UploadMetadata) error {
	// DynamoDBのアイテム形式に変換
	item, err := dynamodbattribute.MarshalMap(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal DynamoDB item: %w", err)
	}

	// DynamoDBにメタデータを保存
	_, err = r.dynamoDBClient.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.metadataTableName),
		Item:      item,
	})

	if err != nil {
		return fmt.Errorf("failed to save metadata to DynamoDB: %w", err)
	}

	return nil
}
