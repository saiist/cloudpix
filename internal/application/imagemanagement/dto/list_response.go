package dto

// ImageMetadataDTO は画像メタデータのデータ転送オブジェクト
type ImageMetadataDTO struct {
	ImageID      string `json:"imageId"`
	FileName     string `json:"fileName"`
	ContentType  string `json:"contentType"`
	Size         int    `json:"size"`
	UploadDate   string `json:"uploadDate"`
	DownloadURL  string `json:"downloadUrl"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
}

// ListResponse は画像一覧のレスポンスを表します
type ListResponse struct {
	Images []ImageMetadataDTO `json:"images"`
	Count  int                `json:"count"`
}
