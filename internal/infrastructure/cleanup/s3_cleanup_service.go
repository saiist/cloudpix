package cleanup

import (
	"cloudpix/internal/domain/imagemanagement/service"
	"cloudpix/internal/logging"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3CleanupService はS3およびDynamoDBのクリーンアップを行う実装
type S3CleanupService struct {
	s3Client      *s3.S3
	dynamoClient  *dynamodb.DynamoDB
	bucketName    string
	metadataTable string
	tagsTableName string
}

// NewS3CleanupService は新しいS3クリーンアップサービスを作成
func NewS3CleanupService(
	s3Client *s3.S3,
	dynamoClient *dynamodb.DynamoDB,
	bucketName string,
	metadataTable string,
	tagsTableName string,
) service.CleanupService {
	return &S3CleanupService{
		s3Client:      s3Client,
		dynamoClient:  dynamoClient,
		bucketName:    bucketName,
		metadataTable: metadataTable,
		tagsTableName: tagsTableName,
	}
}

// getImageMetadata は画像メタデータをDynamoDBから取得する共通メソッド
func (s *S3CleanupService) getImageMetadata(ctx context.Context, imageID string) (map[string]*dynamodb.AttributeValue, error) {
	result, err := s.dynamoClient.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.metadataTable),
		Key: map[string]*dynamodb.AttributeValue{
			"ImageID": {
				S: aws.String(imageID),
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get image metadata: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("image not found: %s", imageID)
	}

	return result.Item, nil
}

// getObjectKey はメタデータからS3オブジェクトキーを抽出する共通メソッド
func (s *S3CleanupService) getObjectKey(metadata map[string]*dynamodb.AttributeValue, imageID string) (string, error) {
	if val, ok := metadata["S3ObjectKey"]; ok && val.S != nil {
		return *val.S, nil
	}
	return "", fmt.Errorf("S3ObjectKey not found for image: %s", imageID)
}

// ArchiveImage は画像をアーカイブバケットに移動
func (s *S3CleanupService) ArchiveImage(ctx context.Context, imageID string) error {
	// DynamoDBから画像メタデータを取得
	metadata, err := s.getImageMetadata(ctx, imageID)
	if err != nil {
		return err
	}

	// オブジェクトキーを取得
	s3ObjectKey, err := s.getObjectKey(metadata, imageID)
	if err != nil {
		return err
	}

	// アーカイブオブジェクトキーを作成
	archiveKey := fmt.Sprintf("archive/%s", s3ObjectKey)

	// オブジェクトをコピー
	_, err = s.s3Client.CopyObjectWithContext(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucketName),
		CopySource: aws.String(fmt.Sprintf("%s/%s", s.bucketName, s3ObjectKey)),
		Key:        aws.String(archiveKey),
	})

	if err != nil {
		return fmt.Errorf("failed to copy object to archive: %w", err)
	}

	// 元のオブジェクトを削除
	_, err = s.s3Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(s3ObjectKey),
	})

	if err != nil {
		return fmt.Errorf("failed to delete original object: %w", err)
	}

	// メタデータを更新
	return s.updateImageStatus(ctx, imageID, archiveKey, "ARCHIVED")
}

// updateImageStatus は画像のステータスを更新する共通メソッド
func (s *S3CleanupService) updateImageStatus(ctx context.Context, imageID string, newKey string, status string) error {
	_, err := s.dynamoClient.UpdateItemWithContext(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.metadataTable),
		Key: map[string]*dynamodb.AttributeValue{
			"ImageID": {
				S: aws.String(imageID),
			},
		},
		UpdateExpression: aws.String("SET S3ObjectKey = :newKey, ImageStatus = :status"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":newKey": {
				S: aws.String(newKey),
			},
			":status": {
				S: aws.String(status),
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	return nil
}

// DeleteImage は画像を完全に削除
func (s *S3CleanupService) DeleteImage(ctx context.Context, imageID string) error {
	logger := logging.FromContext(ctx)

	// DynamoDBから画像メタデータを取得
	metadata, err := s.getImageMetadata(ctx, imageID)
	if err != nil {
		return err
	}

	// オブジェクトキーを取得
	s3ObjectKey, err := s.getObjectKey(metadata, imageID)
	if err != nil {
		return err
	}

	// サムネイルが存在するかチェック
	var hasThumbnail bool
	if val, ok := metadata["HasThumbnail"]; ok && val.BOOL != nil {
		hasThumbnail = *val.BOOL
	}

	// 元の画像オブジェクトを削除
	if err := s.deleteS3Object(ctx, s3ObjectKey); err != nil {
		return err
	}

	// サムネイルが存在する場合は削除
	if hasThumbnail {
		if err := s.deleteThumbnail(ctx, logger, imageID, s3ObjectKey); err != nil {
			// サムネイル削除に失敗してもメインの処理は続行
			logger.Warn(fmt.Sprintf("Failed to delete thumbnail: %v", err), map[string]interface{}{
				"imageid": imageID,
			})
		}
	}

	// タグ情報を削除
	if err := s.deleteImageTags(ctx, imageID); err != nil {
		return err
	}

	// メタデータを削除
	return s.deleteImageMetadata(ctx, imageID)
}

// deleteS3Object はS3オブジェクトを削除する共通メソッド
func (s *S3CleanupService) deleteS3Object(ctx context.Context, objectKey string) error {
	_, err := s.s3Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		return fmt.Errorf("failed to delete S3 object: %w", err)
	}

	return nil
}

// deleteThumbnail はサムネイル画像を削除する共通メソッド
func (s *S3CleanupService) deleteThumbnail(ctx context.Context, logger logging.Logger, imageID string, originalKey string) error {
	// サムネイルのキーを生成 (命名規則に基づいて)
	thumbnailKey := strings.Replace(originalKey, "uploads/", "thumbnails/", 1)

	_, err := s.s3Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(thumbnailKey),
	})

	if err != nil {
		return fmt.Errorf("failed to delete thumbnail: %w", err)
	}

	return nil
}

// deleteImageTags は画像に関連するタグを削除する共通メソッド
func (s *S3CleanupService) deleteImageTags(ctx context.Context, imageID string) error {
	// タグ情報を検索
	queryResult, err := s.dynamoClient.QueryWithContext(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.tagsTableName),
		IndexName:              aws.String("ImageIDIndex"),
		KeyConditionExpression: aws.String("ImageID = :imageId"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":imageId": {
				S: aws.String(imageID),
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to query tags for image: %w", err)
	}

	// タグのバッチ削除（存在する場合）
	if len(queryResult.Items) > 0 {
		// バッチ書き込みリクエストを作成
		var writeRequests []*dynamodb.WriteRequest

		for _, item := range queryResult.Items {
			// TagNameを取得
			var tagName string
			if val, ok := item["TagName"]; ok && val.S != nil {
				tagName = *val.S
			} else {
				continue
			}

			// 削除リクエストを追加
			writeRequests = append(writeRequests, &dynamodb.WriteRequest{
				DeleteRequest: &dynamodb.DeleteRequest{
					Key: map[string]*dynamodb.AttributeValue{
						"TagName": {
							S: aws.String(tagName),
						},
						"ImageID": {
							S: aws.String(imageID),
						},
					},
				},
			})
		}

		// バッチ処理（25件ずつ）
		const batchSize = 25
		for i := 0; i < len(writeRequests); i += batchSize {
			end := min(i+batchSize, len(writeRequests))

			batch := writeRequests[i:end]
			_, err := s.dynamoClient.BatchWriteItemWithContext(ctx, &dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]*dynamodb.WriteRequest{
					s.tagsTableName: batch,
				},
			})

			if err != nil {
				return fmt.Errorf("failed to delete tags: %w", err)
			}
		}
	}

	return nil
}

// deleteImageMetadata は画像メタデータを削除する共通メソッド
func (s *S3CleanupService) deleteImageMetadata(ctx context.Context, imageID string) error {
	_, err := s.dynamoClient.DeleteItemWithContext(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(s.metadataTable),
		Key: map[string]*dynamodb.AttributeValue{
			"ImageID": {
				S: aws.String(imageID),
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to delete image metadata: %w", err)
	}

	return nil
}

// CleanupOldImages は古い画像を一括処理
func (s *S3CleanupService) CleanupOldImages(ctx context.Context, retentionDays int) error {
	logger := logging.FromContext(ctx)

	// 保持期間から日付の閾値を計算
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	cutoffDateStr := cutoffDate.Format("2006-01-02")

	// クエリの準備
	queryInput, err := s.prepareOldImagesQuery(cutoffDateStr)
	if err != nil {
		return err
	}

	// 処理結果の統計
	stats := processingStats{logger: logger}

	// ページング処理を行いながら画像をアーカイブ
	err = s.processImagesWithPaging(ctx, queryInput, func(item map[string]*dynamodb.AttributeValue) error {
		return s.processOldImage(ctx, item, &stats)
	})

	if err != nil {
		return err
	}

	// 結果をログに記録
	logger.Info("Completed cleaning up old images", map[string]interface{}{
		"processedCount": stats.processedCount,
		"errorCount":     stats.errorCount,
		"skippedCount":   stats.skippedCount,
	})

	// エラーがあれば報告
	if stats.errorCount > 0 {
		return fmt.Errorf("completed with %d errors, processed %d images",
			stats.errorCount, stats.processedCount)
	}

	return nil
}

// processingStats は処理統計を追跡するための構造体
type processingStats struct {
	processedCount int
	errorCount     int
	skippedCount   int
	logger         logging.Logger
}

// prepareOldImagesQuery はクエリ入力を準備する共通メソッド
func (s *S3CleanupService) prepareOldImagesQuery(cutoffDateStr string) (*dynamodb.ScanInput, error) {
	return &dynamodb.ScanInput{
		TableName: aws.String(s.metadataTable),
		// UploadDateが存在し、かつ空でなく、基準日以前のアイテムのみをスキャン
		FilterExpression: aws.String("attribute_exists(UploadDate) AND attribute_type(UploadDate, :s) AND UploadDate <= :date"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":date": {
				S: aws.String(cutoffDateStr),
			},
			":s": {
				S: aws.String("S"), // 文字列型のチェック
			},
		},
	}, nil
}

// processImagesWithPaging はページングを行いながら画像を処理する共通メソッド
func (s *S3CleanupService) processImagesWithPaging(
	ctx context.Context,
	scanInput *dynamodb.ScanInput,
	processor func(map[string]*dynamodb.AttributeValue) error,
) error {
	var lastEvaluatedKey map[string]*dynamodb.AttributeValue

	for {
		// 前回のスキャンで最後に評価されたキーがあれば、そこから再開
		if lastEvaluatedKey != nil {
			scanInput.ExclusiveStartKey = lastEvaluatedKey
		}

		// スキャンを実行
		result, err := s.dynamoClient.ScanWithContext(ctx, scanInput)
		if err != nil {
			return fmt.Errorf("failed to scan images: %w", err)
		}

		// 結果が空の場合は終了
		if len(result.Items) == 0 {
			break
		}

		// 各画像を処理
		for _, item := range result.Items {
			if err := processor(item); err != nil {
				// 個別の処理エラーはスキップして続行
				continue
			}
		}

		// 次のページがない場合は終了
		lastEvaluatedKey = result.LastEvaluatedKey
		if lastEvaluatedKey == nil {
			break
		}
	}

	return nil
}

// processOldImage は古い画像を処理する共通メソッド
func (s *S3CleanupService) processOldImage(
	ctx context.Context,
	item map[string]*dynamodb.AttributeValue,
	stats *processingStats,
) error {
	// ImageIDを取得
	var imageID string
	if val, ok := item["ImageID"]; ok && val.S != nil {
		imageID = *val.S
	} else {
		stats.errorCount++
		return fmt.Errorf("invalid image item: ImageID not found")
	}

	// 画像がすでにアーカイブ済みかチェック
	var status string
	if val, ok := item["ImageStatus"]; ok && val.S != nil {
		status = *val.S
	}

	// アーカイブ済みの場合はスキップ
	if status == "ARCHIVED" {
		stats.skippedCount++
		return nil
	}

	// アーカイブ処理を実行
	err := s.ArchiveImage(ctx, imageID)
	if err != nil {
		stats.logger.Error(err, "Error archiving image", map[string]interface{}{
			"imageid":    imageID,
			"bucketName": s.bucketName,
		})
		stats.errorCount++
		return err
	}

	stats.processedCount++
	return nil
}
