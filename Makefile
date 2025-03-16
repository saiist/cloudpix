.PHONY: deploy update-code test test-list test-list-date tf-init tf-plan tf-apply tf-destroy tf-validate tf-fmt tf-clean recreate tf-init-env

# Terraformのディレクトリ
TF_DIR = terraform
# Terraformコマンドのプレフィックス
TF_CMD = cd $(TF_DIR) && terraform

# 完全デプロイ（Terraformを含む）
deploy: tf-apply

# リソースを完全に再作成（destroyしてからapply）
recreate: tf-destroy tf-apply

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

# タグ機能のコード更新
update-tags-code:
	$(eval ECR_REPO := $(call tf_output,ecr_tags_repository_url))
	./build_and_push.sh $(ECR_REPO) ./cmd/tags/main.go
	aws lambda update-function-code \
	  --function-name cloudpix-tags \
	  --image-uri $(ECR_REPO):latest

# 認証処理を含むアップロードテスト
test-upload:
	$(eval API_URL := $(call tf_output,api_url))
	$(eval USER_POOL_ID := $(call tf_output,cognito_user_pool_id))
	$(eval CLIENT_ID := $(call tf_output,cognito_client_id))
	$(eval TEST_EMAIL := test@example.com)
	$(eval TEST_PASSWORD := TestPassword123!)
	
	# テスト画像の準備
	@echo "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==" > /tmp/test_base64.txt
	
	@echo "1. ユーザー存在確認..."
	@aws cognito-idp admin-get-user --user-pool-id $(USER_POOL_ID) --username $(TEST_EMAIL) > /dev/null 2>&1 || \
	(echo "ユーザーが存在しません。新規作成します..." && \
	aws cognito-idp admin-create-user --user-pool-id $(USER_POOL_ID) --username $(TEST_EMAIL) \
	--user-attributes Name=email,Value=$(TEST_EMAIL) Name=email_verified,Value=true --temporary-password $(TEST_PASSWORD) > /dev/null && \
	echo "パスワードを設定しています..." && \
	aws cognito-idp admin-set-user-password --user-pool-id $(USER_POOL_ID) --username $(TEST_EMAIL) \
	--password $(TEST_PASSWORD) --permanent > /dev/null && \
	echo "ユーザー $(TEST_EMAIL) を作成しました" && \
	echo "少し待機します..." && sleep 2)
	
	@echo "2. 認証トークンを取得しています..."
	$(eval TOKEN := $(shell aws cognito-idp initiate-auth --auth-flow USER_PASSWORD_AUTH --client-id $(CLIENT_ID) \
	--auth-parameters USERNAME=$(TEST_EMAIL),PASSWORD=$(TEST_PASSWORD) 2>/dev/null | jq -r '.AuthenticationResult.IdToken'))
	
	@if [ "$(TOKEN)" = "null" ] || [ -z "$(TOKEN)" ]; then \
		echo "トークン取得に失敗しました。もう一度試しています..."; \
		sleep 3; \
		TOKEN=$$(aws cognito-idp initiate-auth --auth-flow USER_PASSWORD_AUTH --client-id $(CLIENT_ID) \
		--auth-parameters USERNAME=$(TEST_EMAIL),PASSWORD=$(TEST_PASSWORD) | jq -r '.AuthenticationResult.IdToken'); \
		if [ "$$TOKEN" = "null" ] || [ -z "$$TOKEN" ]; then \
			echo "再試行してもトークン取得に失敗しました"; \
			exit 1; \
		fi; \
	else \
		echo "トークンを取得しました"; \
	fi
	
	@echo "3. アップロードリクエストを送信しています..."
	@curl -X POST $(API_URL) \
	  -H "Content-Type: application/json" \
	  -H "Authorization: Bearer $$TOKEN" \
	  -d "{\"fileName\":\"test.png\",\"contentType\":\"image/png\",\"data\":\"$$(cat /tmp/test_base64.txt)\"}"

# 画像一覧のテスト
test-list:
	$(eval LIST_API_URL := $(call tf_output,list_api_url))
	curl -X GET $(LIST_API_URL)
	
# 特定の日付の画像一覧のテスト
test-list-date:
	$(eval LIST_API_URL := $(call tf_output,list_api_url))
	curl -X GET "$(LIST_API_URL)?date=$(shell date +%Y-%m-%d)"


# ヘルパー関数 - 最初の画像IDを取得
define get_first_image_id
$(shell curl -s $(call tf_output,list_api_url) | jq -r '.images[0].imageId')
endef

# タグ追加のテスト
test-add-tags:
	$(eval TAGS_API_URL := $(call tf_output,tags_api_url))
	$(eval IMAGE_ID := $(call get_first_image_id))
	@echo "Adding tags to image ID: $(IMAGE_ID)"
	curl -X POST $(TAGS_API_URL) \
	  -H "Content-Type: application/json" \
	  -d "{\"imageId\":\"$(IMAGE_ID)\",\"tags\":[\"nature\",\"landscape\",\"vacation\"]}"

# 画像のタグ取得テスト
test-get-image-tags:
	$(eval TAGS_API_URL := $(call tf_output,tags_api_url))
	$(eval IMAGE_ID := $(call get_first_image_id))
	@echo "Getting tags for image ID: $(IMAGE_ID)"
	curl -X GET $(TAGS_API_URL)/$(IMAGE_ID)

# すべてのタグのリスト取得テスト
test-list-tags:
	$(eval TAGS_API_URL := $(call tf_output,tags_api_url))
	curl -X GET $(TAGS_API_URL)

# タグによる画像検索テスト
test-search-by-tag:
	$(eval LIST_API_URL := $(call tf_output,list_api_url))
	@echo "Enter tag to search for: " && read TAG && \
	curl -X GET "$(LIST_API_URL)?tag=$${TAG}"

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

# 環境設定ファイルの初期化（サンプルをコピー）
tf-init-env:
	@if [ ! -f $(TF_DIR)/terraform.tfvars ]; then \
		if [ -f $(TF_DIR)/terraform.tfvars.example ]; then \
			cp $(TF_DIR)/terraform.tfvars.example $(TF_DIR)/terraform.tfvars; \
			echo "terraform.tfvars created from example. Please edit the file with your settings."; \
		else \
			echo "Error: $(TF_DIR)/terraform.tfvars.example not found."; \
			echo "Please create terraform.tfvars.example first."; \
			exit 1; \
		fi; \
	else \
		echo "terraform.tfvars already exists."; \
	fi