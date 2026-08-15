// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package open

import (
	"context"
	"errors"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/modules/channel"
	"aipivot/internal/repository/chat"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WidgetMessageFeedbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Widget — 提交消息满意度评分
func NewWidgetMessageFeedbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WidgetMessageFeedbackLogic {
	return &WidgetMessageFeedbackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

const widgetFeedbackMaxLen = 500

// WidgetMessageFeedback 校验 sessionToken → 鉴权 → 调 Repo 更新评分。
// 第一波为锁定语义：已评分消息允许重复调用直接覆盖（前端锁定 UI 不会发起第二次），后端不引入额外状态机。
func (l *WidgetMessageFeedbackLogic) WidgetMessageFeedback(req *types.WidgetMessageFeedbackRequest) (resp *types.CommResponse, err error) {
	tenantID := auth.TenantIDFromAPIKeyContext(l.ctx)
	if tenantID == 0 {
		return nil, errorx.NewUnauthError("invalid API key context")
	}
	if !auth.IsPublicApiKey(l.ctx) {
		return nil, errorx.NewBusinessError(errorx.CodeBadRequest, "public key required")
	}

	// 评分白名单
	if req.Rating != "up" && req.Rating != "down" {
		return nil, errorx.NewBusinessError(errorx.CodeBadRequest, "rating 仅支持 up / down")
	}
	// feedback 仅在 down 时有意义，但允许 up 也带空串；超长统一拒绝
	if len(req.Feedback) > widgetFeedbackMaxLen {
		return nil, errorx.NewBusinessError(errorx.CodeBadRequest, "反馈内容超过 500 字")
	}

	// 会话归属校验：sessionToken = conversation UUID
	conv, err := l.svcCtx.ConversationRepo.GetByUUID(l.ctx, req.SessionToken)
	if err != nil || conv == nil {
		return nil, errorx.NewNotFoundError("会话不存在")
	}
	if conv.TenantID != tenantID || conv.Channel != channel.Widget.String() {
		return nil, errorx.NewForbidError("无权访问该会话")
	}

	// 仅 assistant 消息允许评分（业务约束）：直接走 RateMessage，由 SQL 跨租户/不存在兜底
	if err = l.svcCtx.MessageRepo.RateMessage(l.ctx, req.MessageID, tenantID, req.Rating, req.Feedback); err != nil {
		if errors.Is(err, chat.ErrMessageNotFound) {
			return nil, errorx.NewNotFoundError("消息不存在或不可评分")
		}
		l.Logger.Errorf("WidgetMessageFeedback RateMessage err: %v", err)
		return nil, errorx.NewInternalError("评分保存失败")
	}

	return &types.CommResponse{
		Code:      0,
		Msg:       "OK",
		Timestamp: time.Now().Unix(),
	}, nil
}
