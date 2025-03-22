package valueobject

import "errors"

// Dimensions はサムネイルの幅と高さを表す値オブジェクト
type Dimensions struct {
	width  int
	height int
}

// NewDimensions は新しいサイズの値オブジェクトを作成します
func NewDimensions(width, height int) (Dimensions, error) {
	if width <= 0 || height <= 0 {
		return Dimensions{}, errors.New("幅と高さは正の値である必要があります")
	}
	return Dimensions{width: width, height: height}, nil
}

// Width は幅を返します
func (d Dimensions) Width() int {
	return d.width
}

// Height は高さを返します
func (d Dimensions) Height() int {
	return d.height
}

// AspectRatio はアスペクト比（幅/高さ）を返します
func (d Dimensions) AspectRatio() float64 {
	return float64(d.width) / float64(d.height)
}

// IsSquare は正方形かどうかを判定します
func (d Dimensions) IsSquare() bool {
	return d.width == d.height
}

// IsLandscape は横長かどうかを判定します
func (d Dimensions) IsLandscape() bool {
	return d.width > d.height
}

// IsPortrait は縦長かどうかを判定します
func (d Dimensions) IsPortrait() bool {
	return d.height > d.width
}

// Equals は2つのサイズが等しいかどうかを判定します
func (d Dimensions) Equals(other Dimensions) bool {
	return d.width == other.width && d.height == other.height
}

// Scale は指定された幅に合わせて高さを調整した新しいサイズを返します
func (d Dimensions) Scale(targetWidth int) (Dimensions, error) {
	if targetWidth <= 0 {
		return Dimensions{}, errors.New("ターゲット幅は正の値である必要があります")
	}

	ratio := float64(targetWidth) / float64(d.width)
	newHeight := int(float64(d.height) * ratio)

	// 少なくとも1ピクセルの高さを確保
	if newHeight < 1 {
		newHeight = 1
	}

	return NewDimensions(targetWidth, newHeight)
}
