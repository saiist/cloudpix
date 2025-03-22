package main

import (
	"cloudpix/config"
	"cloudpix/internal/adapter/handler"
	"cloudpix/internal/adapter/middleware"
	"cloudpix/internal/infrastructure/persistence"
	"cloudpix/internal/infrastructure/storage"
	"cloudpix/internal/logging"
	"cloudpix/internal/usecase"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	// 環境変数の設定（ロギングシステムの初期化前に必要な場合）
	// 例: 開発環境ではデバッグログを有効化
	if os.Getenv("ENVIRONMENT") == "dev" {
		os.Setenv("LOG_LEVEL", "debug")
	}

	// ロギングの初期化
	logging.InitLogging()
	logger := logging.GetLogger("UploadLambda")

	// 設定の読み込み
	cfg := config.NewConfig()
	logger.Info("Starting Upload Lambda", map[string]interface{}{
		"config": map[string]string{
			"bucketName":    cfg.S3BucketName,
			"metadataTable": cfg.MetadataTableName,
			"environment":   cfg.Environment,
			"enableMetrics": fmt.Sprint(cfg.EnableMetrics),
			"enableXRay":    fmt.Sprint(cfg.EnableXRay),
		},
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
	storageRepo := storage.NewS3StorageRepository(s3Client, cfg.AWSRegion)
	metadataRepo := persistence.NewDynamoDBUploadMetadataRepository(dbClient, cfg.MetadataTableName)

	// ユースケースのセットアップ
	uploadUsecase := usecase.NewUploadUsecase(storageRepo, metadataRepo, cfg.S3BucketName)

	// ハンドラのセットアップ
	uploadHandler := handler.NewUploadHandler(uploadUsecase)

	// ミドルウェア設定の作成
	middlewareCfg := middleware.NewDefaultMiddlewareConfig()
	middlewareCfg.AWSRegion = cfg.AWSRegion
	middlewareCfg.UserPoolID = cfg.UserPoolID
	middlewareCfg.ClientID = cfg.ClientID
	middlewareCfg.ServiceName = "CloudPix"
	middlewareCfg.OperationName = "UploadImage"
	middlewareCfg.FunctionName = "UploadLambda"

	// 環境に基づくログ詳細度の設定
	if cfg.Environment == "dev" {
		middlewareCfg.DetailedRequestLog = true
		middlewareCfg.DetailedResponseLog = true
		middlewareCfg.IncludeHeaders = true
		middlewareCfg.IncludeBody = true
	} else {
		// 本番環境では最小限のログ
		middlewareCfg.DetailedRequestLog = false
		middlewareCfg.DetailedResponseLog = false
		middlewareCfg.IncludeHeaders = false
		middlewareCfg.IncludeBody = false
	}

	// ミドルウェアレジストリの取得
	registry := middleware.GetRegistry()

	// 標準ミドルウェアを登録
	registry.RegisterStandardMiddlewares(sess, middlewareCfg)

	// ミドルウェア名の順序を指定（ロギングが最初、認証が最後）
	middlewareNames := []string{"logging", "metrics", "auth"}

	// ミドルウェアチェーンの構築
	chain := registry.BuildChain(middlewareNames)

	// ハンドラーにミドルウェアを適用
	wrappedHandler := chain.Then(uploadHandler.Handle)

	// Lambda関数のスタート
	lambda.Start(wrappedHandler)
}
