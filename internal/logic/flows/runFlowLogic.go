package flows

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"aipivot/internal/modules/auth"
	flowmod "aipivot/internal/modules/flow"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/shared/po"
	"aipivot/internal/shared/sse"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type RunFlowLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 试运行 Flow（SSE 流式返回每个节点执行过程）
func NewRunFlowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RunFlowLogic {
	return &RunFlowLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// RunFlow 试运行：SSE Writer 先行 → 前置校验 → 建 running 记录 → 引擎执行 → 收尾落库。
// 直接操作 http.ResponseWriter，不走标准 JSON 响应模式。
func (l *RunFlowLogic) RunFlow(w http.ResponseWriter, req *types.RunFlowRequest) {
	tenantID := auth.TenantIDFromContext(l.ctx)

	// SSE Writer 必须在任何错误响应之前初始化（需要特定 header）
	sseWriter, err := sse.NewWriter(w)
	if err != nil {
		l.Logger.Errorf("RunFlow SSE init err: %v", err)
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	if len([]rune(req.Message)) > 2000 {
		sseWriter.WriteError(errorx.CodeBadRequest, "message 长度不能超过 2000 字符")
		return
	}

	fl, err := l.svcCtx.FlowRepo.GetByID(l.ctx, req.ID, tenantID)
	if err != nil {
		l.Logger.Errorf("RunFlow GetFlow err: %v", err)
		sseWriter.WriteError(errorx.CodeFailed, "试运行失败")
		return
	}
	if fl == nil {
		sseWriter.WriteError(errorx.CodeNotFound, "流程不存在")
		return
	}
	// archived 视为不可运行（draft/published 可运行，试运行即调试语义）
	if fl.Status == flowStatusArchived {
		sseWriter.WriteError(errorx.CodeBadRequest, "流程已归档，无法试运行")
		return
	}

	graph, err := flowmod.ParseDefinition(fl.Definition)
	if err != nil {
		l.Logger.Errorf("RunFlow ParseDefinition err: %v", err)
		sseWriter.WriteError(errorx.CodeBadRequest, "流程定义无效: "+err.Error())
		return
	}

	// LLM 网关健康检查（2s）：不可用时直接报错，不产生垃圾 run 记录（对齐 sendMessageStream 的 guard 哲学）
	healthCtx, cancel := context.WithTimeout(l.ctx, 2*time.Second)
	healthErr := l.svcCtx.LLMClient.HealthCheck(healthCtx)
	cancel()
	if healthErr != nil {
		l.Logger.Errorf("RunFlow LLM healthcheck err: %v", healthErr)
		sseWriter.WriteError(errorx.CodeLLMUnavailable, "LLM 网关不可用，无法试运行")
		return
	}

	// 日 Token 配额检查（limit=0 时跳过）
	if rateLimitErr := l.svcCtx.TokenLimiter.Check(l.ctx, tenantID); rateLimitErr != nil {
		sseWriter.WriteError(errorx.CodeTokenExceeded, rateLimitErr.Error())
		return
	}

	// 创建 running 记录（显式设置 UUID，不依赖数据库默认值）
	run := &po.FlowRun{
		UUID:        uuid.New().String(),
		TenantID:    tenantID,
		FlowID:      fl.ID,
		FlowVersion: fl.Version,
		Status:      "running",
		TriggerType: "manual",
		Input:       po.JSONMap{"message": req.Message, "variables": req.Variables},
	}
	if err = l.svcCtx.FlowRunRepo.Create(l.ctx, run); err != nil {
		l.Logger.Errorf("RunFlow CreateRun err: %v", err)
		sseWriter.WriteError(errorx.CodeFailed, "试运行失败")
		return
	}

	_ = sseWriter.WriteEvent("run_start", sse.RunStart{
		RunID:       run.ID,
		FlowID:      fl.ID,
		FlowVersion: fl.Version,
	})

	// 引擎执行（带超时，默认 60s，可配置）
	runTimeout := time.Duration(l.svcCtx.Config.Flow.RunTimeoutSec) * time.Second
	if runTimeout <= 0 {
		runTimeout = 60 * time.Second
	}
	execCtx, execCancel := context.WithTimeout(l.ctx, runTimeout)
	res := l.svcCtx.FlowEngine.Execute(execCtx, graph, flowmod.RunInput{
		TenantID:  tenantID,
		Message:   req.Message,
		Variables: req.Variables,
	}, func(event string, data any) {
		_ = sseWriter.WriteEvent(event, data)
	})
	execCancel()

	nodeResultsJSON, _ := json.Marshal(res.NodeResults)

	// 收尾落库（RowsAffected==0 理论上只在 run 被并发删除时出现，降级为日志）
	if finishErr := l.svcCtx.FlowRunRepo.Finish(l.ctx, run.ID, tenantID, map[string]any{
		"status":       res.Status,
		"output":       res.Output,
		"node_results": string(nodeResultsJSON),
		"error":        res.Error,
		"total_ms":     res.TotalMs,
		"token_count":  res.TokenCount,
	}); finishErr != nil {
		l.Logger.Errorf("RunFlow FinishRun err: %v", finishErr)
	}

	_ = sseWriter.WriteEvent("run_end", sse.RunEnd{
		RunID:      run.ID,
		Status:     res.Status,
		TotalMs:    res.TotalMs,
		TokenCount: res.TokenCount,
		Output:     res.Output,
	})
	_ = sseWriter.WriteDone()

	// 记录当日已用 token（fire-and-forget）
	if res.TokenCount > 0 {
		l.svcCtx.TokenLimiter.Incr(l.ctx, tenantID, res.TokenCount)
	}
}
