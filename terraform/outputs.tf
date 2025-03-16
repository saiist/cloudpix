################################
# Outputs
################################
output "ecr_repository_url" {
  value       = aws_ecr_repository.cloudpix_upload.repository_url
  description = "ECRリポジトリのURL（アップロード機能用）"
}

output "ecr_list_repository_url" {
  value       = aws_ecr_repository.cloudpix_list.repository_url
  description = "ECRリポジトリのURL（一覧取得機能用）"
}

output "s3_bucket_name" {
  value       = aws_s3_bucket.cloudpix_images.bucket
  description = "画像保存用S3バケット名"
}

output "METADATA_TABLE_NAME" {
  value       = aws_dynamodb_table.cloudpix_metadata.name
  description = "メタデータ保存用DynamoDBテーブル名"
}

output "api_url" {
  value       = "${aws_api_gateway_stage.dev.invoke_url}/upload"
  description = "画像アップロードAPIのエンドポイントURL"
}

output "list_api_url" {
  value       = "${aws_api_gateway_stage.dev.invoke_url}/list"
  description = "画像一覧取得APIのエンドポイントURL"
}

output "ecr_thumbnail_repository_url" {
  value       = aws_ecr_repository.cloudpix_thumbnail.repository_url
  description = "ECRリポジトリのURL（サムネイル生成機能用）"
}

output "ecr_tags_repository_url" {
  value       = aws_ecr_repository.cloudpix_tags.repository_url
  description = "ECRリポジトリのURL（タグ管理用）"
}

output "tags_api_url" {
  value       = "${aws_api_gateway_stage.dev.invoke_url}/tags"
  description = "タグ管理APIのエンドポイントURL"
}

# 認証関連の情報
output "cognito_user_pool_id" {
  value       = aws_cognito_user_pool.cloudpix_users.id
  description = "Cognitoユーザープールのid"
}

output "cognito_client_id" {
  value       = aws_cognito_user_pool_client.cloudpix_client.id
  description = "Cognitoアプリクライアントid"
}

output "cognito_domain" {
  value       = "https://${aws_cognito_user_pool_domain.cloudpix_domain.domain}.auth.${var.aws_region}.amazoncognito.com"
  description = "Cognito ホストされたUIドメイン"
}