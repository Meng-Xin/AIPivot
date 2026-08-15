package chat

import (
	"context"

	"aipivot/internal/shared/po"
)

// ========== Repository 接口 ==========

type ConversationRepository interface {
	Create(ctx context.Context, conv *po.Conversation) error
	GetByID(ctx context.Context, id int64) (*po.Conversation, error)
	GetByUUID(ctx context.Context, uuid string) (*po.Conversation, error)
	GetList(ctx context.Context, tenantID int64, page, pageSize int, status string) ([]*po.Conversation, int64, error)
	Close(ctx context.Context, id int64) error
	UpdateStatus(ctx context.Context, id int64, status string) error
	IncrMessageCount(ctx context.Context, id int64) error
}

type MessageRepository interface {
	Create(ctx context.Context, msg *po.Message) error
	GetList(ctx context.Context, convID int64, page, pageSize int) ([]*po.Message, int64, error)
	GetRecentMessages(ctx context.Context, convID int64, limit int) ([]*po.Message, error)
	// RateMessage 按 UUID + tenantID 更新评分列。RowsAffected==0 视为消息不存在或越权。
	RateMessage(ctx context.Context, msgUUID string, tenantID int64, rating string, feedback string) error
}
