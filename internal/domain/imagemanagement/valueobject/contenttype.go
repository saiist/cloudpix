package valueobject

import (
	"errors"
	"strings"
)

// ContentType は画像のコンテンツタイプを表す値オブジェクト
type ContentType struct {
	value string
}

// NewContentType はコンテンツタイプの値オブジェクトを作成し、検証します
func NewContentType(value string) (ContentType, error) {
	if value == "" {
		return ContentType{}, errors.New("コンテンツタイプは空にできません")
	}

	// 画像タイプのみ許可
	if !strings.HasPrefix(value, "image/") {
		return ContentType{}, errors.New("コンテンツタイプは image/ で始まる必要があります")
	}

	return ContentType{value: value}, nil
}

// String はコンテンツタイプを文字列として返します
func (c ContentType) String() string {
	return c.value
}

// Equals は2つのコンテンツタイプが等しいかどうかを判定します
func (c ContentType) Equals(other ContentType) bool {
	return c.value == other.value
}

// IsJPEG はコンテンツタイプがJPEGかどうかを判定します
func (c ContentType) IsJPEG() bool {
	return c.value == "image/jpeg" || c.value == "image/jpg"
}

// IsPNG はコンテンツタイプがPNGかどうかを判定します
func (c ContentType) IsPNG() bool {
	return c.value == "image/png"
}

// IsGIF はコンテンツタイプがGIFかどうかを判定します
func (c ContentType) IsGIF() bool {
	return c.value == "image/gif"
}
