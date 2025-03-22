package main

import (
	"cloudpix/config"
	scheduler_handler "cloudpix/internal/adapter/event/scheduler"
	"cloudpix/internal/adapter/middleware"
	"cloudpix/internal/application/imagemanagement/usecase"
	"cloudpix/internal/domain/shared/event/dispatcher"
	"cloudpix/internal/infrastructure/cleanup"
	"cloudpix/internal/infrastructure/persistence/dynamodb/imagemanagement"
	"cloudpix/internal/infrastructure/persistence/dynamodb/tagmanagement"
	storageS3 "cloudpix/internal/infrastructure/storage/s3"
	"cloudpix/internal/logging"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	// ロギングの初期化
	logging.InitLogging()
	logger := logging.GetLogger("CleanupLambda")

	// 設定の読み込み
	cfg := config.NewConfig()

	logger.Info("Starting Cleanup Lambda", map[string]interface{}{
		"config": map[string]string{
			"bucketName":    cfg.S3BucketName,
			"metadataTable": cfg.MetadataTableName,
			"tagsTable":     cfg.TagsTableName,
		},
		"retentionDays": cfg.ImageRetentionDays,
	})

	// AWS セッションの初期化
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.AWSRegion),
	})
	if err != nil {
		logger.Fatal(err, "Error creating AWS session", nil)
	}

	// クライアントの初期化
	s3Client := s3.New(sess)
	dbClient := dynamodb.New(sess)
	logger.Info("DynamoDB client initialized", map[string]interface{}{
		"tableName": cfg.MetadataTableName,
	})

	// インフラストラクチャレイヤーのセットアップ
	imageRepo := imagemanagement.NewDynamoDBImageRepository(dbClient, cfg.MetadataTableName)
	tagRepo := tagmanagement.NewDynamoDBTagRepository(dbClient, cfg.TagsTableName, cfg.MetadataTableName)
	storageService := storageS3.NewS3StorageService(s3Client, cfg.AWSRegion)
	cleanupService := cleanup.NewS3CleanupService(s3Client, dbClient, cfg.S3BucketName, cfg.MetadataTableName, cfg.TagsTableName)
	eventDispatcher := dispatcher.NewSimpleEventDispatcher()

	// アプリケーションレイヤーのセットアップ
	cleanupUsecase := usecase.NewCleanupUsecase(
		imageRepo,
		tagRepo,
		storageService,
		cleanupService,
		eventDispatcher,
		cfg.ImageRetentionDays,
		logger,
	)

	// ハンドラーのセットアップ
	cleanupHandler := scheduler_handler.NewCleanupHandler(cleanupUsecase, logger)

	// ミドルウェア設定の作成
	middlewareCfg := middleware.NewDefaultMiddlewareConfig()
	middlewareCfg.AWSRegion = cfg.AWSRegion
	middlewareCfg.ServiceName = "CloudPix"
	middlewareCfg.OperationName = "CleanupProcess"
	middlewareCfg.FunctionName = "CleanupLambda"

	// 認証は不要（スケジュールタスクのため）
	middlewareCfg.AuthEnabled = false

	// ミドルウェアレジストリの取得
	registry := middleware.GetRegistry()

	// 標準ミドルウェアを登録（認証なし）
	registry.RegisterStandardMiddlewares(sess, middlewareCfg, nil, logger)

	// イベント用のミドルウェア処理
	handlerFactory := middleware.NewHandlerFactory(middlewareCfg).WithAWSSession(sess)

	// ミドルウェアを適用したスケジュールイベントハンドラーを作成
	wrappedHandler := handlerFactory.WrapCloudWatchEventHandler(cleanupHandler.Handle)

	// Lambda関数のスタート
	lambda.Start(wrappedHandler)
}
