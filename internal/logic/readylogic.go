package logic

import (
	"context"

	"aipivot/internal/infra"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReadyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReadyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReadyLogic {
	return &ReadyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReadyLogic) Ready() (*types.ReadyResponse, bool, error) {
	result := infra.CheckDependencies(l.ctx, l.svcCtx.HealthChecks)
	dependencies := make([]types.DependencyStatus, 0, len(result.Dependencies))

	for _, dep := range result.Dependencies {
		if l.svcCtx.Metrics != nil {
			l.svcCtx.Metrics.SetDependencyReady(dep.Name, dep.Ready)
		}

		dependencies = append(dependencies, types.DependencyStatus{
			Name:  dep.Name,
			Ready: dep.Ready,
			Error: dep.Error,
		})
	}

	return &types.ReadyResponse{
		Status:       result.Status,
		Dependencies: dependencies,
	}, result.Ready, nil
}
