.PHONY: deploy update-code test

# 完全デプロイ（Terraformを含む）
deploy:
	terraform apply -auto-approve

# コードのみを更新
update-code:
	$(eval ECR_REPO := $(shell terraform output -raw ecr_repository_url))
	./build_and_push.sh $(ECR_REPO)
	aws lambda update-function-code \
	  --function-name cloudpix-upload \
	  --image-uri $(ECR_REPO):latest

# APIのテスト (Base64形式での画像アップロード)
test:
	$(eval API_URL := $(shell terraform output -raw api_url))
	# サンプルの小さなPNG画像をBase64エンコードしてアップロード
	echo "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==" > /tmp/test_base64.txt
	curl -X POST $(API_URL) \
	  -H "Content-Type: application/json" \
	  -d "{\"fileName\":\"test.png\",\"contentType\":\"image/png\",\"data\":\"$(shell cat /tmp/test_base64.txt)\"}"

# 画像一覧のテスト
test-list:
	$(eval LIST_API_URL := $(shell terraform output -raw list_api_url))
	curl -X GET $(LIST_API_URL)
	
# 特定の日付の画像一覧のテスト
test-list-date:
	$(eval LIST_API_URL := $(shell terraform output -raw list_api_url))
	curl -X GET "$(LIST_API_URL)?date=$(shell date +%Y-%m-%d)"