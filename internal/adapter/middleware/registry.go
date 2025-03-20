package middleware

import (
	"context"
	"log"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/session"
)

// HandlerFunc はLambdaハンドラー関数の型定義
type HandlerFunc func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

// Middleware はミドルウェア関数の型定義
type Middleware func(HandlerFunc) HandlerFunc

// MiddlewareRegistry はミドルウェアを管理するレジストリ
type MiddlewareRegistry struct {
	middlewares map[string]Middleware
	mu          sync.RWMutex
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
		middlewares: make(map[string]Middleware),
	}
}

// Register はミドルウェアをレジストリに登録する
func (r *MiddlewareRegistry) Register(name string, middleware Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middlewares[name] = middleware
	log.Printf("Registered middleware: %s", name)
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

// RegisterAuthMiddleware は認証ミドルウェアを登録する
func (r *MiddlewareRegistry) RegisterAuthMiddleware(name string, authMiddleware AuthMiddleware) {
	r.Register(name, func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			// 認証処理
			newCtx, userInfo, errResp, err := authMiddleware.Process(ctx, event)
			if err != nil {
				log.Printf("Authentication error: %v", err)
				return events.APIGatewayProxyResponse{StatusCode: 401}, err
			}

			// エラーレスポンスが設定されている場合（認証失敗）
			if errResp.StatusCode != 0 {
				return errResp, nil
			}

			// 認証済みコンテキストでハンドラー実行
			log.Printf("Authenticated user: %s", userInfo.UserID)
			return next(newCtx, event)
		}
	})
}

// RegisterMetricsMiddleware はメトリクスミドルウェアを登録する
func (r *MiddlewareRegistry) RegisterMetricsMiddleware(name string, metricsMiddleware *MetricsMiddleware) {
	r.Register(name, func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			// レスポンスとエラーを格納する変数
			var resp events.APIGatewayProxyResponse
			var err error

			// 処理時間計測を開始し、完了時に計測を終了する
			defer metricsMiddleware.StartTiming(ctx)(ctx, &resp, &err)

			// ハンドラー実行
			resp, err = next(ctx, event)
			return resp, err
		}
	})
}

// RegisterStandardMiddlewares は標準的なミドルウェアを一括で登録する
func (r *MiddlewareRegistry) RegisterStandardMiddlewares(sess *session.Session, cfg *MiddlewareConfig) {
	// 認証ミドルウェアの登録
	if cfg.AuthEnabled {
		authMiddleware := CreateDefaultAuthMiddleware(cfg.AWSRegion, cfg.UserPoolID, cfg.ClientID)
		r.RegisterAuthMiddleware("auth", authMiddleware)
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

	// ログミドルウェアの登録
	r.Register("logging", LoggingMiddleware)
}

// LoggingMiddleware はリクエスト/レスポンスのログを取るミドルウェア
func LoggingMiddleware(next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		log.Printf("Request: %s %s", event.HTTPMethod, event.Path)
		resp, err := next(ctx, event)
		log.Printf("Response: status=%d, error=%v", resp.StatusCode, err)
		return resp, err
	}
}
