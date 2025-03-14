package model

// メタデータ構造体
type ImageMetadata struct {
	ImageID     string `json:"imageId"`
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
	Size        int    `json:"size"`
	UploadDate  string `json:"uploadDate"`
	S3ObjectKey string `json:"s3ObjectKey"`
	DownloadURL string `json:"downloadUrl"`
}

// リストレスポンス構造体
type ListResponse struct {
	Images []ImageMetadata `json:"images"`
	Count  int             `json:"count"`
}
