package webhook

import (
	"context"
	"errors"

	"aipivot/internal/shared/po"
	"aipivot/internal/shared/query"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type WebhookDao struct {
	q  *query.Query
	db *gorm.DB // JSONB 查询需要原生 SQL
}

func NewWebhookDao(q *query.Query, db *gorm.DB) *WebhookDao {
	return &WebhookDao{q: q, db: db}
}

func (d *WebhookDao) Create(ctx context.Context, wh *po.Webhook) error {
	err := d.q.Webhook.WithContext(ctx).Create(wh)
	if err != nil {
		logx.WithContext(ctx).Errorf("WebhookDao.Create err: %v", err)
		return err
	}
	return nil
}

func (d *WebhookDao) GetByID(ctx context.Context, id int64) (*po.Webhook, error) {
	w := d.q.Webhook
	result, err := w.WithContext(ctx).Where(w.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logx.WithContext(ctx).Errorf("WebhookDao.GetByID err: %v", err)
		return nil, err
	}
	return result, nil
}

func (d *WebhookDao) GetListByTenant(ctx context.Context, tenantID int64) ([]*po.Webhook, error) {
	w := d.q.Webhook
	list, err := w.WithContext(ctx).
		Where(w.TenantID.Eq(tenantID)).
		Order(w.CreatedAt.Desc()).
		Find()
	if err != nil {
		logx.WithContext(ctx).Errorf("WebhookDao.GetListByTenant err: %v", err)
		return nil, err
	}
	return list, nil
}

// GetActiveByTenantAndEvent 查询租户下订阅了指定事件类型的所有活跃 Webhook。
// events 使用 PostgreSQL JSONB @> 包含查询，需要原生 SQL。
func (d *WebhookDao) GetActiveByTenantAndEvent(ctx context.Context, tenantID int64, event string) ([]*po.Webhook, error) {
	var list []*po.Webhook
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND status = ? AND events @> ?", tenantID, "active", `["`+event+`"]`).
		Find(&list).Error
	if err != nil {
		logx.WithContext(ctx).Errorf("WebhookDao.GetActiveByTenantAndEvent err: %v", err)
		return nil, err
	}
	return list, nil
}

func (d *WebhookDao) Update(ctx context.Context, wh *po.Webhook) error {
	w := d.q.Webhook
	_, err := w.WithContext(ctx).Where(w.ID.Eq(wh.ID)).Updates(wh)
	if err != nil {
		logx.WithContext(ctx).Errorf("WebhookDao.Update err: %v", err)
		return err
	}
	return nil
}

func (d *WebhookDao) Delete(ctx context.Context, id int64) error {
	w := d.q.Webhook
	_, err := w.WithContext(ctx).Where(w.ID.Eq(id)).Delete()
	if err != nil {
		logx.WithContext(ctx).Errorf("WebhookDao.Delete err: %v", err)
		return err
	}
	return nil
}
