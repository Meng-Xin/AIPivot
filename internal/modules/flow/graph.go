package flow

// RuntimeNode 运行时节点（从 definition JSON 解析）。
type RuntimeNode struct {
	ID     string
	Type   string // trigger / llm / skill / condition / end
	Label  string
	Config map[string]any
}

// RuntimeEdge 运行时连线。SourcePort 为 condition 分支端口（"true"/"false"），其余节点为空。
type RuntimeEdge struct {
	ID         string
	Source     string
	Target     string
	SourcePort string
}

// RuntimeGraph 运行时图：节点表 + 有序边表 + 邻接表。
type RuntimeGraph struct {
	Nodes       map[string]*RuntimeNode
	Edges       []RuntimeEdge
	adjacency   map[string][]int // nodeID -> 边下标（保持 definition 顺序）
	TriggerID   string           // 入口 trigger 节点
}

// OutEdges 返回节点的全部出边（保持 definition 顺序）。
func (g *RuntimeGraph) OutEdges(nodeID string) []RuntimeEdge {
	idx := g.adjacency[nodeID]
	edges := make([]RuntimeEdge, 0, len(idx))
	for _, i := range idx {
		edges = append(edges, g.Edges[i])
	}
	return edges
}

// SelectOutEdge 选择节点的下一条出边。
// condition 节点按分支结果选 port 匹配的边；无 port 标注时确定性兜底：true→第一条，false→第二条。
// 非 condition 节点（decision 恒为 true）取第一条出边。
// 返回 ok=false 表示无出边（自然结束）。
func (g *RuntimeGraph) SelectOutEdge(nodeID string, decision bool) (RuntimeEdge, bool) {
	edges := g.OutEdges(nodeID)
	if len(edges) == 0 {
		return RuntimeEdge{}, false
	}

	if decision {
		for _, e := range edges {
			if e.SourcePort == "true" {
				return e, true
			}
		}
		return edges[0], true
	}

	for _, e := range edges {
		if e.SourcePort == "false" {
			return e, true
		}
	}
	// 无 port 标注时的确定性兜底：false → 第二条出边
	if len(edges) >= 2 {
		return edges[1], true
	}
	return RuntimeEdge{}, false
}
