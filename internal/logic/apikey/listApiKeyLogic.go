// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package apikey

import (
	"context"
	"encoding/json"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListApiKeyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取 API Key 列表
func NewListApiKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListApiKeyLogic {
	return &ListApiKeyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListApiKeyLogic) ListApiKey() (resp *types.ApiKeyListResponse, err error) {
	tenantID := auth.TenantIDFromContext(l.ctx)

	list, err := l.svcCtx.ApiKeyRepo.GetListByTenant(l.ctx, tenantID)
	if err != nil {
		l.Logger.Errorf("ListApiKey err: %v", err)
		return nil, errorx.NewInternalError("获取 API Key 列表失败")
	}

	data := make([]types.ShowApiKey, 0, len(list))
	for _, k := range list {
		var scopes []string
		_ = json.Unmarshal([]byte(k.Scopes), &scopes)

		var lastUsed int64
		if k.LastUsed != nil {
			lastUsed = k.LastUsed.Unix()
		}
		var expiresAt int64
		if k.ExpiresAt != nil {
			expiresAt = k.ExpiresAt.Unix()
		}

		data = append(data, types.ShowApiKey{
			ID:        k.ID,
			Name:      k.Name,
			KeyPrefix: k.KeyPrefix,
			Scopes:    scopes,
			Status:    k.Status,
			LastUsed:  lastUsed,
			ExpiresAt: expiresAt,
			CreatedAt: k.CreatedAt.Unix(),
		})
	}

	return &types.ApiKeyListResponse{
		Code:      0,
		Msg:       "OK",
		Timestamp: time.Now().Unix(),
		Data:      data,
	}, nil
}
