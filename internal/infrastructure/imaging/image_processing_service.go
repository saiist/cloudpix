package imaging

import (
	"bytes"
	"cloudpix/internal/domain/thumbnailmanagement/service"
	"cloudpix/internal/domain/thumbnailmanagement/valueobject"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

// ImageProcessingServiceImpl は画像処理サービスの実装
type ImageProcessingServiceImpl struct {
	supportedFormats map[string]bool
}

// NewImageProcessingService は新しい画像処理サービスを作成します
func NewImageProcessingService() service.ImageProcessingService {
	return &ImageProcessingServiceImpl{
		supportedFormats: map[string]bool{
			"image/jpeg": true,
			"image/jpg":  true,
			"image/png":  true,
			"image/gif":  true,
		},
	}
}

// DecodeImage は画像をデコードして幅と高さを取得します
func (s *ImageProcessingServiceImpl) DecodeImage(data valueobject.ImageData) (int, int, error) {
	// コンテンツタイプをチェック
	if !s.IsSupported(data.ContentType) {
		return 0, 0, fmt.Errorf("unsupported image format: %s", data.ContentType)
	}

	var img image.Image
	var err error

	// フォーマットに応じたデコード
	reader := bytes.NewReader(data.Data)
	if strings.Contains(data.ContentType, "jpeg") || strings.Contains(data.ContentType, "jpg") {
		img, err = jpeg.Decode(reader)
	} else if strings.Contains(data.ContentType, "png") {
		img, err = png.Decode(reader)
	} else {
		// サポートされている他のフォーマットはここで対応
		return 0, 0, fmt.Errorf("format not implemented: %s", data.ContentType)
	}

	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode image: %w", err)
	}

	// 画像サイズを取得
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	return width, height, nil
}

// GenerateThumbnail はサムネイルを生成します
func (s *ImageProcessingServiceImpl) GenerateThumbnail(data valueobject.ImageData, targetWidth int) (valueobject.ImageData, valueobject.Dimensions, error) {
	// コンテンツタイプをチェック
	if !s.IsSupported(data.ContentType) {
		return valueobject.ImageData{}, valueobject.Dimensions{}, fmt.Errorf("unsupported image format: %s", data.ContentType)
	}

	var img image.Image
	var err error

	// フォーマットに応じたデコード
	reader := bytes.NewReader(data.Data)
	if strings.Contains(data.ContentType, "jpeg") || strings.Contains(data.ContentType, "jpg") {
		img, err = jpeg.Decode(reader)
	} else if strings.Contains(data.ContentType, "png") {
		img, err = png.Decode(reader)
	} else {
		return valueobject.ImageData{}, valueobject.Dimensions{}, fmt.Errorf("format not implemented: %s", data.ContentType)
	}

	if err != nil {
		return valueobject.ImageData{}, valueobject.Dimensions{}, fmt.Errorf("failed to decode image: %w", err)
	}

	// サムネイルを生成（幅を指定して縦横比を維持）
	thumbnail := imaging.Resize(img, targetWidth, 0, imaging.Lanczos)

	// サムネイルのサイズを取得
	bounds := thumbnail.Bounds()
	thumbnailWidth := bounds.Dx()
	thumbnailHeight := bounds.Dy()

	// サムネイルをエンコード
	var buf bytes.Buffer
	var encodeErr error

	if strings.Contains(data.ContentType, "jpeg") || strings.Contains(data.ContentType, "jpg") {
		encodeErr = jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: 85})
	} else if strings.Contains(data.ContentType, "png") {
		encodeErr = png.Encode(&buf, thumbnail)
	}

	if encodeErr != nil {
		return valueobject.ImageData{}, valueobject.Dimensions{}, fmt.Errorf("failed to encode thumbnail: %w", encodeErr)
	}

	// サムネイルデータを作成
	thumbnailData := valueobject.NewImageData(buf.Bytes(), data.ContentType)

	// サムネイルのサイズ値オブジェクトを作成
	dimensions, err := valueobject.NewDimensions(thumbnailWidth, thumbnailHeight)
	if err != nil {
		return valueobject.ImageData{}, valueobject.Dimensions{}, fmt.Errorf("invalid thumbnail dimensions: %w", err)
	}

	return thumbnailData, dimensions, nil
}

// ExtractImageID は画像キーから画像IDを抽出します
func (s *ImageProcessingServiceImpl) ExtractImageID(key string) (string, error) {
	// ファイル名部分を取得
	filename := filepath.Base(key)

	// IDを抽出（フォーマット: {ID}-{filename} を想定）
	parts := strings.SplitN(filename, "-", 2)
	if len(parts) < 2 {
		return "", errors.New("invalid filename format: no ID found")
	}

	imageID := parts[0]
	if imageID == "" {
		return "", errors.New("empty image ID")
	}

	return imageID, nil
}

// IsSupported は指定されたコンテンツタイプがサポートされているかをチェックします
func (s *ImageProcessingServiceImpl) IsSupported(contentType string) bool {
	_, supported := s.supportedFormats[contentType]
	return supported
}
