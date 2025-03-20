package handler

import (
	"context"
	"log"

	"cloudpix/internal/adapter/middleware"

	"github.com/aws/aws-lambda-go/events"
)

// HandlerFunc はAPIハンドラー関数の型
type HandlerFunc func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

// WithAuth は認証ミドルウェアを適用するハンドラーラッパー
func WithAuth(authMiddleware middleware.AuthMiddleware, handler HandlerFunc) HandlerFunc {
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
		log.Printf("Authenticated user: %s, Groups: %v", userInfo.UserID, userInfo.Groups)
		return handler(newCtx, event)
	}
}

// WithMetrics はメトリクス収集ミドルウェアを適用するハンドラーラッパー
func WithMetrics(metricsMiddleware *middleware.MetricsMiddleware, handler HandlerFunc) HandlerFunc {
	return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		// レスポンスとエラーを格納する変数
		var resp events.APIGatewayProxyResponse
		var err error

		// 処理時間計測を開始し、完了時に計測を終了する
		defer metricsMiddleware.StartTiming(ctx)(ctx, &resp, &err)

		// ハンドラー実行
		resp, err = handler(ctx, event)
		return resp, err
	}
}

// WithMetricsForThumbnail はメトリクス収集ミドルウェアを適用するハンドラーラッパー
func WithMetricsForThumbnail(metricsMiddleware *middleware.MetricsMiddleware,
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
