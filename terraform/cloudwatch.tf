################################
# CloudWatch Metrics & Monitoring
################################

# CloudWatch Logs グループ - 各Lambda関数用
resource "aws_cloudwatch_log_group" "lambda_upload_logs" {
  name              = "/aws/lambda/${aws_lambda_function.cloudpix_upload.function_name}"
  retention_in_days = var.metrics_retention_days

  # destroy できるようにコメントアウト
  # # 既存のリソースをインポートするため、作成済みのリソースをスキップ
  # lifecycle {
  #   prevent_destroy = true
  #   ignore_changes = [
  #     # 自動作成されたリソースの属性を無視
  #     name,
  #   ]
  # }

  tags = {
    Environment = var.environment
    Application = var.app_name
  }
}

resource "aws_cloudwatch_log_group" "lambda_list_logs" {
  name              = "/aws/lambda/${aws_lambda_function.cloudpix_list.function_name}"
  retention_in_days = var.metrics_retention_days

  # destroy できるようにコメントアウト
  # # 既存のリソースをインポートするため、作成済みのリソースをスキップ
  # lifecycle {
  #   prevent_destroy = true
  #   ignore_changes = [
  #     # 自動作成されたリソースの属性を無視
  #     name,
  #   ]
  # }

  tags = {
    Environment = var.environment
    Application = var.app_name
  }
}

resource "aws_cloudwatch_log_group" "lambda_thumbnail_logs" {
  name              = "/aws/lambda/${aws_lambda_function.cloudpix_thumbnail.function_name}"
  retention_in_days = var.metrics_retention_days

  # destroy できるようにコメントアウト
  # # 既存のリソースをインポートするため、作成済みのリソースをスキップ
  # lifecycle {
  #   prevent_destroy = true
  #   ignore_changes = [
  #     # 自動作成されたリソースの属性を無視
  #     name,
  #   ]
  # }

  tags = {
    Environment = var.environment
    Application = var.app_name
  }
}

resource "aws_cloudwatch_log_group" "lambda_tags_logs" {
  name              = "/aws/lambda/${aws_lambda_function.cloudpix_tags.function_name}"
  retention_in_days = var.metrics_retention_days

  # destroy できるようにコメントアウト
  # # 既存のリソースをインポートするため、作成済みのリソースをスキップ
  # lifecycle {
  #   prevent_destroy = true
  #   ignore_changes = [
  #     # 自動作成されたリソースの属性を無視
  #     name,
  #   ]
  # }

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
          [ "AWS/Lambda", "Invocations", "FunctionName", "${aws_lambda_function.cloudpix_upload.function_name}", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "AWS/Lambda", "Invocations", "FunctionName", "${aws_lambda_function.cloudpix_list.function_name}", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "AWS/Lambda", "Invocations", "FunctionName", "${aws_lambda_function.cloudpix_thumbnail.function_name}", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "AWS/Lambda", "Invocations", "FunctionName", "${aws_lambda_function.cloudpix_tags.function_name}", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "Lambda関数の呼び出し回数",
        "period": ${var.dashboard_refresh_interval}
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
          [ "AWS/Lambda", "Duration", "FunctionName", "${aws_lambda_function.cloudpix_upload.function_name}", { "stat": "Average", "period": ${var.dashboard_refresh_interval} } ],
          [ "AWS/Lambda", "Duration", "FunctionName", "${aws_lambda_function.cloudpix_list.function_name}", { "stat": "Average", "period": ${var.dashboard_refresh_interval} } ],
          [ "AWS/Lambda", "Duration", "FunctionName", "${aws_lambda_function.cloudpix_thumbnail.function_name}", { "stat": "Average", "period": ${var.dashboard_refresh_interval} } ],
          [ "AWS/Lambda", "Duration", "FunctionName", "${aws_lambda_function.cloudpix_tags.function_name}", { "stat": "Average", "period": ${var.dashboard_refresh_interval} } ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "Lambda関数の実行時間",
        "period": ${var.dashboard_refresh_interval}
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
          [ "AWS/Lambda", "Errors", "FunctionName", "${aws_lambda_function.cloudpix_upload.function_name}", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "AWS/Lambda", "Errors", "FunctionName", "${aws_lambda_function.cloudpix_list.function_name}", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "AWS/Lambda", "Errors", "FunctionName", "${aws_lambda_function.cloudpix_thumbnail.function_name}", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "AWS/Lambda", "Errors", "FunctionName", "${aws_lambda_function.cloudpix_tags.function_name}", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "Lambda関数のエラー数",
        "period": ${var.dashboard_refresh_interval}
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
          [ "AWS/Lambda", "Throttles", "FunctionName", "${aws_lambda_function.cloudpix_upload.function_name}", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "AWS/Lambda", "Throttles", "FunctionName", "${aws_lambda_function.cloudpix_list.function_name}", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "AWS/Lambda", "Throttles", "FunctionName", "${aws_lambda_function.cloudpix_thumbnail.function_name}", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "AWS/Lambda", "Throttles", "FunctionName", "${aws_lambda_function.cloudpix_tags.function_name}", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "Lambda関数のスロットリング数",
        "period": ${var.dashboard_refresh_interval}
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
          [ "AWS/ApiGateway", "Count", "ApiName", "${title(var.app_name)}-API", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "API Gatewayのリクエスト数",
        "period": ${var.dashboard_refresh_interval}
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
          [ "AWS/ApiGateway", "Latency", "ApiName", "${title(var.app_name)}-API", { "stat": "Average", "period": ${var.dashboard_refresh_interval} } ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "API Gatewayのレイテンシー",
        "period": ${var.dashboard_refresh_interval}
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
          [ "AWS/DynamoDB", "ConsumedReadCapacityUnits", "TableName", "${aws_dynamodb_table.cloudpix_metadata.name}", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "AWS/DynamoDB", "ConsumedWriteCapacityUnits", "TableName", "${aws_dynamodb_table.cloudpix_metadata.name}", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "AWS/DynamoDB", "ConsumedReadCapacityUnits", "TableName", "${aws_dynamodb_table.cloudpix_tags.name}", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "AWS/DynamoDB", "ConsumedWriteCapacityUnits", "TableName", "${aws_dynamodb_table.cloudpix_tags.name}", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "DynamoDBの消費キャパシティ",
        "period": ${var.dashboard_refresh_interval}
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
    },
    {
      "type": "metric",
      "x": 0,
      "y": 24,
      "width": 24,
      "height": 6,
      "properties": {
        "metrics": [
          [ "CloudPix/Lambda", "Duration", "Service", "CloudPix", "Operation", "UploadImage", { "stat": "Average", "period": ${var.dashboard_refresh_interval} } ],
          [ "CloudPix/Lambda", "Duration", "Service", "CloudPix", "Operation", "ListImages", { "stat": "Average", "period": ${var.dashboard_refresh_interval} } ],
          [ "CloudPix/Lambda", "Duration", "Service", "CloudPix", "Operation", "TagManagement", { "stat": "Average", "period": ${var.dashboard_refresh_interval} } ],
          [ "CloudPix/Lambda", "ProcessingTime", "Service", "CloudPix", "Operation", "ThumbnailProcessing", { "stat": "Average", "period": ${var.dashboard_refresh_interval} } ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "カスタムメトリクス - 処理時間",
        "period": ${var.dashboard_refresh_interval}
      }
    },
    {
      "type": "metric",
      "x": 0,
      "y": 30,
      "width": 24,
      "height": 6,
      "properties": {
        "metrics": [
          [ "CloudPix/Lambda", "UserRequests", "Service", "CloudPix", { "stat": "SampleCount", "period": ${var.dashboard_refresh_interval} } ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "ユーザーリクエスト数",
        "period": ${var.dashboard_refresh_interval}
      }
    },
    {
      "type": "metric",
      "x": 0,
      "y": 36,
      "width": 12,
      "height": 6,
      "properties": {
        "metrics": [
          [ "CloudPix/Logs", "ErrorCount", "Function", "upload", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "CloudPix/Logs", "ErrorCount", "Function", "list", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "CloudPix/Logs", "ErrorCount", "Function", "thumbnail", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "CloudPix/Logs", "ErrorCount", "Function", "tags", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "構造化ログのエラー数",
        "period": ${var.dashboard_refresh_interval}
      }
    },
    {
      "type": "metric",
      "x": 12,
      "y": 36,
      "width": 12,
      "height": 6,
      "properties": {
        "metrics": [
          [ "CloudPix/Logs", "LongDurationCount", "Function", "upload", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "CloudPix/Logs", "LongDurationCount", "Function", "list", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "CloudPix/Logs", "LongDurationCount", "Function", "thumbnail", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "CloudPix/Logs", "LongDurationCount", "Function", "tags", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "長時間実行リクエスト数",
        "period": ${var.dashboard_refresh_interval}
      }
    },
    {
      "type": "metric",
      "x": 0,
      "y": 42,
      "width": 24,
      "height": 6,
      "properties": {
        "metrics": [
          [ "CloudPix/Operations", "OperationCount", "Operation", "UploadImage", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "CloudPix/Operations", "OperationCount", "Operation", "ListImages", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "CloudPix/Operations", "OperationCount", "Operation", "GenerateThumbnail", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "CloudPix/Operations", "OperationCount", "Operation", "ListTags", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "CloudPix/Operations", "OperationCount", "Operation", "AddTags", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "CloudPix/Operations", "OperationCount", "Operation", "GetImageTags", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "CloudPix/Operations", "OperationCount", "Operation", "RemoveTags", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ]
        ],
        "view": "timeSeries",
        "stacked": true,
        "region": "${var.aws_region}",
        "title": "操作別リクエスト数",
        "period": ${var.dashboard_refresh_interval}
      }
    },
    {
      "type": "metric",
      "x": 0,
      "y": 48,
      "width": 12,
      "height": 6,
      "properties": {
        "metrics": [
          [ "CloudPix/Security", "AuthErrorCount", "Function", "upload", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "CloudPix/Security", "AuthErrorCount", "Function", "list", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "CloudPix/Security", "AuthErrorCount", "Function", "tags", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "認証エラー数",
        "period": ${var.dashboard_refresh_interval}
      }
    },
    {
      "type": "metric",
      "x": 12,
      "y": 48,
      "width": 12,
      "height": 6,
      "properties": {
        "metrics": [
          [ "CloudPix/Responses", "Response4xxCount", "Function", "upload", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "CloudPix/Responses", "Response4xxCount", "Function", "list", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "CloudPix/Responses", "Response4xxCount", "Function", "tags", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "CloudPix/Responses", "Response5xxCount", "Function", "upload", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "CloudPix/Responses", "Response5xxCount", "Function", "list", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ],
          [ "CloudPix/Responses", "Response5xxCount", "Function", "tags", { "stat": "Sum", "period": ${var.dashboard_refresh_interval} } ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "エラーレスポンス数 (4xx/5xx)",
        "period": ${var.dashboard_refresh_interval}
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
  threshold           = var.lambda_error_threshold
  alarm_description   = "Lambda関数 ${each.value} で${var.lambda_error_threshold}回以上のエラーが発生しました"
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
    upload    = { name = aws_lambda_function.cloudpix_upload.function_name, threshold = var.lambda_duration_threshold_base }
    list      = { name = aws_lambda_function.cloudpix_list.function_name, threshold = var.lambda_duration_threshold_base }
    thumbnail = { name = aws_lambda_function.cloudpix_thumbnail.function_name, threshold = var.lambda_duration_threshold_thumbnail }
    tags      = { name = aws_lambda_function.cloudpix_tags.function_name, threshold = var.lambda_duration_threshold_base }
  }

  alarm_name          = "${each.value.name}-duration-alarm"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "Duration"
  namespace           = "AWS/Lambda"
  period              = 300
  statistic           = "Average"
  threshold           = each.value.threshold
  alarm_description   = "Lambda関数 ${each.value.name} の平均実行時間が${each.value.threshold}msを超えました"
  treat_missing_data  = "notBreaching"

  dimensions = {
    FunctionName = each.value.name
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
  threshold           = var.api_4xx_error_threshold
  alarm_description   = "API Gatewayで${var.api_4xx_error_threshold}回以上の4xxエラーが発生しました"
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
  threshold           = var.api_5xx_error_threshold
  alarm_description   = "API Gatewayで${var.api_5xx_error_threshold}回以上の5xxエラーが発生しました"
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

resource "aws_sns_topic_subscription" "email_subscription" {
  topic_arn = aws_sns_topic.cloudpix_alerts.arn
  protocol  = "email"
  endpoint  = var.alarm_email
}

################################
# CloudWatch Logs Metrics Filters
################################

# エラーログからのメトリクスフィルター（Lambda関数別）
resource "aws_cloudwatch_log_metric_filter" "upload_error_logs" {
  name           = "${var.app_name}-upload-error-logs"
  pattern        = "{ $.level = \"error\" }"
  log_group_name = aws_cloudwatch_log_group.lambda_upload_logs.name

  metric_transformation {
    name      = "UploadErrorCount"
    namespace = "CloudPix/Logs"
    value     = "1"
  }
}

resource "aws_cloudwatch_log_metric_filter" "list_error_logs" {
  name           = "${var.app_name}-list-error-logs"
  pattern        = "{ $.level = \"error\" }"
  log_group_name = aws_cloudwatch_log_group.lambda_list_logs.name

  metric_transformation {
    name      = "ListErrorCount"
    namespace = "CloudPix/Logs"
    value     = "1"
  }
}

resource "aws_cloudwatch_log_metric_filter" "thumbnail_error_logs" {
  name           = "${var.app_name}-thumbnail-error-logs"
  pattern        = "{ $.level = \"error\" }"
  log_group_name = aws_cloudwatch_log_group.lambda_thumbnail_logs.name

  metric_transformation {
    name      = "ThumbnailErrorCount"
    namespace = "CloudPix/Logs"
    value     = "1"
  }
}

resource "aws_cloudwatch_log_metric_filter" "tags_error_logs" {
  name           = "${var.app_name}-tags-error-logs"
  pattern        = "{ $.level = \"error\" }"
  log_group_name = aws_cloudwatch_log_group.lambda_tags_logs.name

  metric_transformation {
    name      = "TagsErrorCount"
    namespace = "CloudPix/Logs"
    value     = "1"
  }
}

# 処理時間が長いリクエストの検出（Lambda関数別）
resource "aws_cloudwatch_log_metric_filter" "upload_long_duration" {
  name           = "${var.app_name}-upload-long-duration"
  pattern        = "{ $.duration > 1000 }"
  log_group_name = aws_cloudwatch_log_group.lambda_upload_logs.name

  metric_transformation {
    name      = "UploadLongDurationCount"
    namespace = "CloudPix/Logs"
    value     = "1"
  }
}

resource "aws_cloudwatch_log_metric_filter" "list_long_duration" {
  name           = "${var.app_name}-list-long-duration"
  pattern        = "{ $.duration > 500 }"
  log_group_name = aws_cloudwatch_log_group.lambda_list_logs.name

  metric_transformation {
    name      = "ListLongDurationCount"
    namespace = "CloudPix/Logs"
    value     = "1"
  }
}

resource "aws_cloudwatch_log_metric_filter" "thumbnail_long_duration" {
  name           = "${var.app_name}-thumbnail-long-duration"
  pattern        = "{ $.duration > 5000 }"
  log_group_name = aws_cloudwatch_log_group.lambda_thumbnail_logs.name

  metric_transformation {
    name      = "ThumbnailLongDurationCount"
    namespace = "CloudPix/Logs"
    value     = "1"
  }
}

resource "aws_cloudwatch_log_metric_filter" "tags_long_duration" {
  name           = "${var.app_name}-tags-long-duration"
  pattern        = "{ $.duration > 500 }"
  log_group_name = aws_cloudwatch_log_group.lambda_tags_logs.name

  metric_transformation {
    name      = "TagsLongDurationCount"
    namespace = "CloudPix/Logs"
    value     = "1"
  }
}

# 操作別メトリクスカウント（主要な操作のみ）
resource "aws_cloudwatch_log_metric_filter" "upload_image_count" {
  name           = "${var.app_name}-upload-image-count"
  pattern        = "{ $.operation = \"UploadImage\" && $.level = \"info\" && $.message = \"Request: * *\" }"
  log_group_name = aws_cloudwatch_log_group.lambda_upload_logs.name

  metric_transformation {
    name      = "UploadImageCount"
    namespace = "CloudPix/Operations"
    value     = "1"
  }
}

resource "aws_cloudwatch_log_metric_filter" "list_images_count" {
  name           = "${var.app_name}-list-images-count"
  pattern        = "{ $.operation = \"ListImages\" && $.level = \"info\" && $.message = \"Request: * *\" }"
  log_group_name = aws_cloudwatch_log_group.lambda_list_logs.name

  metric_transformation {
    name      = "ListImagesCount"
    namespace = "CloudPix/Operations"
    value     = "1"
  }
}

resource "aws_cloudwatch_log_metric_filter" "thumbnail_count" {
  name           = "${var.app_name}-thumbnail-count"
  pattern        = "{ $.operation = \"GenerateThumbnail\" && $.level = \"info\" }"
  log_group_name = aws_cloudwatch_log_group.lambda_thumbnail_logs.name

  metric_transformation {
    name      = "ThumbnailGenerationCount"
    namespace = "CloudPix/Operations"
    value     = "1"
  }
}

# 認証エラーの検出
resource "aws_cloudwatch_log_metric_filter" "auth_errors" {
  name           = "${var.app_name}-auth-errors"
  pattern        = "{ $.message = \"Authentication error\" || $.message = \"Authentication failed\" }"
  log_group_name = aws_cloudwatch_log_group.lambda_upload_logs.name

  metric_transformation {
    name      = "AuthErrorCount"
    namespace = "CloudPix/Security"
    value     = "1"
  }
}

# DynamoDB関連エラーの検出
resource "aws_cloudwatch_log_metric_filter" "dynamodb_errors" {
  name           = "${var.app_name}-dynamodb-errors"
  pattern        = "{ $.errorType = \"*DB*\" || $.errorType = \"*Dynamo*\" || $.errorMsg = \"*dynamodb*\" }"
  log_group_name = aws_cloudwatch_log_group.lambda_upload_logs.name

  metric_transformation {
    name      = "DynamoDBErrorCount"
    namespace = "CloudPix/Errors"
    value     = "1"
  }
}

# S3関連エラーの検出
resource "aws_cloudwatch_log_metric_filter" "s3_errors" {
  name           = "${var.app_name}-s3-errors"
  pattern        = "{ $.errorType = \"*S3*\" || $.errorMsg = \"*s3*\" || $.errorMsg = \"*bucket*\" }"
  log_group_name = aws_cloudwatch_log_group.lambda_upload_logs.name

  metric_transformation {
    name      = "S3ErrorCount"
    namespace = "CloudPix/Errors"
    value     = "1"
  }
}

# 4xxレスポンスコードの検出
resource "aws_cloudwatch_log_metric_filter" "upload_response_4xx" {
  name           = "${var.app_name}-upload-4xx-responses"
  pattern        = "{ $.message = \"Response: *\" && $.statusCode >= 400 && $.statusCode < 500 }"
  log_group_name = aws_cloudwatch_log_group.lambda_upload_logs.name

  metric_transformation {
    name      = "Upload4xxCount"
    namespace = "CloudPix/Responses"
    value     = "1"
  }
}

# 5xxレスポンスコードの検出
resource "aws_cloudwatch_log_metric_filter" "upload_response_5xx" {
  name           = "${var.app_name}-upload-5xx-responses"
  pattern        = "{ $.message = \"Response: *\" && $.statusCode >= 500 }"
  log_group_name = aws_cloudwatch_log_group.lambda_upload_logs.name

  metric_transformation {
    name      = "Upload5xxCount"
    namespace = "CloudPix/Responses"
    value     = "1"
  }
}


################################
# CloudWatch Metrics Alarm
################################
# ログベースのアラーム設定
resource "aws_cloudwatch_metric_alarm" "auth_errors_alarm" {
  alarm_name          = "${var.app_name}-auth-errors-alarm"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "AuthErrorCount"
  namespace           = "CloudPix/Security"
  period              = 300
  statistic           = "Sum"
  threshold           = 5
  alarm_description   = "5回以上の認証エラーが発生しました。不正アクセスの可能性があります。"
  treat_missing_data  = "notBreaching"

  alarm_actions = [aws_sns_topic.cloudpix_alerts.arn]
  ok_actions    = [aws_sns_topic.cloudpix_alerts.arn]
}

resource "aws_cloudwatch_metric_alarm" "dynamodb_errors_alarm" {
  alarm_name          = "${var.app_name}-dynamodb-errors-alarm"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "DynamoDBErrorCount"
  namespace           = "CloudPix/Errors"
  period              = 300
  statistic           = "Sum"
  threshold           = 3
  alarm_description   = "DynamoDBへのアクセスで連続したエラーが発生しています。"
  treat_missing_data  = "notBreaching"

  alarm_actions = [aws_sns_topic.cloudpix_alerts.arn]
  ok_actions    = [aws_sns_topic.cloudpix_alerts.arn]
}

resource "aws_cloudwatch_metric_alarm" "s3_errors_alarm" {
  alarm_name          = "${var.app_name}-s3-errors-alarm"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "S3ErrorCount"
  namespace           = "CloudPix/Errors"
  period              = 300
  statistic           = "Sum"
  threshold           = 3
  alarm_description   = "S3バケットへのアクセスで連続したエラーが発生しています。"
  treat_missing_data  = "notBreaching"

  alarm_actions = [aws_sns_topic.cloudpix_alerts.arn]
  ok_actions    = [aws_sns_topic.cloudpix_alerts.arn]
}

resource "aws_cloudwatch_metric_alarm" "upload_long_duration_alarm" {
  alarm_name          = "${var.app_name}-upload-long-duration-alarm"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "UploadLongDurationCount"
  namespace           = "CloudPix/Logs"
  period              = 300
  statistic           = "Sum"
  threshold           = 5
  alarm_description   = "アップロード処理で1000ms以上の処理時間が5回以上発生しました。"
  treat_missing_data  = "notBreaching"

  alarm_actions = [aws_sns_topic.cloudpix_alerts.arn]
  ok_actions    = [aws_sns_topic.cloudpix_alerts.arn]
}

resource "aws_cloudwatch_metric_alarm" "list_long_duration_alarm" {
  alarm_name          = "${var.app_name}-list-long-duration-alarm"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ListLongDurationCount"
  namespace           = "CloudPix/Logs"
  period              = 300
  statistic           = "Sum"
  threshold           = 5
  alarm_description   = "リスト取得処理で500ms以上の処理時間が5回以上発生しました。"
  treat_missing_data  = "notBreaching"

  alarm_actions = [aws_sns_topic.cloudpix_alerts.arn]
  ok_actions    = [aws_sns_topic.cloudpix_alerts.arn]
}

resource "aws_cloudwatch_metric_alarm" "upload_response_5xx_alarm" {
  alarm_name          = "${var.app_name}-upload-5xx-responses-alarm"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "Upload5xxCount"
  namespace           = "CloudPix/Responses"
  period              = 300
  statistic           = "Sum"
  threshold           = 3
  alarm_description   = "アップロード処理で3回以上の5xxエラーレスポンスが発生しました。"
  treat_missing_data  = "notBreaching"

  alarm_actions = [aws_sns_topic.cloudpix_alerts.arn]
  ok_actions    = [aws_sns_topic.cloudpix_alerts.arn]
}