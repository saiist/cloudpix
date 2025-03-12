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

# Lambda関数にタグテーブルへのアクセス権限を付与
resource "aws_iam_policy" "lambda_tags_access" {
  name        = "lambda-tags-access-policy"
  description = "Allow Lambda to access Tags DynamoDB table"

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
          aws_dynamodb_table.cloudpix_tags.arn,
          "${aws_dynamodb_table.cloudpix_tags.arn}/index/*"
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

resource "aws_iam_role_policy_attachment" "lambda_dynamodb" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.lambda_dynamodb_access.arn
}

resource "aws_iam_role_policy_attachment" "lambda_tags" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.lambda_tags_access.arn
}