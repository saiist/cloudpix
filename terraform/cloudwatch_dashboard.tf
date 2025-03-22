################################
# CloudWatch ダッシュボード
################################

resource "aws_cloudwatch_dashboard" "cloudpix_dashboard" {
  dashboard_name = "${var.app_name}-dashboard"

  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        x      = 0
        y      = 0
        width  = 12
        height = 6
        properties = {
          metrics = [
            ["AWS/Lambda", "Invocations", "FunctionName", "${aws_lambda_function.cloudpix_upload.function_name}", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["AWS/Lambda", "Invocations", "FunctionName", "${aws_lambda_function.cloudpix_list.function_name}", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["AWS/Lambda", "Invocations", "FunctionName", "${aws_lambda_function.cloudpix_thumbnail.function_name}", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["AWS/Lambda", "Invocations", "FunctionName", "${aws_lambda_function.cloudpix_tags.function_name}", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["AWS/Lambda", "Invocations", "FunctionName", "${aws_lambda_function.cloudpix_cleanup.function_name}", { stat = "Sum", period = var.dashboard_refresh_interval }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = var.aws_region
          title   = "Lambda関数の呼び出し回数"
          period  = var.dashboard_refresh_interval
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 0
        width  = 12
        height = 6
        properties = {
          metrics = [
            ["AWS/Lambda", "Duration", "FunctionName", "${aws_lambda_function.cloudpix_upload.function_name}", { stat = "Average", period = var.dashboard_refresh_interval }],
            ["AWS/Lambda", "Duration", "FunctionName", "${aws_lambda_function.cloudpix_list.function_name}", { stat = "Average", period = var.dashboard_refresh_interval }],
            ["AWS/Lambda", "Duration", "FunctionName", "${aws_lambda_function.cloudpix_thumbnail.function_name}", { stat = "Average", period = var.dashboard_refresh_interval }],
            ["AWS/Lambda", "Duration", "FunctionName", "${aws_lambda_function.cloudpix_tags.function_name}", { stat = "Average", period = var.dashboard_refresh_interval }],
            ["AWS/Lambda", "Duration", "FunctionName", "${aws_lambda_function.cloudpix_cleanup.function_name}", { stat = "Average", period = var.dashboard_refresh_interval }],
          ]
          view    = "timeSeries"
          stacked = false
          region  = var.aws_region
          title   = "Lambda関数の実行時間"
          period  = var.dashboard_refresh_interval
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 6
        width  = 12
        height = 6
        properties = {
          metrics = [
            ["AWS/Lambda", "Errors", "FunctionName", "${aws_lambda_function.cloudpix_upload.function_name}", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["AWS/Lambda", "Errors", "FunctionName", "${aws_lambda_function.cloudpix_list.function_name}", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["AWS/Lambda", "Errors", "FunctionName", "${aws_lambda_function.cloudpix_thumbnail.function_name}", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["AWS/Lambda", "Errors", "FunctionName", "${aws_lambda_function.cloudpix_tags.function_name}", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["AWS/Lambda", "Errors", "FunctionName", "${aws_lambda_function.cloudpix_cleanup.function_name}", { stat = "Sum", period = var.dashboard_refresh_interval }],
          ]
          view    = "timeSeries"
          stacked = false
          region  = var.aws_region
          title   = "Lambda関数のエラー数"
          period  = var.dashboard_refresh_interval
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 6
        width  = 12
        height = 6
        properties = {
          metrics = [
            ["AWS/Lambda", "Throttles", "FunctionName", "${aws_lambda_function.cloudpix_upload.function_name}", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["AWS/Lambda", "Throttles", "FunctionName", "${aws_lambda_function.cloudpix_list.function_name}", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["AWS/Lambda", "Throttles", "FunctionName", "${aws_lambda_function.cloudpix_thumbnail.function_name}", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["AWS/Lambda", "Throttles", "FunctionName", "${aws_lambda_function.cloudpix_tags.function_name}", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["AWS/Lambda", "Throttles", "FunctionName", "${aws_lambda_function.cloudpix_cleanup.function_name}", { stat = "Sum", period = var.dashboard_refresh_interval }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = var.aws_region
          title   = "Lambda関数のスロットリング数"
          period  = var.dashboard_refresh_interval
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 12
        width  = 12
        height = 6
        properties = {
          metrics = [
            ["AWS/ApiGateway", "Count", "ApiName", "${title(var.app_name)}-API", { stat = "Sum", period = var.dashboard_refresh_interval }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = var.aws_region
          title   = "API Gatewayのリクエスト数"
          period  = var.dashboard_refresh_interval
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 12
        width  = 12
        height = 6
        properties = {
          metrics = [
            ["AWS/ApiGateway", "Latency", "ApiName", "${title(var.app_name)}-API", { stat = "Average", period = var.dashboard_refresh_interval }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = var.aws_region
          title   = "API Gatewayのレイテンシー"
          period  = var.dashboard_refresh_interval
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 18
        width  = 12
        height = 6
        properties = {
          metrics = [
            ["AWS/DynamoDB", "ConsumedReadCapacityUnits", "TableName", "${aws_dynamodb_table.cloudpix_metadata.name}", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["AWS/DynamoDB", "ConsumedWriteCapacityUnits", "TableName", "${aws_dynamodb_table.cloudpix_metadata.name}", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["AWS/DynamoDB", "ConsumedReadCapacityUnits", "TableName", "${aws_dynamodb_table.cloudpix_tags.name}", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["AWS/DynamoDB", "ConsumedWriteCapacityUnits", "TableName", "${aws_dynamodb_table.cloudpix_tags.name}", { stat = "Sum", period = var.dashboard_refresh_interval }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = var.aws_region
          title   = "DynamoDBの消費キャパシティ"
          period  = var.dashboard_refresh_interval
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 18
        width  = 12
        height = 6
        properties = {
          metrics = [
            ["AWS/S3", "BucketSizeBytes", "BucketName", "${aws_s3_bucket.cloudpix_images.bucket}", "StorageType", "StandardStorage", { stat = "Maximum", period = 86400 }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = var.aws_region
          title   = "S3バケットのサイズ"
          period  = 86400
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 24
        width  = 24
        height = 6
        properties = {
          metrics = [
            ["CloudPix/Lambda", "Duration", "Service", "CloudPix", "Operation", "UploadImage", { stat = "Average", period = var.dashboard_refresh_interval }],
            ["CloudPix/Lambda", "Duration", "Service", "CloudPix", "Operation", "ListImages", { stat = "Average", period = var.dashboard_refresh_interval }],
            ["CloudPix/Lambda", "Duration", "Service", "CloudPix", "Operation", "TagManagement", { stat = "Average", period = var.dashboard_refresh_interval }],
            ["CloudPix/Lambda", "ProcessingTime", "Service", "CloudPix", "Operation", "ThumbnailProcessing", { stat = "Average", period = var.dashboard_refresh_interval }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = var.aws_region
          title   = "カスタムメトリクス - 処理時間"
          period  = var.dashboard_refresh_interval
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 30
        width  = 24
        height = 6
        properties = {
          metrics = [
            ["CloudPix/Lambda", "UserRequests", "Service", "CloudPix", { stat = "SampleCount", period = var.dashboard_refresh_interval }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = var.aws_region
          title   = "ユーザーリクエスト数"
          period  = var.dashboard_refresh_interval
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 36
        width  = 12
        height = 6
        properties = {
          metrics = [
            ["CloudPix/Logs", "ErrorCount", "Function", "upload", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["CloudPix/Logs", "ErrorCount", "Function", "list", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["CloudPix/Logs", "ErrorCount", "Function", "thumbnail", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["CloudPix/Logs", "ErrorCount", "Function", "tags", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["CloudPix/Logs", "ErrorCount", "Function", "cleanup", { stat = "Sum", period = var.dashboard_refresh_interval }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = var.aws_region
          title   = "構造化ログのエラー数"
          period  = var.dashboard_refresh_interval
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 36
        width  = 12
        height = 6
        properties = {
          metrics = [
            ["CloudPix/Logs", "LongDurationCount", "Function", "upload", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["CloudPix/Logs", "LongDurationCount", "Function", "list", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["CloudPix/Logs", "LongDurationCount", "Function", "thumbnail", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["CloudPix/Logs", "LongDurationCount", "Function", "tags", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["CloudPix/Logs", "LongDurationCount", "Function", "cleanup", { stat = "Sum", period = var.dashboard_refresh_interval }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = var.aws_region
          title   = "長時間実行リクエスト数"
          period  = var.dashboard_refresh_interval
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 42
        width  = 24
        height = 6
        properties = {
          metrics = [
            ["CloudPix/Operations", "OperationCount", "Operation", "UploadImage", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["CloudPix/Operations", "OperationCount", "Operation", "ListImages", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["CloudPix/Operations", "OperationCount", "Operation", "GenerateThumbnail", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["CloudPix/Operations", "OperationCount", "Operation", "ListTags", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["CloudPix/Operations", "OperationCount", "Operation", "AddTags", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["CloudPix/Operations", "OperationCount", "Operation", "GetImageTags", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["CloudPix/Operations", "OperationCount", "Operation", "RemoveTags", { stat = "Sum", period = var.dashboard_refresh_interval }]
          ]
          view    = "timeSeries"
          stacked = true
          region  = var.aws_region
          title   = "操作別リクエスト数"
          period  = var.dashboard_refresh_interval
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 48
        width  = 12
        height = 6
        properties = {
          metrics = [
            ["CloudPix/Security", "AuthErrorCount", "Function", "upload", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["CloudPix/Security", "AuthErrorCount", "Function", "list", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["CloudPix/Security", "AuthErrorCount", "Function", "tags", { stat = "Sum", period = var.dashboard_refresh_interval }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = var.aws_region
          title   = "認証エラー数"
          period  = var.dashboard_refresh_interval
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 48
        width  = 12
        height = 6
        properties = {
          metrics = [
            ["CloudPix/Responses", "Response4xxCount", "Function", "upload", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["CloudPix/Responses", "Response4xxCount", "Function", "list", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["CloudPix/Responses", "Response4xxCount", "Function", "tags", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["CloudPix/Responses", "Response5xxCount", "Function", "upload", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["CloudPix/Responses", "Response5xxCount", "Function", "list", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["CloudPix/Responses", "Response5xxCount", "Function", "tags", { stat = "Sum", period = var.dashboard_refresh_interval }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = var.aws_region
          title   = "エラーレスポンス数 (4xx/5xx)"
          period  = var.dashboard_refresh_interval
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 54
        width  = 24
        height = 6
        properties = {
          metrics = [
            ["CloudPix/Operations", "ArchivedImageCount", { stat = "Sum", period = var.dashboard_refresh_interval }],
            ["CloudPix/Operations", "CleanupProcessCount", { stat = "Sum", period = var.dashboard_refresh_interval }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = var.aws_region
          title   = "アーカイブ処理メトリクス"
          period  = var.dashboard_refresh_interval
        }
      }
    ]
  })
}