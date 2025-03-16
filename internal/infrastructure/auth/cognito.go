package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"cloudpix/internal/domain/model"
	"cloudpix/internal/domain/repository"

	"github.com/aws/aws-lambda-go/events"
	"github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/jwk"
)

// CognitoAuthConfig は認証設定を保持する構造体
type CognitoAuthConfig struct {
	Region      string
	UserPoolID  string
	ClientID    string
	JwksURI     string
	keySet      jwk.Set
	lastUpdated time.Time
}

// CognitoAuthClaims はJWTトークンのクレーム情報
type CognitoAuthClaims struct {
	jwt.RegisteredClaims
	Username string   `json:"cognito:username"`
	Groups   []string `json:"cognito:groups"`
	Email    string   `json:"email"`
}

// CognitoAuthRepository はCognito認証リポジトリの実装
type CognitoAuthRepository struct {
	config *CognitoAuthConfig
}

// NewCognitoAuthRepository は新しいCognito認証リポジトリを作成する
func NewCognitoAuthRepository(region, userPoolID, clientID string) repository.AuthRepository {
	jwksURI := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", region, userPoolID)

	config := &CognitoAuthConfig{
		Region:     region,
		UserPoolID: userPoolID,
		ClientID:   clientID,
		JwksURI:    jwksURI,
	}

	return &CognitoAuthRepository{
		config: config,
	}
}

// getKeySet はJWKSから鍵セットを取得する
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

// VerifyToken はCognitoのJWTトークンを検証する
func (r *CognitoAuthRepository) verifyJWTToken(ctx context.Context, tokenString string) (*jwt.Token, *CognitoAuthClaims, error) {
	keySet, err := r.getKeySet(ctx)
	if err != nil {
		return nil, nil, err
	}

	// トークンの解析と検証
	token, err := jwt.ParseWithClaims(tokenString, &CognitoAuthClaims{}, func(token *jwt.Token) (interface{}, error) {
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
		keys, ok := keySet.LookupKeyID(kid)
		if !ok {
			return nil, errors.New("指定されたkidに対応する鍵が見つかりません")
		}

		var rawKey interface{}
		if err := keys.Raw(&rawKey); err != nil {
			return nil, errors.New("鍵の変換に失敗しました")
		}

		return rawKey, nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("トークン検証エラー: %w", err)
	}

	// クレームのキャスト
	claims, ok := token.Claims.(*CognitoAuthClaims)
	if !ok {
		return nil, nil, errors.New("クレームのキャストに失敗しました")
	}

	// 発行者の確認
	issuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", r.config.Region, r.config.UserPoolID)
	if claims.Issuer != issuer {
		return nil, nil, errors.New("発行者が一致しません")
	}

	// オーディエンスの確認（クライアントID）
	if !claims.VerifyAudience(r.config.ClientID, true) {
		return nil, nil, errors.New("オーディエンスが一致しません")
	}

	return token, claims, nil
}

// VerifyToken はトークンを検証し、ユーザー情報を返す
func (r *CognitoAuthRepository) VerifyToken(ctx context.Context, tokenString string) (*model.UserInfo, error) {
	_, claims, err := r.verifyJWTToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	// ユーザー情報を作成
	userInfo := &model.UserInfo{
		UserID:   claims.Subject,
		Username: claims.Username,
		Email:    claims.Email,
		Groups:   claims.Groups,
	}

	// 管理者かどうかの判定
	for _, group := range claims.Groups {
		if group == "Administrators" {
			userInfo.IsAdmin = true
		}
		if group == "PremiumUsers" || group == "Administrators" {
			userInfo.IsPremium = true
		}
	}

	return userInfo, nil
}

// GetUserInfoFromHeader はAuthorizationヘッダーからユーザー情報を取得する
func (r *CognitoAuthRepository) GetUserInfoFromHeader(ctx context.Context, authHeader string) (*model.UserInfo, error) {
	// Bearer トークンの形式をチェック
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, errors.New("無効な認証フォーマットです")
	}

	// トークンを取得
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// トークンの検証
	return r.VerifyToken(ctx, tokenString)
}

// CheckPermission は指定されたユーザーが操作を実行できるかチェックする
func (r *CognitoAuthRepository) CheckPermission(userInfo *model.UserInfo, ownerID string, isPublic bool) bool {
	// 管理者は常に許可
	if userInfo.IsAdmin {
		return true
	}

	// 自分自身のリソースへのアクセスは許可
	if userInfo.UserID == ownerID {
		return true
	}

	// 公開リソースへのアクセスは許可
	if isPublic {
		return true
	}

	return false
}

// CreateErrorResponse はエラーレスポンスを作成する
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
