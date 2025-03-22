package middleware

import (
	"cloudpix/internal/infrastructure/auth"
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// CreateDefaultAuthMiddleware は環境変数から標準認証ミドルウェアを作成する
func CreateDefaultAuthMiddleware(region, userPoolID, clientID string) AuthMiddleware {
	authRepo := auth.NewCognitoAuthRepository(region, userPoolID, clientID)
	return NewCognitoAuthMiddleware(authRepo)
}

// HandlerFactory はLambdaハンドラーを生成するためのファクトリ
type HandlerFactory struct {
	config     *MiddlewareConfig
	registry   *MiddlewareRegistry
	awsSession *session.Session
}

// NewHandlerFactory は新しいハンドラーファクトリを作成する
func NewHandlerFactory(config *MiddlewareConfig) *HandlerFactory {
	return &HandlerFactory{
		config:   config,
		registry: GetRegistry(),
	}
}

// WithAWSSession はファクトリにAWSセッションを設定する
func (f *HandlerFactory) WithAWSSession(sess *session.Session) *HandlerFactory {
	f.awsSession = sess
	return f
}

// getOrCreateAWSSession はAWSセッションを取得または作成する
func (f *HandlerFactory) getOrCreateAWSSession() *session.Session {
	if f.awsSession != nil {
		return f.awsSession
	}

	// デフォルトのAWSセッションを作成
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(f.config.AWSRegion),
	})
	if err != nil {
		log.Printf("Error creating AWS session: %v", err)
		return nil
	}

	f.awsSession = sess
	return sess
}

// RegisterMiddlewares は設定に基づいてミドルウェアを登録する
func (f *HandlerFactory) RegisterMiddlewares() {
	sess := f.getOrCreateAWSSession()
	if sess == nil {
		log.Printf("Warning: Failed to create AWS session, some middlewares may not work correctly")
	}

	f.registry.RegisterStandardMiddlewares(sess, f.config)
}

// WrapHandler はハンドラーにミドルウェアを適用する
func (f *HandlerFactory) WrapHandler(handler HandlerFunc) HandlerFunc {
	// まだミドルウェアが登録されていない場合は登録する
	if f.registry.Count() == 0 {
		f.RegisterMiddlewares()
	}

	// ミドルウェアチェーンを構築
	middlewareNames := f.config.GetDefaultMiddlewareNames()
	chain := f.registry.BuildChain(middlewareNames)

	// ハンドラーを包む
	return chain.Then(handler)
}

// WrapAPIGatewayHandler はAPI Gateway Lambdaハンドラーにミドルウェアを適用する
func (f *HandlerFactory) WrapAPIGatewayHandler(handler HandlerFunc) func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return f.WrapHandler(handler)
}

// WrapS3EventHandler はS3イベントハンドラーにミドルウェアを適用する
// S3イベントハンドラーは独自の型を持つため、別の関数として実装
func (f *HandlerFactory) WrapS3EventHandler(handler func(context.Context, events.S3Event) error) func(context.Context, events.S3Event) error {
	// S3イベント用のメトリクスミドルウェアを使用
	if f.config.MetricsEnabled {
		sess := f.getOrCreateAWSSession()
		if sess != nil {
			metricsMiddleware := NewMetricsMiddleware(
				sess,
				f.config.ServiceName,
				f.config.OperationName,
				f.config.FunctionName,
				f.config.GetMetricsConfig(),
			)

			return withMetricsForThumbnail(metricsMiddleware, handler)
		}
	}

	// ミドルウェアを適用できない場合は元のハンドラーをそのまま返す
	return handler
}

// withMetricsForThumbnail はメトリクス収集ミドルウェアをS3イベントハンドラーに適用する
func withMetricsForThumbnail(metricsMiddleware *MetricsMiddleware,
	handler func(ctx context.Context, s3Event events.S3Event) error) func(ctx context.Context, s3Event events.S3Event) error {

	return func(ctx context.Context, s3Event events.S3Event) error {
		// エラーを格納する変数
		var err error

		// 処理時間計測を開始し、完了時に計測を終了する
		defer metricsMiddleware.StartTimingForThumbnail(ctx, s3Event)(ctx, &err)

		// ハンドラー実行
		err = handler(ctx, s3Event)
		return err
	}
}
