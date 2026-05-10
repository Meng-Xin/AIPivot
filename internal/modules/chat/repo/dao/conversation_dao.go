package dao

import (
	"context"
	"errors"
	"time"

	"aipivot/internal/shared/po"
	"aipivot/internal/shared/query"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type ConversationDao struct {
	q *query.Query
}

func NewConversationDao(q *query.Query) *ConversationDao {
	return &ConversationDao{q: q}
}

func (d *ConversationDao) WithTx(tx *query.Query) *ConversationDao {
	return &ConversationDao{q: tx}
}

func (d *ConversationDao) Create(ctx context.Context, conv *po.Conversation) error {
	err := d.q.Conversation.WithContext(ctx).Create(conv)
	if err != nil {
		logx.WithContext(ctx).Errorf("ConversationDao.Create err: %v", err)
		return err
	}
	return nil
}

func (d *ConversationDao) GetByID(ctx context.Context, id int64) (*po.Conversation, error) {
	c := d.q.Conversation
	result, err := c.WithContext(ctx).Where(c.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logx.WithContext(ctx).Errorf("ConversationDao.GetByID err: %v", err)
		return nil, err
	}
	return result, nil
}

func (d *ConversationDao) GetList(ctx context.Context, tenantID int64, page, pageSize int, status string) ([]*po.Conversation, int64, error) {
	c := d.q.Conversation
	q := c.WithContext(ctx).Where(c.TenantID.Eq(tenantID))

	if status != "" {
		q = q.Where(c.Status.Eq(status))
	}

	total, err := q.Count()
	if err != nil {
		logx.WithContext(ctx).Errorf("ConversationDao.GetList count err: %v", err)
		return nil, 0, err
	}

	list, err := q.Order(c.CreatedAt.Desc()).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find()
	if err != nil {
		logx.WithContext(ctx).Errorf("ConversationDao.GetList find err: %v", err)
		return nil, 0, err
	}

	return list, total, nil
}

func (d *ConversationDao) Close(ctx context.Context, id int64) error {
	c := d.q.Conversation
	now := time.Now()
	_, err := c.WithContext(ctx).Where(c.ID.Eq(id)).Updates(map[string]any{
		"status":    "closed",
		"closed_at": &now,
	})
	if err != nil {
		logx.WithContext(ctx).Errorf("ConversationDao.Close err: %v", err)
		return err
	}
	return nil
}

// UpdateStatus 更新会话状态（如 active → waiting_human → resolved）。
func (d *ConversationDao) UpdateStatus(ctx context.Context, id int64, status string) error {
	c := d.q.Conversation
	_, err := c.WithContext(ctx).Where(c.ID.Eq(id)).Update(c.Status, status)
	if err != nil {
		logx.WithContext(ctx).Errorf("ConversationDao.UpdateStatus err: %v", err)
		return err
	}
	return nil
}

func (d *ConversationDao) IncrMessageCount(ctx context.Context, id int64) error {
	c := d.q.Conversation
	_, err := c.WithContext(ctx).Where(c.ID.Eq(id)).UpdateSimple(c.MessageCount.Add(1))
	if err != nil {
		logx.WithContext(ctx).Errorf("ConversationDao.IncrMessageCount err: %v", err)
		return err
	}
	return nil
}
