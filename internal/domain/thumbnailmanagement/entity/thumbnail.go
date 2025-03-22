package entity

import (
	"cloudpix/internal/domain/thumbnailmanagement/valueobject"
	"time"
)

// Thumbnail エンティティはサムネイル情報を表します
type Thumbnail struct {
	ImageID      string
	ThumbnailKey string
	ThumbnailURL string
	Dimensions   valueobject.Dimensions
	OriginalKey  string
	ContentType  string
	CreatedAt    time.Time
}

// NewThumbnail は新しいサムネイルエンティティを作成します
func NewThumbnail(
	imageID string,
	thumbnailKey string,
	thumbnailURL string,
	dimensions valueobject.Dimensions,
	originalKey string,
	contentType string,
) *Thumbnail {
	return &Thumbnail{
		ImageID:      imageID,
		ThumbnailKey: thumbnailKey,
		ThumbnailURL: thumbnailURL,
		Dimensions:   dimensions,
		OriginalKey:  originalKey,
		ContentType:  contentType,
		CreatedAt:    time.Now(),
	}
}

// GetWidth はサムネイルの幅を返します
func (t *Thumbnail) GetWidth() int {
	return t.Dimensions.Width()
}

// GetHeight はサムネイルの高さを返します
func (t *Thumbnail) GetHeight() int {
	return t.Dimensions.Height()
}

// IsSquare はサムネイルが正方形かどうかを判定します
func (t *Thumbnail) IsSquare() bool {
	return t.Dimensions.IsSquare()
}
