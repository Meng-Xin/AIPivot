// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package infra

import (
	"context"

	"aipivot/internal/observability"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingLogic {
	return &PingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PingLogic) Ping() (resp *types.PingResponse, err error) {
	return &types.PingResponse{
		Message:   "pong",
		TraceId:   observability.TraceIDFromContext(l.ctx),
		RequestId: observability.RequestIDFromContext(l.ctx),
	}, nil
}
