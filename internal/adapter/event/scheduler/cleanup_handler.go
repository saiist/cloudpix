package scheduler

import (
	"cloudpix/internal/application/imagemanagement/usecase"
	"cloudpix/internal/logging"
	"context"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

// CleanupHandler はスケジュールされたクリーンアップイベントを処理するハンドラー
type CleanupHandler struct {
	cleanupUsecase *usecase.CleanupUsecase
	logger         logging.Logger
}

// NewCleanupHandler は新しいクリーンアップハンドラーを作成します
func NewCleanupHandler(cleanupUsecase *usecase.CleanupUsecase, logger logging.Logger) *CleanupHandler {
	return &CleanupHandler{
		cleanupUsecase: cleanupUsecase,
		logger:         logger,
	}
}

// Handle はEventBridgeスケジュールイベントを処理します
func (h *CleanupHandler) Handle(ctx context.Context, event events.CloudWatchEvent) error {
	startTime := time.Now()
	h.logger.Info("Cleanup process started", map[string]interface{}{
		"event":     event.Source,
		"eventTime": event.Time.String(),
	})

	// クリーンアップ処理を実行
	err := h.cleanupUsecase.ProcessCleanup(ctx)
	if err != nil {
		h.logger.Error(err, "Cleanup process failed", map[string]interface{}{
			"duration": time.Since(startTime).Milliseconds(),
		})
		return err
	}

	// 成功ログの記録
	h.logger.Info("Cleanup process completed successfully", map[string]interface{}{
		"duration": time.Since(startTime).Milliseconds(),
	})
	return nil
}
