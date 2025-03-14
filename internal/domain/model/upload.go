package model

import "time"

// UploadRequest はクライアントからのアップロードリクエスト構造体
type UploadRequest struct {
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
	Data        string `json:"data,omitempty"` // Base64エンコードされた画像データ
}

// UploadMetadata はアップロードされた画像のメタデータ構造体
type UploadMetadata struct {
	ImageID      string    `json:"ImageID"`
	FileName     string    `json:"fileName"`
	ContentType  string    `json:"contentType"`
	Size         int       `json:"size"`
	UploadDate   string    `json:"UploadDate"`
	CreatedAt    time.Time `json:"createdAt"`
	S3ObjectKey  string    `json:"s3ObjectKey"`
	S3BucketName string    `json:"s3BucketName"`
	DownloadURL  string    `json:"downloadUrl"`
}

// UploadResponse はクライアントへのレスポンス構造体
type UploadResponse struct {
	ImageID     string `json:"imageId"`
	UploadURL   string `json:"uploadUrl,omitempty"`
	DownloadURL string `json:"downloadUrl"`
	Message     string `json:"message"`
}
