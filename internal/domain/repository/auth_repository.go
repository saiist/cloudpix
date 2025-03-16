package repository

import (
	"context"

	"cloudpix/internal/domain/model"
)

// AuthRepository は認証に関する操作を定義するインターフェース
type AuthRepository interface {
	// VerifyToken はトークンを検証し、ユーザー情報を返す
	VerifyToken(ctx context.Context, tokenString string) (*model.UserInfo, error)

	// GetUserInfoFromHeader はAuthorizationヘッダーからユーザー情報を取得する
	GetUserInfoFromHeader(ctx context.Context, authHeader string) (*model.UserInfo, error)

	// CheckPermission は指定されたユーザーが操作を実行できるかチェックする
	CheckPermission(userInfo *model.UserInfo, ownerID string, isPublic bool) bool

	// CreateErrorResponse はエラーレスポンスを作成する
	CreateErrorResponse(statusCode int, message string) interface{}
}
