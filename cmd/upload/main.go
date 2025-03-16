package main

import (
	"log"

	"cloudpix/config"
	"cloudpix/internal/adapter/handler"
	"cloudpix/internal/adapter/middleware"
	"cloudpix/internal/infrastructure/persistence"
	"cloudpix/internal/infrastructure/storage"
	"cloudpix/internal/usecase"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	// 設定の読み込み
	cfg := config.NewConfig()

	// AWS セッションの初期化
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.AWSRegion),
	})
	if err != nil {
		log.Printf("Error creating AWS session: %s", err)
	}

	// S3とDynamoDBクライアントの初期化
	s3Client := s3.New(sess)
	dbClient := dynamodb.New(sess)

	log.Printf("Upload Lambda initialized with bucket: %s, table: %s", cfg.S3BucketName, cfg.MetadataTableName)

	// リポジトリのセットアップ
	storageRepo := storage.NewS3StorageRepository(s3Client, cfg.AWSRegion)
	metadataRepo := persistence.NewDynamoDBUploadMetadataRepository(dbClient, cfg.MetadataTableName)

	// ユースケースのセットアップ
	uploadUsecase := usecase.NewUploadUsecase(storageRepo, metadataRepo, cfg.S3BucketName)

	// ミドルウェアの作成
	authMiddleware := middleware.CreateDefaultAuthMiddleware(cfg.AWSRegion, cfg.UserPoolID, cfg.UserPoolID)

	// ハンドラのセットアップ
	uploadHandler := handler.NewUploadHandler(uploadUsecase, authMiddleware)

	// Lambda関数のスタート
	lambda.Start(uploadHandler.Handle)
}
