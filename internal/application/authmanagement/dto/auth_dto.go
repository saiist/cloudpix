package dto

// UserInfoDTO はユーザー情報のデータ転送オブジェクト
type UserInfoDTO struct {
	UserID    string   `json:"userId"`
	Username  string   `json:"username"`
	Email     string   `json:"email,omitempty"`
	Roles     []string `json:"roles"`
	IsAdmin   bool     `json:"isAdmin"`
	IsPremium bool     `json:"isPremium"`
}

// AuthResponseDTO は認証レスポンスのデータ転送オブジェクト
type AuthResponseDTO struct {
	Token     string      `json:"token,omitempty"`
	ExpiresIn int         `json:"expiresIn,omitempty"`
	User      UserInfoDTO `json:"user"`
}

// AuthErrorDTO は認証エラーのデータ転送オブジェクト
type AuthErrorDTO struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}
