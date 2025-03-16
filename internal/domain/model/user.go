package model

// UserInfo はユーザー情報を保持する構造体
type UserInfo struct {
	UserID    string
	Username  string
	Email     string
	Groups    []string
	IsAdmin   bool
	IsPremium bool
}
