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