// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package knowledge

import (
	"context"
	"io"
	"net/http"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/modules/knowledge/domain/assembler"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/shared/po"
	"aipivot/internal/svc"
	"aipivot/internal/types"
	"aipivot/internal/worker"

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

	// 3. 读取文件文本内容（MVP 阶段仅支持纯文本/Markdown，暂存于 file_path 字段）
	content, err := io.ReadAll(file)
	if err != nil {
		l.Logger.Errorf("UploadDocument ReadFile err: %v", err)
		return nil, errorx.NewInternalError("读取文件失败")
	}

	// 4. 创建文档记录（status=pending，后续异步处理切块+Embedding）
	doc := &po.Document{
		KnowledgeBaseID: req.KnowledgeBaseID,
		TenantID:        tenantID,
		Name:            header.Filename,
		ContentType:     header.Header.Get("Content-Type"),
		FileSize:        header.Size,
		FilePath:        string(content), // MVP: 文本内容暂存于此字段，后续迁移到 MinIO/OSS
		Status:          "pending",
	}

	if err = l.svcCtx.DocumentRepo.Create(l.ctx, doc); err != nil {
		l.Logger.Errorf("UploadDocument Create err: %v", err)
		return nil, errorx.NewInternalError("上传文档失败")
	}

	// 5. 提交异步处理任务（切块 → Embedding → pgvector 存储）
	if err = worker.EnqueueDocumentProcess(l.svcCtx.AsynqClient, worker.DocumentProcessPayload{
		DocumentID:      doc.ID,
		KnowledgeBaseID: req.KnowledgeBaseID,
		TenantID:        tenantID,
		EmbeddingModel:  l.svcCtx.Config.LLM.EmbeddingModel,
		EmbeddingDim:    l.svcCtx.Config.LLM.EmbeddingDim,
	}); err != nil {
		// 任务入队失败不影响文档记录，后续可重试
		l.Logger.Errorf("UploadDocument EnqueueTask err: %v", err)
	}

	show := assembler.DocumentPoToShow(doc)
	return &types.DocumentDetailResponse{
		Code:      0,
		Msg:       "上传成功，正在处理中",
		Timestamp: time.Now().Unix(),
		Data:      show,
	}, nil
}
