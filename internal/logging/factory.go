package logging

import (
	"os"
	"sync"
)

var (
	// globalLogger はアプリケーション全体で使用するグローバルロガー
	globalLogger Logger

	// once はグローバルロガーの初期化を一度だけ行うためのガード
	once sync.Once
)

// 環境変数名
const (
	EnvLogLevel     = "LOG_LEVEL"                // ログレベル設定用
	EnvEnv          = "ENVIRONMENT"              // 環境名設定用
	EnvServiceName  = "SERVICE_NAME"             // サービス名設定用
	EnvFunctionName = "AWS_LAMBDA_FUNCTION_NAME" // Lambda関数名
)

// InitLogging はアプリケーションのロギングシステムを初期化する
// main関数の最初に呼び出すべき
func InitLogging() {
	once.Do(func() {
		// 環境変数からログレベルを取得
		logLevelStr := os.Getenv(EnvLogLevel)
		var logLevel LogLevel
		switch logLevelStr {
		case "debug":
			logLevel = LevelDebug
		case "info":
			logLevel = LevelInfo
		case "warn":
			logLevel = LevelWarn
		case "error":
			logLevel = LevelError
		default:
			// デフォルトはINFOレベル
			logLevel = LevelInfo
		}

		// 環境を取得
		env := os.Getenv(EnvEnv)
		if env == "" {
			env = "dev" // デフォルトはdev環境
		}

		// サービス名を取得
		serviceName := os.Getenv(EnvServiceName)
		if serviceName == "" {
			serviceName = "CloudPix" // デフォルトサービス名
		}

		// Lambda関数名を取得
		functionName := os.Getenv(EnvFunctionName)
		if functionName == "" {
			functionName = "unknown-function" // 関数名が不明の場合
		}

		// グローバルロガーを初期化
		globalLogger = NewCloudWatchLogger(
			serviceName,
			functionName,
			WithMinimumLevel(logLevel),
			WithEnvironment(env),
		)
	})
}

// GetLogger は指定された名前のロガーを取得する
// 各コンポーネントはこの関数を使って専用のロガーを取得すべき
func GetLogger(component string) Logger {
	// まだ初期化されていない場合は初期化する
	if globalLogger == nil {
		InitLogging()
	}

	// コンポーネント名でロガーを返す
	return globalLogger.WithFunction(component)
}
