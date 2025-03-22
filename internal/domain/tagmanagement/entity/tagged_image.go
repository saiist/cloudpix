package entity

import (
	"cloudpix/internal/domain/tagmanagement/valueobject"
	"time"
)

// TaggedImage エンティティはタグ付けされた画像を表します
type TaggedImage struct {
	ImageID   string
	Tags      []valueobject.Tag
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewTaggedImage は新しいタグ付き画像エンティティを作成します
func NewTaggedImage(imageID string) *TaggedImage {
	now := time.Now()
	return &TaggedImage{
		ImageID:   imageID,
		Tags:      make([]valueobject.Tag, 0),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddTag は画像にタグを追加します（重複チェックつき）
func (t *TaggedImage) AddTag(tag valueobject.Tag) bool {
	// 重複をチェック
	for _, existingTag := range t.Tags {
		if existingTag.Equals(tag) {
			return false // 既に存在するため追加なし
		}
	}

	// タグを追加
	t.Tags = append(t.Tags, tag)
	t.UpdatedAt = time.Now()

	return true // 正常に追加
}

// RemoveTag は画像からタグを削除します
func (t *TaggedImage) RemoveTag(tag valueobject.Tag) bool {
	initialLength := len(t.Tags)

	// タグをフィルタリングして削除
	newTags := make([]valueobject.Tag, 0, initialLength)
	for _, existingTag := range t.Tags {
		if !existingTag.Equals(tag) {
			newTags = append(newTags, existingTag)
		}
	}

	t.Tags = newTags
	t.UpdatedAt = time.Now()

	// 削除されたかどうかを返す
	return len(t.Tags) < initialLength
}

// HasTag は指定されたタグが存在するかをチェックします
func (t *TaggedImage) HasTag(tag valueobject.Tag) bool {
	for _, existingTag := range t.Tags {
		if existingTag.Equals(tag) {
			return true
		}
	}
	return false
}

// GetTagNames はタグ名の配列を返します
func (t *TaggedImage) GetTagNames() []string {
	names := make([]string, len(t.Tags))
	for i, tag := range t.Tags {
		names[i] = tag.Name()
	}
	return names
}

// ClearTags は全てのタグを削除します
func (t *TaggedImage) ClearTags() int {
	count := len(t.Tags)
	t.Tags = make([]valueobject.Tag, 0)
	t.UpdatedAt = time.Now()
	return count
}
