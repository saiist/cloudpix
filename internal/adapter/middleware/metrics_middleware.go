package middleware

import (
	"cloudpix/internal/infrastructure/metrics"
	"context"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

// MetricsMiddleware はメトリクス収集用ミドルウェア
type MetricsMiddleware struct {
	serviceName    string
	operationName  string
	functionName   string
	metricsService metrics.MetricsService
}

// NewMetricsMiddleware は新しいメトリクス収集ミドルウェアを作成する
func NewMetricsMiddleware(sess *session.Session, serviceName, operationName, functionName string, config *metrics.MetricsConfig) *MetricsMiddleware {
	// 設定の初期化
	cfg := metrics.DefaultMetricsConfig
	if config != nil {
		cfg = *config
	}

	// デフォルトセッションがない場合は作成
	if sess == nil {
		var err error
		sess, err = session.NewSession()
		if err != nil {
			log.Printf("Error creating AWS session for metrics: %v", err)
			return &MetricsMiddleware{
				serviceName:   serviceName,
				operationName: operationName,
				functionName:  functionName,
			}
		}
	}

	// メトリクスサービスを作成
	metricsService := metrics.NewCloudWatchMetricsService(sess, cfg)

	return &MetricsMiddleware{
		serviceName:    serviceName,
		operationName:  operationName,
		functionName:   functionName,
		metricsService: metricsService,
	}
}

// 標準ディメンションを作成
func (m *MetricsMiddleware) createStandardDimensions() []*cloudwatch.Dimension {
	return []*cloudwatch.Dimension{
		{
			Name:  aws.String("Service"),
			Value: aws.String(m.serviceName),
		},
		{
			Name:  aws.String("Operation"),
			Value: aws.String(m.operationName),
		},
		{
			Name:  aws.String("FunctionName"),
			Value: aws.String(m.functionName),
		},
	}
}

// 特定ユーザーのディメンションを作成
func (m *MetricsMiddleware) createUserDimensions(userID string) []*cloudwatch.Dimension {
	return []*cloudwatch.Dimension{
		{
			Name:  aws.String("Service"),
			Value: aws.String(m.serviceName),
		},
		{
			Name:  aws.String("UserID"),
			Value: aws.String(userID),
		},
	}
}

// StartTiming は処理時間計測を開始し、終了関数を返す
func (m *MetricsMiddleware) StartTiming(ctx context.Context) func(context.Context, interface{}, *error) {
	// 処理開始時間を記録
	startTime := time.Now()

	// メトリクスサービスがない場合は空関数を返す
	if m.metricsService == nil {
		return func(ctx context.Context, result interface{}, err *error) {}
	}

	// 標準ディメンション
	dimensions := m.createStandardDimensions()

	// 終了関数を返す
	return func(ctx context.Context, result interface{}, err *error) {
		// 処理時間を計算
		duration := time.Since(startTime)

		// 処理時間メトリクスを追加
		m.metricsService.AddMetric(ctx, "Duration", float64(duration.Milliseconds()), dimensions)

		// 成功/失敗メトリクスの追加
		if err != nil && *err != nil {
			// エラーメトリクス
			m.metricsService.AddMetric(ctx, "Errors", 1.0, dimensions)
		} else {
			// 成功メトリクス
			m.metricsService.AddMetric(ctx, "Successful", 1.0, dimensions)
		}

		// レスポンスのステータスコードに基づくメトリクス（APIGatewayProxyResponseの場合）
		if resp, ok := result.(*events.APIGatewayProxyResponse); ok && resp != nil {
			statusCodeCategory := "2xx"
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				statusCodeCategory = "4xx"
				m.metricsService.AddMetric(ctx, "ClientErrors", 1.0, dimensions)
			} else if resp.StatusCode >= 500 {
				statusCodeCategory = "5xx"
				m.metricsService.AddMetric(ctx, "ServerErrors", 1.0, dimensions)
			}

			// HTTPステータスコード別メトリクス
			m.metricsService.AddMetric(ctx, statusCodeCategory, 1.0, dimensions)
		}

		// ユーザー情報があればユーザー別メトリクスを追加
		userInfo, _ := GetUserInfo(ctx)
		if userInfo != nil {
			// ユーザー別メトリクス
			userDimensions := m.createUserDimensions(userInfo.UserID)
			m.metricsService.AddMetric(ctx, "UserRequests", 1.0, userDimensions)
		}
	}
}

// StartTimingForThumbnail はサムネイル処理用の処理時間計測を開始し、終了関数を返す
func (m *MetricsMiddleware) StartTimingForThumbnail(ctx context.Context, s3Event events.S3Event) func(context.Context, *error) {
	// 処理開始時間を記録
	startTime := time.Now()

	// メトリクスサービスがない場合は空関数を返す
	if m.metricsService == nil {
		return func(ctx context.Context, err *error) {}
	}

	// 標準ディメンション
	dimensions := m.createStandardDimensions()

	// 処理する画像数
	imageCount := len(s3Event.Records)

	// イベント数メトリクスを追加
	m.metricsService.AddMetric(ctx, "Invocations", 1.0, dimensions)

	// 処理する画像数メトリクス
	m.metricsService.AddMetric(ctx, "ProcessedImagesCount", float64(imageCount), dimensions)

	// 終了関数を返す
	return func(ctx context.Context, err *error) {
		// 処理時間を計算
		duration := time.Since(startTime)

		// 処理時間メトリクスを追加
		m.metricsService.AddMetric(ctx, "ProcessingTime", float64(duration.Milliseconds()), dimensions)

		// エラーがあれば記録
		if err != nil && *err != nil {
			m.metricsService.AddMetric(ctx, "Errors", 1.0, dimensions)
		} else {
			m.metricsService.AddMetric(ctx, "Successful", 1.0, dimensions)
		}

		// 平均処理時間（1画像あたり）
		if imageCount > 0 {
			m.metricsService.AddMetric(ctx, "AverageImageProcessingTime",
				float64(duration.Milliseconds())/float64(imageCount), dimensions)
		}
	}
}

// FlushMetrics はバッファのメトリクスを強制的に送信する
func (m *MetricsMiddleware) FlushMetrics(ctx context.Context) error {
	if m.metricsService == nil {
		return nil
	}
	return m.metricsService.Flush(ctx)
}

// Cleanup はリソース解放を行う
// CleanupableMiddlewareインターフェースに対応させる
func (m *MetricsMiddleware) Cleanup() error {
	if m.metricsService != nil {
		if closer, ok := m.metricsService.(interface{ Close() error }); ok {
			return closer.Close()
		}
	}
	return nil
}
