package dao

import (
	"context"
	"errors"

	"aipivot/internal/shared/po"
	"aipivot/internal/shared/query"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DocumentDao struct {
	q *query.Query
}

func NewDocumentDao(q *query.Query) *DocumentDao {
	return &DocumentDao{q: q}
}

func (d *DocumentDao) WithTx(tx *query.Query) *DocumentDao {
	return &DocumentDao{q: tx}
}

func (d *DocumentDao) Create(ctx context.Context, doc *po.Document) error {
	err := d.q.Document.WithContext(ctx).Create(doc)
	if err != nil {
		logx.WithContext(ctx).Errorf("DocumentDao.Create err: %v", err)
		return err
	}
	return nil
}

func (d *DocumentDao) GetByID(ctx context.Context, id int64) (*po.Document, error) {
	doc := d.q.Document
	result, err := doc.WithContext(ctx).Where(doc.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logx.WithContext(ctx).Errorf("DocumentDao.GetByID err: %v", err)
		return nil, err
	}
	return result, nil
}

func (d *DocumentDao) GetList(ctx context.Context, kbID int64, page, pageSize int, status string) ([]*po.Document, int64, error) {
	doc := d.q.Document
	q := doc.WithContext(ctx).Where(doc.KnowledgeBaseID.Eq(kbID))

	if status != "" {
		q = q.Where(doc.Status.Eq(status))
	}

	total, err := q.Count()
	if err != nil {
		logx.WithContext(ctx).Errorf("DocumentDao.GetList count err: %v", err)
		return nil, 0, err
	}

	list, err := q.Order(doc.CreatedAt.Desc()).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find()
	if err != nil {
		logx.WithContext(ctx).Errorf("DocumentDao.GetList find err: %v", err)
		return nil, 0, err
	}

	return list, total, nil
}

func (d *DocumentDao) Delete(ctx context.Context, id int64) error {
	doc := d.q.Document
	_, err := doc.WithContext(ctx).Where(doc.ID.Eq(id)).Delete()
	if err != nil {
		logx.WithContext(ctx).Errorf("DocumentDao.Delete err: %v", err)
		return err
	}
	return nil
}

func (d *DocumentDao) UpdateStatus(ctx context.Context, id int64, status, errorMsg string) error {
	doc := d.q.Document
	_, err := doc.WithContext(ctx).Where(doc.ID.Eq(id)).Updates(map[string]any{
		"status":    status,
		"error_msg": errorMsg,
	})
	if err != nil {
		logx.WithContext(ctx).Errorf("DocumentDao.UpdateStatus err: %v", err)
		return err
	}
	return nil
}
