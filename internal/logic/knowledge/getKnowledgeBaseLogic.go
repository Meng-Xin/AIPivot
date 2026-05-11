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

type GetKnowledgeBaseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取知识库详情
func NewGetKnowledgeBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetKnowledgeBaseLogic {
	return &GetKnowledgeBaseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetKnowledgeBaseLogic) GetKnowledgeBase(req *types.GetKnowledgeBaseRequest) (resp *types.KnowledgeBaseDetailResponse, err error) {
	kb, err := l.svcCtx.KnowledgeBaseRepo.GetByID(l.ctx, req.ID)
	if err != nil {
		l.Logger.Errorf("GetKnowledgeBase err: %v", err)
		return nil, errorx.NewInternalError("查询知识库失败")
	}
	if kb == nil {
		return nil, errorx.NewNotFoundError("知识库不存在")
	}

	show := knowledge.ToShowKB(kb)
	return &types.KnowledgeBaseDetailResponse{
		Code:      0,
		Msg:       "OK",
		Timestamp: time.Now().Unix(),
		Data:      show,
	}, nil
}
