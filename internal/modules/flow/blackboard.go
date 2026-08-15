package flow

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var templatePattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_.]+)\s*\}\}`)

// Blackboard 流程黑板变量：节点间传递数据的共享上下文。
// 一期变量集合：message / confidence（默认 1.0，RAG 暂未暴露检索分）/ lastOutput /
// node.<id>（各节点输出）/ variables.<k>（试运行自定义变量）。
type Blackboard struct {
	message     string
	confidence  float64
	variables   map[string]string
	lastOutput  string
	nodeOutputs map[string]string
}

// NewBlackboard 用试运行输入初始化黑板。
func NewBlackboard(input RunInput) *Blackboard {
	vars := input.Variables
	if vars == nil {
		vars = map[string]string{}
	}
	return &Blackboard{
		message:     input.Message,
		confidence:  1.0, // 一期无真实检索分来源，默认放行（文档注明，二期由 AnswerResult 暴露 top score）
		variables:   vars,
		nodeOutputs: map[string]string{},
	}
}

// SetMessage trigger 节点写入测试消息。
func (b *Blackboard) SetMessage(msg string) {
	b.message = msg
}

// Lookup 按变量名取值，用于表达式求值。ok=false 表示变量不存在。
func (b *Blackboard) Lookup(key string) (any, bool) {
	switch {
	case key == "message":
		return b.message, true
	case key == "lastOutput":
		return b.lastOutput, true
	case key == "confidence":
		return b.confidence, true
	case strings.HasPrefix(key, "variables."):
		v, ok := b.variables[strings.TrimPrefix(key, "variables.")]
		return v, ok
	case strings.HasPrefix(key, "node."):
		v, ok := b.nodeOutputs[strings.TrimPrefix(key, "node.")]
		return v, ok
	default:
		return nil, false
	}
}

// SetNodeOutput 记录节点输出，并同步推进 lastOutput。
func (b *Blackboard) SetNodeOutput(nodeID, output string) {
	b.nodeOutputs[nodeID] = output
	b.lastOutput = output
}

// LastOutput 返回最近一次节点输出。
func (b *Blackboard) LastOutput() string {
	return b.lastOutput
}

// Message 返回当前消息。
func (b *Blackboard) Message() string {
	return b.message
}

// RenderTemplate 渲染 {{var}} 占位符模板（支持 message / lastOutput / node.xxx / variables.xxx）。
// 未知占位符原样保留（fail-soft，让用户在输出里直接看到哪个变量没解析出来）。
func (b *Blackboard) RenderTemplate(tmpl string) string {
	return templatePattern.ReplaceAllStringFunc(tmpl, func(match string) string {
		key := templatePattern.FindStringSubmatch(match)[1]
		v, ok := b.Lookup(key)
		if !ok {
			return match
		}
		switch val := v.(type) {
		case string:
			return val
		case float64:
			return strconv.FormatFloat(val, 'f', -1, 64)
		default:
			return fmt.Sprint(val)
		}
	})
}
