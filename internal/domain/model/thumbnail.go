package model

// サムネイル情報の構造体
type ThumbnailInfo struct {
	ImageID      string `json:"imageId"`      // 元画像のID
	ThumbnailKey string `json:"thumbnailKey"` // サムネイルのS3キー
	ThumbnailURL string `json:"thumbnailUrl"` // サムネイルのURL
	Width        int    `json:"width"`        // サムネイルの幅
	Height       int    `json:"height"`       // サムネイルの高さ
	OriginalKey  string `json:"originalKey"`  // 元画像のS3キー
	ContentType  string `json:"contentType"`  // コンテンツタイプ
}

// イメージデータ構造体
type ImageData struct {
	Data        []byte // 画像のバイナリデータ
	ContentType string // コンテンツタイプ
}
