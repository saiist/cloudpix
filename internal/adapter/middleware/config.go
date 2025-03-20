package middleware

import (
	"cloudpix/internal/infrastructure/metrics"
	"time"
)

// MiddlewareConfig はミドルウェアの設定を保持する構造体
type MiddlewareConfig struct {
	// 共通設定
	AWSRegion string

	// 認証ミドルウェア設定
	AuthEnabled bool
	UserPoolID  string
	ClientID    string

	// メトリクスミドルウェア設定
	MetricsEnabled   bool
	ServiceName      string
	OperationName    string
	FunctionName     string
	MetricsConfig    *metrics.MetricsConfig
	MetricsNamespace string
	FlushInterval    time.Duration
	BatchSize        int

	// ロギングミドルウェア設定
	LoggingEnabled       bool
	DetailedRequestLog   bool
	DetailedResponseLog  bool
	IncludeHeaders       bool
	IncludeQueryParams   bool
	IncludeBody          bool
	MaxBodyLogLength     int
	SensitiveHeaderNames []string

	// 追加のミドルウェア名リスト
	AdditionalMiddlewares []string
}

// NewDefaultMiddlewareConfig はデフォルト設定のミドルウェア設定を作成する
func NewDefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		// デフォルトでは認証とメトリクスを有効にする
		AuthEnabled:    true,
		MetricsEnabled: true,
		LoggingEnabled: true,

		// ロギング設定
		DetailedRequestLog:  false,
		DetailedResponseLog: false,
		IncludeHeaders:      false,
		IncludeQueryParams:  true,
		IncludeBody:         false,
		MaxBodyLogLength:    1000,
		SensitiveHeaderNames: []string{
			"Authorization", "X-Api-Key", "Cookie", "X-Amz-Security-Token",
		},

		// メトリクス設定
		MetricsNamespace: "CloudPix/Lambda",
		BatchSize:        20,
		FlushInterval:    time.Minute,

		// デフォルトミドルウェアの順序
		AdditionalMiddlewares: []string{},
	}
}

// GetDefaultMiddlewareNames は標準ミドルウェア名のリストを返す
func (c *MiddlewareConfig) GetDefaultMiddlewareNames() []string {
	var middlewares []string

	// ログミドルウェアは最初に実行
	if c.LoggingEnabled {
		middlewares = append(middlewares, "logging")
	}

	// 次にメトリクスミドルウェア
	if c.MetricsEnabled {
		middlewares = append(middlewares, "metrics")
	}

	// 最後に認証ミドルウェア（認証成功後にハンドラー処理）
	if c.AuthEnabled {
		middlewares = append(middlewares, "auth")
	}

	// 追加のミドルウェアを適用
	middlewares = append(middlewares, c.AdditionalMiddlewares...)

	return middlewares
}

// GetMetricsConfig はメトリクス設定を返す
func (c *MiddlewareConfig) GetMetricsConfig() *metrics.MetricsConfig {
	if c.MetricsConfig != nil {
		return c.MetricsConfig
	}

	return &metrics.MetricsConfig{
		BatchSize:       c.BatchSize,
		FlushInterval:   c.FlushInterval,
		DetailedMetrics: c.DetailedRequestLog,
		Namespace:       c.MetricsNamespace,
	}
}
