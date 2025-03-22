package middleware

import (
	"cloudpix/internal/application/authmanagement/usecase"
	"cloudpix/internal/contextutil"
	"cloudpix/internal/logging"
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/session"
)

// HandlerFunc はLambdaハンドラー関数の型定義
type HandlerFunc func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

// Middleware はミドルウェア関数の型定義
type Middleware func(HandlerFunc) HandlerFunc

// CleanupableMiddleware はリソース解放が必要なミドルウェアのためのインターフェース
type CleanupableMiddleware interface {
	Cleanup() error
}

// MiddlewareRegistry はミドルウェアを管理するレジストリ
type MiddlewareRegistry struct {
	middlewares        map[string]Middleware
	cleanupMiddlewares map[string]CleanupableMiddleware
	config             *MiddlewareConfig
	mu                 sync.RWMutex
}

// グローバルインスタンス
var (
	globalRegistry *MiddlewareRegistry
	once           sync.Once
)

// GetRegistry はグローバルミドルウェアレジストリのシングルトンインスタンスを返す
func GetRegistry() *MiddlewareRegistry {
	once.Do(func() {
		globalRegistry = NewMiddlewareRegistry()
	})
	return globalRegistry
}

// NewMiddlewareRegistry は新しいミドルウェアレジストリを作成する
func NewMiddlewareRegistry() *MiddlewareRegistry {
	return &MiddlewareRegistry{
		middlewares:        make(map[string]Middleware),
		cleanupMiddlewares: make(map[string]CleanupableMiddleware),
	}
}

// SetConfig はミドルウェアレジストリの設定を更新する
func (r *MiddlewareRegistry) SetConfig(config *MiddlewareConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.config = config
}

// Register はミドルウェアをレジストリに登録する
func (r *MiddlewareRegistry) Register(name string, middleware Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middlewares[name] = middleware
	log.Printf("Registered middleware: %s", name)
}

// RegisterCleanupableMiddleware はクリーンアップ可能なミドルウェアを登録する
func (r *MiddlewareRegistry) RegisterCleanupableMiddleware(name string, middleware Middleware, cleanupable CleanupableMiddleware) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.middlewares[name] = middleware
	r.cleanupMiddlewares[name] = cleanupable
	log.Printf("Registered cleanupable middleware: %s", name)
}

// Get は指定された名前のミドルウェアを取得する
func (r *MiddlewareRegistry) Get(name string) (Middleware, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	middleware, exists := r.middlewares[name]
	return middleware, exists
}

// BuildChain は指定されたミドルウェア名の配列に基づいてミドルウェアチェーンを構築する
func (r *MiddlewareRegistry) BuildChain(middlewareNames []string) *Chain {
	chain := NewChain()

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, name := range middlewareNames {
		if middleware, exists := r.middlewares[name]; exists {
			chain.Use(middleware)
		} else {
			log.Printf("Warning: Middleware '%s' not found in registry", name)
		}
	}

	return chain
}

// Cleanup はすべてのクリーンアップ可能なミドルウェアのリソースを解放する
func (r *MiddlewareRegistry) Cleanup() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, middleware := range r.cleanupMiddlewares {
		if err := middleware.Cleanup(); err != nil {
			log.Printf("Error cleaning up middleware '%s': %v", name, err)
		} else {
			log.Printf("Successfully cleaned up middleware: %s", name)
		}
	}
}

// RegisterLoggingMiddleware はロギングミドルウェアを登録する
func (r *MiddlewareRegistry) RegisterLoggingMiddleware(name string, logger logging.Logger) {
	// ロギング設定
	loggingConfig := logging.NewLoggingMiddlewareConfig(logger)

	// 環境設定に基づいて調整
	if r.config != nil {
		if r.config.DetailedRequestLog {
			loggingConfig.LogRequestBody = true
			loggingConfig.LogRequestHeaders = true
		}
		if r.config.DetailedResponseLog {
			loggingConfig.LogResponseBody = true
		}
		if r.config.IncludeHeaders {
			loggingConfig.LogRequestHeaders = true
		}
		if !r.config.IncludeBody {
			loggingConfig.LogRequestBody = false
			loggingConfig.LogResponseBody = false
		}

		// 機密ヘッダーの設定
		if len(r.config.SensitiveHeaderNames) > 0 {
			loggingConfig.SensitiveHeaders = r.config.SensitiveHeaderNames
		}

		// 最大ボディログ長の設定
		if r.config.MaxBodyLogLength > 0 {
			loggingConfig.MaxBodyLogLength = r.config.MaxBodyLogLength
		}
	}

	loggingMiddleware := LoggingMiddleware(loggingConfig)

	r.Register(name, loggingMiddleware)
}

// RegisterAuthMiddleware は認証ミドルウェアを登録する
func (r *MiddlewareRegistry) RegisterAuthMiddleware(name string, authMiddleware *AuthMiddleware) {
	r.Register(name, func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			// コンテキストからロガーを取得
			logger := logging.FromContext(ctx)

			// 認証処理
			newCtx, errResp, err := authMiddleware.Process(ctx, event)
			if err != nil {
				logger.Error(err, "Authentication error", nil)
				return events.APIGatewayProxyResponse{StatusCode: 401}, err
			}

			// エラーレスポンスが設定されている場合（認証失敗）
			if errResp.StatusCode != 0 {
				logger.Warn("Authentication failed", map[string]interface{}{
					"statusCode": errResp.StatusCode,
				})
				return errResp, nil
			}

			// 認証済みコンテキストでハンドラー実行
			user, _ := contextutil.GetUserInfo(ctx)
			if user != nil {
				// ユーザー情報をロガーに追加
				logger = logger.WithUserID(user.ID.String())

				// ユーザー情報をコンテキストに追加
				newCtx = logging.WithLogger(newCtx, logger)

				logger.Info("User authenticated", map[string]interface{}{
					"username":  user.Username,
					"rols":      user.Roles,
					"isAdmin":   user.IsAdmin,
					"isPremium": user.IsPremium,
				})
			}

			return next(newCtx, event)
		}
	})
}

// RegisterMetricsMiddleware はメトリクスミドルウェアを登録する
func (r *MiddlewareRegistry) RegisterMetricsMiddleware(name string, metricsMiddleware *MetricsMiddleware) {
	middlewareFunc := func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			// コンテキストからロガーを取得
			logger := logging.FromContext(ctx)

			// レスポンスとエラーを格納する変数
			var resp events.APIGatewayProxyResponse
			var err error

			// 処理時間計測を開始し、完了時に計測を終了する
			defer metricsMiddleware.StartTiming(ctx)(ctx, &resp, &err)

			// メトリクス収集開始をログに記録
			logger.Debug("Starting metrics collection", nil)

			// ハンドラー実行
			resp, err = next(ctx, event)

			// メトリクス収集結果をログに記録（デバッグモードの場合）
			if logger.IsLevelEnabled(logging.LevelDebug) {
				logger.Debug("Metrics collected", map[string]interface{}{
					"operation": metricsMiddleware.operationName,
					"function":  metricsMiddleware.functionName,
				})
			}

			return resp, err
		}
	}

	// クリーンアップ可能なミドルウェアとして登録
	r.RegisterCleanupableMiddleware(name, middlewareFunc, metricsMiddleware)
}

// RegisterStandardMiddlewares は標準的なミドルウェアを一括で登録する
func (r *MiddlewareRegistry) RegisterStandardMiddlewares(
	sess *session.Session,
	cfg *MiddlewareConfig,
	authUsecase *usecase.AuthUsecase,
	logger logging.Logger,
) {
	// 設定を保存
	r.SetConfig(cfg)

	// ログミドルウェアの登録（常に最初に実行されるよう登録）
	if cfg.LoggingEnabled {
		r.RegisterLoggingMiddleware("logging", logger)
	}

	// メトリクスミドルウェアの登録
	if cfg.MetricsEnabled {
		metricsMiddleware := NewMetricsMiddleware(
			sess,
			cfg.ServiceName,
			cfg.OperationName,
			cfg.FunctionName,
			cfg.MetricsConfig,
		)
		r.RegisterMetricsMiddleware("metrics", metricsMiddleware)
	}

	// 認証ミドルウェアの登録
	if cfg.AuthEnabled {
		authMiddleware := NewAuthMiddleware(authUsecase, logger)
		r.RegisterAuthMiddleware("auth", authMiddleware)
	}

	// 追加のミドルウェアを登録
	for _, name := range cfg.AdditionalMiddlewares {
		logger.Info(fmt.Sprintf("Registered additional middleware: %s", name), nil)
	}
}

// Count はレジストリに登録されているミドルウェアの数を返す
func (r *MiddlewareRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.middlewares)
}
