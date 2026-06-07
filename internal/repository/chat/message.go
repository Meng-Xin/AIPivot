package chat

import (
	"context"

	"aipivot/internal/shared/po"
	"aipivot/internal/shared/query"
)

// MessageRepo 消息数据仓储实现。
type MessageRepo struct {
	q *query.Query
}

func NewMessageRepo(q *query.Query) *MessageRepo {
	return &MessageRepo{q: q}
}

func (r *MessageRepo) Create(ctx context.Context, msg *po.Message) error {
	if msg.Sources == "" {
		msg.Sources = "[]"
	}
	if msg.Metadata == "" {
		msg.Metadata = "{}"
	}
	return r.q.Message.WithContext(ctx).Create(msg)
}

func (r *MessageRepo) GetList(ctx context.Context, convID int64, page, pageSize int) ([]*po.Message, int64, error) {
	m := r.q.Message
	q := m.WithContext(ctx).Where(m.ConversationID.Eq(convID))

	total, err := q.Count()
	if err != nil {
		return nil, 0, err
	}

	list, err := q.Order(m.CreatedAt).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find()
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// GetRecentMessages 获取最近 N 条消息，用于构建 LLM 上下文窗口
func (r *MessageRepo) GetRecentMessages(ctx context.Context, convID int64, limit int) ([]*po.Message, error) {
	m := r.q.Message
	list, err := m.WithContext(ctx).
		Where(m.ConversationID.Eq(convID)).
		Order(m.CreatedAt.Desc()).
		Limit(limit).
		Find()
	if err != nil {
		return nil, err
	}

	// 反转为时间正序，便于构建 LLM 消息列表
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	return list, nil
}
