package main

import (
	"cloudpix/config"
	"cloudpix/internal/adapter/handler"
	"cloudpix/internal/adapter/middleware"
	"cloudpix/internal/infrastructure/persistence"
	"cloudpix/internal/logging"
	"cloudpix/internal/usecase"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func main() {
	// 環境変数の設定（ロギングシステムの初期化前に必要な場合）
	if os.Getenv("ENVIRONMENT") == "dev" {
		os.Setenv("LOG_LEVEL", "debug")
	}

	// ロギングの初期化
	logging.InitLogging()
	logger := logging.GetLogger("ListLambda")

	// 設定の読み込み
	cfg := config.NewConfig()
	logger.Info("Starting List Lambda", map[string]interface{}{
		"config": map[string]string{
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

	// DynamoDBクライアントの初期化
	dbClient := dynamodb.New(sess)
	logger.Info("DynamoDB client initialized", map[string]interface{}{
		"tableName": cfg.MetadataTableName,
	})

	// リポジトリのセットアップ
	metaRepo := persistence.NewDynamoDBMetadataRepository(dbClient, cfg.MetadataTableName)

	// ユースケースのセットアップ
	metaUsecase := usecase.NewMetadataUsecase(metaRepo)

	// ハンドラのセットアップ
	metaHandler := handler.NewListHandler(metaUsecase)

	// ミドルウェア設定の作成
	middlewareCfg := middleware.NewDefaultMiddlewareConfig()
	middlewareCfg.AWSRegion = cfg.AWSRegion
	middlewareCfg.UserPoolID = cfg.UserPoolID
	middlewareCfg.ClientID = cfg.ClientID
	middlewareCfg.ServiceName = "CloudPix"
	middlewareCfg.OperationName = "ListImages"
	middlewareCfg.FunctionName = "ListLambda"

	// 環境に基づくログ詳細度の設定
	if cfg.Environment == "dev" {
		middlewareCfg.DetailedRequestLog = true
		middlewareCfg.DetailedResponseLog = true
		middlewareCfg.IncludeQueryParams = true
	} else {
		// 本番環境では最小限のログ（ボディなし、クエリパラメータのみ）
		middlewareCfg.DetailedRequestLog = false
		middlewareCfg.DetailedResponseLog = false
		middlewareCfg.IncludeQueryParams = true
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
	wrappedHandler := chain.Then(metaHandler.Handle)

	// Lambda関数のスタート
	lambda.Start(wrappedHandler)
}
