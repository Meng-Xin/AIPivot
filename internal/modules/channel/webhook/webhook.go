package webhook

import (
	"context"

	"aipivot/internal/shared/po"
)

// Repository Webhook 数据仓储接口。
type Repository interface {
	Create(ctx context.Context, wh *po.Webhook) error
	GetByID(ctx context.Context, id int64) (*po.Webhook, error)
	GetListByTenant(ctx context.Context, tenantID int64) ([]*po.Webhook, error)
	GetActiveByTenantAndEvent(ctx context.Context, tenantID int64, event string) ([]*po.Webhook, error)
	Update(ctx context.Context, wh *po.Webhook) error
	Delete(ctx context.Context, id int64) error
}
