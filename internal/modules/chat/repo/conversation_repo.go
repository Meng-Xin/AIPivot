package repo

import (
	"context"

	"aipivot/internal/modules/chat/repo/dao"
	"aipivot/internal/shared/po"
)

type ConversationRepo struct {
	convDao *dao.ConversationDao
}

func NewConversationRepo(convDao *dao.ConversationDao) *ConversationRepo {
	return &ConversationRepo{convDao: convDao}
}

func (r *ConversationRepo) Create(ctx context.Context, conv *po.Conversation) error {
	return r.convDao.Create(ctx, conv)
}

func (r *ConversationRepo) GetByID(ctx context.Context, id int64) (*po.Conversation, error) {
	return r.convDao.GetByID(ctx, id)
}

func (r *ConversationRepo) GetList(ctx context.Context, tenantID int64, page, pageSize int, status string) ([]*po.Conversation, int64, error) {
	return r.convDao.GetList(ctx, tenantID, page, pageSize, status)
}

func (r *ConversationRepo) Close(ctx context.Context, id int64) error {
	return r.convDao.Close(ctx, id)
}

func (r *ConversationRepo) IncrMessageCount(ctx context.Context, id int64) error {
	return r.convDao.IncrMessageCount(ctx, id)
}
