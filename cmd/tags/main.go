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

	// ミドルウェアの作成
	authMiddleware := middleware.CreateDefaultAuthMiddleware(cfg.AWSRegion, cfg.UserPoolID, cfg.ClientID)

	// ハンドラのセットアップ
	tagHandler := handler.NewTagHandler(tagUsecase, authMiddleware)

	// Lambda関数のスタート
	lambda.Start(tagHandler.Handle)
}
