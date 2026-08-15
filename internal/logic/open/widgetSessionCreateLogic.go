// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package open

import (
	"context"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/modules/channel"
	"aipivot/internal/modules/knowledge"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/shared/po"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type WidgetSessionCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Widget — 创建访客会话（返回 sessionToken 作为后续凭证）
func NewWidgetSessionCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WidgetSessionCreateLogic {
	return &WidgetSessionCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// WidgetSessionCreate 为访客创建持久化会话。
// 仅允许 public key 调用，会话绑定 public key 关联的知识库，便于后续 RAG 检索。
func (l *WidgetSessionCreateLogic) WidgetSessionCreate(req *types.WidgetSessionCreateRequest) (resp *types.WidgetSessionResponse, err error) {
	tenantID := auth.TenantIDFromAPIKeyContext(l.ctx)
	if tenantID == 0 {
		return nil, errorx.NewUnauthError("invalid API key context")
	}
	if !auth.IsPublicApiKey(l.ctx) {
		return nil, errorx.NewBusinessError(errorx.CodeBadRequest, "public key required")
	}
	if req.VisitorID == "" {
		return nil, errorx.NewBusinessError(errorx.CodeBadRequest, "visitorId 不能为空")
	}

	title := req.Title
	if title == "" {
		title = "Widget: " + req.VisitorID
	}

	conv := &po.Conversation{
		UUID:           uuid.New().String(), // 显式设置 UUID（项目约定）
		TenantID:       tenantID,
		Title:          title,
		Status:         "active",
		Channel:        channel.Widget.String(),
		ExternalUserID: req.VisitorID,
	}

	// 引导问答：读绑定的 KB.suggested_questions，建会话时一次性返回给 Widget 首屏渲染
	var suggested []string
	if kbID, ok := auth.KnowledgeBaseIDFromApiKeyContext(l.ctx); ok {
		conv.KnowledgeBaseID = &kbID
		if kb, kbErr := l.svcCtx.KnowledgeBaseRepo.GetByID(l.ctx, kbID); kbErr == nil && kb != nil {
			suggested = knowledge.DeserializeSuggestedQuestions(kb.SuggestedQuestions)
		} else if kbErr != nil {
			l.Logger.Errorf("WidgetSessionCreate load KB suggested err: %v", kbErr)
		}
	}

	if err = l.svcCtx.ConversationRepo.Create(l.ctx, conv); err != nil {
		l.Logger.Errorf("WidgetSessionCreate CreateConversation err: %v", err)
		return nil, errorx.NewInternalError("创建会话失败")
	}

	return &types.WidgetSessionResponse{
		Code:      0,
		Msg:       "OK",
		Timestamp: time.Now().Unix(),
		Data: types.WidgetSessionData{
			SessionToken:       conv.UUID,
			ConversationID:     conv.ID,
			VisitorID:          req.VisitorID,
			CreatedAt:          conv.CreatedAt.Unix(),
			SuggestedQuestions: suggested,
		},
	}, nil
}
