// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package apikey

import (
	"context"
	"time"

	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RevokeApiKeyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 吊销 API Key
func NewRevokeApiKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RevokeApiKeyLogic {
	return &RevokeApiKeyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RevokeApiKeyLogic) RevokeApiKey(req *types.RevokeApiKeyRequest) (resp *types.CommResponse, err error) {
	if err = l.svcCtx.ApiKeyRepo.Revoke(l.ctx, req.ID); err != nil {
		l.Logger.Errorf("RevokeApiKey err: %v", err)
		return nil, errorx.NewInternalError("吊销 API Key 失败")
	}

	return &types.CommResponse{
		Code:      0,
		Msg:       "吊销成功",
		Timestamp: time.Now().Unix(),
	}, nil
}
