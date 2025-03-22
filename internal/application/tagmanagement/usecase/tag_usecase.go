package usecase

import (
	"cloudpix/internal/application/tagmanagement/dto"
	"cloudpix/internal/domain/shared/event/dispatcher"
	"cloudpix/internal/domain/tagmanagement/entity"
	"cloudpix/internal/domain/tagmanagement/event"
	"cloudpix/internal/domain/tagmanagement/repository"
	"cloudpix/internal/domain/tagmanagement/valueobject"
	"context"
	"errors"
	"fmt"
)

// 定義済みエラー
var (
	ErrImageNotFound     = errors.New("指定された画像が見つかりません")
	ErrInvalidTag        = errors.New("無効なタグ形式です")
	ErrRepositoryFailure = errors.New("タグリポジトリ操作に失敗しました")
)

// TagUsecase はタグ管理のユースケース
type TagUsecase struct {
	tagRepository   repository.TagRepository
	eventDispatcher dispatcher.EventDispatcher
}

// NewTagUsecase は新しいタグユースケースを作成します
func NewTagUsecase(
	tagRepository repository.TagRepository,
	eventDispatcher dispatcher.EventDispatcher,
) *TagUsecase {
	return &TagUsecase{
		tagRepository:   tagRepository,
		eventDispatcher: eventDispatcher,
	}
}

// ListAllTags は全てのタグを取得します
func (u *TagUsecase) ListAllTags(ctx context.Context) (*dto.TagsResponseDTO, error) {
	// リポジトリから全タグを取得
	tags, err := u.tagRepository.FindAllTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRepositoryFailure, err)
	}

	// タグ名の配列に変換
	tagNames := make([]string, len(tags))
	for i, tag := range tags {
		tagNames[i] = tag.Name()
	}

	return &dto.TagsResponseDTO{
		Tags:  tagNames,
		Count: len(tagNames),
	}, nil
}

// GetImageTags は指定された画像のタグを取得します
func (u *TagUsecase) GetImageTags(ctx context.Context, imageID string) (*dto.TagsResponseDTO, error) {
	// 画像の存在チェック
	exists, err := u.tagRepository.ImageExists(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRepositoryFailure, err)
	}

	if !exists {
		return nil, ErrImageNotFound
	}

	// タグ付き画像情報を取得
	taggedImage, err := u.tagRepository.FindTaggedImage(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRepositoryFailure, err)
	}

	// 見つからない場合は空のリストを返す
	if taggedImage == nil {
		return &dto.TagsResponseDTO{
			Tags:  []string{},
			Count: 0,
		}, nil
	}

	// タグ名のリストを取得
	tagNames := taggedImage.GetTagNames()

	return &dto.TagsResponseDTO{
		Tags:  tagNames,
		Count: len(tagNames),
	}, nil
}

// AddTags は画像にタグを追加します
func (u *TagUsecase) AddTags(ctx context.Context, request *dto.AddTagRequestDTO) (*dto.TagUpdateResponseDTO, error) {
	// 画像の存在チェック
	exists, err := u.tagRepository.ImageExists(ctx, request.ImageID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRepositoryFailure, err)
	}

	if !exists {
		return nil, ErrImageNotFound
	}

	// タグ付き画像情報を取得または作成
	var taggedImage *entity.TaggedImage
	taggedImage, err = u.tagRepository.FindTaggedImage(ctx, request.ImageID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRepositoryFailure, err)
	}

	if taggedImage == nil {
		taggedImage = entity.NewTaggedImage(request.ImageID)
	}

	// タグを追加
	addedCount := 0
	for _, tagName := range request.Tags {
		tag, err := valueobject.NewTag(tagName)
		if err != nil {
			continue // 無効なタグはスキップ
		}

		if taggedImage.AddTag(tag) {
			addedCount++
		}
	}

	// 変更があった場合のみ保存
	if addedCount > 0 {
		err = u.tagRepository.Save(ctx, taggedImage)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrRepositoryFailure, err)
		}

		// イベント発行
		tagsUpdatedEvent := event.NewTagsUpdatedEvent(taggedImage)
		u.eventDispatcher.Dispatch(ctx, tagsUpdatedEvent)
	}

	return &dto.TagUpdateResponseDTO{
		ImageID:  request.ImageID,
		Message:  fmt.Sprintf("%d tags added", addedCount),
		Modified: addedCount,
	}, nil
}

// RemoveTags は画像からタグを削除します
func (u *TagUsecase) RemoveTags(ctx context.Context, request *dto.RemoveTagRequestDTO) (*dto.TagUpdateResponseDTO, error) {
	// 画像の存在チェック
	exists, err := u.tagRepository.ImageExists(ctx, request.ImageID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRepositoryFailure, err)
	}

	if !exists {
		return nil, ErrImageNotFound
	}

	// タグ付き画像情報を取得
	taggedImage, err := u.tagRepository.FindTaggedImage(ctx, request.ImageID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRepositoryFailure, err)
	}

	// タグ情報がない場合は処理不要
	if taggedImage == nil {
		return &dto.TagUpdateResponseDTO{
			ImageID:  request.ImageID,
			Message:  "No tags to remove",
			Modified: 0,
		}, nil
	}

	// タグを削除
	removedCount := 0
	if len(request.Tags) == 0 {
		// タグが指定されていない場合は全て削除
		removedCount = taggedImage.ClearTags()
	} else {
		// 指定されたタグを削除
		for _, tagName := range request.Tags {
			tag, err := valueobject.NewTag(tagName)
			if err != nil {
				continue // 無効なタグはスキップ
			}

			if taggedImage.RemoveTag(tag) {
				removedCount++
			}
		}
	}

	// 変更があった場合のみ保存
	if removedCount > 0 {
		err = u.tagRepository.Save(ctx, taggedImage)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrRepositoryFailure, err)
		}

		// イベント発行
		tagsUpdatedEvent := event.NewTagsUpdatedEvent(taggedImage)
		u.eventDispatcher.Dispatch(ctx, tagsUpdatedEvent)
	}

	return &dto.TagUpdateResponseDTO{
		ImageID:  request.ImageID,
		Message:  fmt.Sprintf("%d tags removed", removedCount),
		Modified: removedCount,
	}, nil
}

// FindImagesByTag はタグで画像を検索します
func (u *TagUsecase) FindImagesByTag(ctx context.Context, tagName string) ([]string, error) {
	// タグの値オブジェクトを作成
	tag, err := valueobject.NewTag(tagName)
	if err != nil {
		return nil, ErrInvalidTag
	}

	// リポジトリから画像IDを取得
	imageIDs, err := u.tagRepository.FindImagesByTag(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRepositoryFailure, err)
	}

	return imageIDs, nil
}
