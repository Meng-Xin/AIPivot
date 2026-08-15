package flow

import (
	"context"
	"fmt"
	"time"

	"aipivot/internal/modules/agent"
	"aipivot/internal/modules/rag"
	"aipivot/internal/shared/sse"
	"aipivot/pkg/llm"
)

// Options 引擎运行参数。
type Options struct {
	MaxSteps int // 单次执行最大步数（防环/防失控，默认 64）
}

// RunInput 单次执行输入。
type RunInput struct {
	TenantID   int64
	Message    string
	Variables  map[string]string
}

// NodeResult 单节点执行结果快照（落库 node_results 数组元素）。
type NodeResult struct {
	NodeID     string         `json:"nodeId"`
	NodeType   string         `json:"nodeType"`
	Label      string         `json:"label"`
	Status     string         `json:"status"` // success / skipped / failed
	DurationMs int64          `json:"durationMs"`
	Summary    map[string]any `json:"summary,omitempty"`
	Warning    string         `json:"warning,omitempty"`
}

// RunResult 整次执行结果聚合。
type RunResult struct {
	Status      string       // success / failed / timeout
	Output      string
	Error       string
	NodeResults []NodeResult
	TotalMs     int64
	TokenCount  int
	Warnings    []string
}

// EventEmitter 引擎事件回调。事件名与 sse 协议对齐（node_start / delta / node_end），
// 未来接入 Asynq / 事件触发时换成其他 sink 即可，引擎不感知 HTTP。
type EventEmitter func(event string, data any)

// Engine Flow 执行引擎：从 trigger 出发单步顺序遍历运行时图。
type Engine struct {
	rag           *rag.Service
	llm           *llm.Client
	skillResolver func(ctx context.Context, tenantID int64) []agent.Tool
	opts          Options
}

func NewEngine(ragService *rag.Service, llmClient *llm.Client, skillResolver func(ctx context.Context, tenantID int64) []agent.Tool, opts Options) *Engine {
	if opts.MaxSteps <= 0 {
		opts.MaxSteps = 64
	}
	return &Engine{rag: ragService, llm: llmClient, skillResolver: skillResolver, opts: opts}
}

// Execute 执行流程。emit 可为 nil（无监听场景）。
// 函数总是返回完整的 RunResult（不会 panic / panic 中断），由调用方负责落库收尾。
func (e *Engine) Execute(ctx context.Context, g *RuntimeGraph, input RunInput, emit EventEmitter) *RunResult {
	if emit == nil {
		emit = func(string, any) {}
	}
	start := time.Now()
	bb := NewBlackboard(input)
	res := &RunResult{Status: "success", NodeResults: make([]NodeResult, 0, len(g.Nodes))}

	visits := make(map[string]int, len(g.Nodes))
	current := g.TriggerID

	for current != "" {
		if err := ctx.Err(); err != nil {
			res.Status = "timeout"
			res.Error = "执行超时被中止"
			return res
		}

		if len(res.NodeResults) >= e.opts.MaxSteps {
			res.Status = "failed"
			res.Error = fmt.Sprintf("超出最大执行步数 %d", e.opts.MaxSteps)
			return res
		}

		node := g.Nodes[current]
		visits[node.ID]++
		if visits[node.ID] > 3 {
			res.Status = "failed"
			res.Error = fmt.Sprintf("节点 %s 被重复执行超过 3 次，疑似死循环", node.Label)
			return res
		}

		emit("node_start", sse.NodeStart{NodeID: node.ID, NodeType: node.Type, Label: node.Label})
		nodeStart := time.Now()

		out := e.executeNode(ctx, node, bb, input, emit)

		nr := NodeResult{
			NodeID:     node.ID,
			NodeType:   node.Type,
			Label:      node.Label,
			Status:     out.Status,
			DurationMs: time.Since(nodeStart).Milliseconds(),
			Summary:    out.Summary,
			Warning:    out.Warning,
		}
		res.NodeResults = append(res.NodeResults, nr)
		res.TokenCount += out.Tokens
		if out.Warning != "" {
			res.Warnings = append(res.Warnings, fmt.Sprintf("节点[%s]: %s", node.Label, out.Warning))
		}

		emit("node_end", sse.NodeEnd{
			NodeID:     node.ID,
			Status:     nr.Status,
			DurationMs: nr.DurationMs,
			Summary:    nr.Summary,
		})

		if out.Status == nodeStatusFailed {
			res.Status = "failed"
			res.Error = out.Error
			return res
		}

		if node.Type == nodeTypeEnd {
			res.Output = out.Output
			res.TotalMs = time.Since(start).Milliseconds()
			return res
		}

		// 仅 condition 节点的分支结果参与选边；其余节点固定走第一条出边
		decision := out.Decision
		if node.Type != nodeTypeCondition {
			decision = true
		}
		next, ok := g.SelectOutEdge(node.ID, decision)
		if !ok {
			// 非 end 节点无出边：视为自然结束（success + warning），不判定为失败
			res.Warnings = append(res.Warnings, fmt.Sprintf("节点[%s]无出边，流程自然结束", node.Label))
			res.Output = bb.LastOutput()
			break
		}
		current = next.Target
	}

	res.TotalMs = time.Since(start).Milliseconds()
	return res
}

// nodeOutput 单个执行器的统一返回结构。
type nodeOutput struct {
	Status   string // success / skipped / failed
	Output   string
	Decision bool         // condition 节点分支结果
	Summary  map[string]any
	Warning  string
	Error    string // Status==failed 时的失败原因
	Tokens   int
}

const (
	nodeTypeTrigger   = "trigger"
	nodeTypeLLM       = "llm"
	nodeTypeSkill     = "skill"
	nodeTypeCondition = "condition"
	nodeTypeEnd       = "end"

	nodeStatusSuccess = "success"
	nodeStatusSkipped = "skipped"
	nodeStatusFailed  = "failed"
)

func (e *Engine) executeNode(ctx context.Context, node *RuntimeNode, bb *Blackboard, input RunInput, emit EventEmitter) *nodeOutput {
	switch node.Type {
	case nodeTypeTrigger:
		return execTrigger(node, bb, input)
	case nodeTypeLLM:
		return e.execLLM(ctx, node, bb, emit)
	case nodeTypeSkill:
		return e.execSkill(ctx, node, bb, input)
	case nodeTypeCondition:
		return execCondition(node, bb)
	case nodeTypeEnd:
		return execEnd(node, bb)
	default:
		// 未知节点类型 fail-soft：跳过并继续走第一条出边
		return &nodeOutput{
			Status:  nodeStatusSkipped,
			Warning: fmt.Sprintf("未知节点类型 %s，已跳过", node.Type),
		}
	}
}
