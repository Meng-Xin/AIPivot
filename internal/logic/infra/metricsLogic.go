// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package infra

import (
	"context"

	"aipivot/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type MetricsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMetricsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MetricsLogic {
	return &MetricsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MetricsLogic) Metrics() error {
	// todo: add your logic here and delete this line

	return nil
}
