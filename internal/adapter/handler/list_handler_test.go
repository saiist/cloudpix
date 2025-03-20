package handler

import (
	"cloudpix/internal/domain/model"
	"cloudpix/internal/mocks/middleware"
	"cloudpix/internal/mocks/repository"
	"cloudpix/internal/usecase"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestListHandler_Handle(t *testing.T) {
	// テストケース定義
	testCases := []struct {
		name           string
		setupMocks     func(*repository.MockMetadataRepository, *middleware.MockAuthMiddleware)
		request        events.APIGatewayProxyRequest
		expectedStatus int
		expectedCount  int
		expectError    bool
	}{
		{
			name: "すべての画像を取得",
			setupMocks: func(mockRepo *repository.MockMetadataRepository, mockAuth *middleware.MockAuthMiddleware) {
				// 認証ミドルウェアのモック設定
				userInfo := &model.UserInfo{
					UserID:  "test-user",
					IsAdmin: true,
				}
				mockAuth.EXPECT().
					Process(gomock.Any(), gomock.Any()).
					Return(context.Background(), userInfo, events.APIGatewayProxyResponse{}, nil)

				// リポジトリのモック設定
				mockRepo.EXPECT().
					Find(gomock.Any()).
					Return([]model.ImageMetadata{
						{
							ImageID:     "image1",
							FileName:    "test1.jpg",
							ContentType: "image/jpeg",
							Size:        1000,
							UploadDate:  "2025-03-01",
							S3ObjectKey: "uploads/image1-test1.jpg",
							DownloadURL: "https://example.com/test1.jpg",
						},
						{
							ImageID:     "image2",
							FileName:    "test2.jpg",
							ContentType: "image/jpeg",
							Size:        2000,
							UploadDate:  "2025-03-02",
							S3ObjectKey: "uploads/image2-test2.jpg",
							DownloadURL: "https://example.com/test2.jpg",
						},
					}, nil)
			},
			request: events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				Path:       "/images",
			},
			expectedStatus: 200,
			expectedCount:  2,
			expectError:    false,
		},
		{
			name: "日付でフィルターされた画像を取得",
			setupMocks: func(mockRepo *repository.MockMetadataRepository, mockAuth *middleware.MockAuthMiddleware) {
				// 認証ミドルウェアのモック設定
				userInfo := &model.UserInfo{
					UserID:  "test-user",
					IsAdmin: false,
				}
				mockAuth.EXPECT().
					Process(gomock.Any(), gomock.Any()).
					Return(context.Background(), userInfo, events.APIGatewayProxyResponse{}, nil)

				// リポジトリのモック設定
				mockRepo.EXPECT().
					FindByDate(gomock.Any(), "2025-03-01").
					Return([]model.ImageMetadata{
						{
							ImageID:     "image1",
							FileName:    "test1.jpg",
							ContentType: "image/jpeg",
							Size:        1000,
							UploadDate:  "2025-03-01",
							S3ObjectKey: "uploads/image1-test1.jpg",
							DownloadURL: "https://example.com/test1.jpg",
						},
					}, nil)
			},
			request: events.APIGatewayProxyRequest{
				HTTPMethod:            "GET",
				Path:                  "/images",
				QueryStringParameters: map[string]string{"date": "2025-03-01"},
			},
			expectedStatus: 200,
			expectedCount:  1,
			expectError:    false,
		},
		{
			name: "リポジトリエラー",
			setupMocks: func(mockRepo *repository.MockMetadataRepository, mockAuth *middleware.MockAuthMiddleware) {
				// 認証ミドルウェアのモック設定
				userInfo := &model.UserInfo{
					UserID:  "test-user",
					IsAdmin: true,
				}
				mockAuth.EXPECT().
					Process(gomock.Any(), gomock.Any()).
					Return(context.Background(), userInfo, events.APIGatewayProxyResponse{}, nil)

				// リポジトリのモック設定（エラーを返す）
				mockRepo.EXPECT().
					Find(gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			request: events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				Path:       "/images",
			},
			expectedStatus: 500,
			expectedCount:  0,
			expectError:    false,
		},
		{
			name: "認証エラー",
			setupMocks: func(mockRepo *repository.MockMetadataRepository, mockAuth *middleware.MockAuthMiddleware) {
				// 認証エラーのシミュレーション
				mockAuth.EXPECT().
					Process(gomock.Any(), gomock.Any()).
					Return(nil, nil, events.APIGatewayProxyResponse{StatusCode: 401}, errors.New("unauthorized"))
			},
			request: events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				Path:       "/images",
			},
			expectedStatus: 401,
			expectError:    true,
		},
	}

	// テストケースを実行
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックの設定
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := repository.NewMockMetadataRepository(ctrl)
			mockAuth := middleware.NewMockAuthMiddleware(ctrl)

			// モックの期待値を設定
			tc.setupMocks(mockRepo, mockAuth)

			// ハンドラを作成
			metadataUsecase := usecase.NewMetadataUsecase(mockRepo)
			handler := NewListHandler(metadataUsecase)

			// テスト実行
			resp, err := handler.Handle(context.Background(), tc.request)

			// 結果の検証
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			// 200 OKの場合はレスポンスの内容を検証
			if tc.expectedStatus == 200 {
				var response model.ListResponse
				err = json.Unmarshal([]byte(resp.Body), &response)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCount, response.Count)
				assert.Len(t, response.Images, tc.expectedCount)
			}
		})
	}
}

func TestListHandler_ErrorResponse(t *testing.T) {
	// モックの設定
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockMetadataRepository(ctrl)

	// ハンドラを作成
	metadataUsecase := usecase.NewMetadataUsecase(mockRepo)
	handler := NewListHandler(metadataUsecase)

	// エラーレスポンスをテスト
	resp, err := handler.errorResponse(500, "Test error message")
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	assert.Contains(t, resp.Body, "Test error message")

	// JSONレスポンスの構造を検証
	var errorResp map[string]string
	err = json.Unmarshal([]byte(resp.Body), &errorResp)
	assert.NoError(t, err)
	assert.Equal(t, "Test error message", errorResp["error"])
}

func TestListHandler_JsonResponse(t *testing.T) {
	// モックの設定
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockMetadataRepository(ctrl)

	// ハンドラを作成
	metadataUsecase := usecase.NewMetadataUsecase(mockRepo)
	handler := NewListHandler(metadataUsecase)

	// テスト用のデータ
	testData := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	// JSONレスポンスをテスト
	resp, err := handler.jsonResponse(200, testData)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Headers["Content-Type"])

	// JSONレスポンスの内容を検証
	var respData map[string]interface{}
	err = json.Unmarshal([]byte(resp.Body), &respData)
	assert.NoError(t, err)
	assert.Equal(t, "value1", respData["key1"])
	assert.Equal(t, float64(123), respData["key2"])
}
