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

# 画像のタグ情報を保存するDynamoDBテーブル
resource "aws_dynamodb_table" "cloudpix_tags" {
  name         = "${var.app_name}-tags"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "TagName"
  range_key    = "ImageID"

  attribute {
    name = "TagName"
    type = "S"
  }

  attribute {
    name = "ImageID"
    type = "S"
  }

  # 逆引き用インデックス（ImageIDからタグを検索）
  global_secondary_index {
    name            = "ImageIDIndex"
    hash_key        = "ImageID"
    projection_type = "ALL"
  }

  tags = {
    Name        = "${var.app_name}-Tags"
    Environment = var.environment
  }
}