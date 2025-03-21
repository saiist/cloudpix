################################
# Docker Build & Push Functions
################################
# アップロード用Lambda関数のイメージのビルドとプッシュ
resource "null_resource" "docker_build_push_upload" {
  depends_on = [aws_ecr_repository.cloudpix_upload]

  # リポジトリURLが変更された場合、またはDockerfileが変更された場合に再実行
  triggers = {
    ecr_repository_url = aws_ecr_repository.cloudpix_upload.repository_url
    dockerfile_hash    = filemd5("${path.module}/../Dockerfile")
    main_go_hash       = filemd5("${path.module}/../cmd/upload/main.go")
    build_script_hash  = filemd5("${path.module}/../build_and_push.sh")
  }

  # シェルスクリプトを実行し、ECRリポジトリURLとビルドパスを渡す
  provisioner "local-exec" {
    command = <<-EOT
      echo "Building upload function image..."
      cd ${path.module}/.. && \
      chmod +x build_and_push.sh && \
      REPO_NAME="cloudpix-upload" ./build_and_push.sh ${aws_ecr_repository.cloudpix_upload.repository_url} ./cmd/upload/main.go
    EOT
  }
}

# リスト用Lambda関数のイメージのビルドとプッシュ
resource "null_resource" "docker_build_push_list" {
  depends_on = [aws_ecr_repository.cloudpix_list]

  # リポジトリURLが変更された場合、またはDockerfileが変更された場合に再実行
  triggers = {
    ecr_repository_url = aws_ecr_repository.cloudpix_list.repository_url
    dockerfile_hash    = filemd5("${path.module}/../Dockerfile")
    main_go_hash       = filemd5("${path.module}/../cmd/list/main.go")
    build_script_hash  = filemd5("${path.module}/../build_and_push.sh")
  }

  # シェルスクリプトを実行し、ECRリポジトリURLとビルドパスを渡す
  provisioner "local-exec" {
    command = <<-EOT
      echo "Building list function image..."
      cd ${path.module}/.. && \
      chmod +x build_and_push.sh && \
      REPO_NAME="cloudpix-list" ./build_and_push.sh ${aws_ecr_repository.cloudpix_list.repository_url} ./cmd/list/main.go
    EOT
  }
}

################################
# Lambda Functions
################################
# アップロード用Lambda関数
resource "aws_lambda_function" "cloudpix_upload" {
  function_name = "${var.app_name}-upload"
  role          = aws_iam_role.lambda_role.arn
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.cloudpix_upload.repository_url}:latest"

  timeout     = var.lambda_timeout
  memory_size = var.lambda_memory_size

  # 環境変数を追加
  environment {
    variables = local.upload_lambda_env_vars
  }

  depends_on = [
    null_resource.docker_build_push_upload
  ]

  # X-Rayトレースを有効化
  tracing_config {
    mode = "Active"
  }
}

# リスト用Lambda関数
resource "aws_lambda_function" "cloudpix_list" {
  function_name = "${var.app_name}-list"
  role          = aws_iam_role.lambda_role.arn
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.cloudpix_list.repository_url}:latest"

  timeout     = var.lambda_timeout
  memory_size = var.lambda_memory_size

  environment {
    variables = local.list_lambda_env_vars
  }

  depends_on = [
    null_resource.docker_build_push_list
  ]

  # X-Rayトレースを有効化
  tracing_config {
    mode = "Active"
  }
}