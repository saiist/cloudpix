package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

// リクエスト構造体
type UploadRequest struct {
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
	Data        string `json:"data,omitempty"` // Base64エンコードされた画像データ
}

// レスポンス構造体
type UploadResponse struct {
	ImageID     string `json:"imageId"`
	UploadURL   string `json:"uploadUrl,omitempty"`
	DownloadURL string `json:"downloadUrl"`
	Message     string `json:"message"`
}

// 環境変数
var (
	s3BucketName = os.Getenv("S3_BUCKET_NAME")
	awsRegion    = os.Getenv("AWS_REGION")
	s3Client     *s3.S3
)

func init() {
	// AWS セッションとS3クライアントの初期化
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	})
	if err != nil {
		log.Printf("Error creating session: %s", err)
	}
	s3Client = s3.New(sess)

	log.Printf("Lambda initialized with bucket: %s", s3BucketName)
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

	// ユニークなファイル名（Object Key）を生成
	imageID := uuid.New().String()
	objectKey := fmt.Sprintf("uploads/%s-%s", imageID, uploadReq.FileName)
	log.Printf("Generated object key: %s", objectKey)

	var uploadResponse UploadResponse

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

		// S3にアップロード
		_, err = s3Client.PutObject(&s3.PutObjectInput{
			Bucket:      aws.String(s3BucketName),
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

		log.Printf("Successfully uploaded to S3: %s/%s", s3BucketName, objectKey)

		uploadResponse = UploadResponse{
			ImageID:     imageID,
			DownloadURL: fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s3BucketName, awsRegion, objectKey),
			Message:     "Image uploaded successfully",
		}
	} else {
		// 直接アップロードされない場合、プレサインドURLを生成して返す
		log.Printf("Generating presigned URL")

		req, _ := s3Client.PutObjectRequest(&s3.PutObjectInput{
			Bucket:      aws.String(s3BucketName),
			Key:         aws.String(objectKey),
			ContentType: aws.String(uploadReq.ContentType),
		})

		// 15分間有効なプレサインドURLを生成
		uploadURL, err := req.Presign(15 * time.Minute)
		if err != nil {
			log.Printf("Error generating presigned URL: %s", err)
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       fmt.Sprintf(`{"error":"Failed to generate presigned URL: %s"}`, err),
			}, nil
		}

		uploadResponse = UploadResponse{
			ImageID:     imageID,
			UploadURL:   uploadURL,
			DownloadURL: fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s3BucketName, awsRegion, objectKey),
			Message:     "Use the uploadUrl to upload your image",
		}
	}

	// レスポンスをJSON形式で返す
	responseJSON, _ := json.Marshal(uploadResponse)
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
