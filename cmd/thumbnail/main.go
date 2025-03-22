// cmd/thumbnail/main.go
package main

import (
	"cloudpix/config"
	s3handler "cloudpix/internal/adapter/event/s3"
	"cloudpix/internal/adapter/middleware"
	"cloudpix/internal/application/thumbnailmanagement/usecase"
	"cloudpix/internal/domain/shared/event/dispatcher"
	"cloudpix/internal/infrastructure/imaging"
	"cloudpix/internal/infrastructure/persistence/dynamodb/thumbnailmanagement"
	s3storage "cloudpix/internal/infrastructure/storage/s3"
	"cloudpix/internal/logging"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
)

// サムネイルサイズ（ピクセル）
const thumbnailSize = 200

func main() {
	// 環境変数の設定（必要に応じて）
	if os.Getenv("ENVIRONMENT") == "dev" {
		os.Setenv("LOG_LEVEL", "debug")
	}

	// ロギングの初期化
	logging.InitLogging()
	logger := logging.GetLogger("ThumbnailLambda")

	// 設定の読み込み
	cfg := config.NewConfig()
	logger.Info("Starting Thumbnail Lambda", map[string]interface{}{
		"config": map[string]string{
			"bucketName":    cfg.S3BucketName,
			"metadataTable": cfg.MetadataTableName,
			"environment":   cfg.Environment,
		},
		"thumbnailSize": thumbnailSize,
	})

	// AWS セッションの初期化
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.AWSRegion),
	})
	if err != nil {
		logger.Fatal(err, "Error creating AWS session", nil)
	}

	// S3とDynamoDBクライアントの初期化
	s3Client := s3.New(sess)
	dbClient := dynamodb.New(sess)

	// インフラストラクチャレイヤーのセットアップ
	thumbnailRepo := thumbnailmanagement.NewDynamoDBThumbnailRepository(dbClient, cfg.MetadataTableName)
	storageService := s3storage.NewS3ThumbnailStorageService(s3Client, cfg.AWSRegion)
	processingService := imaging.NewImageProcessingService()
	eventDispatcher := dispatcher.NewSimpleEventDispatcher()

	// アプリケーションレイヤーのセットアップ
	thumbnailUsecase := usecase.NewThumbnailGenerationUsecase(
		thumbnailRepo,
		storageService,
		processingService,
		eventDispatcher,
		thumbnailSize,
		cfg.AWSRegion,
	)

	// インターフェースレイヤーのセットアップ
	// S3イベント用のハンドラー
	thumbnailHandler := s3handler.NewThumbnailHandler(thumbnailUsecase, logger)

	// ミドルウェア設定の作成
	middlewareCfg := middleware.NewDefaultMiddlewareConfig()
	middlewareCfg.AWSRegion = cfg.AWSRegion
	middlewareCfg.ServiceName = "CloudPix"
	middlewareCfg.OperationName = "GenerateThumbnail"
	middlewareCfg.FunctionName = "ThumbnailLambda"

	// 認証は不要（S3イベント起動のため）
	middlewareCfg.AuthEnabled = false

	// ミドルウェアレジストリの取得
	registry := middleware.GetRegistry()

	// 標準ミドルウェアを登録（認証なし）
	registry.RegisterStandardMiddlewares(sess, middlewareCfg, nil, logger)

	// S3イベント用のミドルウェア処理
	// ここではS3イベント用のカスタムアダプターが必要
	handlerFactory := middleware.NewHandlerFactory(middlewareCfg).WithAWSSession(sess)

	// ミドルウェアを適用したS3イベントハンドラーを作成
	wrappedHandler := handlerFactory.WrapS3EventHandler(thumbnailHandler.Handle)

	// Lambda関数のスタート
	lambda.Start(wrappedHandler)
}
