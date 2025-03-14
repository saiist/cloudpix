################################
# ECR Repository for Thumbnail
################################
# サムネイル生成用のECRリポジトリ
resource "aws_ecr_repository" "cloudpix_thumbnail" {
  name                 = "${var.app_name}-thumbnail"
  image_tag_mutability = "MUTABLE"
  force_delete         = true

  image_scanning_configuration {
    scan_on_push = true
  }
}

################################
# Docker Build & Push - Thumbnail
################################
# サムネイル生成関数のイメージのビルドとプッシュ
resource "null_resource" "docker_build_push_thumbnail" {
  depends_on = [aws_ecr_repository.cloudpix_thumbnail]

  # リポジトリURLが変更された場合、またはDockerfileが変更された場合に再実行
  triggers = {
    ecr_repository_url = aws_ecr_repository.cloudpix_thumbnail.repository_url
    dockerfile_hash    = filemd5("${path.module}/../Dockerfile")
    main_go_hash       = filemd5("${path.module}/../cmd/thumbnail/main.go")
    build_script_hash  = filemd5("${path.module}/../build_and_push.sh")
  }

  # シェルスクリプトを実行し、ECRリポジトリURLとビルドパスを渡す
  provisioner "local-exec" {
    command = <<-EOT
      echo "Building thumbnail function image..."
      cd ${path.module}/.. && \
      chmod +x build_and_push.sh && \
      REPO_NAME="cloudpix-thumbnail" ./build_and_push.sh ${aws_ecr_repository.cloudpix_thumbnail.repository_url} ./cmd/thumbnail/main.go
    EOT
  }
}

################################
# Thumbnail Lambda Function
################################
# サムネイル生成用Lambda関数
resource "aws_lambda_function" "cloudpix_thumbnail" {
  function_name = "${var.app_name}-thumbnail"
  role          = aws_iam_role.lambda_role.arn
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.cloudpix_thumbnail.repository_url}:latest"

  timeout     = 60  # サムネイル生成は時間がかかるため、長めのタイムアウトを設定
  memory_size = 512 # 画像処理には多めのメモリが必要

  environment {
    variables = {
      S3_BUCKET_NAME      = aws_s3_bucket.cloudpix_images.bucket
      METADATA_TABLE_NAME = aws_dynamodb_table.cloudpix_metadata.name
    }
  }

  depends_on = [
    null_resource.docker_build_push_thumbnail
  ]
}

################################
# S3 Event Notification
################################
# S3バケットからLambda関数へのイベント通知
resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = aws_s3_bucket.cloudpix_images.id

  lambda_function {
    lambda_function_arn = aws_lambda_function.cloudpix_thumbnail.arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "uploads/"
    filter_suffix       = ".jpg"
  }

  lambda_function {
    lambda_function_arn = aws_lambda_function.cloudpix_thumbnail.arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "uploads/"
    filter_suffix       = ".jpeg"
  }

  lambda_function {
    lambda_function_arn = aws_lambda_function.cloudpix_thumbnail.arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "uploads/"
    filter_suffix       = ".png"
  }

  depends_on = [
    aws_lambda_permission.allow_bucket
  ]
}

# S3バケットからLambda関数を呼び出す権限
resource "aws_lambda_permission" "allow_bucket" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.cloudpix_thumbnail.function_name
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.cloudpix_images.arn
}