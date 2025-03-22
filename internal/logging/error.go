package logging

import (
	"fmt"
)

// ErrorType は構造化エラータイプを定義
type ErrorType string

// 定義済みエラータイプ
const (
	ErrorTypeValidation     ErrorType = "ValidationError"      // 入力検証エラー
	ErrorTypeNotFound       ErrorType = "NotFoundError"        // リソース未発見エラー
	ErrorTypePermission     ErrorType = "PermissionError"      // 権限エラー
	ErrorTypeAuthentication ErrorType = "AuthenticationError"  // 認証エラー
	ErrorTypeInternal       ErrorType = "InternalError"        // 内部エラー
	ErrorTypeExternal       ErrorType = "ExternalServiceError" // 外部サービスエラー
	ErrorTypeConfiguration  ErrorType = "ConfigurationError"   // 設定エラー
	ErrorTypeResourceLimit  ErrorType = "ResourceLimitError"   // リソース制限エラー
	ErrorTypeFormat         ErrorType = "FormatError"          // フォーマットエラー
	ErrorTypeIO             ErrorType = "IOError"              // 入出力エラー
)

// StructuredError は詳細情報を含む構造化エラー
type StructuredError struct {
	Type      ErrorType              // エラータイプ
	Message   string                 // エラーメッセージ
	Cause     error                  // 元のエラー（ある場合）
	Details   map[string]interface{} // 詳細情報
	RequestID string                 // リクエストID（追跡用）
	Code      string                 // エラーコード（クライアント向け）
}

// NewError は新しい構造化エラーを作成
func NewError(errType ErrorType, message string, cause error) *StructuredError {
	return &StructuredError{
		Type:    errType,
		Message: message,
		Cause:   cause,
		Details: make(map[string]interface{}),
	}
}

// WithDetail はエラーに詳細情報を追加
func (e *StructuredError) WithDetail(key string, value interface{}) *StructuredError {
	e.Details[key] = value
	return e
}

// WithDetails は複数の詳細情報を追加
func (e *StructuredError) WithDetails(details map[string]interface{}) *StructuredError {
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// WithRequestID はリクエストIDを設定
func (e *StructuredError) WithRequestID(requestID string) *StructuredError {
	e.RequestID = requestID
	return e
}

// WithCode はエラーコードを設定
func (e *StructuredError) WithCode(code string) *StructuredError {
	e.Code = code
	return e
}

// Error はエラーインターフェースを実装
func (e *StructuredError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap はエラーアンラップのサポート（Go 1.13+）
func (e *StructuredError) Unwrap() error {
	return e.Cause
}

// IsValidationError は検証エラーかどうかを判定
func IsValidationError(err error) bool {
	structErr, ok := err.(*StructuredError)
	return ok && structErr.Type == ErrorTypeValidation
}

// IsNotFoundError は未発見エラーかどうかを判定
func IsNotFoundError(err error) bool {
	structErr, ok := err.(*StructuredError)
	return ok && structErr.Type == ErrorTypeNotFound
}

// IsPermissionError は権限エラーかどうかを判定
func IsPermissionError(err error) bool {
	structErr, ok := err.(*StructuredError)
	return ok && structErr.Type == ErrorTypePermission
}

// IsAuthenticationError は認証エラーかどうかを判定
func IsAuthenticationError(err error) bool {
	structErr, ok := err.(*StructuredError)
	return ok && structErr.Type == ErrorTypeAuthentication
}

// IsInternalError は内部エラーかどうかを判定
func IsInternalError(err error) bool {
	structErr, ok := err.(*StructuredError)
	return ok && structErr.Type == ErrorTypeInternal
}

// IsExternalServiceError は外部サービスエラーかどうかを判定
func IsExternalServiceError(err error) bool {
	structErr, ok := err.(*StructuredError)
	return ok && structErr.Type == ErrorTypeExternal
}
