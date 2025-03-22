package valueobject

// ImageData はバイナリ画像データと関連情報を表す値オブジェクト
type ImageData struct {
	Data        []byte
	ContentType string
}

// NewImageData は新しい画像データ値オブジェクトを作成します
func NewImageData(data []byte, contentType string) ImageData {
	return ImageData{
		Data:        data,
		ContentType: contentType,
	}
}

// Size はデータサイズをバイト単位で返します
func (i ImageData) Size() int {
	return len(i.Data)
}

// IsEmpty はデータが空かどうかを判定します
func (i ImageData) IsEmpty() bool {
	return len(i.Data) == 0
}
