package cognito

import (
	"cloudpix/internal/domain/authmanagement/entity"
	"cloudpix/internal/domain/authmanagement/service"
	"cloudpix/internal/domain/authmanagement/valueobject"
	"context"
	"strings"
)

// CognitoAuthService はCognito認証サービスの実装
type CognitoAuthService struct {
	authRepository *CognitoAuthRepository
}

// NewCognitoAuthService は新しいCognito認証サービスを作成します
func NewCognitoAuthService(region, userPoolID, clientID string) service.AuthService {
	return &CognitoAuthService{
		authRepository: NewCognitoAuthRepository(region, userPoolID, clientID).(*CognitoAuthRepository),
	}
}

// VerifyCredentials は認証情報を検証し、ユーザー情報を返します
func (s *CognitoAuthService) VerifyCredentials(ctx context.Context, credentials valueobject.Credentials) (*entity.User, error) {
	// クレデンシャルのタイプをチェック
	if credentials.TokenType() != valueobject.BearerToken {
		return nil, service.ErrInvalidCredentials
	}

	// トークンを検証
	user, err := s.authRepository.VerifyToken(ctx, credentials.Token())
	if err != nil {
		return nil, service.ErrInvalidCredentials
	}

	return user, nil
}

// GetUserFromToken はトークンからユーザー情報を取得します
func (s *CognitoAuthService) GetUserFromToken(ctx context.Context, tokenString string) (*entity.User, error) {
	if strings.TrimSpace(tokenString) == "" {
		return nil, service.ErrInvalidCredentials
	}

	// トークンを検証
	user, err := s.authRepository.VerifyToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	return user, nil
}
