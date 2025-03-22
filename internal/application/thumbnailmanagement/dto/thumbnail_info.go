package dto

// ThumbnailInfoDTO はサムネイル情報のデータ転送オブジェクト
type ThumbnailInfoDTO struct {
	ImageID      string `json:"imageId"`
	ThumbnailURL string `json:"thumbnailUrl"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	ContentType  string `json:"contentType"`
}

// ThumbnailGenerationRequestDTO はサムネイル生成リクエストのDTO
type ThumbnailGenerationRequestDTO struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}

// ThumbnailGenerationResponseDTO はサムネイル生成レスポンスのDTO
type ThumbnailGenerationResponseDTO struct {
	Success      bool   `json:"success"`
	ImageID      string `json:"imageId"`
	ThumbnailURL string `json:"thumbnailUrl"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Message      string `json:"message,omitempty"`
}
