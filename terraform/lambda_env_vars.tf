################################
# Lambda Environment Variables
################################

locals {
  common_lambda_env_vars = {
    ENVIRONMENT    = var.environment
    ENABLE_METRICS = var.enable_cloudwatch_metrics ? "true" : "false"
    ENABLE_XRAY    = var.enable_xray_tracing ? "true" : "false"
  }

  upload_lambda_env_vars = merge(local.common_lambda_env_vars, {
    S3_BUCKET_NAME      = aws_s3_bucket.cloudpix_images.bucket
    METADATA_TABLE_NAME = aws_dynamodb_table.cloudpix_metadata.name
    USER_POOL_ID        = aws_cognito_user_pool.cloudpix_users.id
    USER_POOL_CLIENT_ID = aws_cognito_user_pool_client.cloudpix_client.id
  })

  list_lambda_env_vars = merge(local.common_lambda_env_vars, {
    METADATA_TABLE_NAME = aws_dynamodb_table.cloudpix_metadata.name
    USER_POOL_ID        = aws_cognito_user_pool.cloudpix_users.id
    USER_POOL_CLIENT_ID = aws_cognito_user_pool_client.cloudpix_client.id
  })

  thumbnail_lambda_env_vars = merge(local.common_lambda_env_vars, {
    S3_BUCKET_NAME      = aws_s3_bucket.cloudpix_images.bucket
    METADATA_TABLE_NAME = aws_dynamodb_table.cloudpix_metadata.name
  })

  tags_lambda_env_vars = merge(local.common_lambda_env_vars, {
    TAGS_TABLE_NAME     = aws_dynamodb_table.cloudpix_tags.name
    METADATA_TABLE_NAME = aws_dynamodb_table.cloudpix_metadata.name
    USER_POOL_ID        = aws_cognito_user_pool.cloudpix_users.id
    USER_POOL_CLIENT_ID = aws_cognito_user_pool_client.cloudpix_client.id
  })

  cleanup_lambda_env_vars = merge(local.common_lambda_env_vars, {
    S3_BUCKET_NAME      = aws_s3_bucket.cloudpix_images.bucket
    METADATA_TABLE_NAME = aws_dynamodb_table.cloudpix_metadata.name
    TAGS_TABLE_NAME     = aws_dynamodb_table.cloudpix_tags.name
    RETENTION_DAYS      = var.image_retention_days
  })


}
