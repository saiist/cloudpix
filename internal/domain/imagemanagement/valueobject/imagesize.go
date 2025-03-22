package valueobject

import "errors"

// ImageSize は画像のサイズを表す値オブジェクト
type ImageSize struct {
	value int
}

// NewImageSize は画像サイズの値オブジェクトを作成し、検証します
func NewImageSize(value int) (ImageSize, error) {
	if value < 0 {
		return ImageSize{}, errors.New("画像サイズは0以上である必要があります")
	}

	return ImageSize{value: value}, nil
}

// Value はサイズの値を返します
func (s ImageSize) Value() int {
	return s.value
}

// IsEmpty はサイズが0かどうかを判定します
func (s ImageSize) IsEmpty() bool {
	return s.value == 0
}

// Equals は2つのサイズが等しいかどうかを判定します
func (s ImageSize) Equals(other ImageSize) bool {
	return s.value == other.value
}
