package repo

import (
	"context"

	"aipivot/internal/modules/chat/repo/dao"
	"aipivot/internal/shared/po"
)

type MessageRepo struct {
	msgDao *dao.MessageDao
}

func NewMessageRepo(msgDao *dao.MessageDao) *MessageRepo {
	return &MessageRepo{msgDao: msgDao}
}

func (r *MessageRepo) Create(ctx context.Context, msg *po.Message) error {
	return r.msgDao.Create(ctx, msg)
}

func (r *MessageRepo) GetList(ctx context.Context, convID int64, page, pageSize int) ([]*po.Message, int64, error) {
	return r.msgDao.GetList(ctx, convID, page, pageSize)
}

func (r *MessageRepo) GetRecentMessages(ctx context.Context, convID int64, limit int) ([]*po.Message, error) {
	return r.msgDao.GetRecentMessages(ctx, convID, limit)
}
