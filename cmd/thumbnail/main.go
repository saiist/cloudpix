package main

import (
	"log"

	"cloudpix/config"
	"cloudpix/internal/adapter/handler"
	"cloudpix/internal/adapter/middleware"
	"cloudpix/internal/infrastructure/imaging"
	"cloudpix/internal/infrastructure/persistence"
	"cloudpix/internal/infrastructure/storage"
	"cloudpix/internal/usecase"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
)

// サムネイルサイズ（ピクセル）
const thumbnailSize = 200

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

	log.Printf("Thumbnail Lambda initialized with bucket: %s, thumbnail size: %d", cfg.S3BucketName, thumbnailSize)

	// リポジトリのセットアップ
	thumbnailRepo := persistence.NewDynamoDBThumbnailRepository(dbClient, cfg.MetadataTableName)
	storageRepo := storage.NewS3StorageRepository(s3Client, cfg.AWSRegion)

	// サービスのセットアップ
	imageService := imaging.NewImageService()

	// ユースケースのセットアップ
	thumbnailUsecase := usecase.NewThumbnailUsecase(thumbnailRepo, storageRepo, imageService, cfg.AWSRegion, thumbnailSize)

	// ミドルウェア設定の作成
	middlewareCfg := middleware.NewDefaultMiddlewareConfig()
	middlewareCfg.AWSRegion = cfg.AWSRegion
	middlewareCfg.ServiceName = "CloudPix"
	middlewareCfg.OperationName = "GenerateThumbnail"
	middlewareCfg.FunctionName = "ThumbnailLambda"

	// ハンドラーファクトリの作成
	handlerFactory := middleware.NewHandlerFactory(middlewareCfg).WithAWSSession(sess)

	// ハンドラのセットアップ
	thumbnailHandler := handler.NewThumbnailHandler(thumbnailUsecase)

	// ミドルウェアを適用したS3イベントハンドラーを作成
	wrappedHandler := handlerFactory.WrapS3EventHandler(thumbnailHandler.Handle)

	// Lambda関数のスタート
	lambda.Start(wrappedHandler)
}
