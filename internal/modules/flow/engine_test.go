package flow

import (
	"context"
	"testing"

	"aipivot/internal/shared/po"
)

func newBB() *Blackboard {
	return NewBlackboard(RunInput{
		Message:    "你好",
		Variables:  map[string]string{"status": "ok"},
	})
}

func TestEvalExpression(t *testing.T) {
	bb := newBB()

	cases := []struct {
		expr string
		want bool
		err  bool // 预期返回 error（调用方 fail-soft）
	}{
		// 数值语境
		{"confidence < 0.6", false, false},
		{"confidence <= 1.0", true, false},
		{"confidence > 0.6", true, false},
		{"confidence >= 1.0", true, false},
		{"0.5 < 0.6", true, false},
		// 字符串语境（变量 + 字面量）
		{`variables.status == "ok"`, true, false},
		{`variables.status != "ok"`, false, false},
		{"variables.status == 'ok'", true, false},
		{`message != ""`, true, false},
		// 类型不可比较（字符串变量 vs 数值字面量，且字符串不可数值化）
		{"variables.status < 5", false, true},
		// 裸标识符 truthy
		{"message", true, false},
		{"variables.status", true, false},
		{"variables.missing", false, true},
		// 缺失变量 / 畸形
		{"missing < 1", false, true},
		{"", false, true},
		{"== 1", false, true},
		{"a ==", false, true},
		{"confidence >", false, true},
	}
	for _, c := range cases {
		got, err := EvalExpression(c.expr, bb)
		if c.err {
			if err == nil {
				t.Errorf("EvalExpression(%q) 期望报错，实际得到 %v", c.expr, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("EvalExpression(%q) 意外报错: %v", c.expr, err)
			continue
		}
		if got != c.want {
			t.Errorf("EvalExpression(%q) = %v, want %v", c.expr, got, c.want)
		}
	}
}

func TestEvalExpressionStringOrdering(t *testing.T) {
	bb := NewBlackboard(RunInput{Variables: map[string]string{"name": "alice"}})
	// 两侧不可数值化，走字符串字典序
	if got, _ := EvalExpression(`variables.name < "bob"`, bb); !got {
		t.Errorf(`variables.name < "bob" 应为 true`)
	}
	// 字符串两侧数值化比较（"10" > "9" 数值语境为 true）
	bb2 := NewBlackboard(RunInput{Variables: map[string]string{"n": "10"}})
	if got, _ := EvalExpression(`variables.n > 9`, bb2); !got {
		t.Errorf(`variables.n("10") > 9 数值语境应为 true`)
	}
}

func TestBlackboardTemplate(t *testing.T) {
	bb := NewBlackboard(RunInput{Message: "问题", Variables: map[string]string{"env": "prod"}})
	bb.SetNodeOutput("n1", "回答")

	cases := map[string]string{
		"{{message}}":              "问题",
		"前置 {{lastOutput}} 后置":     "前置 回答 后置",
		"{{node.n1}}":              "回答",
		"{{variables.env}}":        "prod",
		"{{ unknown }} 保留":         "{{ unknown }} 保留",
		"无占位符":                     "无占位符",
		"{{ message }} 带空格":        "问题 带空格",
	}
	for tmpl, want := range cases {
		if got := bb.RenderTemplate(tmpl); got != want {
			t.Errorf("RenderTemplate(%q) = %q, want %q", tmpl, got, want)
		}
	}
}

func TestParseDefinitionCycle(t *testing.T) {
	// a -> b -> a 环，应报错
	def := po.JSONMap{
		"nodes": []any{
			map[string]any{"id": "a", "type": "trigger"},
			map[string]any{"id": "b", "type": "llm"},
		},
		"edges": []any{
			map[string]any{"id": "e1", "source": "a", "target": "b"},
			map[string]any{"id": "e2", "source": "b", "target": "a"},
		},
	}
	if _, err := ParseDefinition(def); err == nil {
		t.Errorf("带环 definition 应解析失败")
	}

	// 菱形汇合（trigger -> a/b -> end）合法
	diamond := po.JSONMap{
		"nodes": []any{
			map[string]any{"id": "t", "type": "trigger"},
			map[string]any{"id": "a", "type": "llm"},
			map[string]any{"id": "b", "type": "llm"},
			map[string]any{"id": "e", "type": "end"},
		},
		"edges": []any{
			map[string]any{"id": "e1", "source": "t", "target": "a"},
			map[string]any{"id": "e2", "source": "t", "target": "b"},
			map[string]any{"id": "e3", "source": "a", "target": "e"},
			map[string]any{"id": "e4", "source": "b", "target": "e"},
		},
	}
	if _, err := ParseDefinition(diamond); err != nil {
		t.Errorf("菱形汇合应合法: %v", err)
	}

	// 结构错误：边引用不存在节点
	badEdge := po.JSONMap{
		"nodes": []any{map[string]any{"id": "t", "type": "trigger"}},
		"edges": []any{
			map[string]any{"id": "e1", "source": "t", "target": "ghost"},
		},
	}
	if _, err := ParseDefinition(badEdge); err == nil {
		t.Errorf("孤儿边应解析失败")
	}

	// 无 trigger
	noTrigger := po.JSONMap{
		"nodes": []any{map[string]any{"id": "a", "type": "llm"}},
	}
	if _, err := ParseDefinition(noTrigger); err == nil {
		t.Errorf("无 trigger 应解析失败")
	}
}

func TestSelectOutEdge(t *testing.T) {
	def := po.JSONMap{
		"nodes": []any{
			map[string]any{"id": "trg", "type": "trigger"},
			map[string]any{"id": "c", "type": "condition"},
			map[string]any{"id": "t", "type": "end"},
			map[string]any{"id": "f", "type": "end"},
		},
		"edges": []any{
			map[string]any{"id": "e1", "source": "c", "target": "t", "sourcePort": "true"},
			map[string]any{"id": "e2", "source": "c", "target": "f", "sourcePort": "false"},
		},
	}
	g, err := ParseDefinition(def)
	if err != nil {
		t.Fatalf("ParseDefinition err: %v", err)
	}
	if e, _ := g.SelectOutEdge("c", true); e.Target != "t" {
		t.Errorf("true 分支应选 t, got %s", e.Target)
	}
	if e, _ := g.SelectOutEdge("c", false); e.Target != "f" {
		t.Errorf("false 分支应选 f, got %s", e.Target)
	}

	// 无 port 兜底：true→第一条，false→第二条
	def2 := po.JSONMap{
		"nodes": []any{
			map[string]any{"id": "trg", "type": "trigger"},
			map[string]any{"id": "c", "type": "condition"},
			map[string]any{"id": "a", "type": "end"},
			map[string]any{"id": "b", "type": "end"},
		},
		"edges": []any{
			map[string]any{"id": "e1", "source": "c", "target": "a"},
			map[string]any{"id": "e2", "source": "c", "target": "b"},
		},
	}
	g2, _ := ParseDefinition(def2)
	if e, _ := g2.SelectOutEdge("c", true); e.Target != "a" {
		t.Errorf("无 port 时 true 应兜底第一条, got %s", e.Target)
	}
	if e, _ := g2.SelectOutEdge("c", false); e.Target != "b" {
		t.Errorf("无 port 时 false 应兜底第二条, got %s", e.Target)
	}
}

// trigger -> condition -> end 全链路（不依赖 LLM/Skill 的执行器）
func TestEngineTriggerConditionEnd(t *testing.T) {
	def := po.JSONMap{
		"nodes": []any{
			map[string]any{"id": "t", "type": "trigger", "label": "开始"},
			map[string]any{"id": "c", "type": "condition", "label": "判断", "config": map[string]any{"expression": "confidence >= 0.6"}},
			map[string]any{"id": "yes", "type": "end", "label": "高置信", "config": map[string]any{"reason": "高置信结束"}},
			map[string]any{"id": "no", "type": "end", "label": "低置信", "config": map[string]any{"reason": "低置信结束"}},
		},
		"edges": []any{
			map[string]any{"id": "e1", "source": "t", "target": "c"},
			map[string]any{"id": "e2", "source": "c", "target": "yes", "sourcePort": "true"},
			map[string]any{"id": "e3", "source": "c", "target": "no", "sourcePort": "false"},
		},
	}
	g, err := ParseDefinition(def)
	if err != nil {
		t.Fatalf("ParseDefinition err: %v", err)
	}

	var events []string
	e := NewEngine(nil, nil, nil, Options{})
	res := e.Execute(context.Background(), g, RunInput{Message: "hi"}, func(event string, _ any) {
		events = append(events, event)
	})

	if res.Status != "success" {
		t.Fatalf("status = %s, want success (err=%s)", res.Status, res.Error)
	}
	// confidence 默认 1.0 >= 0.6，走 true 分支；end 输出 lastOutput（trigger 写入的 message）
	if res.Output != "hi" {
		t.Errorf("output = %q, want %q", res.Output, "hi")
	}
	if len(res.NodeResults) != 3 {
		t.Fatalf("node results = %d, want 3", len(res.NodeResults))
	}
	// 事件序列：每节点 start + end
	wantEvents := []string{"node_start", "node_end", "node_start", "node_end", "node_start", "node_end"}
	if len(events) != len(wantEvents) {
		t.Fatalf("events = %v", events)
	}
	for i := range wantEvents {
		if events[i] != wantEvents[i] {
			t.Fatalf("events = %v", events)
		}
	}
}

func TestEngineConditionFalseBranch(t *testing.T) {
	def := po.JSONMap{
		"nodes": []any{
			map[string]any{"id": "t", "type": "trigger"},
			map[string]any{"id": "c", "type": "condition", "config": map[string]any{"expression": "confidence < 0.6"}},
			map[string]any{"id": "yes", "type": "end", "config": map[string]any{"reason": "YES"}},
			map[string]any{"id": "no", "type": "end", "config": map[string]any{"reason": "NO"}},
		},
		"edges": []any{
			map[string]any{"id": "e1", "source": "t", "target": "c"},
			map[string]any{"id": "e2", "source": "c", "target": "yes", "sourcePort": "true"},
			map[string]any{"id": "e3", "source": "c", "target": "no", "sourcePort": "false"},
		},
	}
	g, _ := ParseDefinition(def)
	e := NewEngine(nil, nil, nil, Options{})
	// confidence 默认 1.0，表达式为 false → 走 no 分支；黑板 lastOutput 为空 → 输出 reason
	res := e.Execute(context.Background(), g, RunInput{Message: ""}, nil)
	if res.Output != "NO" {
		t.Errorf("output = %q, want NO", res.Output)
	}
}

func TestEngineConditionFailSoft(t *testing.T) {
	def := po.JSONMap{
		"nodes": []any{
			map[string]any{"id": "t", "type": "trigger"},
			map[string]any{"id": "c", "type": "condition", "config": map[string]any{"expression": "missing_var < 1"}},
			map[string]any{"id": "yes", "type": "end", "config": map[string]any{"reason": "YES"}},
			map[string]any{"id": "no", "type": "end", "config": map[string]any{"reason": "NO"}},
		},
		"edges": []any{
			map[string]any{"id": "e1", "source": "t", "target": "c"},
			map[string]any{"id": "e2", "source": "c", "target": "yes", "sourcePort": "true"},
			map[string]any{"id": "e3", "source": "c", "target": "no", "sourcePort": "false"},
		},
	}
	g, _ := ParseDefinition(def)
	e := NewEngine(nil, nil, nil, Options{})
	res := e.Execute(context.Background(), g, RunInput{Message: "m"}, nil)
	// 表达式失败 fail-soft 走 true 分支，且带 warning
	if res.Status != "success" || res.Output != "m" {
		t.Errorf("fail-soft 应走 true 分支: status=%s output=%q", res.Status, res.Output)
	}
	if len(res.Warnings) == 0 {
		t.Errorf("fail-soft 应产生 warning")
	}
}

func TestEngineMaxSteps(t *testing.T) {
	// 两条边互相来回的图会被解析期环检测拦截，这里直接构造图绕过检测验证运行期保险
	g := &RuntimeGraph{
		Nodes: map[string]*RuntimeNode{
			"a": {ID: "a", Type: nodeTypeTrigger},
			"b": {ID: "b", Type: nodeTypeCondition, Config: map[string]any{"expression": "confidence >= 0.6"}},
		},
		Edges: []RuntimeEdge{
			{ID: "e1", Source: "a", Target: "b"},
			{ID: "e2", Source: "b", Target: "a"},
		},
		adjacency: map[string][]int{"a": {0}, "b": {1}},
		TriggerID: "a",
	}
	e := NewEngine(nil, nil, nil, Options{MaxSteps: 10})
	res := e.Execute(context.Background(), g, RunInput{Message: "m"}, nil)
	if res.Status != "failed" {
		t.Errorf("环图运行期应 fail fast, got %s", res.Status)
	}
}

func TestEngineNaturalEnd(t *testing.T) {
	// 非 end 节点无出边 → success + warning
	def := po.JSONMap{
		"nodes": []any{
			map[string]any{"id": "t", "type": "trigger"},
			map[string]any{"id": "c", "type": "condition", "config": map[string]any{"expression": "confidence < 0.6"}},
		},
		"edges": []any{
			map[string]any{"id": "e1", "source": "t", "target": "c"},
		},
	}
	g, _ := ParseDefinition(def)
	e := NewEngine(nil, nil, nil, Options{})
	res := e.Execute(context.Background(), g, RunInput{Message: "hi"}, nil)
	if res.Status != "success" {
		t.Errorf("自然结束应为 success, got %s", res.Status)
	}
	if len(res.Warnings) == 0 {
		t.Errorf("自然结束应带 warning")
	}
}
