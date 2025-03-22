package valueobject

import (
	"errors"
	"strings"
)

// TokenType は認証トークンの種類を表す列挙型
type TokenType string

const (
	// BearerToken はBearer認証トークン
	BearerToken TokenType = "Bearer"
	// APIKeyToken はAPI Key認証トークン
	APIKeyToken TokenType = "ApiKey"
)

// Credentials は認証情報を表す値オブジェクト
type Credentials struct {
	tokenType TokenType
	token     string
}

// NewBearerTokenCredentials はBearer認証トークンを作成します
func NewBearerTokenCredentials(token string) (Credentials, error) {
	if strings.TrimSpace(token) == "" {
		return Credentials{}, errors.New("認証トークンは空にできません")
	}
	return Credentials{
		tokenType: BearerToken,
		token:     token,
	}, nil
}

// NewAPIKeyCredentials はAPI Key認証トークンを作成します
func NewAPIKeyCredentials(apiKey string) (Credentials, error) {
	if strings.TrimSpace(apiKey) == "" {
		return Credentials{}, errors.New("APIキーは空にできません")
	}
	return Credentials{
		tokenType: APIKeyToken,
		token:     apiKey,
	}, nil
}

// TokenType はトークンタイプを返します
func (c Credentials) TokenType() TokenType {
	return c.tokenType
}

// Token はトークン文字列を返します
func (c Credentials) Token() string {
	return c.token
}

// GetAuthorizationHeader は認証ヘッダー形式を返します
func (c Credentials) GetAuthorizationHeader() string {
	switch c.tokenType {
	case BearerToken:
		return "Bearer " + c.token
	case APIKeyToken:
		return "ApiKey " + c.token
	default:
		return c.token
	}
}
