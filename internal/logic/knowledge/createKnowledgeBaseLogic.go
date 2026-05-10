// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package knowledge

import (
	"context"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/modules/knowledge/domain/assembler"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateKnowledgeBaseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建知识库
func NewCreateKnowledgeBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateKnowledgeBaseLogic {
	return &CreateKnowledgeBaseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateKnowledgeBaseLogic) CreateKnowledgeBase(req *types.CreateKnowledgeBaseRequest) (resp *types.KnowledgeBaseDetailResponse, err error) {
	tenantID := auth.TenantIDFromContext(l.ctx)

	// 1. Request → Domain Model → 校验
	kbModel := assembler.CreateKBRequestToModel(req)
	if err = kbModel.CheckName(); err != nil {
		return nil, errorx.NewInternalError(err.Error())
	}

	// 2. Domain Model → PO
	kbPo := assembler.ModelKBToKnowledgeBasePo(kbModel, tenantID)

	// 3. 持久化
	if err = l.svcCtx.KnowledgeBaseRepo.Create(l.ctx, kbPo); err != nil {
		l.Logger.Errorf("CreateKnowledgeBase err: %v", err)
		return nil, errorx.NewInternalError("创建知识库失败")
	}

	// 4. PO → Show
	show := assembler.KnowledgeBasePoToShow(kbPo)
	return &types.KnowledgeBaseDetailResponse{
		Code:      0,
		Msg:       "创建成功",
		Timestamp: time.Now().Unix(),
		Data:      show,
	}, nil
}
