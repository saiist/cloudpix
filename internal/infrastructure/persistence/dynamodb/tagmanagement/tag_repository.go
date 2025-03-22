package tagmanagement

import (
	"cloudpix/internal/domain/tagmanagement/entity"
	"cloudpix/internal/domain/tagmanagement/repository"
	"cloudpix/internal/domain/tagmanagement/valueobject"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

// DynamoDBTagItem はDynamoDBのタグアイテム表現
type DynamoDBTagItem struct {
	TagName   string `json:"TagName"` // PK
	ImageID   string `json:"ImageID"` // SK
	CreatedAt string `json:"CreatedAt"`
}

// DynamoDBTaggedImageItem はDynamoDBのタグ付き画像アイテム表現
type DynamoDBTaggedImageItem struct {
	ImageID   string   `json:"ImageID"`
	Tags      []string `json:"Tags"`
	CreatedAt string   `json:"CreatedAt"`
	UpdatedAt string   `json:"UpdatedAt"`
}

// DynamoDBTagRepository はDynamoDBを使用したタグリポジトリの実装
type DynamoDBTagRepository struct {
	client            *dynamodb.DynamoDB
	tagsTableName     string
	metadataTableName string
}

// NewDynamoDBTagRepository は新しいDynamoDBタグリポジトリを作成します
func NewDynamoDBTagRepository(
	client *dynamodb.DynamoDB,
	tagsTableName string,
	metadataTableName string,
) repository.TagRepository {
	return &DynamoDBTagRepository{
		client:            client,
		tagsTableName:     tagsTableName,
		metadataTableName: metadataTableName,
	}
}

// FindAllTags は全てのユニークなタグを取得します
func (r *DynamoDBTagRepository) FindAllTags(ctx context.Context) ([]valueobject.Tag, error) {
	// タグテーブルからユニークなタグ名を取得
	proj := expression.NamesList(expression.Name("TagName"))
	expr, err := expression.NewBuilder().WithProjection(proj).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build expression: %w", err)
	}

	input := &dynamodb.ScanInput{
		TableName:                aws.String(r.tagsTableName),
		ExpressionAttributeNames: expr.Names(),
		ProjectionExpression:     expr.Projection(),
	}

	result, err := r.client.ScanWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan tags table: %w", err)
	}

	// 重複を除去するために一時的にマップを使用
	uniqueTags := make(map[string]bool)
	for _, item := range result.Items {
		if tagName, ok := item["TagName"]; ok && tagName.S != nil {
			uniqueTags[*tagName.S] = true
		}
	}

	// ユニークなタグ名を値オブジェクトに変換
	tags := make([]valueobject.Tag, 0, len(uniqueTags))
	for tagName := range uniqueTags {
		tag, err := valueobject.NewTag(tagName)
		if err != nil {
			continue // 無効なタグはスキップ
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// FindTaggedImage は指定された画像IDのタグ情報を取得します
func (r *DynamoDBTagRepository) FindTaggedImage(ctx context.Context, imageID string) (*entity.TaggedImage, error) {
	// DynamoDBから特定の画像IDに関連するタグを取得
	keyCondition := expression.Key("ImageID").Equal(expression.Value(imageID))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build expression: %w", err)
	}

	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(r.tagsTableName),
		IndexName:                 aws.String("ImageIDIndex"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	result, err := r.client.QueryWithContext(ctx, queryInput)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}

	// 結果がない場合はnilを返す
	if len(result.Items) == 0 {
		return nil, nil
	}

	// タグ付き画像エンティティを作成
	taggedImage := entity.NewTaggedImage(imageID)

	// タグを追加
	for _, item := range result.Items {
		if tagNameAttr, ok := item["TagName"]; ok && tagNameAttr.S != nil {
			tag, err := valueobject.NewTag(*tagNameAttr.S)
			if err != nil {
				continue // 無効なタグはスキップ
			}
			taggedImage.AddTag(tag)
		}
	}

	return taggedImage, nil
}

// FindImagesByTag は指定されたタグを持つ画像IDのリストを取得します
func (r *DynamoDBTagRepository) FindImagesByTag(ctx context.Context, tag valueobject.Tag) ([]string, error) {
	// タグ名で検索するためのキー条件を作成
	keyCondition := expression.Key("TagName").Equal(expression.Value(tag.Name()))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build expression: %w", err)
	}

	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(r.tagsTableName),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	result, err := r.client.QueryWithContext(ctx, queryInput)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags by tag name: %w", err)
	}

	// 画像IDのリストを抽出
	imageIDs := make([]string, 0, len(result.Items))
	for _, item := range result.Items {
		if imageIDAttr, ok := item["ImageID"]; ok && imageIDAttr.S != nil {
			imageIDs = append(imageIDs, *imageIDAttr.S)
		}
	}

	return imageIDs, nil
}

// Save はタグ付き画像情報を保存します
func (r *DynamoDBTagRepository) Save(ctx context.Context, taggedImage *entity.TaggedImage) error {
	// 既存のタグを一度取得
	existingTagged, err := r.FindTaggedImage(ctx, taggedImage.ImageID)
	if err != nil {
		return fmt.Errorf("failed to get existing tags: %w", err)
	}

	// 既存のタグセット
	existingTagMap := make(map[string]bool)
	if existingTagged != nil {
		for _, tag := range existingTagged.Tags {
			existingTagMap[tag.Name()] = true
		}
	}

	// 新しいタグセット
	newTagMap := make(map[string]bool)
	for _, tag := range taggedImage.Tags {
		newTagMap[tag.Name()] = true
	}

	// DynamoDBのトランザクションアイテムを準備
	var transactItems []*dynamodb.TransactWriteItem

	// 1. 削除されたタグを削除
	for tagName := range existingTagMap {
		if !newTagMap[tagName] {
			// このタグは削除する必要がある
			deleteItem := &dynamodb.TransactWriteItem{
				Delete: &dynamodb.Delete{
					TableName: aws.String(r.tagsTableName),
					Key: map[string]*dynamodb.AttributeValue{
						"TagName": {S: aws.String(tagName)},
						"ImageID": {S: aws.String(taggedImage.ImageID)},
					},
				},
			}
			transactItems = append(transactItems, deleteItem)
		}
	}

	// 2. 新しいタグを追加
	for tagName := range newTagMap {
		if !existingTagMap[tagName] {
			// このタグは追加する必要がある
			item := DynamoDBTagItem{
				TagName:   tagName,
				ImageID:   taggedImage.ImageID,
				CreatedAt: time.Now().Format(time.RFC3339),
			}

			av, err := dynamodbattribute.MarshalMap(item)
			if err != nil {
				return fmt.Errorf("failed to marshal tag item: %w", err)
			}

			putItem := &dynamodb.TransactWriteItem{
				Put: &dynamodb.Put{
					TableName: aws.String(r.tagsTableName),
					Item:      av,
				},
			}
			transactItems = append(transactItems, putItem)
		}
	}

	// トランザクションアイテムがない場合は処理しない
	if len(transactItems) == 0 {
		return nil
	}

	// トランザクションを実行
	_, err = r.client.TransactWriteItemsWithContext(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: transactItems,
	})
	if err != nil {
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	return nil
}

// Delete はタグ付き画像情報を削除します
func (r *DynamoDBTagRepository) Delete(ctx context.Context, imageID string) error {
	// 特定の画像IDに関連するタグを取得
	existingTagged, err := r.FindTaggedImage(ctx, imageID)
	if err != nil {
		return fmt.Errorf("failed to get existing tags: %w", err)
	}

	// タグがない場合は何もしない
	if existingTagged == nil || len(existingTagged.Tags) == 0 {
		return nil
	}

	// DynamoDBのバッチ書き込みアイテムを準備
	var writeRequests []*dynamodb.WriteRequest

	// 各タグの削除リクエストを作成
	for _, tag := range existingTagged.Tags {
		deleteRequest := &dynamodb.WriteRequest{
			DeleteRequest: &dynamodb.DeleteRequest{
				Key: map[string]*dynamodb.AttributeValue{
					"TagName": {S: aws.String(tag.Name())},
					"ImageID": {S: aws.String(imageID)},
				},
			},
		}
		writeRequests = append(writeRequests, deleteRequest)
	}

	// バッチ削除実行（最大25アイテムずつ）
	const batchSize = 25
	for i := 0; i < len(writeRequests); i += batchSize {
		end := i + batchSize
		if end > len(writeRequests) {
			end = len(writeRequests)
		}

		batch := writeRequests[i:end]
		_, err := r.client.BatchWriteItemWithContext(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]*dynamodb.WriteRequest{
				r.tagsTableName: batch,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to batch delete tags: %w", err)
		}
	}

	return nil
}

// ImageExists は画像が存在するか確認します
func (r *DynamoDBTagRepository) ImageExists(ctx context.Context, imageID string) (bool, error) {
	// メタデータテーブルで画像IDをチェック
	result, err := r.client.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.metadataTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"ImageID": {S: aws.String(imageID)},
		},
		ProjectionExpression: aws.String("ImageID"), // 必要な属性のみを取得
	})
	if err != nil {
		return false, fmt.Errorf("failed to check image existence: %w", err)
	}

	// アイテムが存在する場合はtrueを返す
	return len(result.Item) > 0, nil
}
