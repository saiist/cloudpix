package main

import (
	"bytes"
	"cloudpix/config"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

// リクエスト構造体
type UploadRequest struct {
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
	Data        string `json:"data,omitempty"` // Base64エンコードされた画像データ
}

// メタデータ構造体（DynamoDBに保存）
type ImageMetadata struct {
	ImageID      string    `json:"ImageID"`
	FileName     string    `json:"fileName"`
	ContentType  string    `json:"contentType"`
	Size         int       `json:"size"`
	UploadDate   string    `json:"UploadDate"`
	CreatedAt    time.Time `json:"createdAt"`
	S3ObjectKey  string    `json:"s3ObjectKey"`
	S3BucketName string    `json:"s3BucketName"`
	DownloadURL  string    `json:"downloadUrl"`
}

// レスポンス構造体
type UploadResponse struct {
	ImageID     string `json:"imageId"`
	UploadURL   string `json:"uploadUrl,omitempty"`
	DownloadURL string `json:"downloadUrl"`
	Message     string `json:"message"`
}

var (
	cfg            = config.NewConfig()
	s3Client       *s3.S3
	dynamoDBClient *dynamodb.DynamoDB
)

func init() {
	// AWS セッションの初期化
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.AWSRegion),
	})
	if err != nil {
		log.Printf("Error creating session: %s", err)
	}

	// クライアントの初期化
	s3Client = s3.New(sess)
	dynamoDBClient = dynamodb.New(sess)

	log.Printf("Lambda initialized with bucket: %s, DynamoDB table: %s", cfg.S3BucketName, cfg.MetadataTableName)
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Received request: %s", request.Body)

	// リクエストボディをパース
	var uploadReq UploadRequest
	err := json.Unmarshal([]byte(request.Body), &uploadReq)
	if err != nil {
		log.Printf("Error parsing request: %s", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       fmt.Sprintf(`{"error":"Invalid request format: %s"}`, err),
		}, nil
	}

	// ユニークなIDを生成
	imageID := uuid.New().String()
	objectKey := fmt.Sprintf("uploads/%s-%s", imageID, uploadReq.FileName)
	log.Printf("Generated object key: %s", objectKey)

	var downloadURL string
	var uploadURL string
	now := time.Now()
	todayDate := now.Format("2006-01-02") // YYYY-MM-DD形式
	var imageSize int

	// Base64エンコードされた画像データがある場合、直接S3にアップロード
	if uploadReq.Data != "" {
		log.Printf("Uploading base64 data to S3")

		// Base64デコード
		imageData, err := base64.StdEncoding.DecodeString(uploadReq.Data)
		if err != nil {
			log.Printf("Error decoding base64 data: %s", err)
			return events.APIGatewayProxyResponse{
				StatusCode: 400,
				Body:       fmt.Sprintf(`{"error":"Invalid base64 data: %s"}`, err),
			}, nil
		}

		imageSize = len(imageData)

		// S3にアップロード
		_, err = s3Client.PutObject(&s3.PutObjectInput{
			Bucket:      aws.String(cfg.S3BucketName),
			Key:         aws.String(objectKey),
			Body:        bytes.NewReader(imageData),
			ContentType: aws.String(uploadReq.ContentType),
		})

		if err != nil {
			log.Printf("Error uploading to S3: %s", err)
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       fmt.Sprintf(`{"error":"Failed to upload to S3: %s"}`, err),
			}, nil
		}

		log.Printf("Successfully uploaded to S3: %s/%s", cfg.S3BucketName, objectKey)
		downloadURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.S3BucketName, cfg.AWSRegion, objectKey)
	} else {
		// 直接アップロードされない場合、プレサインドURLを生成して返す
		log.Printf("Generating presigned URL")

		req, _ := s3Client.PutObjectRequest(&s3.PutObjectInput{
			Bucket:      aws.String(cfg.S3BucketName),
			Key:         aws.String(objectKey),
			ContentType: aws.String(uploadReq.ContentType),
		})

		// 15分間有効なプレサインドURLを生成
		signedURL, err := req.Presign(15 * time.Minute)
		if err != nil {
			log.Printf("Error generating presigned URL: %s", err)
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       fmt.Sprintf(`{"error":"Failed to generate presigned URL: %s"}`, err),
			}, nil
		}

		uploadURL = signedURL
		downloadURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.S3BucketName, cfg.AWSRegion, objectKey)

		// プレサインドURLの場合、サイズは不明
		imageSize = 0
	}

	// メタデータをDynamoDBに保存
	metadata := ImageMetadata{
		ImageID:      imageID,
		FileName:     uploadReq.FileName,
		ContentType:  uploadReq.ContentType,
		Size:         imageSize,
		UploadDate:   todayDate,
		CreatedAt:    now,
		S3ObjectKey:  objectKey,
		S3BucketName: cfg.S3BucketName,
		DownloadURL:  downloadURL,
	}

	// DynamoDBのアイテム形式に変換
	item, err := dynamodbattribute.MarshalMap(metadata)
	if err != nil {
		log.Printf("Error marshalling DynamoDB item: %s", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf(`{"error":"Failed to prepare metadata: %s"}`, err),
		}, nil
	}

	// DynamoDBにメタデータを保存
	_, err = dynamoDBClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(cfg.MetadataTableName),
		Item:      item,
	})

	if err != nil {
		log.Printf("Error saving to DynamoDB: %s", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf(`{"error":"Failed to save metadata: %s"}`, err),
		}, nil
	}

	log.Printf("Successfully saved metadata to DynamoDB")

	// レスポンスを作成
	var response UploadResponse
	if uploadURL != "" {
		response = UploadResponse{
			ImageID:     imageID,
			UploadURL:   uploadURL,
			DownloadURL: downloadURL,
			Message:     "Use the uploadUrl to upload your image",
		}
	} else {
		response = UploadResponse{
			ImageID:     imageID,
			DownloadURL: downloadURL,
			Message:     "Image uploaded successfully",
		}
	}

	// レスポンスをJSON形式で返す
	responseJSON, _ := json.Marshal(response)
	log.Printf("Returning response: %s", string(responseJSON))

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseJSON),
	}, nil
}

func main() {
	lambda.Start(handler)
}
