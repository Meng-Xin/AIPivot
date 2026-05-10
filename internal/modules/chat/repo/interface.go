package repo

import (
	"context"

	"aipivot/internal/shared/po"
)

type ConversationRepository interface {
	Create(ctx context.Context, conv *po.Conversation) error
	GetByID(ctx context.Context, id int64) (*po.Conversation, error)
	GetList(ctx context.Context, tenantID int64, page, pageSize int, status string) ([]*po.Conversation, int64, error)
	Close(ctx context.Context, id int64) error
	UpdateStatus(ctx context.Context, id int64, status string) error
	IncrMessageCount(ctx context.Context, id int64) error
}

type MessageRepository interface {
	Create(ctx context.Context, msg *po.Message) error
	GetList(ctx context.Context, convID int64, page, pageSize int) ([]*po.Message, int64, error)
	GetRecentMessages(ctx context.Context, convID int64, limit int) ([]*po.Message, error)
}
