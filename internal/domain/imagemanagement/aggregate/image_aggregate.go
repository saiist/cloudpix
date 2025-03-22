package aggregate

import (
	"cloudpix/internal/domain/imagemanagement/entity"
	"errors"
)

// ImageAggregate は画像とその関連情報を含む集約ルート
type ImageAggregate struct {
	Image           *entity.Image
	ThumbnailURL    string
	ThumbnailWidth  int
	ThumbnailHeight int
	Tags            []string
}

// NewImageAggregate は新しい画像集約を作成します
func NewImageAggregate(image *entity.Image) *ImageAggregate {
	return &ImageAggregate{
		Image: image,
		Tags:  make([]string, 0),
	}
}

// SetThumbnail はサムネイル情報を設定します
func (a *ImageAggregate) SetThumbnail(url string, width, height int) {
	a.ThumbnailURL = url
	a.ThumbnailWidth = width
	a.ThumbnailHeight = height
	a.Image.SetThumbnail(true)
}

// AddTag はタグを追加します（重複チェック付き）
func (a *ImageAggregate) AddTag(tag string) error {
	if tag == "" {
		return errors.New("空のタグは追加できません")
	}

	// 重複チェック
	for _, existingTag := range a.Tags {
		if existingTag == tag {
			return nil // 既に存在するタグなので何もしない
		}
	}

	a.Tags = append(a.Tags, tag)
	return nil
}

// RemoveTag はタグを削除します
func (a *ImageAggregate) RemoveTag(tag string) {
	var newTags []string
	for _, existingTag := range a.Tags {
		if existingTag != tag {
			newTags = append(newTags, existingTag)
		}
	}
	a.Tags = newTags
}

// GetImageID は画像IDを返します
func (a *ImageAggregate) GetImageID() string {
	return a.Image.ID
}
