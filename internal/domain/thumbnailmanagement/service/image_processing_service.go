package service

import (
	"cloudpix/internal/domain/thumbnailmanagement/valueobject"
)

// ImageProcessingService は画像処理サービスのインターフェース
type ImageProcessingService interface {
	// DecodeImage は画像データをデコードして幅と高さを取得します
	DecodeImage(data valueobject.ImageData) (int, int, error)

	// GenerateThumbnail はサムネイルを生成します
	GenerateThumbnail(data valueobject.ImageData, targetWidth int) (valueobject.ImageData, valueobject.Dimensions, error)

	// ExtractImageID は画像キーから画像IDを抽出します
	ExtractImageID(key string) (string, error)

	// IsSupported は指定されたコンテンツタイプがサポートされているかをチェックします
	IsSupported(contentType string) bool
}
