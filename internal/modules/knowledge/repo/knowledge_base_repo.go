package repo

import (
	"context"

	"aipivot/internal/modules/knowledge/repo/dao"
	"aipivot/internal/shared/po"
)

type KnowledgeBaseRepo struct {
	kbDao *dao.KnowledgeBaseDao
}

func NewKnowledgeBaseRepo(kbDao *dao.KnowledgeBaseDao) *KnowledgeBaseRepo {
	return &KnowledgeBaseRepo{kbDao: kbDao}
}

func (r *KnowledgeBaseRepo) Create(ctx context.Context, kb *po.KnowledgeBase) error {
	return r.kbDao.Create(ctx, kb)
}

func (r *KnowledgeBaseRepo) GetByID(ctx context.Context, id int64) (*po.KnowledgeBase, error) {
	return r.kbDao.GetByID(ctx, id)
}

func (r *KnowledgeBaseRepo) GetList(ctx context.Context, tenantID int64, page, pageSize int, name string) ([]*po.KnowledgeBase, int64, error) {
	return r.kbDao.GetList(ctx, tenantID, page, pageSize, name)
}

func (r *KnowledgeBaseRepo) Update(ctx context.Context, id int64, updates map[string]any) error {
	return r.kbDao.Update(ctx, id, updates)
}

func (r *KnowledgeBaseRepo) Delete(ctx context.Context, id int64) error {
	return r.kbDao.Delete(ctx, id)
}
