package imaging

import (
	"bytes"
	"cloudpix/internal/domain/model"
	"cloudpix/internal/domain/service"
	"image"
	"image/jpeg"
	"image/png"
	"strings"

	"github.com/disintegration/imaging"
)

type ImageServiceImpl struct{}

func NewImageService() service.ImageService {
	return &ImageServiceImpl{}
}

// 画像をデコードする
func (s *ImageServiceImpl) DecodeImage(data *model.ImageData) (width int, height int, err error) {
	// コンテンツタイプをチェック
	contentType := data.ContentType

	var img image.Image

	// フォーマットに応じたデコード
	if strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg") {
		img, err = jpeg.Decode(bytes.NewReader(data.Data))
	} else if strings.Contains(contentType, "png") {
		img, err = png.Decode(bytes.NewReader(data.Data))
	} else {
		return 0, 0, &service.UnsupportedFormatError{ContentType: contentType}
	}

	if err != nil {
		return 0, 0, err
	}

	// 画像サイズを取得
	bounds := img.Bounds()
	width = bounds.Dx()
	height = bounds.Dy()

	return width, height, nil
}

// サムネイルを生成する
func (s *ImageServiceImpl) GenerateThumbnail(data *model.ImageData, width int) (*model.ImageData, int, int, error) {
	// 元画像をデコード
	contentType := data.ContentType

	var img image.Image
	var err error

	// フォーマットに応じたデコード
	if strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg") {
		img, err = jpeg.Decode(bytes.NewReader(data.Data))
	} else if strings.Contains(contentType, "png") {
		img, err = png.Decode(bytes.NewReader(data.Data))
	} else {
		return nil, 0, 0, &service.UnsupportedFormatError{ContentType: contentType}
	}

	if err != nil {
		return nil, 0, 0, err
	}

	// サムネイルを生成
	thumbnail := imaging.Resize(img, width, 0, imaging.Lanczos)

	// サムネイルのサイズを取得
	bounds := thumbnail.Bounds()
	thumbnailWidth := bounds.Dx()
	thumbnailHeight := bounds.Dy()

	// サムネイルをエンコード
	var buf bytes.Buffer
	var encodeErr error

	if strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg") {
		encodeErr = jpeg.Encode(&buf, thumbnail, nil)
	} else if strings.Contains(contentType, "png") {
		encodeErr = png.Encode(&buf, thumbnail)
	}

	if encodeErr != nil {
		return nil, 0, 0, encodeErr
	}

	// サムネイルデータを返す
	return &model.ImageData{
		Data:        buf.Bytes(),
		ContentType: contentType,
	}, thumbnailWidth, thumbnailHeight, nil
}

// 画像IDをファイル名から抽出する
func (s *ImageServiceImpl) ExtractImageID(filename string) (string, error) {
	parts := strings.Split(filename, "-")
	if len(parts) < 2 {
		return "", &service.InvalidFilenameError{Filename: filename}
	}

	return parts[0], nil
}
