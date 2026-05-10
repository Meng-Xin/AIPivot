package dao

import (
	"context"

	"aipivot/internal/shared/po"
	"aipivot/internal/shared/query"

	"github.com/zeromicro/go-zero/core/logx"
)

type MessageDao struct {
	q *query.Query
}

func NewMessageDao(q *query.Query) *MessageDao {
	return &MessageDao{q: q}
}

func (d *MessageDao) WithTx(tx *query.Query) *MessageDao {
	return &MessageDao{q: tx}
}

func (d *MessageDao) Create(ctx context.Context, msg *po.Message) error {
	err := d.q.Message.WithContext(ctx).Create(msg)
	if err != nil {
		logx.WithContext(ctx).Errorf("MessageDao.Create err: %v", err)
		return err
	}
	return nil
}

func (d *MessageDao) GetList(ctx context.Context, convID int64, page, pageSize int) ([]*po.Message, int64, error) {
	m := d.q.Message
	q := m.WithContext(ctx).Where(m.ConversationID.Eq(convID))

	total, err := q.Count()
	if err != nil {
		logx.WithContext(ctx).Errorf("MessageDao.GetList count err: %v", err)
		return nil, 0, err
	}

	list, err := q.Order(m.CreatedAt).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find()
	if err != nil {
		logx.WithContext(ctx).Errorf("MessageDao.GetList find err: %v", err)
		return nil, 0, err
	}

	return list, total, nil
}

// GetRecentMessages 获取最近 N 条消息，用于构建 LLM 上下文窗口
func (d *MessageDao) GetRecentMessages(ctx context.Context, convID int64, limit int) ([]*po.Message, error) {
	m := d.q.Message
	list, err := m.WithContext(ctx).
		Where(m.ConversationID.Eq(convID)).
		Order(m.CreatedAt.Desc()).
		Limit(limit).
		Find()
	if err != nil {
		logx.WithContext(ctx).Errorf("MessageDao.GetRecentMessages err: %v", err)
		return nil, err
	}

	// 反转为时间正序，便于构建 LLM 消息列表
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	return list, nil
}
