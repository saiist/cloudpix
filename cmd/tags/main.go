// cmd/tags/main.go
package main

import (
	"cloudpix/cmd/shared"
	"cloudpix/config"
	"cloudpix/internal/adapter/api/handler"
	"cloudpix/internal/adapter/middleware"
	"cloudpix/internal/application/tagmanagement/usecase"
	"cloudpix/internal/domain/shared/event/dispatcher"
	"cloudpix/internal/infrastructure/persistence/dynamodb/tagmanagement"
	"cloudpix/internal/logging"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func main() {
	// 環境変数の設定（必要に応じて）
	if os.Getenv("ENVIRONMENT") == "dev" {
		os.Setenv("LOG_LEVEL", "debug")
	}

	// ロギングの初期化
	logging.InitLogging()
	logger := logging.GetLogger("TagsLambda")

	// 設定の読み込み
	cfg := config.NewConfig()
	logger.Info("Starting Tags Lambda", map[string]interface{}{
		"config": map[string]string{
			"tagsTable":     cfg.TagsTableName,
			"metadataTable": cfg.MetadataTableName,
			"environment":   cfg.Environment,
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
		"tagsTable":     cfg.TagsTableName,
		"metadataTable": cfg.MetadataTableName,
	})

	// インフラストラクチャレイヤーのセットアップ
	tagRepo := tagmanagement.NewDynamoDBTagRepository(dbClient, cfg.TagsTableName, cfg.MetadataTableName)
	eventDispatcher := dispatcher.NewSimpleEventDispatcher()

	// アプリケーションレイヤーのセットアップ
	tagUsecase := usecase.NewTagUsecase(tagRepo, eventDispatcher)

	// インターフェースレイヤーのセットアップ
	tagHandler := handler.NewTagHandler(tagUsecase)

	// ミドルウェア設定の作成
	middlewareCfg := middleware.NewDefaultMiddlewareConfig()
	middlewareCfg.AWSRegion = cfg.AWSRegion
	middlewareCfg.UserPoolID = cfg.UserPoolID
	middlewareCfg.ClientID = cfg.ClientID
	middlewareCfg.ServiceName = "CloudPix"
	middlewareCfg.OperationName = "TagManagement"
	middlewareCfg.FunctionName = "TagsLambda"

	// 環境に基づくログ詳細度の設定
	if cfg.Environment == "dev" {
		middlewareCfg.DetailedRequestLog = true
		middlewareCfg.DetailedResponseLog = true
		middlewareCfg.IncludeBody = true
	} else {
		// 本番環境では最小限のログ
		middlewareCfg.DetailedRequestLog = false
		middlewareCfg.DetailedResponseLog = false
		middlewareCfg.IncludeBody = false
	}

	// 認証コンポーネントの初期化
	authUsecase := shared.InitAuth(cfg, sess, logger)

	// ミドルウェアレジストリの取得
	registry := middleware.GetRegistry()

	// 標準ミドルウェアを登録
	registry.RegisterStandardMiddlewares(sess, middlewareCfg, authUsecase, logger)

	// ミドルウェア名の順序を指定（ロギングが最初、認証が最後）
	middlewareNames := []string{"logging", "metrics", "auth"}

	// ミドルウェアチェーンの構築
	chain := registry.BuildChain(middlewareNames)

	// ハンドラーにミドルウェアを適用
	wrappedHandler := chain.Then(tagHandler.Handle)

	// Lambda関数のスタート
	lambda.Start(wrappedHandler)
}
