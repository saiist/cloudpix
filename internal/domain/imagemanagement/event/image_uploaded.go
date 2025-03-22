package event

import (
	"cloudpix/internal/domain/imagemanagement/entity"
	"time"
)

// ImageUploadedEvent は画像がアップロードされたときに発行されるイベント
type ImageUploadedEvent struct {
	ImageID     string
	FileName    string
	ContentType string
	Size        int
	UploadDate  string
	S3ObjectKey string
	Bucket      string
	UploadTime  time.Time
}

// NewImageUploadedEvent は画像エンティティから新しいイベントを作成します
func NewImageUploadedEvent(image *entity.Image, bucket string) *ImageUploadedEvent {
	return &ImageUploadedEvent{
		ImageID:     image.ID,
		FileName:    image.FileName.String(),
		ContentType: image.ContentType.String(),
		Size:        image.Size.Value(),
		UploadDate:  image.UploadDate.String(),
		S3ObjectKey: image.S3ObjectKey,
		Bucket:      bucket,
		UploadTime:  time.Now(),
	}
}

// EventType はイベントタイプを返します
func (e *ImageUploadedEvent) EventType() string {
	return "image.uploaded"
}

// OccurredAt はイベント発生時刻を返します
func (e *ImageUploadedEvent) OccurredAt() time.Time {
	return e.UploadTime
}

// AggregateID は集約IDを返します
func (e *ImageUploadedEvent) AggregateID() string {
	return e.ImageID
}
