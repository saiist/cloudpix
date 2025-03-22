package middleware

import (
	"cloudpix/internal/application/authmanagement/usecase"
	"cloudpix/internal/contextutil"
	"cloudpix/internal/domain/authmanagement/entity"
	"cloudpix/internal/domain/authmanagement/service"
	"cloudpix/internal/domain/authmanagement/valueobject"
	"cloudpix/internal/logging"
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// AuthMiddleware は認証ミドルウェア
type AuthMiddleware struct {
	authUsecase *usecase.AuthUsecase
	logger      logging.Logger
}

// NewAuthMiddleware は新しい認証ミドルウェアを作成します
func NewAuthMiddleware(authUsecase *usecase.AuthUsecase, logger logging.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authUsecase: authUsecase,
		logger:      logger,
	}
}

// Process は認証処理を行い、認証済みのコンテキストを返します
func (m *AuthMiddleware) Process(ctx context.Context, event events.APIGatewayProxyRequest) (context.Context, events.APIGatewayProxyResponse, error) {
	// Authorization ヘッダーの取得
	authHeader, ok := event.Headers["Authorization"]
	if !ok {
		// 認証エラーレスポンス
		m.logger.Warn("認証ヘッダーがありません", nil)
		return nil, m.CreateErrorResponse(401, "認証ヘッダーがありません"), nil
	}

	// Bearer トークンの抽出
	if !strings.HasPrefix(authHeader, "Bearer ") {
		m.logger.Warn("無効な認証フォーマットです", nil)
		return nil, m.CreateErrorResponse(401, "無効な認証フォーマットです"), nil
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// ユーザー認証
	userDTO, err := m.authUsecase.AuthenticateUser(ctx, tokenString)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			m.logger.Warn("無効な認証情報です", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, m.CreateErrorResponse(401, "認証エラー: 無効なトークン"), nil
		}
		if errors.Is(err, service.ErrUserNotFound) {
			m.logger.Warn("ユーザーが見つかりません", nil)
			return nil, m.CreateErrorResponse(401, "認証エラー: ユーザーが見つかりません"), nil
		}
		m.logger.Error(err, "認証処理中にエラーが発生しました", nil)
		return nil, m.CreateErrorResponse(500, "サーバーエラー"), nil
	}

	// userDTOからentity.Userに変換
	userID, err := valueobject.NewUserID(userDTO.UserID)
	if err != nil {
		m.logger.Error(err, "ユーザーIDの変換に失敗しました", nil)
		return nil, m.CreateErrorResponse(500, "サーバーエラー"), nil
	}

	// 文字列のロールをエンティティのUserRoleに変換
	roles := make([]entity.UserRole, 0, len(userDTO.Roles))
	for _, roleStr := range userDTO.Roles {
		switch roleStr {
		case "Admin":
			roles = append(roles, entity.RoleAdmin)
		case "Premium":
			roles = append(roles, entity.RolePremium)
		default:
			roles = append(roles, entity.RoleStandard)
		}
	}

	// User エンティティを作成
	user := entity.NewUser(
		userID,
		userDTO.Username,
		userDTO.Email,
		roles,
	)

	// ユーザー情報をコンテキストに追加
	newCtx := contextutil.WithUserInfo(ctx, user)

	// ロガーにユーザー情報を追加
	m.logger.Info("ユーザー認証成功", map[string]interface{}{
		"userId":   userDTO.UserID,
		"username": userDTO.Username,
		"isAdmin":  userDTO.IsAdmin,
	})

	return newCtx, events.APIGatewayProxyResponse{}, nil
}

// CreateErrorResponse はエラーレスポンスを作成します
func (m *AuthMiddleware) CreateErrorResponse(statusCode int, message string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"error":"` + message + `"}`,
	}
}
