################################
# CloudPix - クラウドベース画像管理システム
################################
# このプロジェクトは、S3、DynamoDB、Lambda、API Gatewayを使用して
# 画像のアップロードと管理を行うサーバーレスアプリケーションです。
#
# 主な機能：
# - 画像のアップロード（Base64形式またはプレサインドURL）
# - 画像一覧の取得（日付フィルタリング可能）
#
# リソースは以下のファイルに分割されています：
# - providers.tf - AWSプロバイダー設定
# - variables.tf - 設定変数
# - outputs.tf - 出力値
# - s3.tf - S3バケット関連リソース
# - dynamodb.tf - DynamoDBテーブル
# - ecr.tf - ECRリポジトリ
# - lambda.tf - Lambda関数とビルドプロセス
# - iam.tf - IAMロールとポリシー
# - api_gateway.tf - API Gateway設定