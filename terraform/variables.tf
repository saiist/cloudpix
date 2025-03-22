################################
# Variables
################################
variable "aws_region" {
  description = "AWSリージョン"
  type        = string
  default     = "ap-northeast-1"
}

variable "app_name" {
  description = "アプリケーション名"
  type        = string
  default     = "cloudpix"
}

variable "environment" {
  description = "デプロイ環境"
  type        = string
  default     = "dev"
}

variable "lambda_memory_size" {
  description = "Lambda関数のメモリサイズ (MB)"
  type        = number
  default     = 128
}

variable "lambda_timeout" {
  description = "Lambda関数のタイムアウト時間 (秒)"
  type        = number
  default     = 30
}

variable "app_frontend_url" {
  description = "フロントエンドアプリケーションのURL"
  type        = string
  default     = "https://dev.cloudpix.example.com"
}

################################
# Monitoring Variables
################################
variable "enable_cloudwatch_metrics" {
  description = "CloudWatch Metricsの有効化"
  type        = bool
  default     = true
}

variable "enable_xray_tracing" {
  description = "AWS X-Rayトレースの有効化"
  type        = bool
  default     = true
}

variable "metrics_retention_days" {
  description = "CloudWatch Logsの保持期間（日数）"
  type        = number
  default     = 14
}

variable "alarm_email" {
  description = "アラート通知先のメールアドレス"
  type        = string
  default     = "admin@example.com"
}

variable "dashboard_refresh_interval" {
  description = "CloudWatchダッシュボードの自動更新間隔（秒）"
  type        = number
  default     = 300
}

variable "lambda_error_threshold" {
  description = "Lambda関数のエラーアラートしきい値"
  type        = number
  default     = 5
}

variable "lambda_duration_threshold_base" {
  description = "Lambda関数の実行時間アラートしきい値（標準、ミリ秒）"
  type        = number
  default     = 3000
}

variable "lambda_duration_threshold_thumbnail" {
  description = "サムネイル処理Lambda関数の実行時間アラートしきい値（ミリ秒）"
  type        = number
  default     = 10000
}

variable "api_4xx_error_threshold" {
  description = "API Gateway 4xxエラーアラートしきい値"
  type        = number
  default     = 10
}

variable "api_5xx_error_threshold" {
  description = "API Gateway 5xxエラーアラートしきい値"
  type        = number
  default     = 5
}

variable "image_retention_days" {
  description = "画像の保持日数"
  type        = number
  default     = 10
}