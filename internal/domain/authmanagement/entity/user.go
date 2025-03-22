package entity

import (
	"cloudpix/internal/domain/authmanagement/valueobject"
	"time"
)

// UserRole はユーザーのロールを表す型
type UserRole string

const (
	// RoleAdmin は管理者ロール
	RoleAdmin UserRole = "Admin"
	// RolePremium はプレミアムユーザーロール
	RolePremium UserRole = "Premium"
	// RoleStandard は標準ユーザーロール
	RoleStandard UserRole = "Standard"
)

// User エンティティはユーザー情報を表します
type User struct {
	ID        valueobject.UserID
	Username  string
	Email     string
	Roles     []UserRole
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUser は新しいユーザーエンティティを作成します
func NewUser(
	id valueobject.UserID,
	username string,
	email string,
	roles []UserRole,
) *User {
	now := time.Now()
	return &User{
		ID:        id,
		Username:  username,
		Email:     email,
		Roles:     roles,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// HasRole は指定されたロールを持っているかをチェックします
func (u *User) HasRole(role UserRole) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// IsAdmin は管理者かどうかをチェックします
func (u *User) IsAdmin() bool {
	return u.HasRole(RoleAdmin)
}

// IsPremium はプレミアムユーザーかどうかをチェックします
func (u *User) IsPremium() bool {
	return u.HasRole(RolePremium) || u.HasRole(RoleAdmin)
}

// AddRole はロールを追加します
func (u *User) AddRole(role UserRole) {
	// 既に持っているロールはスキップ
	if u.HasRole(role) {
		return
	}
	u.Roles = append(u.Roles, role)
	u.UpdatedAt = time.Now()
}

// RemoveRole はロールを削除します
func (u *User) RemoveRole(role UserRole) {
	newRoles := make([]UserRole, 0, len(u.Roles))
	for _, r := range u.Roles {
		if r != role {
			newRoles = append(newRoles, r)
		}
	}

	// ロールが変更された場合のみ更新
	if len(newRoles) != len(u.Roles) {
		u.Roles = newRoles
		u.UpdatedAt = time.Now()
	}
}
