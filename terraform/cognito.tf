################################
# Cognito User Pool
################################
resource "aws_cognito_user_pool" "cloudpix_users" {
  name = "${var.app_name}-users"

  # ユーザー名属性の設定
  username_attributes      = ["email"]
  auto_verified_attributes = ["email"]

  # パスワードポリシー
  password_policy {
    minimum_length    = 8
    require_lowercase = true
    require_numbers   = true
    require_symbols   = true
    require_uppercase = true
  }

  # メール設定
  email_configuration {
    email_sending_account = "COGNITO_DEFAULT"
  }

  # メール検証設定
  verification_message_template {
    default_email_option = "CONFIRM_WITH_CODE"
    email_subject        = "${title(var.app_name)} - アカウント確認コード"
    email_message        = "${title(var.app_name)}へようこそ！確認コード: {####}"
  }

  # ユーザー属性スキーマ
  schema {
    name                = "name"
    attribute_data_type = "String"
    mutable             = true
    required            = true
  }

  # カスタム属性：プランタイプ
  schema {
    name                = "plan_type"
    attribute_data_type = "String"
    mutable             = true
    required            = false
    string_attribute_constraints {
      min_length = 1
      max_length = 20
    }
  }

  tags = {
    Name        = "${var.app_name}-Users"
    Environment = var.environment
  }
}

################################
# Cognito User Groups
################################
# 管理者グループ
resource "aws_cognito_user_group" "admin_group" {
  name         = "Administrators"
  user_pool_id = aws_cognito_user_pool.cloudpix_users.id
  description  = "${title(var.app_name)}の管理者グループ"
  precedence   = 1
}

# プレミアムユーザーグループ
resource "aws_cognito_user_group" "premium_group" {
  name         = "PremiumUsers"
  user_pool_id = aws_cognito_user_pool.cloudpix_users.id
  description  = "有料プランのユーザー"
  precedence   = 2
}

# 一般ユーザーグループ
resource "aws_cognito_user_group" "standard_group" {
  name         = "StandardUsers"
  user_pool_id = aws_cognito_user_pool.cloudpix_users.id
  description  = "無料プランのユーザー"
  precedence   = 3
}

################################
# Cognito App Client
################################
resource "aws_cognito_user_pool_client" "cloudpix_client" {
  name = "${var.app_name}-app-client"

  user_pool_id = aws_cognito_user_pool.cloudpix_users.id

  # トークン設定
  refresh_token_validity = 30
  access_token_validity  = 1
  id_token_validity      = 1
  token_validity_units {
    access_token  = "hours"
    id_token      = "hours"
    refresh_token = "days"
  }

  # 認証フロー設定
  explicit_auth_flows = [
    "ALLOW_USER_SRP_AUTH",
    "ALLOW_REFRESH_TOKEN_AUTH",
    "ALLOW_USER_PASSWORD_AUTH"
  ]

  # コールバックURL設定 - code/implicit フローには必須
  callback_urls        = ["${var.app_frontend_url}/callback", "${var.app_frontend_url}/auth"]
  logout_urls          = ["${var.app_frontend_url}/logout"]
  default_redirect_uri = "${var.app_frontend_url}/callback"

  # セキュリティ設定
  prevent_user_existence_errors = "ENABLED"

  # OAuthスコープ設定
  allowed_oauth_flows = ["code", "implicit"]
  allowed_oauth_scopes = [
    "phone",
    "email",
    "openid",
    "profile",
    "aws.cognito.signin.user.admin"
  ]
  allowed_oauth_flows_user_pool_client = true

  # シークレット生成（サーバーサイドアプリケーション用）
  generate_secret = false
}

################################
# Cognito Domain
################################
resource "aws_cognito_user_pool_domain" "cloudpix_domain" {
  domain       = "${var.app_name}-auth"
  user_pool_id = aws_cognito_user_pool.cloudpix_users.id
}