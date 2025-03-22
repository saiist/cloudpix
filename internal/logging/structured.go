package logging

// LogEntry は共通の構造化ログエントリを定義します
type LogEntry struct {
	// 共通フィールド
	Timestamp string `json:"timestamp"`           // ISO8601形式（2025-03-22T14:30:00Z）
	Level     string `json:"level"`               // "debug", "info", "warn", "error", "fatal"
	RequestID string `json:"requestId,omitempty"` // Lambda/APIGatewayのリクエストID
	Service   string `json:"service"`             // "CloudPix"
	Function  string `json:"function"`            // Lambda関数名
	Operation string `json:"operation,omitempty"` // 実行している操作（"UploadImage", "ListImages"など）
	UserID    string `json:"userId,omitempty"`    // ユーザーID（認証済みの場合）
	Message   string `json:"message"`             // ログメッセージ
	Duration  int    `json:"duration,omitempty"`  // 処理時間（ミリ秒）
	Env       string `json:"env"`                 // 環境（dev, prod）

	// エラー関連（エラー発生時のみ）
	ErrorType  string `json:"errorType,omitempty"`  // エラータイプ
	ErrorMsg   string `json:"errorMsg,omitempty"`   // エラーメッセージ
	StackTrace string `json:"stackTrace,omitempty"` // スタックトレース（必要な場合）

	// コンテキスト情報（操作に特有のデータ）
	Context map[string]interface{} `json:"context,omitempty"` // 追加のコンテキスト情報
}

// 操作固有のコンテキスト構造体

// UploadContext はアップロード操作のコンテキスト情報
type UploadContext struct {
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
	ImageID     string `json:"imageId,omitempty"`
	Size        int    `json:"size,omitempty"`
	S3ObjectKey string `json:"s3ObjectKey,omitempty"`
}

// ListContext は一覧取得操作のコンテキスト情報
type ListContext struct {
	DateFilter string `json:"dateFilter,omitempty"`
	TagFilter  string `json:"tagFilter,omitempty"`
	Count      int    `json:"count,omitempty"`
}

// ThumbnailContext はサムネイル生成操作のコンテキスト情報
type ThumbnailContext struct {
	OriginalKey  string `json:"originalKey"`
	ThumbnailKey string `json:"thumbnailKey,omitempty"`
	ImageID      string `json:"imageId,omitempty"`
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
	ProcessTime  int    `json:"processTime,omitempty"` // 画像処理時間（ミリ秒）
}

// TagContext はタグ操作のコンテキスト情報
type TagContext struct {
	ImageID   string   `json:"imageId,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Operation string   `json:"operation"` // "add", "remove", "list", "get"
	Count     int      `json:"count,omitempty"`
}
