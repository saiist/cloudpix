package dto

// TagsResponseDTO はタグ一覧のレスポンスDTO
type TagsResponseDTO struct {
	Tags  []string `json:"tags"`
	Count int      `json:"count"`
}

// AddTagRequestDTO はタグ追加リクエストのDTO
type AddTagRequestDTO struct {
	ImageID string   `json:"imageId"`
	Tags    []string `json:"tags"`
}

// RemoveTagRequestDTO はタグ削除リクエストのDTO
type RemoveTagRequestDTO struct {
	ImageID string   `json:"imageId"`
	Tags    []string `json:"tags"`
}

// TagUpdateResponseDTO はタグ更新のレスポンスDTO
type TagUpdateResponseDTO struct {
	ImageID  string `json:"imageId"`
	Message  string `json:"message"`
	Modified int    `json:"modified"`
}
