################################
# CloudWatch Metrics & Monitoring
################################

# CloudWatch Logs グループ - 各Lambda関数用
resource "aws_cloudwatch_log_group" "lambda_upload_logs" {
  name              = "/aws/lambda/${aws_lambda_function.cloudpix_upload.function_name}"
  retention_in_days = 14

  # 既存のリソースをインポートするため、作成済みのリソースをスキップ
  lifecycle {
    prevent_destroy = true
    ignore_changes = [
      # 自動作成されたリソースの属性を無視
      name,
    ]
  }

  tags = {
    Environment = var.environment
    Application = var.app_name
  }
}

resource "aws_cloudwatch_log_group" "lambda_list_logs" {
  name              = "/aws/lambda/${aws_lambda_function.cloudpix_list.function_name}"
  retention_in_days = 14

  # 既存のリソースをインポートするため、作成済みのリソースをスキップ
  lifecycle {
    prevent_destroy = true
    ignore_changes = [
      # 自動作成されたリソースの属性を無視
      name,
    ]
  }

  tags = {
    Environment = var.environment
    Application = var.app_name
  }
}

resource "aws_cloudwatch_log_group" "lambda_thumbnail_logs" {
  name              = "/aws/lambda/${aws_lambda_function.cloudpix_thumbnail.function_name}"
  retention_in_days = 14

  # 既存のリソースをインポートするため、作成済みのリソースをスキップ
  lifecycle {
    prevent_destroy = true
    ignore_changes = [
      # 自動作成されたリソースの属性を無視
      name,
    ]
  }

  tags = {
    Environment = var.environment
    Application = var.app_name
  }
}

resource "aws_cloudwatch_log_group" "lambda_tags_logs" {
  name              = "/aws/lambda/${aws_lambda_function.cloudpix_tags.function_name}"
  retention_in_days = 14

  # 既存のリソースをインポートするため、作成済みのリソースをスキップ
  lifecycle {
    prevent_destroy = true
    ignore_changes = [
      # 自動作成されたリソースの属性を無視
      name,
    ]
  }

  tags = {
    Environment = var.environment
    Application = var.app_name
  }
}

################################
# CloudWatch ダッシュボード
################################
resource "aws_cloudwatch_dashboard" "cloudpix_dashboard" {
  dashboard_name = "${var.app_name}-dashboard"

  dashboard_body = <<EOF
{
  "widgets": [
    {
      "type": "metric",
      "x": 0,
      "y": 0,
      "width": 12,
      "height": 6,
      "properties": {
        "metrics": [
          [ "AWS/Lambda", "Invocations", "FunctionName", "${aws_lambda_function.cloudpix_upload.function_name}", { "stat": "Sum", "period": 300 } ],
          [ "AWS/Lambda", "Invocations", "FunctionName", "${aws_lambda_function.cloudpix_list.function_name}", { "stat": "Sum", "period": 300 } ],
          [ "AWS/Lambda", "Invocations", "FunctionName", "${aws_lambda_function.cloudpix_thumbnail.function_name}", { "stat": "Sum", "period": 300 } ],
          [ "AWS/Lambda", "Invocations", "FunctionName", "${aws_lambda_function.cloudpix_tags.function_name}", { "stat": "Sum", "period": 300 } ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "Lambda関数の呼び出し回数",
        "period": 300
      }
    },
    {
      "type": "metric",
      "x": 12,
      "y": 0,
      "width": 12,
      "height": 6,
      "properties": {
        "metrics": [
          [ "AWS/Lambda", "Duration", "FunctionName", "${aws_lambda_function.cloudpix_upload.function_name}", { "stat": "Average", "period": 300 } ],
          [ "AWS/Lambda", "Duration", "FunctionName", "${aws_lambda_function.cloudpix_list.function_name}", { "stat": "Average", "period": 300 } ],
          [ "AWS/Lambda", "Duration", "FunctionName", "${aws_lambda_function.cloudpix_thumbnail.function_name}", { "stat": "Average", "period": 300 } ],
          [ "AWS/Lambda", "Duration", "FunctionName", "${aws_lambda_function.cloudpix_tags.function_name}", { "stat": "Average", "period": 300 } ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "Lambda関数の実行時間",
        "period": 300
      }
    },
    {
      "type": "metric",
      "x": 0,
      "y": 6,
      "width": 12,
      "height": 6,
      "properties": {
        "metrics": [
          [ "AWS/Lambda", "Errors", "FunctionName", "${aws_lambda_function.cloudpix_upload.function_name}", { "stat": "Sum", "period": 300 } ],
          [ "AWS/Lambda", "Errors", "FunctionName", "${aws_lambda_function.cloudpix_list.function_name}", { "stat": "Sum", "period": 300 } ],
          [ "AWS/Lambda", "Errors", "FunctionName", "${aws_lambda_function.cloudpix_thumbnail.function_name}", { "stat": "Sum", "period": 300 } ],
          [ "AWS/Lambda", "Errors", "FunctionName", "${aws_lambda_function.cloudpix_tags.function_name}", { "stat": "Sum", "period": 300 } ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "Lambda関数のエラー数",
        "period": 300
      }
    },
    {
      "type": "metric",
      "x": 12,
      "y": 6,
      "width": 12,
      "height": 6,
      "properties": {
        "metrics": [
          [ "AWS/Lambda", "Throttles", "FunctionName", "${aws_lambda_function.cloudpix_upload.function_name}", { "stat": "Sum", "period": 300 } ],
          [ "AWS/Lambda", "Throttles", "FunctionName", "${aws_lambda_function.cloudpix_list.function_name}", { "stat": "Sum", "period": 300 } ],
          [ "AWS/Lambda", "Throttles", "FunctionName", "${aws_lambda_function.cloudpix_thumbnail.function_name}", { "stat": "Sum", "period": 300 } ],
          [ "AWS/Lambda", "Throttles", "FunctionName", "${aws_lambda_function.cloudpix_tags.function_name}", { "stat": "Sum", "period": 300 } ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "Lambda関数のスロットリング数",
        "period": 300
      }
    },
    {
      "type": "metric",
      "x": 0,
      "y": 12,
      "width": 12,
      "height": 6,
      "properties": {
        "metrics": [
          [ "AWS/ApiGateway", "Count", "ApiName", "${title(var.app_name)}-API", { "stat": "Sum", "period": 300 } ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "API Gatewayのリクエスト数",
        "period": 300
      }
    },
    {
      "type": "metric",
      "x": 12,
      "y": 12,
      "width": 12,
      "height": 6,
      "properties": {
        "metrics": [
          [ "AWS/ApiGateway", "Latency", "ApiName", "${title(var.app_name)}-API", { "stat": "Average", "period": 300 } ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "API Gatewayのレイテンシー",
        "period": 300
      }
    },
    {
      "type": "metric",
      "x": 0,
      "y": 18,
      "width": 12,
      "height": 6,
      "properties": {
        "metrics": [
          [ "AWS/DynamoDB", "ConsumedReadCapacityUnits", "TableName", "${aws_dynamodb_table.cloudpix_metadata.name}", { "stat": "Sum", "period": 300 } ],
          [ "AWS/DynamoDB", "ConsumedWriteCapacityUnits", "TableName", "${aws_dynamodb_table.cloudpix_metadata.name}", { "stat": "Sum", "period": 300 } ],
          [ "AWS/DynamoDB", "ConsumedReadCapacityUnits", "TableName", "${aws_dynamodb_table.cloudpix_tags.name}", { "stat": "Sum", "period": 300 } ],
          [ "AWS/DynamoDB", "ConsumedWriteCapacityUnits", "TableName", "${aws_dynamodb_table.cloudpix_tags.name}", { "stat": "Sum", "period": 300 } ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "DynamoDBの消費キャパシティ",
        "period": 300
      }
    },
    {
      "type": "metric",
      "x": 12,
      "y": 18,
      "width": 12,
      "height": 6,
      "properties": {
        "metrics": [
          [ "AWS/S3", "BucketSizeBytes", "BucketName", "${aws_s3_bucket.cloudpix_images.bucket}", "StorageType", "StandardStorage", { "stat": "Maximum", "period": 86400 } ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "S3バケットのサイズ",
        "period": 86400
      }
    }
  ]
}
EOF
}

################################
# CloudWatch アラーム
################################
# Lambda関数のエラーアラーム
resource "aws_cloudwatch_metric_alarm" "lambda_error_alarm" {
  for_each = {
    upload    = aws_lambda_function.cloudpix_upload.function_name
    list      = aws_lambda_function.cloudpix_list.function_name
    thumbnail = aws_lambda_function.cloudpix_thumbnail.function_name
    tags      = aws_lambda_function.cloudpix_tags.function_name
  }

  alarm_name          = "${each.value}-error-alarm"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "Errors"
  namespace           = "AWS/Lambda"
  period              = 300
  statistic           = "Sum"
  threshold           = 5
  alarm_description   = "Lambda関数 ${each.value} で5回以上のエラーが発生しました"
  treat_missing_data  = "notBreaching"

  dimensions = {
    FunctionName = each.value
  }

  alarm_actions = [aws_sns_topic.cloudpix_alerts.arn]
  ok_actions    = [aws_sns_topic.cloudpix_alerts.arn]
}

# Lambda関数の実行時間アラーム
resource "aws_cloudwatch_metric_alarm" "lambda_duration_alarm" {
  for_each = {
    upload    = aws_lambda_function.cloudpix_upload.function_name
    list      = aws_lambda_function.cloudpix_list.function_name
    thumbnail = aws_lambda_function.cloudpix_thumbnail.function_name
    tags      = aws_lambda_function.cloudpix_tags.function_name
  }

  alarm_name          = "${each.value}-duration-alarm"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "Duration"
  namespace           = "AWS/Lambda"
  period              = 300
  statistic           = "Average"
  threshold           = each.key == "thumbnail" ? 10000 : 3000 # サムネイル処理は長めに設定
  alarm_description   = "Lambda関数 ${each.value} の平均実行時間が閾値を超えました"
  treat_missing_data  = "notBreaching"

  dimensions = {
    FunctionName = each.value
  }

  alarm_actions = [aws_sns_topic.cloudpix_alerts.arn]
  ok_actions    = [aws_sns_topic.cloudpix_alerts.arn]
}

# API Gatewayの4xxエラーアラーム
resource "aws_cloudwatch_metric_alarm" "api_gateway_4xx_alarm" {
  alarm_name          = "${var.app_name}-api-4xx-error-alarm"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "4XXError"
  namespace           = "AWS/ApiGateway"
  period              = 300
  statistic           = "Sum"
  threshold           = 10
  alarm_description   = "API Gatewayで10回以上の4xxエラーが発生しました"
  treat_missing_data  = "notBreaching"

  dimensions = {
    ApiName = "${title(var.app_name)}-API"
  }

  alarm_actions = [aws_sns_topic.cloudpix_alerts.arn]
  ok_actions    = [aws_sns_topic.cloudpix_alerts.arn]
}

# API Gatewayの5xxエラーアラーム
resource "aws_cloudwatch_metric_alarm" "api_gateway_5xx_alarm" {
  alarm_name          = "${var.app_name}-api-5xx-error-alarm"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "5XXError"
  namespace           = "AWS/ApiGateway"
  period              = 300
  statistic           = "Sum"
  threshold           = 5
  alarm_description   = "API Gatewayで5回以上の5xxエラーが発生しました"
  treat_missing_data  = "notBreaching"

  dimensions = {
    ApiName = "${title(var.app_name)}-API"
  }

  alarm_actions = [aws_sns_topic.cloudpix_alerts.arn]
  ok_actions    = [aws_sns_topic.cloudpix_alerts.arn]
}

################################
# SNS Topic for Alerts
################################
resource "aws_sns_topic" "cloudpix_alerts" {
  name = "${var.app_name}-alerts"
}

# メールサブスクリプション - 変数を追加
variable "alert_email" {
  description = "アラート通知先のメールアドレス"
  type        = string
  default     = "admin@example.com"
}

resource "aws_sns_topic_subscription" "email_subscription" {
  topic_arn = aws_sns_topic.cloudpix_alerts.arn
  protocol  = "email"
  endpoint  = var.alert_email
}