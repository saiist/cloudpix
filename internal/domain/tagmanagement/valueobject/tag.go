package valueobject

import (
	"errors"
	"strings"
)

// Tag はタグを表す値オブジェクト
type Tag struct {
	name string
}

// NewTag は新しいタグを作成します
func NewTag(name string) (Tag, error) {
	// タグ名をトリムして正規化
	normalizedName := strings.TrimSpace(strings.ToLower(name))

	// 空のタグをチェック
	if normalizedName == "" {
		return Tag{}, errors.New("タグ名は空にできません")
	}

	// タグの長さをチェック
	if len(normalizedName) > 50 {
		return Tag{}, errors.New("タグ名は50文字以下である必要があります")
	}

	// 不正な文字をチェック
	for _, c := range normalizedName {
		if !isValidTagChar(c) {
			return Tag{}, errors.New("タグ名に不正な文字が含まれています")
		}
	}

	return Tag{name: normalizedName}, nil
}

// isValidTagChar はタグ名に使用できる文字かどうかをチェックします
func isValidTagChar(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '_' || c == '.'
}

// Name はタグ名を返します
func (t Tag) Name() string {
	return t.name
}

// Equals は2つのタグが等しいかどうかを判定します
func (t Tag) Equals(other Tag) bool {
	return t.name == other.name
}
