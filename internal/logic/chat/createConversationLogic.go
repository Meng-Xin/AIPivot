// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package chat

import (
	"context"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/modules/chat/domain/assembler"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/shared/po"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateConversationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建会话
func NewCreateConversationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateConversationLogic {
	return &CreateConversationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateConversationLogic) CreateConversation(req *types.CreateConversationRequest) (resp *types.ConversationDetailResponse, err error) {
	tenantID := auth.TenantIDFromContext(l.ctx)
	userID := auth.UserIDFromContext(l.ctx)

	conv := &po.Conversation{
		UUID:     uuid.New().String(),
		TenantID: tenantID,
		UserID:   &userID,
		Title:    req.Title,
		Status:   "active",
		Channel:  "web",
	}

	if req.KnowledgeBaseID > 0 {
		conv.KnowledgeBaseID = &req.KnowledgeBaseID
	}

	if conv.Title == "" {
		conv.Title = "新对话"
	}

	if err = l.svcCtx.ConversationRepo.Create(l.ctx, conv); err != nil {
		l.Logger.Errorf("CreateConversation err: %v", err)
		return nil, errorx.NewInternalError("创建会话失败")
	}

	show := assembler.ConversationPoToShow(conv)
	return &types.ConversationDetailResponse{
		Code:      0,
		Msg:       "创建成功",
		Timestamp: time.Now().Unix(),
		Data:      show,
	}, nil
}
