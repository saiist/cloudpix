################################
# Random Resource
################################
# バケットポリシーのためのランダムサフィックス
resource "random_string" "bucket_suffix" {
  length  = 8
  special = false
  upper   = false
}

################################
# S3 Bucket
################################
# 画像保存用のS3バケット
resource "aws_s3_bucket" "cloudpix_images" {
  bucket        = "${var.app_name}-images-${random_string.bucket_suffix.result}"
  force_destroy = true # デモ用：削除時にバケット内のオブジェクトも削除

  tags = {
    Name        = "${var.app_name}-Images"
    Environment = var.environment
  }
}

# パブリックアクセスブロック設定
resource "aws_s3_bucket_public_access_block" "cloudpix_images" {
  bucket = aws_s3_bucket.cloudpix_images.id

  block_public_acls       = true
  block_public_policy     = false # ポリシーによる公開アクセスを許可
  ignore_public_acls      = true
  restrict_public_buckets = false # パブリックポリシーを持つバケットへのアクセスを許可
}

# S3バケットポリシー - uploads/ フォルダの読み取りを許可
resource "aws_s3_bucket_policy" "allow_public_read" {
  # パブリックアクセスブロック設定の後に適用されるように依存関係を明示
  depends_on = [aws_s3_bucket_public_access_block.cloudpix_images]

  bucket = aws_s3_bucket.cloudpix_images.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "PublicReadForUploads"
        Effect    = "Allow"
        Principal = "*"
        Action    = "s3:GetObject"
        Resource  = "${aws_s3_bucket.cloudpix_images.arn}/uploads/*"
      }
    ]
  })
}

# CORSの設定
resource "aws_s3_bucket_cors_configuration" "cloudpix_images" {
  bucket = aws_s3_bucket.cloudpix_images.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "PUT", "POST", "DELETE", "HEAD"]
    allowed_origins = ["*"] # 本番環境では特定のオリジンに制限すべき
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }
}