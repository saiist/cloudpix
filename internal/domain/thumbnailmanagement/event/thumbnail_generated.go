package event

import (
	"cloudpix/internal/domain/thumbnailmanagement/entity"
	"time"
)

// ThumbnailGeneratedEvent はサムネイルが生成されたイベント
type ThumbnailGeneratedEvent struct {
	ImageID       string
	ThumbnailURL  string
	Width         int
	Height        int
	ContentType   string
	GeneratedTime time.Time
}

// NewThumbnailGeneratedEvent はサムネイルエンティティから新しいイベントを作成します
func NewThumbnailGeneratedEvent(thumbnail *entity.Thumbnail) *ThumbnailGeneratedEvent {
	return &ThumbnailGeneratedEvent{
		ImageID:       thumbnail.ImageID,
		ThumbnailURL:  thumbnail.ThumbnailURL,
		Width:         thumbnail.GetWidth(),
		Height:        thumbnail.GetHeight(),
		ContentType:   thumbnail.ContentType,
		GeneratedTime: time.Now(),
	}
}

// EventType はイベントタイプを返します
func (e *ThumbnailGeneratedEvent) EventType() string {
	return "thumbnail.generated"
}

// OccurredAt はイベント発生時刻を返します
func (e *ThumbnailGeneratedEvent) OccurredAt() time.Time {
	return e.GeneratedTime
}

// AggregateID は集約IDを返します
func (e *ThumbnailGeneratedEvent) AggregateID() string {
	return e.ImageID
}
