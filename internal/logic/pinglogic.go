package logic

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

func (l *PingLogic) Ping() (*types.PingResponse, error) {
	return &types.PingResponse{
		Message:   "pong",
		TraceID:   observability.TraceIDFromContext(l.ctx),
		RequestID: observability.RequestIDFromContext(l.ctx),
	}, nil
}
