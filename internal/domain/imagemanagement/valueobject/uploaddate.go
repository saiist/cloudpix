package valueobject

import (
	"errors"
	"regexp"
	"time"
)

// UploadDate は画像のアップロード日を表す値オブジェクト
type UploadDate struct {
	value string // YYYY-MM-DD 形式
}

// NewUploadDate はアップロード日の値オブジェクトを作成し、検証します
func NewUploadDate(value string) (UploadDate, error) {
	// YYYY-MM-DD 形式かどうかチェック
	matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, value)
	if !matched {
		return UploadDate{}, errors.New("アップロード日はYYYY-MM-DD形式である必要があります")
	}

	// 日付として有効かどうかチェック
	_, err := time.Parse("2006-01-02", value)
	if err != nil {
		return UploadDate{}, errors.New("無効な日付形式です")
	}

	return UploadDate{value: value}, nil
}

// Today は今日の日付の UploadDate を返します
func Today() UploadDate {
	today := time.Now().Format("2006-01-02")
	date, _ := NewUploadDate(today)
	return date
}

// String はアップロード日を文字列として返します
func (d UploadDate) String() string {
	return d.value
}

// Time は time.Time 型に変換して返します
func (d UploadDate) Time() (time.Time, error) {
	return time.Parse("2006-01-02", d.value)
}

// Equals は2つのアップロード日が等しいかどうかを判定します
func (d UploadDate) Equals(other UploadDate) bool {
	return d.value == other.value
}
