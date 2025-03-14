package main

import (
	"bytes"
	"cloudpix/config"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"path/filepath"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/disintegration/imaging"
)

var (
	cfg            = config.NewConfig()
	thumbnailSize  = 200              // サムネイルのサイズ（ピクセル）
	s3Client       *s3.S3             // S3クライアント
	dynamoDBClient *dynamodb.DynamoDB // DynamoDBクライアント
)

// サムネイル情報構造体
type ThumbnailInfo struct {
	ImageID      string `json:"ImageID"`      // 元画像のID
	ThumbnailKey string `json:"thumbnailKey"` // サムネイルのS3キー
	ThumbnailURL string `json:"thumbnailUrl"` // サムネイルのURL
	Width        int    `json:"width"`        // サムネイルの幅
	Height       int    `json:"height"`       // サムネイルの高さ
	OriginalKey  string `json:"originalKey"`  // 元画像のS3キー
}

func init() {
	// AWS セッションの初期化
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.AWSRegion),
	})
	if err != nil {
		log.Printf("Error creating AWS session: %s", err)
	}

	// クライアントの初期化
	s3Client = s3.New(sess)
	dynamoDBClient = dynamodb.New(sess)

	log.Printf("Thumbnail Lambda initialized with bucket: %s, thumbnail size: %d", cfg.S3BucketName, thumbnailSize)
}

func handler(ctx context.Context, s3Event events.S3Event) error {
	// イベントの処理
	for _, record := range s3Event.Records {
		// S3バケットとオブジェクトキーを取得
		bucket := record.S3.Bucket.Name
		key := record.S3.Object.Key

		// uploadsディレクトリの画像のみ処理
		if !strings.HasPrefix(key, "uploads/") {
			log.Printf("Skipping non-upload file: %s", key)
			continue
		}

		log.Printf("Processing image from bucket: %s, key: %s", bucket, key)

		// 画像を取得
		resp, err := s3Client.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			log.Printf("Error getting S3 object: %s", err)
			return err
		}
		defer resp.Body.Close()

		// 画像をデコード
		var img image.Image
		contentType := *resp.ContentType

		if strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg") {
			img, err = jpeg.Decode(resp.Body)
		} else if strings.Contains(contentType, "png") {
			img, err = png.Decode(resp.Body)
		} else {
			log.Printf("Unsupported image format: %s", contentType)
			continue
		}

		if err != nil {
			log.Printf("Error decoding image: %s", err)
			continue
		}

		// サムネイルを生成
		thumbnail := imaging.Resize(img, thumbnailSize, 0, imaging.Lanczos)

		// サムネイル用のバッファを作成
		var buf bytes.Buffer
		var encodeErr error

		if strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg") {
			encodeErr = jpeg.Encode(&buf, thumbnail, nil)
		} else if strings.Contains(contentType, "png") {
			encodeErr = png.Encode(&buf, thumbnail)
		}

		if encodeErr != nil {
			log.Printf("Error encoding thumbnail: %s", encodeErr)
			continue
		}

		// サムネイルのS3キーを生成
		filename := filepath.Base(key)
		thumbnailKey := fmt.Sprintf("thumbnails/%s", filename)

		// サムネイルをS3にアップロード
		_, uploadErr := s3Client.PutObject(&s3.PutObjectInput{
			Bucket:      aws.String(bucket),
			Key:         aws.String(thumbnailKey),
			Body:        bytes.NewReader(buf.Bytes()),
			ContentType: resp.ContentType,
		})

		if uploadErr != nil {
			log.Printf("Error uploading thumbnail: %s", uploadErr)
			continue
		}

		log.Printf("Successfully created thumbnail: %s", thumbnailKey)

		// 画像IDの抽出（ファイル名から）
		parts := strings.Split(filename, "-")
		if len(parts) < 2 {
			log.Printf("Could not extract image ID from filename: %s", filename)
			continue
		}
		imageID := parts[0]

		// サムネイル情報を作成
		thumbnailInfo := ThumbnailInfo{
			ImageID:      imageID,
			ThumbnailKey: thumbnailKey,
			ThumbnailURL: fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, cfg.AWSRegion, thumbnailKey),
			Width:        thumbnailSize,
			Height:       thumbnail.Bounds().Dy(),
			OriginalKey:  key,
		}

		// DynamoDBに既存のアイテムを更新
		_, updateErr := dynamoDBClient.UpdateItem(&dynamodb.UpdateItemInput{
			TableName: aws.String(cfg.DynamoDBTableName),
			Key: map[string]*dynamodb.AttributeValue{
				"ImageID": {
					S: aws.String(thumbnailInfo.ImageID),
				},
			},
			UpdateExpression: aws.String("SET thumbnailKey = :tk, thumbnailUrl = :tu, thumbnailWidth = :tw, thumbnailHeight = :th"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":tk": {
					S: aws.String(thumbnailInfo.ThumbnailKey),
				},
				":tu": {
					S: aws.String(thumbnailInfo.ThumbnailURL),
				},
				":tw": {
					N: aws.String(fmt.Sprintf("%d", thumbnailInfo.Width)),
				},
				":th": {
					N: aws.String(fmt.Sprintf("%d", thumbnailInfo.Height)),
				},
			},
		})

		if updateErr != nil {
			log.Printf("Error updating DynamoDB item: %s", updateErr)
			continue
		}

		log.Printf("Successfully updated metadata for image: %s", imageID)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
