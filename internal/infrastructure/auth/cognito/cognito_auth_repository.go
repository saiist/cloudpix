package cognito

import (
	"cloudpix/internal/domain/authmanagement/entity"
	"cloudpix/internal/domain/authmanagement/repository"
	"cloudpix/internal/domain/authmanagement/valueobject"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/jwk"
)

// CognitoConfig はCognito設定を保持する構造体
type CognitoConfig struct {
	Region      string
	UserPoolID  string
	ClientID    string
	JwksURI     string
	keySet      jwk.Set
	lastUpdated time.Time
}

// CognitoClaims はJWTトークンのクレーム情報
type CognitoClaims struct {
	jwt.RegisteredClaims
	Username string   `json:"cognito:username"`
	Groups   []string `json:"cognito:groups"`
	Email    string   `json:"email"`
}

// CognitoAuthRepository はCognito認証リポジトリの実装
type CognitoAuthRepository struct {
	config *CognitoConfig
}

// NewCognitoAuthRepository は新しいCognito認証リポジトリを作成します
func NewCognitoAuthRepository(region, userPoolID, clientID string) repository.AuthRepository {
	jwksURI := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", region, userPoolID)

	config := &CognitoConfig{
		Region:     region,
		UserPoolID: userPoolID,
		ClientID:   clientID,
		JwksURI:    jwksURI,
	}

	return &CognitoAuthRepository{
		config: config,
	}
}

// getKeySet はJWKSから鍵セットを取得します
func (r *CognitoAuthRepository) getKeySet(ctx context.Context) (jwk.Set, error) {
	// 最後の更新から1時間以内ならキャッシュを使用
	if r.config.keySet != nil && time.Since(r.config.lastUpdated) < time.Hour {
		return r.config.keySet, nil
	}

	keySet, err := jwk.Fetch(ctx, r.config.JwksURI)
	if err != nil {
		return nil, fmt.Errorf("JWKS取得エラー: %w", err)
	}

	r.config.keySet = keySet
	r.config.lastUpdated = time.Now()
	return keySet, nil
}

// FindUserByID はIDに基づいてユーザーを検索します
func (r *CognitoAuthRepository) FindUserByID(ctx context.Context, userID valueobject.UserID) (*entity.User, error) {
	// この実装では、Cognitoからユーザー情報を取得する方法がないため、
	// トークンの検証とユーザー情報の取得が必要です。
	// TODO: 通常、トークンからユーザー情報を取得するため、実装はスキップします。
	return nil, errors.New("FindUserByID is not implemented for Cognito")
}

// FindUserByUsername はユーザー名に基づいてユーザーを検索します
func (r *CognitoAuthRepository) FindUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	// Cognitoからユーザー情報を取得するAPIを呼び出す必要があります
	// TODO: この実装では、直接ユーザー情報を取得する方法がないため、実装はスキップします。
	return nil, errors.New("FindUserByUsername is not implemented for Cognito")
}

// VerifyToken はJWTトークンを検証し、ユーザー情報を返します
func (r *CognitoAuthRepository) VerifyToken(ctx context.Context, tokenString string) (*entity.User, error) {
	// JWKSを取得
	keySet, err := r.getKeySet(ctx)
	if err != nil {
		return nil, err
	}

	// トークンの解析と検証
	token, err := jwt.ParseWithClaims(tokenString, &CognitoClaims{}, func(token *jwt.Token) (interface{}, error) {
		// アルゴリズムの確認
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("予期しない署名方式: %v", token.Header["alg"])
		}

		// kidの取得
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("トークンのkidが見つかりません")
		}

		// JWKSから適切な鍵を取得
		key, ok := keySet.LookupKeyID(kid)
		if !ok {
			return nil, errors.New("指定されたkidに対応する鍵が見つかりません")
		}

		var rawKey interface{}
		if err := key.Raw(&rawKey); err != nil {
			return nil, errors.New("鍵の変換に失敗しました")
		}

		return rawKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("トークン検証エラー: %w", err)
	}

	// クレームのキャスト
	claims, ok := token.Claims.(*CognitoClaims)
	if !ok {
		return nil, errors.New("クレームのキャストに失敗しました")
	}

	// 発行者の確認
	issuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", r.config.Region, r.config.UserPoolID)
	if claims.Issuer != issuer {
		return nil, errors.New("発行者が一致しません")
	}

	// オーディエンスの確認（クライアントID）
	if !claims.VerifyAudience(r.config.ClientID, true) {
		return nil, errors.New("オーディエンスが一致しません")
	}

	// ユーザー情報からエンティティを作成
	userID, _ := valueobject.NewUserID(claims.Subject)

	// ロールの変換
	roles := make([]entity.UserRole, 0)
	for _, group := range claims.Groups {
		switch group {
		case "Administrators":
			roles = append(roles, entity.RoleAdmin)
		case "PremiumUsers":
			roles = append(roles, entity.RolePremium)
		default:
			// その他のグループは標準ユーザーとして扱う
			roles = append(roles, entity.RoleStandard)
		}
	}

	// ユーザーエンティティの作成
	user := entity.NewUser(
		userID,
		claims.Username,
		claims.Email,
		roles,
	)

	return user, nil
}

// CreateErrorResponse はエラーレスポンスを作成します
func (r *CognitoAuthRepository) CreateErrorResponse(statusCode int, message string) interface{} {
	body := map[string]string{"message": message}
	bodyJSON, _ := json.Marshal(body)

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(bodyJSON),
	}
}
