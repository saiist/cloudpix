package middleware

// Chain はミドルウェアのチェーンを管理する構造体
type Chain struct {
	middlewares []Middleware
}

// NewChain は新しいミドルウェアチェーンを作成する
func NewChain(middlewares ...Middleware) *Chain {
	return &Chain{middlewares: middlewares}
}

// Use はミドルウェアチェーンに新しいミドルウェアを追加する
func (c *Chain) Use(middleware Middleware) *Chain {
	c.middlewares = append(c.middlewares, middleware)
	return c
}

// Then はミドルウェアチェーンを最終ハンドラーと結合し、完全な処理パイプラインを返す
func (c *Chain) Then(handler HandlerFunc) HandlerFunc {
	// ミドルウェアを逆順に適用（最初に登録したミドルウェアが最も外側になる）
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		handler = c.middlewares[i](handler)
	}
	return handler
}

// Append は現在のチェーンに別のチェーンのミドルウェアを追加する
func (c *Chain) Append(other *Chain) *Chain {
	if other == nil {
		return c
	}

	result := NewChain()
	result.middlewares = append(result.middlewares, c.middlewares...)
	result.middlewares = append(result.middlewares, other.middlewares...)
	return result
}

// Clone は現在のチェーンのコピーを作成する
func (c *Chain) Clone() *Chain {
	result := NewChain()
	result.middlewares = make([]Middleware, len(c.middlewares))
	copy(result.middlewares, c.middlewares)
	return result
}

// Count はチェーン内のミドルウェア数を返す
func (c *Chain) Count() int {
	return len(c.middlewares)
}
