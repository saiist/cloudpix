package shared

import (
	"cloudpix/config"
	"cloudpix/internal/application/authmanagement/usecase"
	"cloudpix/internal/infrastructure/auth/cognito"
	"cloudpix/internal/logging"

	"github.com/aws/aws-sdk-go/aws/session"
)

// InitAuth は認証関連のコンポーネントを初期化します
func InitAuth(cfg *config.Config, sess *session.Session, logger logging.Logger) *usecase.AuthUsecase {
	// 認証サービスの初期化
	authService := cognito.NewCognitoAuthService(
		cfg.AWSRegion,
		cfg.UserPoolID,
		cfg.ClientID,
	)

	// 認証リポジトリの初期化
	authRepository := cognito.NewCognitoAuthRepository(
		cfg.AWSRegion,
		cfg.UserPoolID,
		cfg.ClientID,
	)

	// 認証ユースケースの初期化
	authUsecase := usecase.NewAuthUsecase(
		authRepository,
		authService,
	)

	return authUsecase
}
