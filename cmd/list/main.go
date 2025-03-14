package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"cloudpix/config"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

// メタデータ構造体
type ImageMetadata struct {
	ImageID     string `json:"imageId"`
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
	Size        int    `json:"size"`
	UploadDate  string `json:"uploadDate"`
	S3ObjectKey string `json:"s3ObjectKey"`
	DownloadURL string `json:"downloadUrl"`
}

// リストレスポンス構造体
type ListResponse struct {
	Images []ImageMetadata `json:"images"`
	Count  int             `json:"count"`
}

// 環境変数
var (
	cfg            = config.NewConfig()
	dynamoDBClient *dynamodb.DynamoDB
)

func init() {

	// AWS セッションの初期化
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.AWSRegion),
	})
	if err != nil {
		log.Printf("Error creating session: %s", err)
	}

	// DynamoDBクライアントの初期化
	dynamoDBClient = dynamodb.New(sess)

	log.Printf("Lambda initialized with DynamoDB table: %s", cfg.DynamoDBTableName)
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Processing image list request")

	// クエリパラメータからフィルターを取得
	date := request.QueryStringParameters["date"]

	var result *dynamodb.ScanOutput
	var err error

	if date != "" {
		// 日付によるフィルタリング
		log.Printf("Filtering by date: %s", date)

		// フィルタ式を作成
		filt := expression.Name("UploadDate").Equal(expression.Value(date))
		expr, err := expression.NewBuilder().WithFilter(filt).Build()

		if err != nil {
			log.Printf("Error building expression: %s", err)
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       fmt.Sprintf(`{"error":"Failed to build query: %s"}`, err),
			}, nil
		}

		// DynamoDBをクエリ
		result, err = dynamoDBClient.Scan(&dynamodb.ScanInput{
			TableName:                 aws.String(cfg.DynamoDBTableName),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			FilterExpression:          expr.Filter(),
		})
	} else {
		// フィルターなしで全件取得
		result, err = dynamoDBClient.Scan(&dynamodb.ScanInput{
			TableName: aws.String(cfg.DynamoDBTableName),
		})
	}

	if err != nil {
		log.Printf("Error scanning DynamoDB: %s", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf(`{"error":"Failed to retrieve images: %s"}`, err),
		}, nil
	}

	// 結果を構造体の配列にアンマーシャル
	var images []ImageMetadata
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &images)
	if err != nil {
		log.Printf("Error unmarshalling DynamoDB result: %s", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf(`{"error":"Failed to parse result: %s"}`, err),
		}, nil
	}

	// レスポンスを作成
	response := ListResponse{
		Images: images,
		Count:  len(images),
	}

	// レスポンスをJSON形式で返す
	responseJSON, _ := json.Marshal(response)
	log.Printf("Returning response with %d images", len(images))

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseJSON),
	}, nil
}

func main() {
	lambda.Start(handler)
}
