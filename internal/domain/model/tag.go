package model

// タグ情報の構造体
type TagItem struct {
	TagName string `json:"tagName"`
	ImageID string `json:"imageId"`
}

// タグ追加リクエスト
type AddTagRequest struct {
	ImageID string   `json:"imageId"`
	Tags    []string `json:"tags"`
}

// タグ削除リクエスト
type RemoveTagRequest struct {
	ImageID string   `json:"imageId"`
	Tags    []string `json:"tags"`
}

// タグ一覧レスポンス
type TagsResponse struct {
	Tags  []string `json:"tags"`
	Count int      `json:"count"`
}
