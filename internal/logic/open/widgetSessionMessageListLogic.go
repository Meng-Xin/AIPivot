// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package open

import (
	"context"
	"encoding/json"

	"aipivot/internal/modules/auth"
	"aipivot/internal/modules/channel"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WidgetSessionMessageListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Widget — 拉取会话历史消息
func NewWidgetSessionMessageListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WidgetSessionMessageListLogic {
	return &WidgetSessionMessageListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// WidgetSessionMessageList 拉取访客会话历史消息。
// 通过 sessionToken（= conversation UUID）定位会话，校验同租户 + widget 渠道。
func (l *WidgetSessionMessageListLogic) WidgetSessionMessageList(req *types.WidgetMessageListRequest) (resp *types.WidgetMessageListResponse, err error) {
	tenantID := auth.TenantIDFromAPIKeyContext(l.ctx)
	if tenantID == 0 {
		return nil, errorx.NewUnauthError("invalid API key context")
	}

	conv, err := l.svcCtx.ConversationRepo.GetByUUID(l.ctx, req.SessionToken)
	if err != nil || conv == nil {
		return nil, errorx.NewNotFoundError("会话不存在")
	}
	if conv.TenantID != tenantID || conv.Channel != channel.Widget.String() {
		return nil, errorx.NewForbidError("无权访问该会话")
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	msgs, _, err := l.svcCtx.MessageRepo.GetList(l.ctx, conv.ID, page, pageSize)
	if err != nil {
		l.Logger.Errorf("WidgetSessionMessageList GetList err: %v", err)
		return nil, errorx.NewInternalError("加载历史消息失败")
	}

	items := make([]types.WidgetMessageItem, 0, len(msgs))
	for _, m := range msgs {
		var sources []string
		if m.Sources != "" {
			_ = json.Unmarshal([]byte(m.Sources), &sources)
		}
		items = append(items, types.WidgetMessageItem{
			UUID:           m.UUID,
			Role:           m.Role,
			Content:        m.Content,
			ContentType:    m.ContentType,
			TokenCount:     m.TokenCount,
			Model:          m.Model,
			Sources:        sources,
			Rating:         m.Rating,
			RatingFeedback: m.RatingFeedback,
			CreatedAt:      m.CreatedAt.Unix(),
		})
	}

	return &types.WidgetMessageListResponse{
		Code: 0,
		Msg:  "OK",
		Data: items,
	}, nil
}
