package valueobject

import (
	"errors"
	"strings"
)

// UserID はユーザーIDを表す値オブジェクト
type UserID struct {
	value string
}

// NewUserID は新しいユーザーIDを作成します
func NewUserID(value string) (UserID, error) {
	if strings.TrimSpace(value) == "" {
		return UserID{}, errors.New("ユーザーIDは空にできません")
	}
	return UserID{value: value}, nil
}

// String はユーザーIDを文字列として返します
func (u UserID) String() string {
	return u.value
}

// Equals は2つのユーザーIDが等しいかどうかを判定します
func (u UserID) Equals(other UserID) bool {
	return u.value == other.value
}
