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
# ECR Repository
################################
# ECRリポジトリの作成
resource "aws_ecr_repository" "cloudpix_upload" {
  name                 = "${var.app_name}-upload"
  image_tag_mutability = "MUTABLE"
  force_delete         = true

  image_scanning_configuration {
    scan_on_push = true
  }
}

# 一覧取得用のECRリポジトリ
resource "aws_ecr_repository" "cloudpix_list" {
  name                 = "${var.app_name}-list"
  image_tag_mutability = "MUTABLE"
  force_delete         = true

  image_scanning_configuration {
    scan_on_push = true
  }
}

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

################################
# IAM Configuration
################################
# Lambda実行用のIAMロール
resource "aws_iam_role" "lambda_role" {
  name = "lambda-${var.app_name}-role"

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

# Lambda関数にS3アクセス権限を付与
resource "aws_iam_policy" "lambda_s3_access" {
  name        = "lambda-s3-access-policy"
  description = "Allow Lambda to access S3 bucket"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ]
        Effect = "Allow"
        Resource = [
          aws_s3_bucket.cloudpix_images.arn,
          "${aws_s3_bucket.cloudpix_images.arn}/*"
        ]
      }
    ]
  })
}

# Lambda関数にDynamoDBアクセス権限を付与
resource "aws_iam_policy" "lambda_dynamodb_access" {
  name        = "lambda-dynamodb-access-policy"
  description = "Allow Lambda to access DynamoDB table"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "dynamodb:PutItem",
          "dynamodb:GetItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem",
          "dynamodb:Query",
          "dynamodb:Scan"
        ]
        Effect = "Allow"
        Resource = [
          aws_dynamodb_table.cloudpix_metadata.arn,
          "${aws_dynamodb_table.cloudpix_metadata.arn}/index/*"
        ]
      }
    ]
  })
}

# IAMポリシーをLambdaロールにアタッチ
resource "aws_iam_role_policy_attachment" "lambda_s3" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.lambda_s3_access.arn
}

# IAMポリシーをLambdaロールにアタッチ
resource "aws_iam_role_policy_attachment" "lambda_dynamodb" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.lambda_dynamodb_access.arn
}

################################
# Docker Build & Push - Upload Function
################################
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

################################
# Upload Lambda Function
################################
# Lambda関数の作成
resource "aws_lambda_function" "cloudpix_upload" {
  function_name = "${var.app_name}-upload"
  role          = aws_iam_role.lambda_role.arn
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.cloudpix_upload.repository_url}:latest"

  timeout     = var.lambda_timeout
  memory_size = var.lambda_memory_size

  # 環境変数を追加
  environment {
    variables = {
      S3_BUCKET_NAME      = aws_s3_bucket.cloudpix_images.bucket
      DYNAMODB_TABLE_NAME = aws_dynamodb_table.cloudpix_metadata.name
    }
  }

  depends_on = [
    null_resource.docker_build_push
  ]
}

################################
# Docker Build & Push - List Function
################################
# イメージのビルドとプッシュを行うnull_resource (list関数用)
resource "null_resource" "docker_build_push_list" {
  depends_on = [aws_ecr_repository.cloudpix_list]

  # リポジトリURLが変更された場合、またはDockerfileが変更された場合に再実行
  triggers = {
    ecr_repository_url = aws_ecr_repository.cloudpix_list.repository_url
    dockerfile_hash    = filemd5("${path.module}/Dockerfile")
    main_go_hash       = filemd5("${path.module}/cmd/list/main.go")
    build_script_hash  = filemd5("${path.module}/build_and_push.sh")
  }

  # シェルスクリプトを実行し、ECRリポジトリURLを渡す
  provisioner "local-exec" {
    command = "cd ${path.module} && cp cmd/list/main.go cmd/upload/main.go.bak && cp cmd/list/main.go cmd/upload/main.go && chmod +x ${path.module}/build_and_push.sh && ${path.module}/build_and_push.sh ${aws_ecr_repository.cloudpix_list.repository_url} && mv cmd/upload/main.go.bak cmd/upload/main.go"
  }
}

################################
# List Lambda Function
################################
# 画像一覧取得用Lambda関数
resource "aws_lambda_function" "cloudpix_list" {
  function_name = "${var.app_name}-list"
  role          = aws_iam_role.lambda_role.arn
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.cloudpix_list.repository_url}:latest"

  timeout     = var.lambda_timeout
  memory_size = var.lambda_memory_size

  environment {
    variables = {
      DYNAMODB_TABLE_NAME = aws_dynamodb_table.cloudpix_metadata.name
    }
  }

  depends_on = [
    null_resource.docker_build_push_list
  ]
}

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
  authorization = "NONE"
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
    aws_api_gateway_integration.list_lambda_integration
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