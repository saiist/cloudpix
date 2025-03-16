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
	@echo "アップロードコードを更新しています..."
	@./build_and_push.sh $(ECR_REPO) ./cmd/upload/main.go
	@aws lambda update-function-code \
	  --function-name cloudpix-upload \
	  --image-uri $(ECR_REPO):latest

# リストのコードを更新
update-list-code:
	$(eval ECR_REPO := $(call tf_output,ecr_list_repository_url))
	@echo "リストコードを更新しています..."
	@./build_and_push.sh $(ECR_REPO) ./cmd/list/main.go
	@aws lambda update-function-code \
	  --function-name cloudpix-list \
	  --image-uri $(ECR_REPO):latest

# サムネイル生成コードの更新
update-thumbnail-code:
	$(eval ECR_REPO := $(call tf_output,ecr_thumbnail_repository_url))
	@echo "サムネイルコードを更新しています..."
	@./build_and_push.sh $(ECR_REPO) ./cmd/thumbnail/main.go
	@aws lambda update-function-code \
	  --function-name cloudpix-thumbnail \
	  --image-uri $(ECR_REPO):latest

# タグ機能のコード更新
update-tags-code:
	$(eval ECR_REPO := $(call tf_output,ecr_tags_repository_url))
	@echo "タグ機能コードを更新しています..."
	@./build_and_push.sh $(ECR_REPO) ./cmd/tags/main.go
	@aws lambda update-function-code \
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
	
	@echo "1. ユーザー存在確認と作成..."
	@aws cognito-idp admin-get-user --user-pool-id $(USER_POOL_ID) --username $(TEST_EMAIL) > /dev/null 2>&1 || \
	(aws cognito-idp admin-create-user --user-pool-id $(USER_POOL_ID) --username $(TEST_EMAIL) \
	--user-attributes Name=email,Value=$(TEST_EMAIL) Name=email_verified,Value=true --temporary-password $(TEST_PASSWORD) > /dev/null && \
	aws cognito-idp admin-set-user-password --user-pool-id $(USER_POOL_ID) --username $(TEST_EMAIL) \
	--password $(TEST_PASSWORD) --permanent > /dev/null && sleep 2)
	
	@echo "2. 認証トークンを取得しています..."
	@aws cognito-idp initiate-auth --auth-flow USER_PASSWORD_AUTH --client-id $(CLIENT_ID) \
	--auth-parameters USERNAME=$(TEST_EMAIL),PASSWORD=$(TEST_PASSWORD) > /tmp/auth_response.json || true
	$(eval TOKEN := $(shell cat /tmp/auth_response.json 2>/dev/null | jq -r '.AuthenticationResult.IdToken' 2>/dev/null || echo ""))
	
	@if [ -z "$(TOKEN)" ] || [ "$(TOKEN)" = "null" ]; then \
		sleep 3; \
		aws cognito-idp initiate-auth --auth-flow USER_PASSWORD_AUTH --client-id $(CLIENT_ID) \
		--auth-parameters USERNAME=$(TEST_EMAIL),PASSWORD=$(TEST_PASSWORD) > /tmp/auth_response_retry.json || true; \
		TOKEN=`cat /tmp/auth_response_retry.json 2>/dev/null | jq -r '.AuthenticationResult.IdToken' 2>/dev/null || echo ""`; \
		if [ -z "$$TOKEN" ] || [ "$$TOKEN" = "null" ]; then \
			echo "認証トークンの取得に失敗しました"; \
			exit 1; \
		fi; \
	fi
	
	@echo "3. アップロードリクエストを送信しています..."
	@curl -s -X POST $(API_URL) \
	  -H "Content-Type: application/json" \
	  -H "Authorization: Bearer $(TOKEN)" \
	  -d "{\"fileName\":\"test.png\",\"contentType\":\"image/png\",\"data\":\"`cat /tmp/test_base64.txt`\"}" | jq .

# 画像一覧のテスト
test-list:
	$(eval LIST_API_URL := $(call tf_output,list_api_url))
	@echo "画像一覧を取得しています..."
	@curl -s -X GET $(LIST_API_URL) | jq .
	
# 特定の日付の画像一覧のテスト
test-list-date:
	$(eval LIST_API_URL := $(call tf_output,list_api_url))
	$(eval TODAY := $(shell date +%Y-%m-%d))
	@echo "日付: $(TODAY) の画像一覧を取得しています..."
	@curl -s -X GET "$(LIST_API_URL)?date=$(TODAY)" | jq .

# ヘルパー関数 - 最初の画像IDを取得
define get_first_image_id
$(shell curl -s $(call tf_output,list_api_url) | jq -r '.images[0].imageId')
endef

# タグ追加のテスト
test-add-tags:
	$(eval TAGS_API_URL := $(call tf_output,tags_api_url))
	$(eval IMAGE_ID := $(call get_first_image_id))
	@echo "画像ID: $(IMAGE_ID) にタグを追加しています..."
	@curl -s -X POST $(TAGS_API_URL) \
	  -H "Content-Type: application/json" \
	  -d "{\"imageId\":\"$(IMAGE_ID)\",\"tags\":[\"nature\",\"landscape\",\"vacation\"]}" | jq .

# 画像のタグ取得テスト
test-get-image-tags:
	$(eval TAGS_API_URL := $(call tf_output,tags_api_url))
	$(eval IMAGE_ID := $(call get_first_image_id))
	@echo "画像ID: $(IMAGE_ID) のタグを取得しています..."
	@curl -s -X GET $(TAGS_API_URL)/$(IMAGE_ID) | jq .

# すべてのタグのリスト取得テスト
test-list-tags:
	$(eval TAGS_API_URL := $(call tf_output,tags_api_url))
	@echo "すべてのタグを取得しています..."
	@curl -s -X GET $(TAGS_API_URL) | jq .

# タグによる画像検索テスト
test-search-by-tag:
	$(eval LIST_API_URL := $(call tf_output,list_api_url))
	@echo "検索するタグを入力してください: " && read TAG && \
	echo "タグ: $${TAG} による画像検索結果:" && \
	curl -s -X GET "$(LIST_API_URL)?tag=$${TAG}" | jq .

# Terraformの初期化
tf-init:
	@echo "Terraformを初期化しています..."
	@$(TF_CMD) init

# Terraformのプラン
tf-plan:
	@echo "変更プランを表示します..."
	@$(TF_CMD) plan

# Terraformの適用
tf-apply:
	@echo "変更を適用しています..."
	@$(TF_CMD) apply -auto-approve

# Terraformのリソース削除
tf-destroy:
	@echo "リソースを削除しています..."
	@$(TF_CMD) destroy -auto-approve

# Terraformの設定ファイル検証
tf-validate:
	@echo "設定ファイルを検証しています..."
	@$(TF_CMD) validate

# Terraformの設定ファイルフォーマット
tf-fmt:
	@echo "設定ファイルをフォーマットしています..."
	@$(TF_CMD) fmt

# Terraformのキャッシュなどをクリーン
tf-clean:
	@echo "Terraformキャッシュをクリーンアップしています..."
	@rm -rf $(TF_DIR)/.terraform/
	@rm -f $(TF_DIR)/.terraform.lock.hcl
	@rm -f $(TF_DIR)/terraform.tfstate*

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