package contextutil

import (
	"cloudpix/internal/domain/model"
	"context"
)

// ContextKey はコンテキストキーの型
type ContextKey string

const (
	// UserInfoKey はコンテキスト内のユーザー情報のキー
	UserInfoKey ContextKey = "userInfo"
)

// WithUserInfo はコンテキストにユーザー情報を追加する
func WithUserInfo(ctx context.Context, userInfo *model.UserInfo) context.Context {
	return context.WithValue(ctx, UserInfoKey, userInfo)
}

// GetUserInfo はコンテキストからユーザー情報を取得する
func GetUserInfo(ctx context.Context) (*model.UserInfo, bool) {
	userInfo, ok := ctx.Value(UserInfoKey).(*model.UserInfo)
	return userInfo, ok
}
