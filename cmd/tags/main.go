package main

import (
	"log"

	"cloudpix/config"
	"cloudpix/internal/adapter/handler"
	"cloudpix/internal/adapter/middleware"
	"cloudpix/internal/infrastructure/persistence"
	"cloudpix/internal/usecase"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
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

	// DynamoDBクライアントの初期化
	dbClient := dynamodb.New(sess)
	log.Printf("Tags Lambda initialized with tables: Tags=%s, Metadata=%s", cfg.TagsTableName, cfg.MetadataTableName)

	// リポジトリのセットアップ
	tagRepo := persistence.NewDynamoDBTagRepository(dbClient, cfg.TagsTableName, cfg.MetadataTableName)

	// ユースケースのセットアップ
	tagUsecase := usecase.NewTagUsecase(tagRepo)

	// ハンドラのセットアップ
	tagHandler := handler.NewTagHandler(tagUsecase)

	// ミドルウェア設定の作成
	middlewareCfg := middleware.NewDefaultMiddlewareConfig()
	middlewareCfg.AWSRegion = cfg.AWSRegion
	middlewareCfg.UserPoolID = cfg.UserPoolID
	middlewareCfg.ClientID = cfg.ClientID
	middlewareCfg.ServiceName = "CloudPix"
	middlewareCfg.OperationName = "ListImages"
	middlewareCfg.FunctionName = "ListLambda"

	// ハンドラーファクトリの作成
	handlerFactory := middleware.NewHandlerFactory(middlewareCfg).WithAWSSession(sess)

	// ミドルウェアを適用したハンドラーを作成
	wrappedHandler := handlerFactory.WrapAPIGatewayHandler(tagHandler.Handle)

	// Lambda関数のスタート
	lambda.Start(wrappedHandler)

}
