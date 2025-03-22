package usecase

import (
	"cloudpix/internal/application/authmanagement/dto"
	"cloudpix/internal/domain/authmanagement/entity"
	"cloudpix/internal/domain/authmanagement/policy"
	"cloudpix/internal/domain/authmanagement/repository"
	"cloudpix/internal/domain/authmanagement/service"
	"cloudpix/internal/domain/authmanagement/valueobject"
	"context"
	"errors"
	"strings"
)

// AuthUsecase は認証ユースケース
type AuthUsecase struct {
	authRepository repository.AuthRepository
	authService    service.AuthService
	accessControl  *policy.AccessControl
}

// NewAuthUsecase は新しい認証ユースケースを作成します
func NewAuthUsecase(
	authRepository repository.AuthRepository,
	authService service.AuthService,
) *AuthUsecase {
	return &AuthUsecase{
		authRepository: authRepository,
		authService:    authService,
		accessControl:  policy.NewAccessControl(),
	}
}

// AuthenticateUser はユーザーを認証し、ユーザー情報とトークンを返します
func (u *AuthUsecase) AuthenticateUser(ctx context.Context, tokenString string) (*dto.UserInfoDTO, error) {
	// トークン形式を検証
	if strings.TrimSpace(tokenString) == "" {
		return nil, errors.New("認証トークンが指定されていません")
	}

	// トークンからユーザー情報を取得
	user, err := u.authService.GetUserFromToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	// DTOに変換
	userDTO := u.mapUserToDTO(user)
	return userDTO, nil
}

// CheckPermission はユーザーのリソースアクセス権限をチェックします
func (u *AuthUsecase) CheckPermission(
	ctx context.Context,
	userID string,
	resourceType string,
	resourceOwnerID string,
	operation string,
) (bool, error) {
	// ユーザーIDを値オブジェクトに変換
	userIDObj, err := valueobject.NewUserID(userID)
	if err != nil {
		return false, err
	}

	// リソース所有者IDを値オブジェクトに変換
	ownerIDObj, err := valueobject.NewUserID(resourceOwnerID)
	if err != nil {
		return false, err
	}

	// ユーザー情報を取得
	user, err := u.authRepository.FindUserByID(ctx, userIDObj)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, errors.New("ユーザーが見つかりません")
	}

	// リソースタイプを変換
	var resType policy.ResourceType
	switch resourceType {
	case "image":
		resType = policy.ResourceImage
	case "tag":
		resType = policy.ResourceTag
	case "user":
		resType = policy.ResourceUser
	default:
		return false, errors.New("不明なリソースタイプです")
	}

	// 操作タイプを変換
	var op policy.Operation
	switch operation {
	case "read":
		op = policy.OperationRead
	case "write":
		op = policy.OperationWrite
	case "delete":
		op = policy.OperationDelete
	default:
		return false, errors.New("不明な操作タイプです")
	}

	// アクセス制御チェック
	allowed := u.accessControl.IsAllowed(user, resType, ownerIDObj, op)
	return allowed, nil
}

// mapUserToDTO はユーザーエンティティをDTOに変換します
func (u *AuthUsecase) mapUserToDTO(user *entity.User) *dto.UserInfoDTO {
	// ロールを文字列配列に変換
	roles := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = string(role)
	}

	return &dto.UserInfoDTO{
		UserID:    user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		Roles:     roles,
		IsAdmin:   user.IsAdmin(),
		IsPremium: user.IsPremium(),
	}
}
