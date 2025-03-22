package policy

import (
	"cloudpix/internal/domain/authmanagement/entity"
	"cloudpix/internal/domain/authmanagement/valueobject"
)

// ResourceType はリソースタイプを表す型
type ResourceType string

const (
	// ResourceImage は画像リソース
	ResourceImage ResourceType = "image"
	// ResourceTag はタグリソース
	ResourceTag ResourceType = "tag"
	// ResourceUser はユーザーリソース
	ResourceUser ResourceType = "user"
)

// Operation は操作タイプを表す型
type Operation string

const (
	// OperationRead は読み取り操作
	OperationRead Operation = "read"
	// OperationWrite は書き込み操作
	OperationWrite Operation = "write"
	// OperationDelete は削除操作
	OperationDelete Operation = "delete"
)

// AccessPolicy はアクセスポリシーを表す
type AccessPolicy struct {
	// ポリシールール
}

// AccessControl はアクセス制御を担当する
type AccessControl struct {
	// アクセスポリシーを適用するロジック
}

// NewAccessControl は新しいアクセス制御を作成します
func NewAccessControl() *AccessControl {
	return &AccessControl{}
}

// IsAllowed はユーザーが特定のリソースに対する操作を許可されているかをチェックします
func (ac *AccessControl) IsAllowed(user *entity.User, resourceType ResourceType,
	resourceOwnerID valueobject.UserID, operation Operation) bool {

	// 管理者は全ての操作が許可される
	if user.IsAdmin() {
		return true
	}

	// 自分自身のリソースに対しては全ての操作が許可される
	if user.ID.Equals(resourceOwnerID) {
		return true
	}

	// リソースタイプと操作に基づいたルール
	switch resourceType {
	case ResourceImage:
		// 画像リソースのアクセスルール
		return ac.isImageAccessAllowed(user, operation)

	case ResourceTag:
		// タグリソースのアクセスルール
		return ac.isTagAccessAllowed(user, operation)

	case ResourceUser:
		// ユーザーリソースのアクセスルール
		return ac.isUserAccessAllowed(user, resourceOwnerID, operation)
	}

	// デフォルトは拒否
	return false
}

// 画像リソースへのアクセス許可ルール
func (ac *AccessControl) isImageAccessAllowed(user *entity.User, operation Operation) bool {
	// 読み取りは全てのユーザーに許可
	if operation == OperationRead {
		return true
	}

	// 書き込みと削除はプレミアムユーザーのみ許可
	if (operation == OperationWrite || operation == OperationDelete) && user.IsPremium() {
		return true
	}

	return false
}

// タグリソースへのアクセス許可ルール
func (ac *AccessControl) isTagAccessAllowed(user *entity.User, operation Operation) bool {
	// 読み取りは全てのユーザーに許可
	if operation == OperationRead {
		return true
	}

	// 書き込みと削除はプレミアムユーザーのみ許可
	if (operation == OperationWrite || operation == OperationDelete) && user.IsPremium() {
		return true
	}

	return false
}

// ユーザーリソースへのアクセス許可ルール
func (ac *AccessControl) isUserAccessAllowed(user *entity.User,
	targetUserID valueobject.UserID, operation Operation) bool {

	// 自分自身の情報は読み書き可能
	if user.ID.Equals(targetUserID) {
		return true
	}

	// 他のユーザーの情報は管理者のみ可能
	return user.IsAdmin()
}
