#!/bin/bash
set -e

# 引数からECRリポジトリURLを取得
ECR_REPO=$1
if [ -z "$ECR_REPO" ]; then
  echo "ECRリポジトリURLを指定してください"
  exit 1
fi

# 引数からビルドパスを取得（指定がなければデフォルト値を使用）
BUILD_PATH=${2:-"./cmd/upload/main.go"}
echo "ビルドするファイル: $BUILD_PATH"

# リージョンを設定
REGION="ap-northeast-1"
# カスタムリポジトリ名が指定されていなければデフォルト値を使用
REPO_NAME=${REPO_NAME:-"cloudpix-upload"}

echo "ECR Repository: $ECR_REPO"

# ECRにログイン
aws ecr get-login-password --region $REGION | docker login --username AWS --password-stdin $(echo $ECR_REPO | cut -d'/' -f1)

# イメージをビルド
echo "Dockerイメージをビルド中..."
docker buildx build --platform linux/amd64 --provenance=false \
  --build-arg BUILD_PATH=$BUILD_PATH \
  -t $REPO_NAME:latest .

# ECRリポジトリ用にタグ付け
docker tag $REPO_NAME:latest $ECR_REPO:latest

# イメージをプッシュ
echo "イメージをECRにプッシュ中..."
docker push $ECR_REPO:latest

# イメージURIを出力
echo "Image URI: $ECR_REPO:latest"