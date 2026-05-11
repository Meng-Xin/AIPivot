package webhook

import (
	"context"
	"errors"

	"aipivot/internal/shared/po"
	"aipivot/internal/shared/query"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// WebhookRepo Webhook 数据仓储实现（合并原 DAO + Repo 两层）。
type WebhookRepo struct {
	q  *query.Query
	db *gorm.DB // JSONB 查询需要原生 SQL
}

func NewWebhookRepo(q *query.Query, db *gorm.DB) *WebhookRepo {
	return &WebhookRepo{q: q, db: db}
}

func (r *WebhookRepo) Create(ctx context.Context, wh *po.Webhook) error {
	err := r.q.Webhook.WithContext(ctx).Create(wh)
	if err != nil {
		logx.WithContext(ctx).Errorf("WebhookRepo.Create err: %v", err)
		return err
	}
	return nil
}

func (r *WebhookRepo) GetByID(ctx context.Context, id int64) (*po.Webhook, error) {
	w := r.q.Webhook
	result, err := w.WithContext(ctx).Where(w.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logx.WithContext(ctx).Errorf("WebhookRepo.GetByID err: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *WebhookRepo) GetListByTenant(ctx context.Context, tenantID int64) ([]*po.Webhook, error) {
	w := r.q.Webhook
	list, err := w.WithContext(ctx).
		Where(w.TenantID.Eq(tenantID)).
		Order(w.CreatedAt.Desc()).
		Find()
	if err != nil {
		logx.WithContext(ctx).Errorf("WebhookRepo.GetListByTenant err: %v", err)
		return nil, err
	}
	return list, nil
}

// GetActiveByTenantAndEvent 查询租户下订阅了指定事件类型的所有活跃 Webhook。
// events 使用 PostgreSQL JSONB @> 包含查询，需要原生 SQL。
func (r *WebhookRepo) GetActiveByTenantAndEvent(ctx context.Context, tenantID int64, event string) ([]*po.Webhook, error) {
	var list []*po.Webhook
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND status = ? AND events @> ?", tenantID, "active", `["`+event+`"]`).
		Find(&list).Error
	if err != nil {
		logx.WithContext(ctx).Errorf("WebhookRepo.GetActiveByTenantAndEvent err: %v", err)
		return nil, err
	}
	return list, nil
}

func (r *WebhookRepo) Update(ctx context.Context, wh *po.Webhook) error {
	w := r.q.Webhook
	_, err := w.WithContext(ctx).Where(w.ID.Eq(wh.ID)).Updates(wh)
	if err != nil {
		logx.WithContext(ctx).Errorf("WebhookRepo.Update err: %v", err)
		return err
	}
	return nil
}

func (r *WebhookRepo) Delete(ctx context.Context, id int64) error {
	w := r.q.Webhook
	_, err := w.WithContext(ctx).Where(w.ID.Eq(id)).Delete()
	if err != nil {
		logx.WithContext(ctx).Errorf("WebhookRepo.Delete err: %v", err)
		return err
	}
	return nil
}
