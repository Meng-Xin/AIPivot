// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package knowledge

import (
	"context"
	"net/http"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/modules/knowledge/domain/assembler"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/shared/po"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadDocumentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 上传文档到知识库
func NewUploadDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadDocumentLogic {
	return &UploadDocumentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadDocumentLogic) UploadDocument(req *types.UploadDocumentRequest, r *http.Request) (resp *types.DocumentDetailResponse, err error) {
	tenantID := auth.TenantIDFromContext(l.ctx)

	// 1. 校验知识库存在
	kb, err := l.svcCtx.KnowledgeBaseRepo.GetByID(l.ctx, req.KnowledgeBaseID)
	if err != nil {
		l.Logger.Errorf("UploadDocument GetKB err: %v", err)
		return nil, errorx.NewInternalError("上传文档失败")
	}
	if kb == nil {
		return nil, errorx.NewNotFoundError("知识库不存在")
	}

	// 2. 解析上传文件
	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, errorx.NewInternalError("请上传文件")
	}
	defer file.Close()

	// 3. 创建文档记录（status=pending，后续异步处理切块+Embedding）
	doc := &po.Document{
		KnowledgeBaseID: req.KnowledgeBaseID,
		TenantID:        tenantID,
		Name:            header.Filename,
		ContentType:     header.Header.Get("Content-Type"),
		FileSize:        header.Size,
		Status:          "pending",
	}

	if err = l.svcCtx.DocumentRepo.Create(l.ctx, doc); err != nil {
		l.Logger.Errorf("UploadDocument Create err: %v", err)
		return nil, errorx.NewInternalError("上传文档失败")
	}

	// TODO: 触发异步任务 → 解析文档 → 切块 → Embedding → 存储 chunks

	show := assembler.DocumentPoToShow(doc)
	return &types.DocumentDetailResponse{
		Code:      0,
		Msg:       "上传成功，正在处理中",
		Timestamp: time.Now().Unix(),
		Data:      show,
	}, nil
}
