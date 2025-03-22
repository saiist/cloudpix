package dto

// UploadRequest はクライアントからのアップロードリクエストを表します
type UploadRequest struct {
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
	Data        string `json:"data,omitempty"` // Base64エンコードされた画像データ
}

// UploadResponse はアップロード操作のレスポンスを表します
type UploadResponse struct {
	ImageID     string `json:"imageId"`
	UploadURL   string `json:"uploadUrl,omitempty"`
	DownloadURL string `json:"downloadUrl"`
	Message     string `json:"message"`
}
