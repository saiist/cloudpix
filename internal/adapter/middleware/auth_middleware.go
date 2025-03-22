package middleware

import (
	"context"

	"cloudpix/internal/contextutil"
	"cloudpix/internal/domain/model"
	"cloudpix/internal/domain/repository"

	"github.com/aws/aws-lambda-go/events"
)

// AuthContext はユーザー情報を含むコンテキストのキー
type AuthContext struct{}

// AuthMiddleware は認証ミドルウェアのインターフェース
type AuthMiddleware interface {
	Process(ctx context.Context, event events.APIGatewayProxyRequest) (context.Context, *model.UserInfo, events.APIGatewayProxyResponse, error)
	CreateErrorResponse(statusCode int, message string) events.APIGatewayProxyResponse
}

// CognitoAuthMiddleware はCognito認証のミドルウェア実装
type CognitoAuthMiddleware struct {
	authRepo repository.AuthRepository
}

// NewCognitoAuthMiddleware は新しいCognito認証ミドルウェアを作成する
func NewCognitoAuthMiddleware(authRepo repository.AuthRepository) AuthMiddleware {
	return &CognitoAuthMiddleware{
		authRepo: authRepo,
	}
}

// Process は認証処理を行い、認証済みのコンテキストを返す
func (m *CognitoAuthMiddleware) Process(ctx context.Context, event events.APIGatewayProxyRequest) (context.Context, *model.UserInfo, events.APIGatewayProxyResponse, error) {
	// Authorization ヘッダーの取得
	authHeader, ok := event.Headers["Authorization"]
	if !ok {
		// 認証エラーレスポンス
		resp := m.authRepo.CreateErrorResponse(401, "認証ヘッダーがありません").(events.APIGatewayProxyResponse)
		return nil, nil, resp, nil
	}

	// ユーザー認証
	userInfo, err := m.authRepo.GetUserInfoFromHeader(ctx, authHeader)
	if err != nil {
		// 認証エラーレスポンス
		resp := m.authRepo.CreateErrorResponse(401, "認証エラー: "+err.Error()).(events.APIGatewayProxyResponse)
		return nil, nil, resp, nil
	}

	// ユーザー情報をコンテキストに追加
	newCtx := contextutil.WithUserInfo(ctx, userInfo)

	return newCtx, userInfo, events.APIGatewayProxyResponse{}, nil
}

// CreateErrorResponse はエラーレスポンスを作成する
func (m *CognitoAuthMiddleware) CreateErrorResponse(statusCode int, message string) events.APIGatewayProxyResponse {
	return m.authRepo.CreateErrorResponse(statusCode, message).(events.APIGatewayProxyResponse)
}
