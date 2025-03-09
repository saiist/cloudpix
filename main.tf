provider "aws" {
  region = "ap-northeast-1"
}

# ECRリポジトリの作成
resource "aws_ecr_repository" "cloudpix_upload" {
  name                 = "cloudpix-upload"
  image_tag_mutability = "MUTABLE"
  force_delete         = true

  image_scanning_configuration {
    scan_on_push = true
  }
}

# イメージのビルドとプッシュを行うnull_resource
resource "null_resource" "docker_build_push" {
  depends_on = [aws_ecr_repository.cloudpix_upload]

  # リポジトリURLが変更された場合、またはDockerfileが変更された場合に再実行
  triggers = {
    ecr_repository_url = aws_ecr_repository.cloudpix_upload.repository_url
    dockerfile_hash    = filemd5("${path.module}/Dockerfile")
    main_go_hash       = filemd5("${path.module}/cmd/upload/main.go")
    build_script_hash  = filemd5("${path.module}/build_and_push.sh")
  }

  # シェルスクリプトを実行し、ECRリポジトリURLを渡す
  provisioner "local-exec" {
    command = "chmod +x ${path.module}/build_and_push.sh && ${path.module}/build_and_push.sh ${aws_ecr_repository.cloudpix_upload.repository_url}"
  }
}

# Lambda実行用のIAMロール
resource "aws_iam_role" "lambda_role" {
  name = "lambda-cloudpix-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

# Lambda基本実行ポリシーのアタッチ
resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# Lambda関数の作成
resource "aws_lambda_function" "cloudpix_upload" {
  function_name = "cloudpix-upload"
  role          = aws_iam_role.lambda_role.arn
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.cloudpix_upload.repository_url}:latest"

  timeout     = 30
  memory_size = 128

  depends_on = [
    null_resource.docker_build_push
  ]
}

# API Gateway
resource "aws_api_gateway_rest_api" "cloudpix_api" {
  name        = "CloudPix-API"
  description = "CloudPix API Gateway"

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}

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
  authorization = "NONE"
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

# APIのデプロイ
resource "aws_api_gateway_deployment" "cloudpix" {
  depends_on = [
    aws_api_gateway_integration.lambda_integration
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
  stage_name    = "dev"
}

# 出力値の定義
output "ecr_repository_url" {
  value = aws_ecr_repository.cloudpix_upload.repository_url
}

output "api_url" {
  value = "${aws_api_gateway_stage.dev.invoke_url}/upload"
}