package logging

// LoggingMiddlewareConfig はロギングミドルウェアの設定
type LoggingMiddlewareConfig struct {
	Logger            Logger   // 使用するロガー
	LogRequestBody    bool     // リクエストボディをログに含めるか
	LogResponseBody   bool     // レスポンスボディをログに含めるか
	MaxBodyLogLength  int      // ボディログの最大長（バイト）
	LogRequestHeaders bool     // リクエストヘッダーをログに含めるか
	SensitiveHeaders  []string // 機密ヘッダー（値をマスク）
}

// NewLoggingMiddlewareConfig はデフォルト設定を持つ設定オブジェクトを作成
func NewLoggingMiddlewareConfig(logger Logger) *LoggingMiddlewareConfig {
	return &LoggingMiddlewareConfig{
		Logger:            logger,
		LogRequestBody:    false, // 標準では含めない（セキュリティ上の理由）
		LogResponseBody:   false, // 標準では含めない（パフォーマンス上の理由）
		MaxBodyLogLength:  1024,  // 1KB
		LogRequestHeaders: true,
		SensitiveHeaders:  []string{"Authorization", "X-Api-Key", "Cookie"},
	}
}
