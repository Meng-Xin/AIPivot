package flow

import (
	"fmt"

	"aipivot/internal/shared/po"
)

// ParseDefinition 将 flows.definition JSONB 解析为运行时图，并做结构校验：
// 节点 id 唯一 / 边引用存在 / 恰好存在 trigger 入口 / 从 trigger 出发不可达环（菱形汇合合法）。
func ParseDefinition(def po.JSONMap) (*RuntimeGraph, error) {
	if def == nil {
		return nil, fmt.Errorf("definition 为空")
	}

	rawNodes, _ := def["nodes"].([]interface{})
	if len(rawNodes) == 0 {
		return nil, fmt.Errorf("definition 缺少节点")
	}

	g := &RuntimeGraph{
		Nodes:     make(map[string]*RuntimeNode, len(rawNodes)),
		adjacency: make(map[string][]int),
	}

	for _, raw := range rawNodes {
		m, ok := raw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("节点必须是 JSON 对象")
		}
		node := &RuntimeNode{
			ID:     stringOf(m["id"]),
			Type:   stringOf(m["type"]),
			Label:  stringOf(m["label"]),
			Config: mapOf(m["config"]),
		}
		if node.ID == "" {
			return nil, fmt.Errorf("存在缺少 id 的节点")
		}
		if _, dup := g.Nodes[node.ID]; dup {
			return nil, fmt.Errorf("节点 id 重复: %s", node.ID)
		}
		if node.Type == "" {
			node.Type = "llm"
		}
		g.Nodes[node.ID] = node
		if node.Type == "trigger" && g.TriggerID == "" {
			g.TriggerID = node.ID
		}
	}
	if g.TriggerID == "" {
		return nil, fmt.Errorf("缺少 trigger 入口节点")
	}

	rawEdges, _ := def["edges"].([]interface{})
	for _, raw := range rawEdges {
		m, ok := raw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("连线必须是 JSON 对象")
		}
		edge := RuntimeEdge{
			ID:         stringOf(m["id"]),
			Source:     stringOf(m["source"]),
			Target:     stringOf(m["target"]),
			SourcePort: stringOf(m["sourcePort"]),
		}
		if _, ok := g.Nodes[edge.Source]; !ok {
			return nil, fmt.Errorf("连线 %s 引用了不存在的源节点 %s", edge.ID, edge.Source)
		}
		if _, ok := g.Nodes[edge.Target]; !ok {
			return nil, fmt.Errorf("连线 %s 引用了不存在的目标节点 %s", edge.ID, edge.Target)
		}
		g.Edges = append(g.Edges, edge)
		g.adjacency[edge.Source] = append(g.adjacency[edge.Source], len(g.Edges)-1)
	}

	if err := g.detectCycle(); err != nil {
		return nil, err
	}
	return g, nil
}

// detectCycle 从 trigger 出发做路径环检测（DFS + on-path 标记）。
// 菱形汇合（同一节点经两条不同路径到达后再汇合）合法，因为回溯时会从路径中移除。
func (g *RuntimeGraph) detectCycle() error {
	const (
		white = 0 // 未访问
		gray  = 1 // 在当前 DFS 路径上
		black = 2 // 已完整探索
	)
	color := make(map[string]int, len(g.Nodes))
	var visit func(id string, path []string) error
	visit = func(id string, path []string) error {
		color[id] = gray
		path = append(path, id)
		for _, e := range g.OutEdges(id) {
			switch color[e.Target] {
			case gray:
				return fmt.Errorf("流程存在环: %s -> %s", id, e.Target)
			case white:
				if err := visit(e.Target, path); err != nil {
					return err
				}
			}
		}
		color[id] = black
		return nil
	}
	return visit(g.TriggerID, nil)
}

// stringOf / mapOf / floatOf / boolOf 是 definition JSON 宽松取值辅助（缺省返回零值，不报错）。
func stringOf(v any) string {
	s, _ := v.(string)
	return s
}

func mapOf(v any) map[string]any {
	m, _ := v.(map[string]any)
	if m == nil {
		return map[string]any{}
	}
	return m
}

func floatOf(v any) (float64, bool) {
	f, ok := v.(float64)
	return f, ok
}

func boolOf(v any) (bool, bool) {
	b, ok := v.(bool)
	return b, ok
}
