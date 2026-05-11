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

type ListDocumentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取知识库文档列表
func NewListDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDocumentLogic {
	return &ListDocumentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListDocumentLogic) ListDocument(req *types.ListDocumentRequest) (resp *types.DocumentListResponse, err error) {
	list, total, err := l.svcCtx.DocumentRepo.GetList(l.ctx, req.KnowledgeBaseID, req.Page, req.PageSize, req.Status)
	if err != nil {
		l.Logger.Errorf("ListDocument err: %v", err)
		return nil, errorx.NewInternalError("查询文档列表失败")
	}

	return &types.DocumentListResponse{
		Code:      0,
		Msg:       "OK",
		Timestamp: time.Now().Unix(),
		Data: types.DocumentListData{
			Total: total,
			List:  knowledge.ToShowDocumentList(list),
		},
	}, nil
}
