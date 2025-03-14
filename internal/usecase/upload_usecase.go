package usecase

import (
	"cloudpix/internal/domain/model"
	"cloudpix/internal/domain/repository"
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type UploadUsecase struct {
	storageRepo  repository.StorageRepository
	metadataRepo repository.UploadMetadataRepository
	bucketName   string
}

func NewUploadUsecase(
	storageRepo repository.StorageRepository,
	metadataRepo repository.UploadMetadataRepository,
	bucketName string,
) *UploadUsecase {
	return &UploadUsecase{
		storageRepo:  storageRepo,
		metadataRepo: metadataRepo,
		bucketName:   bucketName,
	}
}

// ProcessUpload はアップロードリクエストを処理します
func (u *UploadUsecase) ProcessUpload(ctx context.Context, request *model.UploadRequest) (*model.UploadResponse, error) {
	// ユニークなIDを生成
	imageID := uuid.New().String()
	now := time.Now()
	todayDate := now.Format("2006-01-02") // YYYY-MM-DD形式

	// オブジェクトキーを生成
	objectKey := fmt.Sprintf("uploads/%s-%s", imageID, request.FileName)

	var downloadURL string
	var uploadURL string
	var imageSize int
	var message string

	// Base64エンコードされたデータがある場合は直接アップロード
	if request.Data != "" {
		var err error
		downloadURL, err = u.storageRepo.UploadImage(
			ctx,
			u.bucketName,
			objectKey,
			request.ContentType,
			request.Data,
		)
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
		uploadURL, downloadURL, err = u.storageRepo.GeneratePresignedURL(
			ctx,
			u.bucketName,
			objectKey,
			request.ContentType,
			15*time.Minute,
		)
		if err != nil {
			return nil, err
		}
		message = "Use the uploadUrl to upload your image"
		imageSize = 0 // プレサインドURLの場合、サイズは不明
	}

	// メタデータを作成して保存
	metadata := &model.UploadMetadata{
		ImageID:      imageID,
		FileName:     request.FileName,
		ContentType:  request.ContentType,
		Size:         imageSize,
		UploadDate:   todayDate,
		CreatedAt:    now,
		S3ObjectKey:  objectKey,
		S3BucketName: u.bucketName,
		DownloadURL:  downloadURL,
	}

	// メタデータを保存
	err := u.metadataRepo.SaveMetadata(ctx, metadata)
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
