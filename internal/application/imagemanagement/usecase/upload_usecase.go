package usecase

import (
	"cloudpix/internal/application/imagemanagement/dto"
	"cloudpix/internal/domain/imagemanagement/aggregate"
	"cloudpix/internal/domain/imagemanagement/entity"
	"cloudpix/internal/domain/imagemanagement/event"
	"cloudpix/internal/domain/imagemanagement/repository"
	"cloudpix/internal/domain/imagemanagement/service"
	"cloudpix/internal/domain/imagemanagement/valueobject"
	"cloudpix/internal/domain/shared/event/dispatcher"
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// UploadUsecase は画像アップロードのユースケースを実装します
type UploadUsecase struct {
	imageRepository repository.ImageRepository
	storageService  service.StorageService
	eventDispatcher dispatcher.EventDispatcher
	bucketName      string
}

// NewUploadUsecase は新しいアップロードユースケースを作成します
func NewUploadUsecase(
	imageRepository repository.ImageRepository,
	storageService service.StorageService,
	eventDispatcher dispatcher.EventDispatcher,
	bucketName string,
) *UploadUsecase {
	return &UploadUsecase{
		imageRepository: imageRepository,
		storageService:  storageService,
		eventDispatcher: eventDispatcher,
		bucketName:      bucketName,
	}
}

// ProcessUpload はアップロードリクエストを処理します
func (u *UploadUsecase) ProcessUpload(ctx context.Context, request *dto.UploadRequest) (*dto.UploadResponse, error) {
	// 値オブジェクトの作成と検証
	fileName, err := valueobject.NewFileName(request.FileName)
	if err != nil {
		return nil, err
	}

	contentType, err := valueobject.NewContentType(request.ContentType)
	if err != nil {
		return nil, err
	}

	// ユニークなIDを生成
	imageID := uuid.New().String()
	today := valueobject.Today()

	// オブジェクトキーを生成
	objectKey := fmt.Sprintf("uploads/%s-%s", imageID, fileName.String())

	var downloadURL string
	var uploadURL string
	var imageSize valueobject.ImageSize
	var message string

	// Base64エンコードされたデータがある場合は直接アップロード
	if request.Data != "" {
		// 画像サイズを計算
		data, err := base64.StdEncoding.DecodeString(request.Data)
		if err != nil {
			return nil, err
		}
		imageSize, _ = valueobject.NewImageSize(len(data))

		// S3にアップロード
		downloadURL, err = u.storageService.StoreImage(
			ctx,
			u.bucketName,
			objectKey,
			contentType.String(),
			request.Data,
		)
		if err != nil {
			return nil, err
		}
		message = "Image uploaded successfully"
	} else {
		// プレサインドURLを生成
		imageSize, _ = valueobject.NewImageSize(0) // サイズ不明
		uploadURL, downloadURL, err = u.storageService.GenerateImageURL(
			ctx,
			u.bucketName,
			objectKey,
			contentType.String(),
			15*time.Minute,
		)
		if err != nil {
			return nil, err
		}
		message = "Use the uploadUrl to upload your image"
	}

	// エンティティを作成
	image := entity.NewImage(
		imageID,
		fileName,
		contentType,
		imageSize,
		today,
		objectKey,
		downloadURL,
	)

	// 集約を作成
	imageAggregate := aggregate.NewImageAggregate(image)

	// リポジトリに保存
	err = u.imageRepository.Save(ctx, imageAggregate)
	if err != nil {
		return nil, err
	}

	// イベントを発行
	uploadEvent := event.NewImageUploadedEvent(image, u.bucketName)
	err = u.eventDispatcher.Dispatch(ctx, uploadEvent)
	if err != nil {
		// イベント発行に失敗してもアップロードは成功とみなす
		// ログだけ記録する
	}

	// レスポンスを作成
	response := &dto.UploadResponse{
		ImageID:     imageID,
		UploadURL:   uploadURL,
		DownloadURL: downloadURL,
		Message:     message,
	}

	return response, nil
}
