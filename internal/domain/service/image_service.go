package service

import (
	"cloudpix/internal/domain/model"
	"fmt"
)

// ImageService は画像処理に関するドメインサービス
type ImageService interface {
	// 画像をデコードする
	DecodeImage(data *model.ImageData) (width int, height int, err error)

	// サムネイルを生成する
	GenerateThumbnail(data *model.ImageData, width int) (*model.ImageData, int, int, error)

	// 画像IDをファイル名から抽出する
	ExtractImageID(filename string) (string, error)
}

// サポートされていない画像フォーマットエラー
type UnsupportedFormatError struct {
	ContentType string
}

func (e *UnsupportedFormatError) Error() string {
	return fmt.Sprintf("unsupported image format: %s", e.ContentType)
}

// IDが抽出できないエラー
type InvalidFilenameError struct {
	Filename string
}

func (e *InvalidFilenameError) Error() string {
	return fmt.Sprintf("could not extract image ID from filename: %s", e.Filename)
}
