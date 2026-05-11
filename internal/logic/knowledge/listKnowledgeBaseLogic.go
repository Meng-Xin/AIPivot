// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package knowledge

import (
	"context"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/modules/knowledge"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListKnowledgeBaseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取知识库列表
func NewListKnowledgeBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListKnowledgeBaseLogic {
	return &ListKnowledgeBaseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListKnowledgeBaseLogic) ListKnowledgeBase(req *types.ListKnowledgeBaseRequest) (resp *types.KnowledgeBaseListResponse, err error) {
	tenantID := auth.TenantIDFromContext(l.ctx)

	list, total, err := l.svcCtx.KnowledgeBaseRepo.GetList(l.ctx, tenantID, req.Page, req.PageSize, req.Name)
	if err != nil {
		l.Logger.Errorf("ListKnowledgeBase err: %v", err)
		return nil, errorx.NewInternalError("查询知识库列表失败")
	}

	return &types.KnowledgeBaseListResponse{
		Code:      0,
		Msg:       "OK",
		Timestamp: time.Now().Unix(),
		Data: types.KnowledgeBaseListData{
			Total: total,
			List:  knowledge.ToShowKBList(list),
		},
	}, nil
}
