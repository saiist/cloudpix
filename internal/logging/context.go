package logging

import "context"

// ContextKey はコンテキストキーの型
type ContextKey string

const (
	// LoggerKey はコンテキスト内のロガーのキー
	LoggerKey ContextKey = "logger"
)

// WithLogger はコンテキストにロガーを追加する
func WithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, logger)
}

// FromContext はコンテキストからロガーを取得する
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(LoggerKey).(Logger); ok {
		return logger
	}

	// コンテキストにロガーがない場合は新しく作成
	return GetLogger("default")
}
