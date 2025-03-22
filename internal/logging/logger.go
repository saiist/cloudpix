package logging

import (
	"cloudpix/internal/contextutil"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogLevel はログレベルを表す型
type LogLevel string

// 利用可能なログレベル
const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
	LevelFatal LogLevel = "fatal"
)

// Logger はプロジェクト全体で使用されるロガーインターフェース
type Logger interface {
	Debug(msg string, fields map[string]interface{})
	Info(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
	Error(err error, msg string, fields map[string]interface{})
	Fatal(err error, msg string, fields map[string]interface{})
	WithRequestID(requestID string) Logger
	WithUserID(userID string) Logger
	WithOperation(operation string) Logger
	WithFunction(function string) Logger
	WithContext(ctx context.Context) Logger
	WithField(key string, value interface{}) Logger
	IsLevelEnabled(level LogLevel) bool
}

// CloudWatchLogger はCloudWatch Logsに出力するロガー実装
type CloudWatchLogger struct {
	service    string                 // サービス名
	function   string                 // Lambda関数名
	requestID  string                 // リクエストID
	userID     string                 // ユーザーID
	operation  string                 // 実行操作名（UploadImage, ListImagesなど）
	defaultCtx map[string]interface{} // デフォルトコンテキスト情報
	minLevel   LogLevel               // 最小ログレベル
	output     *log.Logger            // 出力先
	env        string                 // 環境（dev, prod）
}

// LoggerOption はロガー作成時のオプション設定関数
type LoggerOption func(*CloudWatchLogger)

// WithMinimumLevel は最小ログレベルを設定するオプション
func WithMinimumLevel(level LogLevel) LoggerOption {
	return func(l *CloudWatchLogger) {
		l.minLevel = level
	}
}

// WithOutput は出力先を設定するオプション
func WithOutput(output *log.Logger) LoggerOption {
	return func(l *CloudWatchLogger) {
		l.output = output
	}
}

// WithEnvironment は環境を設定するオプション
func WithEnvironment(env string) LoggerOption {
	return func(l *CloudWatchLogger) {
		l.env = env
	}
}

// NewCloudWatchLogger は新しいCloudWatchLoggerを作成します
func NewCloudWatchLogger(service, function string, opts ...LoggerOption) *CloudWatchLogger {
	// デフォルト設定
	logger := &CloudWatchLogger{
		service:    service,
		function:   function,
		defaultCtx: make(map[string]interface{}),
		minLevel:   LevelDebug,
		output:     log.New(os.Stdout, "", 0), // 標準出力（CloudWatchにリダイレクトされる）
		env:        "dev",                     // デフォルト環境
	}

	// オプションを適用
	for _, opt := range opts {
		opt(logger)
	}

	return logger
}

// shouldLog は指定されたレベルのログを出力すべきかを判定
func (l *CloudWatchLogger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
		LevelFatal: 4,
	}

	return levels[level] >= levels[l.minLevel]
}

// log は実際のログエントリを作成して出力
func (l *CloudWatchLogger) log(level LogLevel, msg string, err error, fields map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}

	// タイムスタンプはUTC
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// ベースのログエントリを作成
	entry := LogEntry{
		Timestamp: timestamp,
		Level:     string(level),
		RequestID: l.requestID,
		Service:   l.service,
		Function:  l.function,
		Operation: l.operation,
		UserID:    l.userID,
		Message:   msg,
		Env:       l.env,
	}

	// フィールドをマージ
	contextMap := mergeContexts(l.defaultCtx, fields)
	if len(contextMap) > 0 {
		entry.Context = contextMap
	}

	// エラー情報を追加
	if err != nil {
		entry.ErrorType = extractErrorType(err)
		entry.ErrorMsg = err.Error()

		// スタックトレース情報を追加（開発環境またはエラーレベル以上の場合）
		if l.env == "dev" || level == LevelError || level == LevelFatal {
			entry.StackTrace = captureStackTrace(2) // 呼び出し元関数のスタックから
		}
	}

	// JSONにシリアル化
	jsonBytes, jsonErr := json.Marshal(entry)
	if jsonErr != nil {
		// マーシャリングエラーの場合は最低限の情報をプレーンテキストで出力
		l.output.Printf("ERROR: Failed to marshal log entry: %v. Original message: %s", jsonErr, msg)
		return
	}

	// ログ出力
	l.output.Println(string(jsonBytes))

	// Fatalレベルの場合はプログラムを終了
	if level == LevelFatal {
		os.Exit(1)
	}
}

// Debug はDEBUGレベルのログを出力します
func (l *CloudWatchLogger) Debug(msg string, fields map[string]interface{}) {
	l.log(LevelDebug, msg, nil, fields)
}

// Info はINFOレベルのログを出力します
func (l *CloudWatchLogger) Info(msg string, fields map[string]interface{}) {
	l.log(LevelInfo, msg, nil, fields)
}

// Warn はWARNレベルのログを出力します
func (l *CloudWatchLogger) Warn(msg string, fields map[string]interface{}) {
	l.log(LevelWarn, msg, nil, fields)
}

// Error はERRORレベルのログを出力します
func (l *CloudWatchLogger) Error(err error, msg string, fields map[string]interface{}) {
	l.log(LevelError, msg, err, fields)
}

// Fatal はFATALレベルのログを出力し、プログラムを終了します
func (l *CloudWatchLogger) Fatal(err error, msg string, fields map[string]interface{}) {
	l.log(LevelFatal, msg, err, fields)
}

// WithRequestID はリクエストIDを設定した新しいロガーを返します
func (l *CloudWatchLogger) WithRequestID(requestID string) Logger {
	newLogger := *l
	newLogger.requestID = requestID
	return &newLogger
}

// WithUserID はユーザーIDを設定した新しいロガーを返します
func (l *CloudWatchLogger) WithUserID(userID string) Logger {
	newLogger := *l
	newLogger.userID = userID
	return &newLogger
}

// WithOperation は操作名を設定した新しいロガーを返します
func (l *CloudWatchLogger) WithOperation(operation string) Logger {
	newLogger := *l
	newLogger.operation = operation
	return &newLogger
}

// WithFunction は関数名を設定した新しいロガーを返します
func (l *CloudWatchLogger) WithFunction(function string) Logger {
	newLogger := *l
	newLogger.function = function
	return &newLogger
}

// WithContext はコンテキストから情報を抽出した新しいロガーを返します
func (l *CloudWatchLogger) WithContext(ctx context.Context) Logger {
	newLogger := *l

	// コンテキストからリクエストIDを抽出（例: x-ray-trace-id）
	if traceID := ctx.Value("X-Amzn-Trace-Id"); traceID != nil {
		if traceIDStr, ok := traceID.(string); ok {
			newLogger.requestID = traceIDStr
		}
	}

	// ユーザー情報をコンテキストから抽出
	userInfo, ok := contextutil.GetUserInfo(ctx)
	if ok && userInfo != nil {
		newLogger.userID = userInfo.UserID

		// ユーザー関連の追加情報を defaultCtx に追加
		newCtx := mergeContexts(newLogger.defaultCtx, nil)
		newCtx["userGroups"] = userInfo.Groups
		if userInfo.IsAdmin {
			newCtx["isAdmin"] = true
		}
		if userInfo.IsPremium {
			newCtx["isPremium"] = true
		}
		newLogger.defaultCtx = newCtx
	}

	return &newLogger
}

// WithField は単一のフィールドを追加した新しいロガーを返します
func (l *CloudWatchLogger) WithField(key string, value interface{}) Logger {
	newLogger := *l
	newCtx := mergeContexts(newLogger.defaultCtx, nil)
	newCtx[key] = value
	newLogger.defaultCtx = newCtx
	return &newLogger
}

// IsLevelEnabled は指定されたレベルのログが有効か判定します
func (l *CloudWatchLogger) IsLevelEnabled(level LogLevel) bool {
	levels := map[LogLevel]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
		LevelFatal: 4,
	}

	return levels[level] >= levels[l.minLevel]
}

// mergeContexts は複数のコンテキストマップをマージします
func mergeContexts(base map[string]interface{}, additional map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// ベースをコピー
	for k, v := range base {
		result[k] = v
	}

	// 追加情報をマージ
	for k, v := range additional {
		result[k] = v
	}

	return result
}

// extractErrorType はエラーの型名を抽出します
func extractErrorType(err error) string {
	if err == nil {
		return ""
	}

	// 構造化エラーの場合はそのタイプを使用
	if structErr, ok := err.(*StructuredError); ok {
		return string(structErr.Type)
	}

	// それ以外は型名を抽出
	errorType := fmt.Sprintf("%T", err)

	// パッケージ名を除去して型名のみを取得
	if parts := strings.Split(errorType, "."); len(parts) > 1 {
		return parts[len(parts)-1]
	}

	return errorType
}

// captureStackTrace はスタックトレースを文字列として取得します
func captureStackTrace(skip int) string {
	buffer := make([]byte, 2048)
	n := runtime.Stack(buffer, false)
	stack := string(buffer[:n])

	// スキップする行数分を除去
	lines := strings.Split(stack, "\n")
	if len(lines) > skip*2 {
		stack = strings.Join(lines[skip*2:], "\n")
	}

	return stack
}

// NoopLogger は何も出力しないロガー実装（テスト用）
type NoopLogger struct{}

func NewNoopLogger() *NoopLogger {
	return &NoopLogger{}
}

func (l *NoopLogger) Debug(msg string, fields map[string]interface{})            {}
func (l *NoopLogger) Info(msg string, fields map[string]interface{})             {}
func (l *NoopLogger) Warn(msg string, fields map[string]interface{})             {}
func (l *NoopLogger) Error(err error, msg string, fields map[string]interface{}) {}
func (l *NoopLogger) Fatal(err error, msg string, fields map[string]interface{}) {}
func (l *NoopLogger) WithRequestID(requestID string) Logger                      { return l }
func (l *NoopLogger) WithUserID(userID string) Logger                            { return l }
func (l *NoopLogger) WithOperation(operation string) Logger                      { return l }
func (l *NoopLogger) WithFunction(function string) Logger                        { return l }
func (l *NoopLogger) WithContext(ctx context.Context) Logger                     { return l }
func (l *NoopLogger) WithField(key string, value interface{}) Logger             { return l }
func (l *NoopLogger) IsLevelEnabled(level LogLevel) bool                         { return false }
