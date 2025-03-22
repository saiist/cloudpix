################################
# Cleanup Lambda + EventBridge (CloudWatch Events)
################################

# クリーンアップ用のECRリポジトリ
resource "aws_ecr_repository" "cloudpix_cleanup" {
  name                 = "${var.app_name}-cleanup"
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
resource "null_resource" "docker_build_push_cleanup" {
  depends_on = [aws_ecr_repository.cloudpix_cleanup]

  triggers = {
    ecr_repository_url = aws_ecr_repository.cloudpix_cleanup.repository_url
    dockerfile_hash    = filemd5("${path.module}/../Dockerfile")
    main_go_hash       = filemd5("${path.module}/../cmd/cleanup/main.go")
    build_script_hash  = filemd5("${path.module}/../build_and_push.sh")
  }

  provisioner "local-exec" {
    command = <<-EOT
      echo "Building tags function image..."
      cd ${path.module}/.. && \
      chmod +x build_and_push.sh && \
      REPO_NAME="cloudpix-cleanup" ./build_and_push.sh ${aws_ecr_repository.cloudpix_cleanup.repository_url} ./cmd/cleanup/main.go
    EOT
  }
}

################################
# Cleanup Lambda Function
################################
# クリーンアップ用Lambda関数
resource "aws_lambda_function" "cloudpix_cleanup" {
  function_name = "${var.app_name}-cleanup"
  role          = aws_iam_role.lambda_role.arn
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.cloudpix_cleanup.repository_url}:latest"

  timeout     = 300 # 長めに設定（5分）
  memory_size = var.lambda_memory_size

  environment {
    variables = local.cleanup_lambda_env_vars
  }

  depends_on = [
    null_resource.docker_build_push_cleanup
  ]

  # X-Rayトレースを有効化
  tracing_config {
    mode = "Active"
  }
}

# EventBridgeルール（毎日0時に実行）
resource "aws_cloudwatch_event_rule" "daily_cleanup" {
  name                = "${var.app_name}-daily-cleanup"
  description         = "古い画像を毎日クリーンアップ"
  schedule_expression = "cron(0 0 * * ? *)"
}

# EventBridgeターゲット（Lambda関数を呼び出す）
resource "aws_cloudwatch_event_target" "cleanup_lambda" {
  rule      = aws_cloudwatch_event_rule.daily_cleanup.name
  target_id = "TriggerCleanupLambda"
  arn       = aws_lambda_function.cloudpix_cleanup.arn
}

# Lambda実行権限
resource "aws_lambda_permission" "allow_eventbridge" {
  statement_id  = "AllowExecutionFromEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.cloudpix_cleanup.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.daily_cleanup.arn
}