// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package knowledge

import (
	"context"
	"time"

	"aipivot/internal/modules/knowledge"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateKnowledgeBaseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新知识库
func NewUpdateKnowledgeBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateKnowledgeBaseLogic {
	return &UpdateKnowledgeBaseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateKnowledgeBaseLogic) UpdateKnowledgeBase(req *types.UpdateKnowledgeBaseRequest) (resp *types.CommResponse, err error) {
	kb, err := l.svcCtx.KnowledgeBaseRepo.GetByID(l.ctx, req.ID)
	if err != nil {
		l.Logger.Errorf("UpdateKnowledgeBase GetByID err: %v", err)
		return nil, errorx.NewInternalError("更新知识库失败")
	}
	if kb == nil {
		return nil, errorx.NewNotFoundError("知识库不存在")
	}

	updates := make(map[string]any)
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	// nil 切片表示不更新；非 nil（含空数组）才落库，便于清空引导问答
	if req.SuggestedQuestions != nil {
		updates["suggested_questions"] = knowledge.SerializeSuggestedQuestions(
			knowledge.NormalizeSuggestedQuestions(req.SuggestedQuestions),
		)
	}

	if len(updates) > 0 {
		if err = l.svcCtx.KnowledgeBaseRepo.Update(l.ctx, req.ID, updates); err != nil {
			l.Logger.Errorf("UpdateKnowledgeBase err: %v", err)
			return nil, errorx.NewInternalError("更新知识库失败")
		}
	}

	return &types.CommResponse{
		Code:      0,
		Msg:       "更新成功",
		Timestamp: time.Now().Unix(),
	}, nil
}
