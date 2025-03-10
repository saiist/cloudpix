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
