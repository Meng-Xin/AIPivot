// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package knowledge

import (
	"context"
	"time"

	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteKnowledgeBaseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除知识库
func NewDeleteKnowledgeBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteKnowledgeBaseLogic {
	return &DeleteKnowledgeBaseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteKnowledgeBaseLogic) DeleteKnowledgeBase(req *types.DeleteKnowledgeBaseRequest) (resp *types.CommResponse, err error) {
	kb, err := l.svcCtx.KnowledgeBaseRepo.GetByID(l.ctx, req.ID)
	if err != nil {
		l.Logger.Errorf("DeleteKnowledgeBase GetByID err: %v", err)
		return nil, errorx.NewInternalError("删除知识库失败")
	}
	if kb == nil {
		return nil, errorx.NewNotFoundError("知识库不存在")
	}

	// 级联删除由数据库外键 ON DELETE CASCADE 保证
	if err = l.svcCtx.KnowledgeBaseRepo.Delete(l.ctx, req.ID); err != nil {
		l.Logger.Errorf("DeleteKnowledgeBase err: %v", err)
		return nil, errorx.NewInternalError("删除知识库失败")
	}

	return &types.CommResponse{
		Code:      0,
		Msg:       "删除成功",
		Timestamp: time.Now().Unix(),
	}, nil
}
