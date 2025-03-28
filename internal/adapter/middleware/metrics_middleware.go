package middleware

import (
	"cloudpix/internal/contextutil"
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

// 共通の時間計測基本処理を実装
func (m *MetricsMiddleware) baseTimingMetrics(ctx context.Context, dimensions []*cloudwatch.Dimension) (time.Time, func(duration time.Duration, err *error)) {
	// 処理開始時間を記録
	startTime := time.Now()

	// メトリクスサービスがない場合は空の関数を返す
	if m.metricsService == nil {
		return startTime, func(duration time.Duration, err *error) {}
	}

	// 終了関数を返す
	return startTime, func(duration time.Duration, err *error) {
		// 処理時間メトリクスを追加
		m.metricsService.AddMetric(ctx, "ProcessingTime", float64(duration.Milliseconds()), dimensions)

		// 成功/失敗メトリクスの追加
		if err != nil && *err != nil {
			// エラーメトリクス
			m.metricsService.AddMetric(ctx, "Errors", 1.0, dimensions)
		} else {
			// 成功メトリクス
			m.metricsService.AddMetric(ctx, "Successful", 1.0, dimensions)
		}
	}
}

// StartTiming は処理時間計測を開始し、終了関数を返す
func (m *MetricsMiddleware) StartTiming(ctx context.Context) func(context.Context, interface{}, *error) {
	// メトリクスサービスがない場合は空関数を返す
	if m.metricsService == nil {
		return func(ctx context.Context, result interface{}, err *error) {}
	}

	// 標準ディメンション
	dimensions := m.createStandardDimensions()

	// 基本計測の開始
	startTime, baseMetricsFunc := m.baseTimingMetrics(ctx, dimensions)

	// 終了関数を返す
	return func(ctx context.Context, result interface{}, err *error) {
		// 処理時間を計算
		duration := time.Since(startTime)

		// 基本メトリクスを記録
		baseMetricsFunc(duration, err)

		// API Gateway固有のメトリクス
		m.recordAPIGatewayMetrics(ctx, dimensions, result)

		// ユーザー情報に基づくメトリクス
		m.recordUserMetrics(ctx)
	}
}

// レスポンスのステータスコードに基づくメトリクスを記録
func (m *MetricsMiddleware) recordAPIGatewayMetrics(ctx context.Context, dimensions []*cloudwatch.Dimension, result interface{}) {
	if m.metricsService == nil {
		return
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
}

// ユーザー情報に基づくメトリクスを記録
func (m *MetricsMiddleware) recordUserMetrics(ctx context.Context) {
	if m.metricsService == nil {
		return
	}

	// ユーザー情報があればユーザー別メトリクスを追加
	user, _ := contextutil.GetUserInfo(ctx)
	if user != nil {
		// ユーザー別メトリクス
		userDimensions := m.createUserDimensions(user.ID.String())
		m.metricsService.AddMetric(ctx, "UserRequests", 1.0, userDimensions)
	}
}

// StartTimingForThumbnail はサムネイル処理用の処理時間計測を開始し、終了関数を返す
func (m *MetricsMiddleware) StartTimingForThumbnail(ctx context.Context, s3Event events.S3Event) func(context.Context, *error) {
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

	// 基本計測の開始
	startTime, baseMetricsFunc := m.baseTimingMetrics(ctx, dimensions)

	// 終了関数を返す
	return func(ctx context.Context, err *error) {
		// 処理時間を計算
		duration := time.Since(startTime)

		// 基本メトリクスを記録
		baseMetricsFunc(duration, err)

		// 平均処理時間（1画像あたり）
		if imageCount > 0 {
			m.metricsService.AddMetric(ctx, "AverageImageProcessingTime",
				float64(duration.Milliseconds())/float64(imageCount), dimensions)
		}
	}
}

// StartTimingForCloudWatchEvent はCloudWatchイベント処理用の処理時間計測を開始し、終了関数を返す
func (m *MetricsMiddleware) StartTimingForCloudWatchEvent(ctx context.Context, event events.CloudWatchEvent) func(context.Context, *error) {
	// メトリクスサービスがない場合は空関数を返す
	if m.metricsService == nil {
		return func(ctx context.Context, err *error) {}
	}

	// 標準ディメンション
	dimensions := m.createStandardDimensions()

	// イベントソース情報を追加
	sourceDimension := &cloudwatch.Dimension{
		Name:  aws.String("EventSource"),
		Value: aws.String(event.Source),
	}
	dimensions = append(dimensions, sourceDimension)

	// イベント数メトリクスを追加
	m.metricsService.AddMetric(ctx, "Invocations", 1.0, dimensions)

	// 基本計測の開始
	startTime, baseMetricsFunc := m.baseTimingMetrics(ctx, dimensions)

	// 終了関数を返す
	return func(ctx context.Context, err *error) {
		// 処理時間を計算
		duration := time.Since(startTime)

		// 基本メトリクスを記録
		baseMetricsFunc(duration, err)
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
