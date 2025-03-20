################################
# ECR Repository for Tags
################################
# タグ管理用のECRリポジトリ
resource "aws_ecr_repository" "cloudpix_tags" {
  name                 = "${var.app_name}-tags"
  image_tag_mutability = "MUTABLE"
  force_delete         = true

  image_scanning_configuration {
    scan_on_push = true
  }
}

################################
# Docker Build & Push - Tags
################################
# タグ管理関数のイメージのビルドとプッシュ
resource "null_resource" "docker_build_push_tags" {
  depends_on = [aws_ecr_repository.cloudpix_tags]

  triggers = {
    ecr_repository_url = aws_ecr_repository.cloudpix_tags.repository_url
    dockerfile_hash    = filemd5("${path.module}/../Dockerfile")
    main_go_hash       = filemd5("${path.module}/../cmd/tags/main.go")
    build_script_hash  = filemd5("${path.module}/../build_and_push.sh")
  }

  provisioner "local-exec" {
    command = <<-EOT
      echo "Building tags function image..."
      cd ${path.module}/.. && \
      chmod +x build_and_push.sh && \
      REPO_NAME="cloudpix-tags" ./build_and_push.sh ${aws_ecr_repository.cloudpix_tags.repository_url} ./cmd/tags/main.go
    EOT
  }
}

################################
# Tags Lambda Function
################################
# タグ管理用Lambda関数
resource "aws_lambda_function" "cloudpix_tags" {
  function_name = "${var.app_name}-tags"
  role          = aws_iam_role.lambda_role.arn
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.cloudpix_tags.repository_url}:latest"

  timeout     = 30
  memory_size = 128

  environment {
    variables = {
      TAGS_TABLE_NAME     = aws_dynamodb_table.cloudpix_tags.name
      METADATA_TABLE_NAME = aws_dynamodb_table.cloudpix_metadata.name
      USER_POOL_ID        = aws_cognito_user_pool.cloudpix_users.id
      USER_POOL_CLIENT_ID = aws_cognito_user_pool_client.cloudpix_client.id
    }
  }

  depends_on = [
    null_resource.docker_build_push_tags
  ]

  # X-Rayトレースを有効化
  tracing_config {
    mode = "Active"
  }
}

################################
# API Gateway - Tags Endpoints
################################
# /tags リソースの作成
resource "aws_api_gateway_resource" "tags" {
  rest_api_id = aws_api_gateway_rest_api.cloudpix_api.id
  parent_id   = aws_api_gateway_rest_api.cloudpix_api.root_resource_id
  path_part   = "tags"
}

# /tags/{imageId} リソースの作成
resource "aws_api_gateway_resource" "tags_image" {
  rest_api_id = aws_api_gateway_rest_api.cloudpix_api.id
  parent_id   = aws_api_gateway_resource.tags.id
  path_part   = "{imageId}"
}

# GET /tags メソッド - すべてのタグのリスト取得
resource "aws_api_gateway_method" "tags_get" {
  rest_api_id   = aws_api_gateway_rest_api.cloudpix_api.id
  resource_id   = aws_api_gateway_resource.tags.id
  http_method   = "GET"
  authorization = "COGNITO_USER_POOLS"
  authorizer_id = aws_api_gateway_authorizer.cloudpix_cognito_authorizer.id
}

# POST /tags メソッド - タグの追加
resource "aws_api_gateway_method" "tags_post" {
  rest_api_id   = aws_api_gateway_rest_api.cloudpix_api.id
  resource_id   = aws_api_gateway_resource.tags.id
  http_method   = "POST"
  authorization = "COGNITO_USER_POOLS"
  authorizer_id = aws_api_gateway_authorizer.cloudpix_cognito_authorizer.id
}

# GET /tags/{imageId} メソッド - 画像のタグ取得
resource "aws_api_gateway_method" "tags_image_get" {
  rest_api_id   = aws_api_gateway_rest_api.cloudpix_api.id
  resource_id   = aws_api_gateway_resource.tags_image.id
  http_method   = "GET"
  authorization = "COGNITO_USER_POOLS"
  authorizer_id = aws_api_gateway_authorizer.cloudpix_cognito_authorizer.id
}

# DELETE /tags/{imageId} メソッド - 画像のタグ削除
resource "aws_api_gateway_method" "tags_image_delete" {
  rest_api_id   = aws_api_gateway_rest_api.cloudpix_api.id
  resource_id   = aws_api_gateway_resource.tags_image.id
  http_method   = "DELETE"
  authorization = "COGNITO_USER_POOLS"
  authorizer_id = aws_api_gateway_authorizer.cloudpix_cognito_authorizer.id
}

# GET /tags との統合
resource "aws_api_gateway_integration" "tags_get_integration" {
  rest_api_id = aws_api_gateway_rest_api.cloudpix_api.id
  resource_id = aws_api_gateway_resource.tags.id
  http_method = aws_api_gateway_method.tags_get.http_method

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.cloudpix_tags.invoke_arn
}

# POST /tags との統合
resource "aws_api_gateway_integration" "tags_post_integration" {
  rest_api_id = aws_api_gateway_rest_api.cloudpix_api.id
  resource_id = aws_api_gateway_resource.tags.id
  http_method = aws_api_gateway_method.tags_post.http_method

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.cloudpix_tags.invoke_arn
}

# GET /tags/{imageId} との統合
resource "aws_api_gateway_integration" "tags_image_get_integration" {
  rest_api_id = aws_api_gateway_rest_api.cloudpix_api.id
  resource_id = aws_api_gateway_resource.tags_image.id
  http_method = aws_api_gateway_method.tags_image_get.http_method

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.cloudpix_tags.invoke_arn
}

# DELETE /tags/{imageId} との統合
resource "aws_api_gateway_integration" "tags_image_delete_integration" {
  rest_api_id = aws_api_gateway_rest_api.cloudpix_api.id
  resource_id = aws_api_gateway_resource.tags_image.id
  http_method = aws_api_gateway_method.tags_image_delete.http_method

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.cloudpix_tags.invoke_arn
}

# Lambda実行権限の付与
resource "aws_lambda_permission" "tags_api_gateway" {
  statement_id  = "AllowExecutionFromAPIGatewayForTags"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.cloudpix_tags.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_api_gateway_rest_api.cloudpix_api.execution_arn}/*/*"
}
