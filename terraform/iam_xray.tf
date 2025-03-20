################################
# X-Ray and CloudWatch Metrics Permission
################################

# Lambda関数にX-Rayアクセス権限を付与
resource "aws_iam_policy" "lambda_xray_access" {
  name        = "lambda-xray-access-policy"
  description = "Allow Lambda to send traces to X-Ray"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "xray:PutTraceSegments",
          "xray:PutTelemetryRecords",
          "xray:GetSamplingRules",
          "xray:GetSamplingTargets",
          "xray:GetSamplingStatisticSummaries"
        ]
        Effect   = "Allow"
        Resource = "*"
      }
    ]
  })
}

# Lambda関数にCloudWatch Metricsへの書き込み権限を付与
resource "aws_iam_policy" "lambda_cloudwatch_metrics_access" {
  name        = "lambda-cloudwatch-metrics-access-policy"
  description = "Allow Lambda to put metrics to CloudWatch"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "cloudwatch:PutMetricData"
        ]
        Effect   = "Allow"
        Resource = "*"
      }
    ]
  })
}

# IAMポリシーをLambdaロールにアタッチ
resource "aws_iam_role_policy_attachment" "lambda_xray" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.lambda_xray_access.arn
}

resource "aws_iam_role_policy_attachment" "lambda_cloudwatch_metrics" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.lambda_cloudwatch_metrics_access.arn
}