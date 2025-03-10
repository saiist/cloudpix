################################
# DynamoDB Table
################################
# 画像メタデータ管理用のDynamoDBテーブル
resource "aws_dynamodb_table" "cloudpix_metadata" {
  name         = "${var.app_name}-metadata"
  billing_mode = "PAY_PER_REQUEST" # オンデマンドキャパシティモード
  hash_key     = "ImageID"         # パーティションキー

  attribute {
    name = "ImageID"
    type = "S"
  }

  attribute {
    name = "UploadDate"
    type = "S"
  }

  # UploadDateによるクエリ用のGSI
  global_secondary_index {
    name            = "UploadDateIndex"
    hash_key        = "UploadDate"
    projection_type = "ALL"
  }

  tags = {
    Name        = "${var.app_name}-Metadata"
    Environment = var.environment
  }
}