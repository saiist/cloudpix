package usecase

import (
	"cloudpix/internal/domain/model"
	"cloudpix/internal/domain/repository"
	"context"
	"strings"
)

type TagUsecase struct {
	tagRepo repository.TagRepository
}

func NewTagUsecase(tagRepo repository.TagRepository) *TagUsecase {
	return &TagUsecase{
		tagRepo: tagRepo,
	}
}

// タグ一覧を取得
func (u *TagUsecase) ListTags(ctx context.Context) (*model.TagsResponse, error) {
	tags, err := u.tagRepo.ListTags(ctx)
	if err != nil {
		return nil, err
	}

	return &model.TagsResponse{
		Tags:  tags,
		Count: len(tags),
	}, nil
}

// 特定の画像のタグを取得
func (u *TagUsecase) GetImageTags(ctx context.Context, imageID string) (*model.TagsResponse, error) {
	tags, err := u.tagRepo.GetImageTags(ctx, imageID)
	if err != nil {
		return nil, err
	}

	return &model.TagsResponse{
		Tags:  tags,
		Count: len(tags),
	}, nil
}

// タグを追加
func (u *TagUsecase) AddTags(ctx context.Context, req *model.AddTagRequest) (int, error) {
	// 画像の存在を確認
	exists, err := u.tagRepo.VerifyImageExists(ctx, req.ImageID)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, ErrImageNotFound
	}

	// タグを正規化してから追加
	normalizedTags := u.normalizeTags(req.Tags)

	return u.tagRepo.AddTags(ctx, req.ImageID, normalizedTags)
}

// 特定のタグを削除
func (u *TagUsecase) RemoveTags(ctx context.Context, req *model.RemoveTagRequest) (int, error) {
	// タグを正規化してから削除
	normalizedTags := u.normalizeTags(req.Tags)

	return u.tagRepo.RemoveTags(ctx, req.ImageID, normalizedTags)
}

// 画像のすべてのタグを削除
func (u *TagUsecase) RemoveAllTags(ctx context.Context, imageID string) (int, error) {
	return u.tagRepo.RemoveAllTags(ctx, imageID)
}

// タグを正規化（トリミングして小文字に）
func (u *TagUsecase) normalizeTags(tags []string) []string {
	var normalizedTags []string

	for _, tag := range tags {
		normalizedTag := strings.ToLower(strings.TrimSpace(tag))
		if normalizedTag != "" {
			normalizedTags = append(normalizedTags, normalizedTag)
		}
	}

	return normalizedTags
}

// エラー定義
var (
	ErrImageNotFound = NewError("image not found")
)

type Error struct {
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

func NewError(message string) *Error {
	return &Error{
		Message: message,
	}
}
