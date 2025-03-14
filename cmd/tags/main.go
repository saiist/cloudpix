package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"cloudpix/config"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// 環境変数
var (
	cfg            = config.NewConfig()
	dynamoDBClient *dynamodb.DynamoDB
)

// タグ情報の構造体
type TagItem struct {
	TagName string `json:"tagName"`
	ImageID string `json:"imageId"`
}

// タグ追加リクエスト
type AddTagRequest struct {
	ImageID string   `json:"imageId"`
	Tags    []string `json:"tags"`
}

// タグ削除リクエスト
type RemoveTagRequest struct {
	ImageID string   `json:"imageId"`
	Tags    []string `json:"tags"`
}

// タグ一覧レスポンス
type TagsResponse struct {
	Tags  []string `json:"tags"`
	Count int      `json:"count"`
}

// 初期化関数
func init() {
	// AWS セッションの初期化
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.AWSRegion),
	})
	if err != nil {
		log.Printf("Error creating AWS session: %s", err)
	}

	// DynamoDBクライアントの初期化
	dynamoDBClient = dynamodb.New(sess)

	log.Printf("Tags Lambda initialized with tables: Tags=%s, Metadata=%s", cfg.TagsTableName, cfg.MetaTableName)
}

// ハンドラー関数
func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Processing request: %s %s", request.HTTPMethod, request.Path)

	// パスとメソッドに基づいてルーティング
	if request.Resource == "/tags" {
		if request.HTTPMethod == "GET" {
			// タグの一覧を取得
			return listTags(request)
		} else if request.HTTPMethod == "POST" {
			// タグを追加
			return addTags(request)
		}
	} else if request.Resource == "/tags/{imageId}" {
		if request.HTTPMethod == "GET" {
			// 特定の画像のタグを取得
			return getImageTags(request)
		} else if request.HTTPMethod == "DELETE" {
			// タグを削除
			return removeTags(request)
		}
	}

	// 未対応のパス・メソッド
	return events.APIGatewayProxyResponse{
		StatusCode: 404,
		Body:       `{"error":"Not Found"}`,
	}, nil
}

// すべてのタグのリストを取得
func listTags(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// タグの一覧を取得（重複なし）
	result, err := dynamoDBClient.Scan(&dynamodb.ScanInput{
		TableName:            aws.String(cfg.TagsTableName),
		ProjectionExpression: aws.String("TagName"),
	})

	if err != nil {
		log.Printf("Error scanning DynamoDB: %s", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf(`{"error":"Failed to retrieve tags: %s"}`, err),
		}, nil
	}

	// ユニークなタグを抽出
	uniqueTags := make(map[string]bool)
	for _, item := range result.Items {
		if tagName, ok := item["TagName"]; ok {
			uniqueTags[*tagName.S] = true
		}
	}

	// マップからスライスに変換
	tags := make([]string, 0, len(uniqueTags))
	for tag := range uniqueTags {
		tags = append(tags, tag)
	}

	// レスポンスを作成
	response := TagsResponse{
		Tags:  tags,
		Count: len(tags),
	}

	responseJSON, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseJSON),
	}, nil
}

// 特定の画像のタグを取得
func getImageTags(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// パスパラメータから画像IDを取得
	imageID := request.PathParameters["imageId"]

	// 指定された画像IDのタグを検索
	result, err := dynamoDBClient.Query(&dynamodb.QueryInput{
		TableName:              aws.String(cfg.TagsTableName),
		IndexName:              aws.String("ImageIDIndex"),
		KeyConditionExpression: aws.String("ImageID = :imageId"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":imageId": {
				S: aws.String(imageID),
			},
		},
	})

	if err != nil {
		log.Printf("Error querying DynamoDB: %s", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf(`{"error":"Failed to retrieve image tags: %s"}`, err),
		}, nil
	}

	// タグのリストを抽出
	var tags []string
	for _, item := range result.Items {
		if tagName, ok := item["TagName"]; ok {
			tags = append(tags, *tagName.S)
		}
	}

	// レスポンスを作成
	response := TagsResponse{
		Tags:  tags,
		Count: len(tags),
	}

	responseJSON, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseJSON),
	}, nil
}

// タグを追加
func addTags(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// リクエストボディをパース
	var tagRequest AddTagRequest
	err := json.Unmarshal([]byte(request.Body), &tagRequest)
	if err != nil {
		log.Printf("Error parsing request: %s", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       fmt.Sprintf(`{"error":"Invalid request format: %s"}`, err),
		}, nil
	}

	// 画像IDの存在確認
	_, err = dynamoDBClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(cfg.MetaTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"ImageID": {
				S: aws.String(tagRequest.ImageID),
			},
		},
	})

	if err != nil {
		log.Printf("Error verifying image existence: %s", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       `{"error":"Image not found"}`,
		}, nil
	}

	// タグを追加
	addedTags := 0
	for _, tag := range tagRequest.Tags {
		// タグを正規化（トリミングして小文字に）
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "" {
			continue // 空のタグはスキップ
		}

		// タグをDynamoDBに追加
		_, err := dynamoDBClient.PutItem(&dynamodb.PutItemInput{
			TableName: aws.String(cfg.TagsTableName),
			Item: map[string]*dynamodb.AttributeValue{
				"TagName": {
					S: aws.String(tag),
				},
				"ImageID": {
					S: aws.String(tagRequest.ImageID),
				},
			},
			// 同じタグが既に存在する場合は上書きしない
			ConditionExpression: aws.String("attribute_not_exists(TagName) AND attribute_not_exists(ImageID)"),
		})

		// 既存のタグエラーは無視（冪等性を確保）
		if err != nil && !strings.Contains(err.Error(), "ConditionalCheckFailedException") {
			log.Printf("Error adding tag: %s", err)
			continue
		}

		addedTags++
	}

	// レスポンスを作成
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: fmt.Sprintf(`{"message":"Added %d tags to image %s"}`, addedTags, tagRequest.ImageID),
	}, nil
}

// タグを削除
func removeTags(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// パスパラメータから画像IDを取得
	imageID := request.PathParameters["imageId"]

	// リクエストボディをパース
	var tagRequest RemoveTagRequest
	err := json.Unmarshal([]byte(request.Body), &tagRequest)
	if err != nil {
		// ボディが空またはJSONでない場合は、すべてのタグを削除
		if request.Body == "" {
			return removeAllTags(imageID)
		}

		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       fmt.Sprintf(`{"error":"Invalid request format: %s"}`, err),
		}, nil
	}

	// 特定のタグのみを削除
	removedTags := 0
	for _, tag := range tagRequest.Tags {
		// タグを正規化
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "" {
			continue
		}

		// タグをDynamoDBから削除
		_, err := dynamoDBClient.DeleteItem(&dynamodb.DeleteItemInput{
			TableName: aws.String(cfg.TagsTableName),
			Key: map[string]*dynamodb.AttributeValue{
				"TagName": {
					S: aws.String(tag),
				},
				"ImageID": {
					S: aws.String(imageID),
				},
			},
		})

		if err != nil {
			log.Printf("Error removing tag: %s", err)
			continue
		}

		removedTags++
	}

	// レスポンスを作成
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: fmt.Sprintf(`{"message":"Removed %d tags from image %s"}`, removedTags, imageID),
	}, nil
}

// 画像のすべてのタグを削除
func removeAllTags(imageID string) (events.APIGatewayProxyResponse, error) {
	// 画像のすべてのタグを検索
	result, err := dynamoDBClient.Query(&dynamodb.QueryInput{
		TableName:              aws.String(cfg.TagsTableName),
		IndexName:              aws.String("ImageIDIndex"),
		KeyConditionExpression: aws.String("ImageID = :imageId"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":imageId": {
				S: aws.String(imageID),
			},
		},
	})

	if err != nil {
		log.Printf("Error querying tags: %s", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf(`{"error":"Failed to retrieve tags: %s"}`, err),
		}, nil
	}

	// すべてのタグを削除
	removedTags := 0
	for _, item := range result.Items {
		tagName := item["TagName"].S

		_, err := dynamoDBClient.DeleteItem(&dynamodb.DeleteItemInput{
			TableName: aws.String(cfg.TagsTableName),
			Key: map[string]*dynamodb.AttributeValue{
				"TagName": {
					S: tagName,
				},
				"ImageID": {
					S: aws.String(imageID),
				},
			},
		})

		if err != nil {
			log.Printf("Error removing tag: %s", err)
			continue
		}

		removedTags++
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: fmt.Sprintf(`{"message":"Removed all %d tags from image %s"}`, removedTags, imageID),
	}, nil
}

func main() {
	lambda.Start(handler)
}
