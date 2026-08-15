package flows

import (
	"context"
	"encoding/json"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/shared/po"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFlowRunsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取 Flow 执行历史
func NewListFlowRunsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFlowRunsLogic {
	return &ListFlowRunsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListFlowRunsLogic) ListFlowRuns(req *types.ListFlowRunsRequest) (resp *types.FlowRunListResponse, err error) {
	tenantID := auth.TenantIDFromContext(l.ctx)
	if tenantID == 0 {
		return nil, errorx.NewUnauthError("unauthenticated")
	}

	// 先确认 Flow 归属当前租户
	fl, err := l.svcCtx.FlowRepo.GetByID(l.ctx, req.ID, tenantID)
	if err != nil {
		l.Logger.Errorf("ListFlowRuns GetFlow err: %v", err)
		return nil, errorx.NewInternalError("获取执行历史失败")
	}
	if fl == nil {
		return nil, errorx.NewNotFoundError("流程不存在")
	}

	list, err := l.svcCtx.FlowRunRepo.ListByFlow(l.ctx, req.ID, tenantID, req.Limit, req.Offset)
	if err != nil {
		l.Logger.Errorf("ListFlowRuns repo err: %v", err)
		return nil, errorx.NewInternalError("获取执行历史失败")
	}

	data := make([]types.ShowFlowRun, 0, len(list))
	for _, run := range list {
		data = append(data, toShowFlowRun(run))
	}

	return &types.FlowRunListResponse{
		Code:      0,
		Msg:       "OK",
		Timestamp: time.Now().Unix(),
		Data:      data,
	}, nil
}

// toShowFlowRun PO → 展示对象。JSON 字段序列化失败兜底 "{}"/"[]"，不让单条脏数据打断列表。
func toShowFlowRun(run *po.FlowRun) types.ShowFlowRun {
	inputJSON, err := json.Marshal(run.Input)
	if err != nil || string(inputJSON) == "null" {
		inputJSON = []byte("{}")
	}
	nodeResults := run.NodeResults
	if nodeResults == "" {
		nodeResults = "[]"
	}
	return types.ShowFlowRun{
		ID:          run.ID,
		UUID:        run.UUID,
		FlowID:      run.FlowID,
		FlowVersion: run.FlowVersion,
		Status:      run.Status,
		TriggerType: run.TriggerType,
		Input:       string(inputJSON),
		Output:      run.Output,
		NodeResults: nodeResults,
		Error:       run.Error,
		TotalMs:     run.TotalMs,
		TokenCount:  run.TokenCount,
		CreatedAt:   run.CreatedAt.Unix(),
	}
}
