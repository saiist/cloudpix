.PHONY: deploy update-code test test-list test-list-date tf-init tf-plan tf-apply tf-destroy tf-validate tf-fmt tf-clean

# Terraformのディレクトリ
TF_DIR = terraform
# Terraformコマンドのプレフィックス
TF_CMD = cd $(TF_DIR) && terraform

# 完全デプロイ（Terraformを含む）
deploy: tf-apply

# Terraformの出力を取得するヘルパー関数
define tf_output
$(shell $(TF_CMD) output -raw $(1))
endef

# アップロードのコードのみを更新
update-code:
	$(eval ECR_REPO := $(call tf_output,ecr_repository_url))
	./build_and_push.sh $(ECR_REPO) ./cmd/upload/main.go
	aws lambda update-function-code \
	  --function-name cloudpix-upload \
	  --image-uri $(ECR_REPO):latest

# リストのコードを更新
update-list-code:
	$(eval ECR_REPO := $(call tf_output,ecr_list_repository_url))
	./build_and_push.sh $(ECR_REPO) ./cmd/list/main.go
	aws lambda update-function-code \
	  --function-name cloudpix-list \
	  --image-uri $(ECR_REPO):latest

# サムネイル生成コードの更新
update-thumbnail-code:
	$(eval ECR_REPO := $(call tf_output,ecr_thumbnail_repository_url))
	./build_and_push.sh $(ECR_REPO) ./cmd/thumbnail/main.go
	aws lambda update-function-code \
	  --function-name cloudpix-thumbnail \
	  --image-uri $(ECR_REPO):latest

# APIのテスト (Base64形式での画像アップロード)
test:
	$(eval API_URL := $(call tf_output,api_url))
	# サンプルの小さなPNG画像をBase64エンコードしてアップロード
	echo "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==" > /tmp/test_base64.txt
	curl -X POST $(API_URL) \
	  -H "Content-Type: application/json" \
	  -d "{\"fileName\":\"test.png\",\"contentType\":\"image/png\",\"data\":\"$(shell cat /tmp/test_base64.txt)\"}"

# 画像一覧のテスト
test-list:
	$(eval LIST_API_URL := $(call tf_output,list_api_url))
	curl -X GET $(LIST_API_URL)
	
# 特定の日付の画像一覧のテスト
test-list-date:
	$(eval LIST_API_URL := $(call tf_output,list_api_url))
	curl -X GET "$(LIST_API_URL)?date=$(shell date +%Y-%m-%d)"

# Terraformの初期化
tf-init:
	$(TF_CMD) init

# Terraformのプラン
tf-plan:
	$(TF_CMD) plan

# Terraformの適用
tf-apply:
	$(TF_CMD) apply -auto-approve

# Terraformのリソース削除
tf-destroy:
	$(TF_CMD) destroy -auto-approve

# Terraformの設定ファイル検証
tf-validate:
	$(TF_CMD) validate

# Terraformの設定ファイルフォーマット
tf-fmt:
	$(TF_CMD) fmt

# Terraformのキャッシュなどをクリーン
tf-clean:
	rm -rf $(TF_DIR)/.terraform/
	rm -f $(TF_DIR)/.terraform.lock.hcl
	rm -f $(TF_DIR)/terraform.tfstate*