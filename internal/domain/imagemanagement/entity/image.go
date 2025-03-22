package entity

import (
	"cloudpix/internal/domain/imagemanagement/valueobject"
	"time"
)

// Image エンティティは画像の基本情報を表します
type Image struct {
	ID           string
	FileName     valueobject.FileName
	ContentType  valueobject.ContentType
	Size         valueobject.ImageSize
	UploadDate   valueobject.UploadDate
	S3ObjectKey  string
	DownloadURL  string
	CreatedAt    time.Time
	ModifiedAt   time.Time
	HasThumbnail bool
}

// NewImage は新しい画像エンティティを作成します
func NewImage(
	id string,
	fileName valueobject.FileName,
	contentType valueobject.ContentType,
	size valueobject.ImageSize,
	uploadDate valueobject.UploadDate,
	s3ObjectKey string,
	downloadURL string,
) *Image {
	now := time.Now()
	return &Image{
		ID:          id,
		FileName:    fileName,
		ContentType: contentType,
		Size:        size,
		UploadDate:  uploadDate,
		S3ObjectKey: s3ObjectKey,
		DownloadURL: downloadURL,
		CreatedAt:   now,
		ModifiedAt:  now,
	}
}

// SetThumbnail はサムネイルが生成されたことを記録します
func (i *Image) SetThumbnail(hasThumbnail bool) {
	i.HasThumbnail = hasThumbnail
	i.ModifiedAt = time.Now()
}

// IsImage は有効な画像かどうかを判定します
func (i *Image) IsImage() bool {
	return i.ContentType.IsJPEG() || i.ContentType.IsPNG() || i.ContentType.IsGIF()
}
