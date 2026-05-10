package repo

import (
	"context"

	"aipivot/internal/modules/knowledge/repo/dao"
	"aipivot/internal/shared/po"
)

type DocumentRepo struct {
	docDao *dao.DocumentDao
}

func NewDocumentRepo(docDao *dao.DocumentDao) *DocumentRepo {
	return &DocumentRepo{docDao: docDao}
}

func (r *DocumentRepo) Create(ctx context.Context, doc *po.Document) error {
	return r.docDao.Create(ctx, doc)
}

func (r *DocumentRepo) GetByID(ctx context.Context, id int64) (*po.Document, error) {
	return r.docDao.GetByID(ctx, id)
}

func (r *DocumentRepo) GetList(ctx context.Context, kbID int64, page, pageSize int, status string) ([]*po.Document, int64, error) {
	return r.docDao.GetList(ctx, kbID, page, pageSize, status)
}

func (r *DocumentRepo) Delete(ctx context.Context, id int64) error {
	return r.docDao.Delete(ctx, id)
}

func (r *DocumentRepo) UpdateStatus(ctx context.Context, id int64, status, errorMsg string) error {
	return r.docDao.UpdateStatus(ctx, id, status, errorMsg)
}
