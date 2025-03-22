package valueobject

import (
	"errors"
	"path/filepath"
	"strings"
)

// FileName は画像のファイル名を表す値オブジェクト
type FileName struct {
	value string
}

// NewFileName はファイル名の値オブジェクトを作成し、検証します
func NewFileName(value string) (FileName, error) {
	if value == "" {
		return FileName{}, errors.New("ファイル名は空にできません")
	}

	// 不正な文字が含まれていないか確認
	if strings.ContainsAny(value, "\\/:*?\"<>|") {
		return FileName{}, errors.New("ファイル名に不正な文字が含まれています")
	}

	return FileName{value: value}, nil
}

// String はファイル名を文字列として返します
func (f FileName) String() string {
	return f.value
}

// Extension はファイルの拡張子を返します
func (f FileName) Extension() string {
	return strings.ToLower(filepath.Ext(f.value))
}

// BaseName は拡張子を除いたファイル名を返します
func (f FileName) BaseName() string {
	return strings.TrimSuffix(f.value, f.Extension())
}

// Equals は2つのファイル名が等しいかどうかを判定します
func (f FileName) Equals(other FileName) bool {
	return f.value == other.value
}
