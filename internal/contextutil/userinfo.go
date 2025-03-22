package contextutil

import (
	"cloudpix/internal/domain/authmanagement/entity"
	"context"
)

// ContextKey はコンテキストキーの型
type ContextKey string

const (
	// UserInfoKey はコンテキスト内のユーザー情報のキー
	UserInfoKey ContextKey = "userInfo"
)

// WithUserInfo はコンテキストにユーザー情報を追加する
func WithUserInfo(ctx context.Context, userInfo *entity.User) context.Context {
	return context.WithValue(ctx, UserInfoKey, userInfo)
}

// GetUserInfo はコンテキストからユーザー情報を取得する
func GetUserInfo(ctx context.Context) (*entity.User, bool) {
	userInfo, ok := ctx.Value(UserInfoKey).(*entity.User)
	return userInfo, ok
}
