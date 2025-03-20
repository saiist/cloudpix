package metrics

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

// MetricsService はメトリクス収集サービスのインターフェース
type MetricsService interface {
	// AddMetric はメトリクスを追加する
	AddMetric(ctx context.Context, metricName string, value float64, dimensions []*cloudwatch.Dimension)

	// Flush はバッファのメトリクスを送信する
	Flush(ctx context.Context) error

	// Close はメトリクスサービスを適切に終了する
	Close() error
}

// MetricsConfig はメトリクス収集の設定
type MetricsConfig struct {
	BatchSize       int
	FlushInterval   time.Duration
	DetailedMetrics bool
	Namespace       string
}

// DefaultMetricsConfig はデフォルト設定
var DefaultMetricsConfig = MetricsConfig{
	BatchSize:       20,
	FlushInterval:   time.Minute,
	DetailedMetrics: false,
	Namespace:       "CloudPix/Lambda",
}

// CloudWatchMetricsService はCloudWatchを使用したメトリクス収集サービス
type CloudWatchMetricsService struct {
	namespace        string
	cloudWatchClient *cloudwatch.CloudWatch
	mutex            sync.RWMutex
	batchData        []*cloudwatch.MetricDatum
	flushThreshold   int
	stopCh           chan struct{} // 停止用チャネル
}

// NewCloudWatchMetricsService は新しいCloudWatchメトリクスサービスを作成する
func NewCloudWatchMetricsService(sess *session.Session, config MetricsConfig) MetricsService {
	// CloudWatchクライアントを作成
	cwClient := cloudwatch.New(sess)

	service := &CloudWatchMetricsService{
		namespace:        config.Namespace,
		cloudWatchClient: cwClient,
		flushThreshold:   config.BatchSize,
		batchData:        make([]*cloudwatch.MetricDatum, 0, config.BatchSize),
		stopCh:           make(chan struct{}),
	}

	// 定期的にバッファをフラッシュするゴルーチン
	if config.FlushInterval > 0 {
		go func() {
			ticker := time.NewTicker(config.FlushInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					if err := service.Flush(context.Background()); err != nil {
						log.Printf("Error flushing metrics: %v", err)
					}
				case <-service.stopCh:
					// 最後に残っているメトリクスをフラッシュして終了
					service.Flush(context.Background())
					return
				}
			}
		}()
	}

	return service
}

// AddMetric はメトリクスデータをバッファに追加する
func (s *CloudWatchMetricsService) AddMetric(ctx context.Context, metricName string, value float64, dimensions []*cloudwatch.Dimension) {
	// CloudWatchクライアントがない場合は何もしない
	if s.cloudWatchClient == nil {
		return
	}

	metricDatum := &cloudwatch.MetricDatum{
		MetricName: aws.String(metricName),
		Value:      aws.Float64(value),
		Unit:       s.getUnit(metricName),
		Dimensions: dimensions,
		Timestamp:  aws.Time(time.Now()),
	}

	// スレッドセーフにバッファにメトリクスを追加
	s.mutex.Lock()
	s.batchData = append(s.batchData, metricDatum)

	// バッファサイズが閾値を超えた場合はフラッシュをトリガー
	if len(s.batchData) >= s.flushThreshold {
		go func() {
			if err := s.Flush(context.Background()); err != nil {
				log.Printf("Error flushing metrics: %v", err)
			}
		}()
	}

	s.mutex.Unlock()
}

// getUnit はメトリクス名に対応する単位を返す
func (s *CloudWatchMetricsService) getUnit(metricName string) *string {
	switch metricName {
	case "Duration", "ProcessingTime", "AverageProcessingTime", "Latency":
		return aws.String(cloudwatch.StandardUnitMilliseconds)
	default:
		return aws.String(cloudwatch.StandardUnitCount)
	}
}

// Flush はバッファのメトリクスをCloudWatchに送信する
func (s *CloudWatchMetricsService) Flush(ctx context.Context) error {
	if s.cloudWatchClient == nil || len(s.batchData) == 0 {
		return nil
	}

	// バッファのデータをスレッドセーフに取得
	s.mutex.Lock()
	metrics := make([]*cloudwatch.MetricDatum, len(s.batchData))
	copy(metrics, s.batchData)
	s.batchData = s.batchData[:0] // バッファをクリア
	s.mutex.Unlock()

	// CloudWatchにメトリクスを送信（最大サイズを考慮してバッチ処理）
	const maxMetricsPerRequest = 20
	for i := 0; i < len(metrics); i += maxMetricsPerRequest {
		end := min(i+maxMetricsPerRequest, len(metrics))

		batch := metrics[i:end]
		_, err := s.cloudWatchClient.PutMetricDataWithContext(ctx, &cloudwatch.PutMetricDataInput{
			Namespace:  aws.String(s.namespace),
			MetricData: batch,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// Close はメトリクスサービスを適切に終了する
func (s *CloudWatchMetricsService) Close() error {
	close(s.stopCh)
	return nil
}
