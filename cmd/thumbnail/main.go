package main

import (
	"cloudpix/config"
	"cloudpix/internal/adapter/handler"
	"cloudpix/internal/adapter/middleware"
	"cloudpix/internal/infrastructure/imaging"
	"cloudpix/internal/infrastructure/persistence"
	"cloudpix/internal/infrastructure/storage"
	"cloudpix/internal/logging"
	"cloudpix/internal/usecase"
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

	// リポジトリのセットアップ
	thumbnailRepo := persistence.NewDynamoDBThumbnailRepository(dbClient, cfg.MetadataTableName)
	storageRepo := storage.NewS3StorageRepository(s3Client, cfg.AWSRegion)

	// サービスのセットアップ
	imageService := imaging.NewImageService()

	// ユースケースのセットアップ
	thumbnailUsecase := usecase.NewThumbnailUsecase(thumbnailRepo, storageRepo, imageService, cfg.AWSRegion, thumbnailSize)

	// ハンドラのセットアップ
	thumbnailHandler := handler.NewThumbnailHandler(thumbnailUsecase)

	// ミドルウェア設定の作成
	// 注意: サムネイル関数はS3イベントを受け取るため、認証ミドルウェアは不要
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
	registry.RegisterStandardMiddlewares(sess, middlewareCfg)

	// S3イベント用のミドルウェア処理
	// ここではS3イベント用のカスタムアダプターが必要
	handlerFactory := middleware.NewHandlerFactory(middlewareCfg).WithAWSSession(sess)

	// ミドルウェアを適用したS3イベントハンドラーを作成
	wrappedHandler := handlerFactory.WrapS3EventHandler(thumbnailHandler.Handle)

	// Lambda関数のスタート
	lambda.Start(wrappedHandler)
}
