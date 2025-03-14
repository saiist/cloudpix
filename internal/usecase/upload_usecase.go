package usecase

import (
	"cloudpix/internal/domain/model"
	"cloudpix/internal/domain/repository"
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type UploadUsecase struct {
	uploadRepo repository.UploadRepository
}

func NewUploadUsecase(uploadRepo repository.UploadRepository) *UploadUsecase {
	return &UploadUsecase{
		uploadRepo: uploadRepo,
	}
}

// ProcessUpload はアップロードリクエストを処理します
func (u *UploadUsecase) ProcessUpload(ctx context.Context, request *model.UploadRequest) (*model.UploadResponse, error) {
	// ユニークなIDを生成
	imageID := uuid.New().String()
	now := time.Now()
	todayDate := now.Format("2006-01-02") // YYYY-MM-DD形式

	var downloadURL string
	var uploadURL string
	var imageSize int
	var message string

	// Base64エンコードされたデータがある場合は直接アップロード
	if request.Data != "" {
		var err error
		downloadURL, err = u.uploadRepo.UploadImage(ctx, imageID, request)
		if err != nil {
			return nil, err
		}
		message = "Image uploaded successfully"

		// Base64デコードしてサイズを計算
		rawData, _ := base64Decode(request.Data)
		imageSize = len(rawData)
	} else {
		// プレサインドURLを生成
		var err error
		uploadURL, downloadURL, err = u.uploadRepo.GeneratePresignedURL(ctx, imageID, request, 15*time.Minute)
		if err != nil {
			return nil, err
		}
		message = "Use the uploadUrl to upload your image"
		imageSize = 0 // プレサインドURLの場合、サイズは不明
	}

	// オブジェクトキーを生成
	objectKey := getObjectKey(imageID, request.FileName)

	// メタデータを作成して保存
	metadata := &model.UploadMetadata{
		ImageID:      imageID,
		FileName:     request.FileName,
		ContentType:  request.ContentType,
		Size:         imageSize,
		UploadDate:   todayDate,
		CreatedAt:    now,
		S3ObjectKey:  objectKey,
		S3BucketName: getBucketNameFromURL(downloadURL),
		DownloadURL:  downloadURL,
	}

	// メタデータを保存
	err := u.uploadRepo.SaveMetadata(ctx, metadata)
	if err != nil {
		return nil, err
	}

	// レスポンスを作成
	response := &model.UploadResponse{
		ImageID:     imageID,
		UploadURL:   uploadURL,
		DownloadURL: downloadURL,
		Message:     message,
	}

	return response, nil
}

// base64Decode はBase64エンコードされた文字列をデコードします
func base64Decode(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}

// getObjectKey はS3のオブジェクトキーを生成します
func getObjectKey(imageID, fileName string) string {
	return fmt.Sprintf("uploads/%s-%s", imageID, fileName)
}

// getBucketNameFromURL はURLからバケット名を抽出します
func getBucketNameFromURL(url string) string {
	// URLの形式: https://bucket-name.s3.region.amazonaws.com/key
	// 簡易的なパース処理
	parts := strings.Split(url, ".")
	if len(parts) >= 2 {
		// https://bucket-name を取得して先頭の https:// を削除
		bucketWithProtocol := parts[0]
		return strings.Replace(bucketWithProtocol, "https://", "", 1)
	}
	return ""
}
