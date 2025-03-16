.PHONY: deploy update-code api-test api-test-list api-test-list-date tf-init tf-plan tf-apply tf-destroy tf-validate tf-fmt tf-clean recreate tf-init-env go-test go-test-verbose go-test-coverage

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

## -- テスト用コマンド群 --  ##

# 共有の認証トークン取得関数
define get_auth_token
$(eval USER_POOL_ID := $(call tf_output,cognito_user_pool_id))
$(eval CLIENT_ID := $(call tf_output,cognito_client_id))
$(eval TEST_EMAIL := test@example.com)
$(eval TEST_PASSWORD := TestPassword123!)

@echo "認証処理を実行中..."
@# ユーザーの存在確認と作成（必要な場合）
@aws cognito-idp admin-get-user --user-pool-id $(USER_POOL_ID) --username $(TEST_EMAIL) > /dev/null 2>&1 || \
(aws cognito-idp admin-create-user --user-pool-id $(USER_POOL_ID) --username $(TEST_EMAIL) \
--user-attributes Name=email,Value=$(TEST_EMAIL) Name=email_verified,Value=true --temporary-password $(TEST_PASSWORD) > /dev/null && \
aws cognito-idp admin-set-user-password --user-pool-id $(USER_POOL_ID) --username $(TEST_EMAIL) \
--password $(TEST_PASSWORD) --permanent > /dev/null && sleep 5)

@# 認証トークン取得
@aws cognito-idp initiate-auth --auth-flow USER_PASSWORD_AUTH --client-id $(CLIENT_ID) \
--auth-parameters USERNAME=$(TEST_EMAIL),PASSWORD=$(TEST_PASSWORD) > /tmp/auth_response.json 2>/dev/null
@TOKEN=`cat /tmp/auth_response.json | jq -r '.AuthenticationResult.IdToken' 2>/dev/null || echo ""`; \
echo "AUTH_TOKEN=$$TOKEN" > /tmp/auth_env.sh
endef

# 認証処理を含むアップロードテスト
api-test-upload:
	$(eval API_URL := $(call tf_output,api_url))
	# テスト画像の準備
	@echo "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==" > /tmp/test_base64.txt
	
	# 認証トークン取得
	$(call get_auth_token)
	
	@echo "アップロードリクエストを送信しています..."
	@. /tmp/auth_env.sh && \
	curl -s -X POST $(API_URL) \
	  -H "Content-Type: application/json" \
	  -H "Authorization: Bearer $$AUTH_TOKEN" \
	  -d "{\"fileName\":\"test.png\",\"contentType\":\"image/png\",\"data\":\"`cat /tmp/test_base64.txt`\"}" | jq .

# 画像一覧のテスト（認証付き）
api-test-list:
	$(eval LIST_API_URL := $(call tf_output,list_api_url))
	
	# 認証トークン取得
	$(call get_auth_token)
	
	@echo "画像一覧を取得しています..."
	@. /tmp/auth_env.sh && \
	curl -s -X GET $(LIST_API_URL) \
	  -H "Authorization: Bearer $$AUTH_TOKEN" | jq .
	
# 特定の日付の画像一覧のテスト（認証付き）
api-test-list-date:
	$(eval LIST_API_URL := $(call tf_output,list_api_url))
	$(eval TODAY := $(shell date +%Y-%m-%d))
	
	# 認証トークン取得
	$(call get_auth_token)
	
	@echo "日付: $(TODAY) の画像一覧を取得しています..."
	@. /tmp/auth_env.sh && \
	curl -s -X GET "$(LIST_API_URL)?date=$(TODAY)" \
	  -H "Authorization: Bearer $$AUTH_TOKEN" | jq .

# タグ追加のテスト（認証付き）
api-test-add-tags:
	$(eval TAGS_API_URL := $(call tf_output,tags_api_url))
	
	# 認証トークン取得
	$(call get_auth_token)
	
	# 画像IDの取得
	@echo "画像一覧を取得して最初の画像IDを抽出します..."
	@. /tmp/auth_env.sh && \
	curl -s -X GET $(call tf_output,list_api_url) \
	  -H "Authorization: Bearer $$AUTH_TOKEN" > /tmp/image_list.json
	@IMAGE_ID=`cat /tmp/image_list.json | jq -r '.images[0].imageId'` && \
	echo "IMAGE_ID=$$IMAGE_ID" > /tmp/image_env.sh
	
	@echo "タグを追加しています..."
	@. /tmp/auth_env.sh && . /tmp/image_env.sh && \
	echo "画像ID: $$IMAGE_ID" && \
	curl -s -X POST $(TAGS_API_URL) \
	  -H "Content-Type: application/json" \
	  -H "Authorization: Bearer $$AUTH_TOKEN" \
	  -d "{\"imageId\":\"$$IMAGE_ID\",\"tags\":[\"nature\",\"landscape\",\"vacation\"]}" | jq .

# 画像のタグ取得テスト（認証付き）
api-test-get-image-tags:
	$(eval TAGS_API_URL := $(call tf_output,tags_api_url))
	
	# 認証トークン取得
	$(call get_auth_token)
	
	# 画像IDの取得
	@echo "画像一覧を取得して最初の画像IDを抽出します..."
	@. /tmp/auth_env.sh && \
	curl -s -X GET $(call tf_output,list_api_url) \
	  -H "Authorization: Bearer $$AUTH_TOKEN" > /tmp/image_list.json
	@IMAGE_ID=`cat /tmp/image_list.json | jq -r '.images[0].imageId'` && \
	echo "IMAGE_ID=$$IMAGE_ID" > /tmp/image_env.sh
	
	@echo "タグを取得しています..."
	@. /tmp/auth_env.sh && . /tmp/image_env.sh && \
	echo "画像ID: $$IMAGE_ID" && \
	curl -s -X GET $(TAGS_API_URL)/$$IMAGE_ID \
	  -H "Authorization: Bearer $$AUTH_TOKEN" | jq .

# すべてのタグのリスト取得テスト（認証付き）
api-test-list-tags:
	$(eval TAGS_API_URL := $(call tf_output,tags_api_url))
	
	# 認証トークン取得
	$(call get_auth_token)
	
	@echo "すべてのタグを取得しています..."
	@. /tmp/auth_env.sh && \
	curl -s -X GET $(TAGS_API_URL) \
	  -H "Authorization: Bearer $$AUTH_TOKEN" | jq .

# タグによる画像検索テスト（認証付き）
api-test-search-by-tag:
	$(eval LIST_API_URL := $(call tf_output,list_api_url))
	
	# 認証トークン取得
	$(call get_auth_token)
	
	@echo "検索するタグを入力してください: " && read TAG && \
	. /tmp/auth_env.sh && \
	echo "タグ: $$TAG による画像検索結果:" && \
	curl -s -X GET "$(LIST_API_URL)?tag=$$TAG" \
	  -H "Authorization: Bearer $$AUTH_TOKEN" | jq .

## -- Goテスト関連コマンド -- ##

# 基本的な単体テスト実行
go-test:
	@echo "単体テストを実行しています..."
	@go test ./internal/...

# 詳細出力ありの単体テスト実行
go-test-verbose:
	@echo "詳細出力ありで単体テストを実行しています..."
	@go test -v ./internal/...

# カバレッジレポート付きのテスト実行
go-test-coverage:
	@echo "テストカバレッジを計測しています..."
	@go test ./internal/... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "coverage.htmlにカバレッジレポートを出力しました"
	@echo "ブラウザでcoverage.htmlを開いてカバレッジを確認してください"
