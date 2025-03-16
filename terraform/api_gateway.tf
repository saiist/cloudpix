################################
# API Gateway - Common
################################
# API Gateway
resource "aws_api_gateway_rest_api" "cloudpix_api" {
  name        = "${title(var.app_name)}-API"
  description = "${title(var.app_name)} API Gateway"

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}

################################
# API Gateway - Cognito Authorizer
################################
resource "aws_api_gateway_authorizer" "cloudpix_cognito_authorizer" {
  name            = "${var.app_name}-cognito-authorizer"
  rest_api_id     = aws_api_gateway_rest_api.cloudpix_api.id
  type            = "COGNITO_USER_POOLS"
  provider_arns   = [aws_cognito_user_pool.cloudpix_users.arn]
  identity_source = "method.request.header.Authorization"
}

################################
# API Gateway - Upload Endpoint
################################
# /upload リソースの作成
resource "aws_api_gateway_resource" "upload" {
  rest_api_id = aws_api_gateway_rest_api.cloudpix_api.id
  parent_id   = aws_api_gateway_rest_api.cloudpix_api.root_resource_id
  path_part   = "upload"
}

# POST メソッドの設定
resource "aws_api_gateway_method" "upload_post" {
  rest_api_id   = aws_api_gateway_rest_api.cloudpix_api.id
  resource_id   = aws_api_gateway_resource.upload.id
  http_method   = "POST"
  authorization = "COGNITO_USER_POOLS"
  authorizer_id = aws_api_gateway_authorizer.cloudpix_cognito_authorizer.id
}

# Lambda関数との統合
resource "aws_api_gateway_integration" "lambda_integration" {
  rest_api_id = aws_api_gateway_rest_api.cloudpix_api.id
  resource_id = aws_api_gateway_resource.upload.id
  http_method = aws_api_gateway_method.upload_post.http_method

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.cloudpix_upload.invoke_arn
}

# Lambda実行権限の付与
resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.cloudpix_upload.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_api_gateway_rest_api.cloudpix_api.execution_arn}/*/*"
}

################################
# API Gateway - List Endpoint
################################
# /list リソースの作成
resource "aws_api_gateway_resource" "list" {
  rest_api_id = aws_api_gateway_rest_api.cloudpix_api.id
  parent_id   = aws_api_gateway_rest_api.cloudpix_api.root_resource_id
  path_part   = "list"
}

# GET メソッドの設定
resource "aws_api_gateway_method" "list_get" {
  rest_api_id   = aws_api_gateway_rest_api.cloudpix_api.id
  resource_id   = aws_api_gateway_resource.list.id
  http_method   = "GET"
  authorization = "COGNITO_USER_POOLS"
  authorizer_id = aws_api_gateway_authorizer.cloudpix_cognito_authorizer.id
}

# Lambda関数との統合
resource "aws_api_gateway_integration" "list_lambda_integration" {
  rest_api_id = aws_api_gateway_rest_api.cloudpix_api.id
  resource_id = aws_api_gateway_resource.list.id
  http_method = aws_api_gateway_method.list_get.http_method

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.cloudpix_list.invoke_arn
}

# Lambda実行権限の付与
resource "aws_lambda_permission" "list_api_gateway" {
  statement_id  = "AllowExecutionFromAPIGatewayForList"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.cloudpix_list.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_api_gateway_rest_api.cloudpix_api.execution_arn}/*/*"
}

################################
# API Gateway - Deployment
################################
# APIのデプロイ
resource "aws_api_gateway_deployment" "cloudpix" {
  depends_on = [
    aws_api_gateway_integration.lambda_integration,
    aws_api_gateway_integration.list_lambda_integration,
    aws_api_gateway_integration.tags_get_integration,
    aws_api_gateway_integration.tags_post_integration,
    aws_api_gateway_integration.tags_image_get_integration,
    aws_api_gateway_integration.tags_image_delete_integration
  ]

  rest_api_id = aws_api_gateway_rest_api.cloudpix_api.id

  lifecycle {
    create_before_destroy = true
  }
}

# APIステージの作成
resource "aws_api_gateway_stage" "dev" {
  deployment_id = aws_api_gateway_deployment.cloudpix.id
  rest_api_id   = aws_api_gateway_rest_api.cloudpix_api.id
  stage_name    = var.environment
}