package dispatcher

import (
	"context"
	"time"
)

// DomainEvent はすべてのドメインイベントの基本インターフェース
type DomainEvent interface {
	EventType() string
	OccurredAt() time.Time
	AggregateID() string
}

// EventHandler はイベントハンドラーのインターフェース
type EventHandler interface {
	HandleEvent(ctx context.Context, event DomainEvent) error
	EventType() string
}

// EventDispatcher はイベントディスパッチャーのインターフェース
type EventDispatcher interface {
	Register(handler EventHandler)
	Dispatch(ctx context.Context, event DomainEvent) error
}

// SimpleEventDispatcher はインメモリイベントディスパッチャーの実装
type SimpleEventDispatcher struct {
	handlers map[string][]EventHandler
}

// NewSimpleEventDispatcher は新しいイベントディスパッチャーを作成します
func NewSimpleEventDispatcher() *SimpleEventDispatcher {
	return &SimpleEventDispatcher{
		handlers: make(map[string][]EventHandler),
	}
}

// Register はイベントハンドラーを登録します
func (d *SimpleEventDispatcher) Register(handler EventHandler) {
	eventType := handler.EventType()
	if _, exists := d.handlers[eventType]; !exists {
		d.handlers[eventType] = make([]EventHandler, 0)
	}
	d.handlers[eventType] = append(d.handlers[eventType], handler)
}

// Dispatch はイベントをディスパッチします
func (d *SimpleEventDispatcher) Dispatch(ctx context.Context, event DomainEvent) error {
	eventType := event.EventType()
	handlers, exists := d.handlers[eventType]
	if !exists {
		// ハンドラーが登録されていない場合は何もしない
		return nil
	}

	// すべてのハンドラーにイベントを配信
	for _, handler := range handlers {
		if err := handler.HandleEvent(ctx, event); err != nil {
			return err
		}
	}

	return nil
}
