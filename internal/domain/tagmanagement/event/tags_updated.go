package event

import (
	"cloudpix/internal/domain/tagmanagement/entity"
	"time"
)

// TagsUpdatedEvent はタグが更新されたイベント
type TagsUpdatedEvent struct {
	ImageID     string
	Tags        []string
	UpdatedTime time.Time
}

// NewTagsUpdatedEvent はタグ更新イベントを作成します
func NewTagsUpdatedEvent(taggedImage *entity.TaggedImage) *TagsUpdatedEvent {
	return &TagsUpdatedEvent{
		ImageID:     taggedImage.ImageID,
		Tags:        taggedImage.GetTagNames(),
		UpdatedTime: time.Now(),
	}
}

// EventType はイベントタイプを返します
func (e *TagsUpdatedEvent) EventType() string {
	return "tags.updated"
}

// OccurredAt はイベント発生時刻を返します
func (e *TagsUpdatedEvent) OccurredAt() time.Time {
	return e.UpdatedTime
}

// AggregateID は集約IDを返します
func (e *TagsUpdatedEvent) AggregateID() string {
	return e.ImageID
}
