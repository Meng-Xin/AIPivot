package chat

import (
	"context"
	"errors"
	"time"

	"aipivot/internal/shared/po"
	"aipivot/internal/shared/query"

	"gorm.io/gorm"
)

// ConversationRepo 会话数据仓储实现。
type ConversationRepo struct {
	q *query.Query
}

func NewConversationRepo(q *query.Query) *ConversationRepo {
	return &ConversationRepo{q: q}
}

func (r *ConversationRepo) Create(ctx context.Context, conv *po.Conversation) error {
	return r.q.Conversation.WithContext(ctx).Create(conv)
}

func (r *ConversationRepo) GetByID(ctx context.Context, id int64) (*po.Conversation, error) {
	c := r.q.Conversation
	result, err := c.WithContext(ctx).Where(c.ID.Eq(id)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return result, err
}

func (r *ConversationRepo) GetList(ctx context.Context, tenantID int64, page, pageSize int, status string) ([]*po.Conversation, int64, error) {
	c := r.q.Conversation
	q := c.WithContext(ctx).Where(c.TenantID.Eq(tenantID))

	if status != "" {
		q = q.Where(c.Status.Eq(status))
	}

	total, err := q.Count()
	if err != nil {
		return nil, 0, err
	}

	list, err := q.Order(c.CreatedAt.Desc()).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find()
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *ConversationRepo) Close(ctx context.Context, id int64) error {
	c := r.q.Conversation
	now := time.Now()
	_, err := c.WithContext(ctx).Where(c.ID.Eq(id)).Updates(map[string]any{
		"status":    "closed",
		"closed_at": &now,
	})
	return err
}

// UpdateStatus 更新会话状态（如 active → waiting_human → resolved）。
func (r *ConversationRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	c := r.q.Conversation
	_, err := c.WithContext(ctx).Where(c.ID.Eq(id)).Update(c.Status, status)
	return err
}

func (r *ConversationRepo) IncrMessageCount(ctx context.Context, id int64) error {
	c := r.q.Conversation
	_, err := c.WithContext(ctx).Where(c.ID.Eq(id)).UpdateSimple(c.MessageCount.Add(1))
	return err
}
