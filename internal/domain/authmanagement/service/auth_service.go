package service

import (
	"cloudpix/internal/domain/authmanagement/entity"
	"cloudpix/internal/domain/authmanagement/valueobject"
	"context"
	"errors"
)

// ErrInvalidCredentials は無効な認証情報エラー
var ErrInvalidCredentials = errors.New("無効な認証情報です")

// ErrUserNotFound はユーザーが見つからないエラー
var ErrUserNotFound = errors.New("ユーザーが見つかりません")

// AuthService は認証サービスのインターフェース
type AuthService interface {
	// VerifyCredentials は認証情報を検証し、ユーザーを取得します
	VerifyCredentials(ctx context.Context, credentials valueobject.Credentials) (*entity.User, error)

	// GetUserFromToken はトークンからユーザー情報を取得します
	GetUserFromToken(ctx context.Context, tokenString string) (*entity.User, error)
}
