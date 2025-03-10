# CloudPix - サーバーレス画像管理システム

CloudPixは、AWSのサーバーレスサービスを活用した画像アップロードと管理を行うシステムです。画像のアップロード、保存、メタデータ管理、一覧取得の機能を備えています。

## アーキテクチャ図

![Architecture Diagram](architecture-diagram.svg)


## 主要コンポーネント

### 1. API Gateway
- RESTful APIエンドポイントを提供
- `/upload` - 画像アップロード用エンドポイント
- `/list` - 画像一覧取得用エンドポイント

### 2. Lambda 関数
- **cloudpix-upload** - 画像アップロード、S3保存、メタデータ登録を行う関数
- **cloudpix-list** - DynamoDBからメタデータを取得し画像一覧を提供する関数

### 3. S3バケット
- **cloudpix-images-{random_suffix}** - アップロードされた画像を保存

### 4. DynamoDBテーブル
- **cloudpix-metadata** - 画像のメタデータを保存
  - `ImageID` (パーティションキー) - 画像の一意識別子
  - `UploadDate` (GSIキー) - アップロード日付によるクエリを可能にする

### 5. ECRリポジトリ
- **cloudpix-upload** - アップロード関数用のコンテナイメージを格納
- **cloudpix-list** - 一覧表示関数用のコンテナイメージを格納

## データフロー

### 画像アップロードフロー
1. クライアントが `/upload` エンドポイントにPOSTリクエストを送信
2. API GatewayがリクエストをLambda関数に転送
3. Lambda関数が:
   - 画像データを受け取る (Base64エンコード形式)
   - S3バケットに画像を保存
   - メタデータを生成しDynamoDBに保存
4. 保存結果とダウンロードURLをクライアントに返却

### 画像一覧取得フロー
1. クライアントが `/list` エンドポイントにGETリクエストを送信
2. API GatewayがリクエストをLambda関数に転送
3. Lambda関数がDynamoDBからメタデータを取得
4. 画像メタデータの一覧をクライアントに返却

## プロジェクト構成

```
cloudpix/
  ├── terraform/           # Terraformによるインフラ定義
  │   ├── api_gateway.tf   # API Gateway定義
  │   ├── dynamodb.tf      # DynamoDBテーブル定義
  │   ├── ecr.tf           # ECRリポジトリ定義
  │   ├── iam.tf           # IAMロールとポリシー定義
  │   ├── lambda.tf        # Lambda関数定義
  │   ├── main.tf          # メインのTerraform設定
  │   ├── outputs.tf       # 出力値の定義
  │   ├── providers.tf     # プロバイダー設定
  │   ├── s3.tf            # S3バケット設定
  │   └── variables.tf     # 変数定義
  ├── cmd/                 # Goのソースコード
  │   ├── upload/          # アップロード関数
  │   │   └── main.go
  │   └── list/            # 一覧取得関数
  │       └── main.go
  ├── Dockerfile           # コンテナイメージ定義
  ├── build_and_push.sh    # イメージビルドスクリプト
  ├── go.mod               # Go依存関係
  └── Makefile             # タスク自動化
```

## 実装機能

- **画像アップロード** - Base64エンコードされた画像データをアップロード
- **プレサインドURL** - S3への直接アップロード用URLの生成
- **メタデータ管理** - 画像のファイル名、サイズ、コンテンツタイプなどを管理
- **画像一覧取得** - アップロードされた画像の一覧取得
- **日付フィルタリング** - アップロード日付による画像の絞り込み

## 今後の拡張予定

- 画像処理（サムネイル生成、リサイズ）機能
- 画像タグ付け機能
- ユーザー認証
- フロントエンドインターフェース
- 画像検索機能
