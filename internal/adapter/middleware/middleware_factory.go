package middleware

import (
	"cloudpix/internal/infrastructure/auth"
)

// CreateDefaultAuthMiddleware は環境変数から標準認証ミドルウェアを作成する
func CreateDefaultAuthMiddleware(region, userPoolID, clientID string) AuthMiddleware {
	authRepo := auth.NewCognitoAuthRepository(region, userPoolID, clientID)
	return NewCognitoAuthMiddleware(authRepo)
}
