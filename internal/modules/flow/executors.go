package flow

import (
	"context"
	"fmt"
	"strings"

	"aipivot/internal/modules/agent"
	"aipivot/internal/shared/sse"
	"aipivot/pkg/llm"
)

// execTrigger trigger 节点：把试运行消息写入黑板，纯透传。
func execTrigger(node *RuntimeNode, bb *Blackboard, input RunInput) *nodeOutput {
	bb.SetMessage(input.Message)
	bb.SetNodeOutput(node.ID, input.Message)
	return &nodeOutput{
		Status:  nodeStatusSuccess,
		Output:  input.Message,
		Summary: map[string]any{"event": stringOf(node.Config["event"])},
	}
}

// execLLM LLM 节点：
//   - mode=rag（缺省）：走 RAGService.AnswerStream，kbId 取 config.knowledgeBaseId，缺省 0 纯 LLM
//   - mode=direct：直接 llmClient.ChatCompletionStream
//
// prompt 经黑板模板渲染（缺省 {{message}}），流式 token 逐个发射 delta 事件。
func (e *Engine) execLLM(ctx context.Context, node *RuntimeNode, bb *Blackboard, emit EventEmitter) *nodeOutput {
	promptTmpl := stringOf(node.Config["prompt"])
	if promptTmpl == "" {
		promptTmpl = "{{message}}"
	}
	prompt := bb.RenderTemplate(promptTmpl)

	mode := stringOf(node.Config["mode"])
	var stream <-chan llm.StreamEvent
	var err error

	if mode == "direct" {
		stream, err = e.llm.ChatCompletionStream(ctx, &llm.ChatRequest{
			Messages: []llm.ChatMessage{{Role: "user", Content: prompt}},
		})
	} else {
		// rag 缺省模式；kbId<=0 时 AnswerStream 内部跳过检索
		var kbID int64
		if f, ok := floatOf(node.Config["knowledgeBaseId"]); ok && f > 0 {
			kbID = int64(f)
		}
		stream, _, err = e.rag.AnswerStream(ctx, kbID, prompt, nil, stringOf(node.Config["model"]), nil)
	}
	if err != nil {
		return &nodeOutput{Status: nodeStatusFailed, Error: fmt.Sprintf("LLM 调用失败: %v", err)}
	}

	var content strings.Builder
	var model string
	var usage *llm.ChatUsage
	for evt := range stream {
		if evt.Err != nil {
			return &nodeOutput{Status: nodeStatusFailed, Error: fmt.Sprintf("LLM 流中断: %v", evt.Err)}
		}
		if evt.Content != "" {
			content.WriteString(evt.Content)
			emit("delta", sse.FlowDelta{NodeID: node.ID, Content: evt.Content})
		}
		if evt.Model != "" {
			model = evt.Model
		}
		if evt.Usage != nil {
			usage = evt.Usage
		}
		if evt.Done {
			break
		}
	}

	tokens := 0
	if usage != nil {
		tokens = usage.TotalTokens
	}
	output := content.String()
	bb.SetNodeOutput(node.ID, output)
	return &nodeOutput{
		Status:  nodeStatusSuccess,
		Output:  output,
		Summary: map[string]any{"model": model, "mode": mode, "tokens": tokens},
		Tokens:  tokens,
	}
}

// execSkill skill 节点：按 config.skillName 在租户工具列表中匹配并执行。
//   - skillName 找不到 → skipped + warning 继续走（软失败，工具下线不应阻断流程）
//   - Execute 报错 → 节点 failed（run 失败）
func (e *Engine) execSkill(ctx context.Context, node *RuntimeNode, bb *Blackboard, input RunInput) *nodeOutput {
	skillName := stringOf(node.Config["skillName"])
	if skillName == "" {
		return &nodeOutput{
			Status:  nodeStatusSkipped,
			Warning: "未配置 skillName，节点已跳过",
		}
	}
	if e.skillResolver == nil {
		return &nodeOutput{
			Status:  nodeStatusSkipped,
			Warning: fmt.Sprintf("Skill 服务不可用，节点 %s 已跳过", skillName),
		}
	}

	tools := e.skillResolver(ctx, input.TenantID)
	var tool agent.Tool
	for _, t := range tools {
		if t.Name() == skillName {
			tool = t
			break
		}
	}
	if tool == nil {
		return &nodeOutput{
			Status:  nodeStatusSkipped,
			Warning: fmt.Sprintf("Skill %q 不存在或未启用，节点已跳过", skillName),
		}
	}

	argsFrom := stringOf(node.Config["argumentsFrom"])
	if argsFrom == "" {
		argsFrom = "message"
	}
	var args string
	if v, ok := bb.Lookup(argsFrom); ok {
		if s, isStr := v.(string); isStr {
			args = s
		} else {
			args = fmt.Sprint(v)
		}
	} else {
		// 指定来源不存在时兜底用 message，避免空参数打挂下游
		args = bb.Message()
	}

	result, err := tool.Execute(ctx, args)
	if err != nil {
		return &nodeOutput{Status: nodeStatusFailed, Error: fmt.Sprintf("Skill %q 执行失败: %v", skillName, err)}
	}

	bb.SetNodeOutput(node.ID, result)
	return &nodeOutput{
		Status:  nodeStatusSuccess,
		Output:  result,
		Summary: map[string]any{"skill": skillName},
	}
}

// execCondition condition 节点：求值 config.expression（fail-soft，失败视为 true + warning），
// Decision 交给引擎选分支边，结果与选中 port 记入 Summary。
func execCondition(node *RuntimeNode, bb *Blackboard) *nodeOutput {
	expr := stringOf(node.Config["expression"])
	if expr == "" {
		// 空表达式默认走 true 分支，不阻断
		return &nodeOutput{
			Status:   nodeStatusSuccess,
			Decision: true,
			Warning:  "condition 未配置表达式，默认走 true 分支",
			Summary:  map[string]any{"expression": "", "branch": "true"},
		}
	}

	decision, err := EvalExpression(expr, bb)
	if err != nil {
		return &nodeOutput{
			Status:   nodeStatusSuccess,
			Decision: true,
			Warning:  fmt.Sprintf("表达式 %q 求值失败（%v），默认走 true 分支", expr, err),
			Summary:  map[string]any{"expression": expr, "branch": "true"},
		}
	}

	branch := "false"
	if decision {
		branch = "true"
	}
	return &nodeOutput{
		Status:   nodeStatusSuccess,
		Decision: decision,
		Summary:  map[string]any{"expression": expr, "branch": branch},
	}
}

// execEnd end 节点：输出最近节点结果；黑板为空时输出 config.reason。
func execEnd(node *RuntimeNode, bb *Blackboard) *nodeOutput {
	output := bb.LastOutput()
	if output == "" {
		output = stringOf(node.Config["reason"])
	}
	return &nodeOutput{
		Status:  nodeStatusSuccess,
		Output:  output,
		Summary: map[string]any{},
	}
}
