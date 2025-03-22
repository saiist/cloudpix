package middleware

import (
	"cloudpix/internal/contextutil"
	"cloudpix/internal/logging"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

// LoggingMiddleware はリクエスト・レスポンスをログに記録するミドルウェア
func LoggingMiddleware(config *logging.LoggingMiddlewareConfig) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			startTime := time.Now()

			// リクエスト情報
			requestID := event.RequestContext.RequestID
			path := event.Path
			method := event.HTTPMethod
			sourceIP := event.RequestContext.Identity.SourceIP

			// ロガーを取得または作成
			logger := config.Logger.
				WithRequestID(requestID).
				WithFunction("APIHandler")

			// ユーザー情報の取得
			userID := ""
			if userInfo, ok := contextutil.GetUserInfo(ctx); ok && userInfo != nil {
				userID = userInfo.UserID
				logger = logger.WithUserID(userID)
			}

			// 操作名を推測（パスに基づいて）
			operation := guessOperation(path, method)
			logger = logger.WithOperation(operation)

			// ロガーをコンテキストに注入
			ctx = logging.WithLogger(ctx, logger)

			// リクエストコンテキスト情報を構築
			requestContext := map[string]interface{}{
				"httpMethod": method,
				"path":       path,
				"sourceIP":   sourceIP,
			}

			// クエリパラメータが存在する場合
			if len(event.QueryStringParameters) > 0 {
				requestContext["queryParams"] = event.QueryStringParameters
			}

			// ヘッダーの処理
			if config.LogRequestHeaders && len(event.Headers) > 0 {
				headers := make(map[string]string)

				for k, v := range event.Headers {
					// 機密ヘッダーの値をマスク
					isSensitive := false
					for _, sensitive := range config.SensitiveHeaders {
						if k == sensitive {
							headers[k] = "******"
							isSensitive = true
							break
						}
					}

					if !isSensitive {
						headers[k] = v
					}
				}

				requestContext["headers"] = headers
			}

			// リクエストボディの処理
			if config.LogRequestBody && event.Body != "" {
				body := event.Body
				if len(body) > config.MaxBodyLogLength {
					body = body[:config.MaxBodyLogLength] + "... [truncated]"
				}

				// JSONとして解析を試みる
				var jsonBody interface{}
				if err := json.Unmarshal([]byte(body), &jsonBody); err == nil {
					requestContext["body"] = jsonBody
				} else {
					requestContext["body"] = body
				}
			}

			// リクエストログ
			logger.Info(fmt.Sprintf("Request: %s %s", method, path), requestContext)

			// ハンドラー実行
			resp, err := next(ctx, event)

			// 処理時間を計算
			duration := time.Since(startTime).Milliseconds()

			// レスポンスコンテキスト情報
			responseContext := map[string]interface{}{
				"statusCode": resp.StatusCode,
				"duration":   duration,
			}

			// レスポンスヘッダー
			if len(resp.Headers) > 0 {
				responseContext["headers"] = resp.Headers
			}

			// レスポンスボディの処理
			if config.LogResponseBody && resp.Body != "" {
				body := resp.Body
				if len(body) > config.MaxBodyLogLength {
					body = body[:config.MaxBodyLogLength] + "... [truncated]"
				}

				// JSONとして解析を試みる
				var jsonBody interface{}
				if err := json.Unmarshal([]byte(body), &jsonBody); err == nil {
					responseContext["body"] = jsonBody
				} else {
					responseContext["body"] = body
				}
			}

			// エラーログまたは成功ログ
			if err != nil {
				logger.Error(err, "Error processing request", responseContext)
			} else {
				logger.Info(fmt.Sprintf("Response: %d (%dms)", resp.StatusCode, duration), responseContext)
			}

			return resp, err
		}
	}
}

// guessOperation はパスとメソッドから操作名を推測する
func guessOperation(path, method string) string {
	// パスに基づいて操作名を推測
	if path == "/upload" && method == "POST" {
		return "UploadImage"
	} else if path == "/list" && method == "GET" {
		return "ListImages"
	} else if path == "/tags" {
		if method == "GET" {
			return "ListTags"
		} else if method == "POST" {
			return "AddTags"
		}
	} else if strings.HasPrefix(path, "/tags/") {
		if method == "GET" {
			return "GetImageTags"
		} else if method == "DELETE" {
			return "RemoveTags"
		}
	}

	// デフォルト
	return method + path
}
