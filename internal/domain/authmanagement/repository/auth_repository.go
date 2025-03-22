package repository

import (
	"cloudpix/internal/domain/authmanagement/entity"
	"cloudpix/internal/domain/authmanagement/valueobject"
	"context"
)

// AuthRepository は認証情報のリポジトリインターフェース
type AuthRepository interface {
	// FindUserByID はIDに基づいてユーザーを検索します
	FindUserByID(ctx context.Context, userID valueobject.UserID) (*entity.User, error)

	// FindUserByUsername はユーザー名に基づいてユーザーを検索します
	FindUserByUsername(ctx context.Context, username string) (*entity.User, error)

	// VerifyToken はトークンを検証し、ユーザー情報を返します
	VerifyToken(ctx context.Context, tokenString string) (*entity.User, error)
}

// ErrorResponse はエラーレスポンスを生成するインターフェース
type ErrorResponseGenerator interface {
	// CreateErrorResponse はエラーレスポンスを作成します
	CreateErrorResponse(statusCode int, message string) interface{}
}
