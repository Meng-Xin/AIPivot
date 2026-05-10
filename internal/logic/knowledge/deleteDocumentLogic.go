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

type DeleteDocumentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除文档
func NewDeleteDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDocumentLogic {
	return &DeleteDocumentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteDocumentLogic) DeleteDocument(req *types.DeleteDocumentRequest) (resp *types.CommResponse, err error) {
	doc, err := l.svcCtx.DocumentRepo.GetByID(l.ctx, req.ID)
	if err != nil {
		l.Logger.Errorf("DeleteDocument GetByID err: %v", err)
		return nil, errorx.NewInternalError("删除文档失败")
	}
	if doc == nil {
		return nil, errorx.NewNotFoundError("文档不存在")
	}

	// 级联删除 chunks 由数据库外键 ON DELETE CASCADE 保证
	if err = l.svcCtx.DocumentRepo.Delete(l.ctx, req.ID); err != nil {
		l.Logger.Errorf("DeleteDocument err: %v", err)
		return nil, errorx.NewInternalError("删除文档失败")
	}

	return &types.CommResponse{
		Code:      0,
		Msg:       "删除成功",
		Timestamp: time.Now().Unix(),
	}, nil
}
