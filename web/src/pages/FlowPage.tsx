import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  AlertCircle,
  Bot,
  CheckCircle2,
  CircleDot,
  GitBranch,
  MousePointer2,
  Play,
  Plus,
  RefreshCw,
  Save,
  Trash2,
  Wrench,
  X,
} from "lucide-react";
import {
  createFlow,
  deleteFlow,
  listFlows,
  ShowFlow,
  updateFlow,
} from "../lib/api";
import { useAuthStore } from "../store/auth";

type FlowNodeType = "trigger" | "llm" | "skill" | "condition" | "end";

type FlowNode = {
  id: string;
  type: FlowNodeType;
  label: string;
  x: number;
  y: number;
  config: Record<string, unknown>;
};

type FlowEdge = {
  id: string;
  source: string;
  target: string;
};

type FlowDefinition = {
  nodes: FlowNode[];
  edges: FlowEdge[];
  viewport?: { x: number; y: number; zoom: number };
};

const nodeMeta: Record<
  FlowNodeType,
  { label: string; icon: JSX.Element; color: string; defaultConfig: Record<string, unknown> }
> = {
  trigger: {
    label: "触发",
    icon: <Play className="h-4 w-4" />,
    color: "border-emerald-200 bg-emerald-50 text-emerald-700",
    defaultConfig: { event: "conversation.message" },
  },
  llm: {
    label: "LLM",
    icon: <Bot className="h-4 w-4" />,
    color: "border-indigo-200 bg-indigo-50 text-indigo-700",
    defaultConfig: { mode: "rag", prompt: "根据上下文回复用户" },
  },
  skill: {
    label: "Skill",
    icon: <Wrench className="h-4 w-4" />,
    color: "border-amber-200 bg-amber-50 text-amber-700",
    defaultConfig: { skillName: "", argumentsFrom: "message" },
  },
  condition: {
    label: "条件",
    icon: <GitBranch className="h-4 w-4" />,
    color: "border-sky-200 bg-sky-50 text-sky-700",
    defaultConfig: { expression: "confidence < 0.6" },
  },
  end: {
    label: "结束",
    icon: <CheckCircle2 className="h-4 w-4" />,
    color: "border-slate-200 bg-slate-100 text-slate-700",
    defaultConfig: { reason: "completed" },
  },
};

function newId(prefix: string) {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return `${prefix}-${crypto.randomUUID().slice(0, 8)}`;
  }
  return `${prefix}-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 6)}`;
}

function defaultDefinition(): FlowDefinition {
  return {
    nodes: [
      {
        id: "start",
        type: "trigger",
        label: "开始",
        x: 110,
        y: 140,
        config: { event: "conversation.message" },
      },
      {
        id: "reply",
        type: "llm",
        label: "AI 回复",
        x: 430,
        y: 140,
        config: { mode: "rag" },
      },
    ],
    edges: [{ id: "edge-start-reply", source: "start", target: "reply" }],
    viewport: { x: 0, y: 0, zoom: 1 },
  };
}

function parseDefinition(raw?: string): FlowDefinition {
  if (!raw) return defaultDefinition();
  try {
    const parsed = JSON.parse(raw);
    return {
      nodes: Array.isArray(parsed.nodes) ? parsed.nodes : [],
      edges: Array.isArray(parsed.edges) ? parsed.edges : [],
      viewport: parsed.viewport ?? { x: 0, y: 0, zoom: 1 },
    };
  } catch {
    return defaultDefinition();
  }
}

function stringifyDefinition(definition: FlowDefinition) {
  return JSON.stringify(definition, null, 2);
}

function formatTime(sec?: number) {
  if (!sec) return "-";
  return new Date(sec * 1000).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function statusBadge(status: string) {
  const cls =
    status === "published"
      ? "bg-emerald-50 text-emerald-700 ring-emerald-600/20"
      : status === "archived"
        ? "bg-slate-100 text-slate-500 ring-slate-200"
        : "bg-amber-50 text-amber-700 ring-amber-600/20";
  const label = status === "published" ? "已发布" : status === "archived" ? "已归档" : "草稿";
  return <span className={`rounded-full px-2 py-0.5 text-xs font-medium ring-1 ring-inset ${cls}`}>{label}</span>;
}

function nodeCenter(node: FlowNode) {
  return { x: node.x + 86, y: node.y + 31 };
}

export default function FlowPage() {
  const { token } = useAuthStore();
  const canvasRef = useRef<HTMLDivElement>(null);
  const [flows, setFlows] = useState<ShowFlow[]>([]);
  const [current, setCurrent] = useState<ShowFlow | null>(null);
  const [definition, setDefinition] = useState<FlowDefinition>(defaultDefinition());
  const [selectedNodeId, setSelectedNodeId] = useState("start");
  const [connectSource, setConnectSource] = useState<string | null>(null);
  const [dragging, setDragging] = useState<{ id: string; dx: number; dy: number } | null>(null);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [dirty, setDirty] = useState(false);
  const [newFlowName, setNewFlowName] = useState("客服问答流程");

  const selectedNode = useMemo(
    () => definition.nodes.find((node) => node.id === selectedNodeId) ?? null,
    [definition.nodes, selectedNodeId]
  );

  const refresh = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    setError("");
    try {
      const resp = await listFlows(token);
      const list = resp.data ?? [];
      setFlows(list);
      const first = list[0];
      if (!current && first) {
        setCurrent(first);
        const next = parseDefinition(first.definition);
        setDefinition(next);
        setSelectedNodeId(next.nodes[0]?.id ?? "");
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : "加载 Flow 失败");
    } finally {
      setLoading(false);
    }
  }, [current, token]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  useEffect(() => {
    if (!dragging) return;

    const onMove = (event: MouseEvent) => {
      const rect = canvasRef.current?.getBoundingClientRect();
      if (!rect) return;
      const x = Math.max(20, event.clientX - rect.left - dragging.dx);
      const y = Math.max(20, event.clientY - rect.top - dragging.dy);
      setDefinition((prev) => ({
        ...prev,
        nodes: prev.nodes.map((node) => (node.id === dragging.id ? { ...node, x, y } : node)),
      }));
      setDirty(true);
    };

    const onUp = () => setDragging(null);
    window.addEventListener("mousemove", onMove);
    window.addEventListener("mouseup", onUp);
    return () => {
      window.removeEventListener("mousemove", onMove);
      window.removeEventListener("mouseup", onUp);
    };
  }, [dragging]);

  const selectFlow = (flow: ShowFlow) => {
    setCurrent(flow);
    const next = parseDefinition(flow.definition);
    setDefinition(next);
    setSelectedNodeId(next.nodes[0]?.id ?? "");
    setConnectSource(null);
    setDirty(false);
  };

  const createNewFlow = async () => {
    if (!token) return;
    const name = newFlowName.trim() || "客服问答流程";
    setSaving(true);
    setError("");
    try {
      const resp = await createFlow(token, {
        name,
        description: "可视化编排客服回复链路",
        definition: stringifyDefinition(defaultDefinition()),
        status: "draft",
      });
      setFlows((prev) => [resp.data, ...prev]);
      selectFlow(resp.data);
      setNewFlowName("");
    } catch (e) {
      setError(e instanceof Error ? e.message : "创建 Flow 失败");
    } finally {
      setSaving(false);
    }
  };

  const saveFlow = async () => {
    if (!token || !current) return;
    setSaving(true);
    setError("");
    try {
      const resp = await updateFlow(token, current.id, {
        name: current.name,
        description: current.description,
        status: current.status === "archived" ? "archived" : current.status === "published" ? "published" : "draft",
        definition: stringifyDefinition(definition),
      });
      setCurrent(resp.data);
      setFlows((prev) => prev.map((flow) => (flow.id === resp.data.id ? resp.data : flow)));
      setDirty(false);
    } catch (e) {
      setError(e instanceof Error ? e.message : "保存 Flow 失败");
    } finally {
      setSaving(false);
    }
  };

  const removeFlow = async () => {
    if (!token || !current) return;
    if (!window.confirm(`确认删除 Flow「${current.name}」？`)) return;
    await deleteFlow(token, current.id);
    const rest = flows.filter((flow) => flow.id !== current.id);
    setFlows(rest);
    const next = rest[0];
    if (next) {
      selectFlow(next);
    } else {
      setCurrent(null);
      setDefinition(defaultDefinition());
      setSelectedNodeId("");
    }
  };

  const addNode = (type: FlowNodeType) => {
    const count = definition.nodes.length;
    const node: FlowNode = {
      id: newId(type),
      type,
      label: nodeMeta[type].label,
      x: 180 + (count % 3) * 210,
      y: 120 + Math.floor(count / 3) * 130,
      config: nodeMeta[type].defaultConfig,
    };
    setDefinition((prev) => ({ ...prev, nodes: [...prev.nodes, node] }));
    setSelectedNodeId(node.id);
    setDirty(true);
  };

  const removeNode = (id: string) => {
    setDefinition((prev) => ({
      ...prev,
      nodes: prev.nodes.filter((node) => node.id !== id),
      edges: prev.edges.filter((edge) => edge.source !== id && edge.target !== id),
    }));
    setSelectedNodeId("");
    setDirty(true);
  };

  const updateNode = (id: string, patch: Partial<FlowNode>) => {
    setDefinition((prev) => ({
      ...prev,
      nodes: prev.nodes.map((node) => (node.id === id ? { ...node, ...patch } : node)),
    }));
    setDirty(true);
  };

  const handleNodeClick = (node: FlowNode) => {
    if (connectSource && connectSource !== node.id) {
      const exists = definition.edges.some((edge) => edge.source === connectSource && edge.target === node.id);
      if (!exists) {
        setDefinition((prev) => ({
          ...prev,
          edges: [...prev.edges, { id: newId("edge"), source: connectSource, target: node.id }],
        }));
        setDirty(true);
      }
      setConnectSource(null);
      setSelectedNodeId(node.id);
      return;
    }
    setSelectedNodeId(node.id);
  };

  const removeEdge = (id: string) => {
    setDefinition((prev) => ({ ...prev, edges: prev.edges.filter((edge) => edge.id !== id) }));
    setDirty(true);
  };

  const updateConfigFromText = (text: string) => {
    if (!selectedNode) return;
    try {
      const parsed = JSON.parse(text || "{}");
      updateNode(selectedNode.id, { config: parsed });
      setError("");
    } catch {
      setError("节点配置不是有效 JSON");
    }
  };

  return (
    <div className="flex h-full bg-slate-50">
      <aside className="flex w-72 flex-col border-r border-slate-200 bg-white">
        <div className="border-b border-slate-200 px-4 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-lg font-semibold text-slate-900">Flow</h1>
              <p className="text-sm text-slate-500">可视化流程编排</p>
            </div>
            <button
              type="button"
              title="刷新"
              onClick={refresh}
              className="rounded p-2 text-slate-500 hover:bg-slate-100 hover:text-slate-800"
            >
              <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
            </button>
          </div>
          <div className="mt-4 flex gap-2">
            <input
              value={newFlowName}
              onChange={(e) => setNewFlowName(e.target.value)}
              className="min-w-0 flex-1 rounded-md border border-slate-300 px-3 py-2 text-sm outline-none focus:border-indigo-500 focus:ring-2 focus:ring-indigo-100"
              placeholder="新流程名称"
            />
            <button
              type="button"
              title="新建"
              onClick={createNewFlow}
              disabled={saving}
              className="rounded-md bg-indigo-600 px-3 text-white hover:bg-indigo-700 disabled:opacity-60"
            >
              <Plus className="h-4 w-4" />
            </button>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto p-3">
          {flows.length === 0 ? (
            <div className="rounded-lg border border-dashed border-slate-300 px-4 py-10 text-center">
              <GitBranch className="mx-auto h-8 w-8 text-slate-300" />
              <p className="mt-2 text-sm font-medium text-slate-700">暂无 Flow</p>
            </div>
          ) : (
            <div className="space-y-2">
              {flows.map((flow) => (
                <button
                  key={flow.id}
                  type="button"
                  onClick={() => selectFlow(flow)}
                  className={`w-full rounded-lg border px-3 py-3 text-left transition ${
                    current?.id === flow.id
                      ? "border-indigo-200 bg-indigo-50"
                      : "border-slate-200 bg-white hover:bg-slate-50"
                  }`}
                >
                  <div className="flex items-center justify-between gap-2">
                    <span className="truncate text-sm font-medium text-slate-900">{flow.name}</span>
                    {statusBadge(flow.status)}
                  </div>
                  <div className="mt-1 text-xs text-slate-500">v{flow.version} · {formatTime(flow.updatedAt)}</div>
                </button>
              ))}
            </div>
          )}
        </div>
      </aside>

      <main className="flex min-w-0 flex-1 flex-col">
        <header className="flex items-center justify-between border-b border-slate-200 bg-white px-5 py-3">
          <div className="min-w-0">
            <input
              value={current?.name ?? ""}
              disabled={!current}
              onChange={(e) => current && setCurrent({ ...current, name: e.target.value })}
              className="w-full border-0 bg-transparent text-lg font-semibold text-slate-900 outline-none disabled:text-slate-400"
              placeholder="请选择或新建 Flow"
            />
            <input
              value={current?.description ?? ""}
              disabled={!current}
              onChange={(e) => current && setCurrent({ ...current, description: e.target.value })}
              className="mt-1 w-full border-0 bg-transparent text-sm text-slate-500 outline-none disabled:text-slate-400"
              placeholder="流程描述"
            />
          </div>
          <div className="flex items-center gap-2">
            {error && (
              <div className="hidden max-w-xs items-center gap-1.5 rounded-md bg-red-50 px-2 py-1 text-xs text-red-600 lg:flex">
                <AlertCircle className="h-3.5 w-3.5" />
                <span className="truncate">{error}</span>
              </div>
            )}
            <select
              value={current?.status ?? "draft"}
              disabled={!current}
              onChange={(e) => current && setCurrent({ ...current, status: e.target.value })}
              className="h-9 rounded-md border border-slate-300 bg-white px-2 text-sm text-slate-700"
            >
              <option value="draft">草稿</option>
              <option value="published">发布</option>
              <option value="archived">归档</option>
            </select>
            <button
              type="button"
              disabled={!current || saving}
              onClick={saveFlow}
              className="inline-flex h-9 items-center gap-2 rounded-md bg-indigo-600 px-3 text-sm font-medium text-white hover:bg-indigo-700 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {saving ? <RefreshCw className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
              {dirty ? "保存*" : "保存"}
            </button>
            <button
              type="button"
              title="删除 Flow"
              disabled={!current}
              onClick={removeFlow}
              className="h-9 rounded-md border border-slate-300 px-2 text-slate-500 hover:bg-red-50 hover:text-red-600 disabled:opacity-50"
            >
              <Trash2 className="h-4 w-4" />
            </button>
          </div>
        </header>

        <div className="flex min-h-0 flex-1">
          <section className="flex min-w-0 flex-1 flex-col">
            <div className="flex items-center gap-2 border-b border-slate-200 bg-white px-4 py-2">
              <span className="text-xs font-medium text-slate-500">节点</span>
              {(Object.keys(nodeMeta) as FlowNodeType[]).map((type) => (
                <button
                  key={type}
                  type="button"
                  onClick={() => addNode(type)}
                  disabled={!current}
                  className="inline-flex h-8 items-center gap-1.5 rounded-md border border-slate-200 px-2 text-xs font-medium text-slate-700 hover:bg-slate-50 disabled:opacity-50"
                >
                  {nodeMeta[type].icon}
                  {nodeMeta[type].label}
                </button>
              ))}
              {connectSource && (
                <span className="ml-auto inline-flex items-center gap-1.5 rounded-md bg-indigo-50 px-2 py-1 text-xs text-indigo-700">
                  <MousePointer2 className="h-3.5 w-3.5" />
                  选择目标节点
                </span>
              )}
            </div>

            <div
              ref={canvasRef}
              className="relative flex-1 overflow-auto bg-[linear-gradient(#e2e8f0_1px,transparent_1px),linear-gradient(90deg,#e2e8f0_1px,transparent_1px)] bg-[size:24px_24px]"
            >
              <svg className="pointer-events-none absolute inset-0 h-[1600px] w-[2200px]">
                <defs>
                  <marker id="arrow" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
                    <path d="M 0 0 L 10 5 L 0 10 z" fill="#64748b" />
                  </marker>
                </defs>
                {definition.edges.map((edge) => {
                  const source = definition.nodes.find((node) => node.id === edge.source);
                  const target = definition.nodes.find((node) => node.id === edge.target);
                  if (!source || !target) return null;
                  const a = nodeCenter(source);
                  const b = nodeCenter(target);
                  const midX = (a.x + b.x) / 2;
                  return (
                    <path
                      key={edge.id}
                      d={`M ${a.x} ${a.y} C ${midX} ${a.y}, ${midX} ${b.y}, ${b.x} ${b.y}`}
                      fill="none"
                      stroke="#64748b"
                      strokeWidth="2"
                      markerEnd="url(#arrow)"
                    />
                  );
                })}
              </svg>

              {definition.nodes.map((node) => {
                const meta = nodeMeta[node.type] ?? nodeMeta.llm;
                const active = selectedNodeId === node.id;
                return (
                  <button
                    key={node.id}
                    type="button"
                    onClick={() => handleNodeClick(node)}
                    onMouseDown={(event) => {
                      const rect = (event.currentTarget as HTMLButtonElement).getBoundingClientRect();
                      setDragging({ id: node.id, dx: event.clientX - rect.left, dy: event.clientY - rect.top });
                    }}
                    className={`absolute w-[172px] rounded-lg border bg-white px-3 py-2 text-left shadow-sm transition hover:shadow-md ${
                      active ? "border-indigo-400 ring-2 ring-indigo-100" : "border-slate-200"
                    } ${connectSource === node.id ? "ring-2 ring-amber-200" : ""}`}
                    style={{ left: node.x, top: node.y }}
                  >
                    <div className="flex items-center gap-2">
                      <span className={`inline-flex h-7 w-7 items-center justify-center rounded-md border ${meta.color}`}>
                        {meta.icon}
                      </span>
                      <div className="min-w-0">
                        <div className="truncate text-sm font-semibold text-slate-900">{node.label}</div>
                        <div className="text-xs text-slate-500">{meta.label}</div>
                      </div>
                    </div>
                  </button>
                );
              })}
            </div>
          </section>

          <aside className="flex w-80 flex-col border-l border-slate-200 bg-white">
            <div className="border-b border-slate-200 px-4 py-3">
              <h2 className="text-sm font-semibold text-slate-900">Inspector</h2>
              <p className="text-xs text-slate-500">{selectedNode ? selectedNode.id : "未选择节点"}</p>
            </div>

            {selectedNode ? (
              <div className="flex-1 overflow-y-auto px-4 py-4">
                <label className="block">
                  <span className="mb-1 block text-xs font-medium text-slate-500">标签</span>
                  <input
                    value={selectedNode.label}
                    onChange={(e) => updateNode(selectedNode.id, { label: e.target.value })}
                    className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm outline-none focus:border-indigo-500 focus:ring-2 focus:ring-indigo-100"
                  />
                </label>

                <label className="mt-4 block">
                  <span className="mb-1 block text-xs font-medium text-slate-500">类型</span>
                  <select
                    value={selectedNode.type}
                    onChange={(e) => {
                      const type = e.target.value as FlowNodeType;
                      updateNode(selectedNode.id, { type, config: nodeMeta[type].defaultConfig });
                    }}
                    className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm outline-none focus:border-indigo-500 focus:ring-2 focus:ring-indigo-100"
                  >
                    {(Object.keys(nodeMeta) as FlowNodeType[]).map((type) => (
                      <option key={type} value={type}>
                        {nodeMeta[type].label}
                      </option>
                    ))}
                  </select>
                </label>

                <label className="mt-4 block">
                  <span className="mb-1 block text-xs font-medium text-slate-500">配置 JSON</span>
                  <textarea
                    key={selectedNode.id}
                    defaultValue={JSON.stringify(selectedNode.config ?? {}, null, 2)}
                    onBlur={(e) => updateConfigFromText(e.target.value)}
                    rows={10}
                    spellCheck={false}
                    className="w-full resize-y rounded-md border border-slate-300 bg-slate-950 px-3 py-2 font-mono text-xs leading-5 text-slate-100 outline-none focus:border-indigo-500 focus:ring-2 focus:ring-indigo-100"
                  />
                </label>

                <div className="mt-4 grid grid-cols-2 gap-2">
                  <button
                    type="button"
                    onClick={() => setConnectSource(selectedNode.id)}
                    className="inline-flex items-center justify-center gap-1.5 rounded-md border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50"
                  >
                    <CircleDot className="h-4 w-4" />
                    连线
                  </button>
                  <button
                    type="button"
                    onClick={() => removeNode(selectedNode.id)}
                    className="inline-flex items-center justify-center gap-1.5 rounded-md border border-red-200 px-3 py-2 text-sm font-medium text-red-600 hover:bg-red-50"
                  >
                    <Trash2 className="h-4 w-4" />
                    删除
                  </button>
                </div>

                <div className="mt-6">
                  <div className="mb-2 text-xs font-medium text-slate-500">相关连线</div>
                  <div className="space-y-2">
                    {definition.edges
                      .filter((edge) => edge.source === selectedNode.id || edge.target === selectedNode.id)
                      .map((edge) => (
                        <div key={edge.id} className="flex items-center justify-between rounded-md border border-slate-200 px-2 py-1.5 text-xs text-slate-600">
                          <span className="truncate">
                            {edge.source} → {edge.target}
                          </span>
                          <button
                            type="button"
                            title="删除连线"
                            onClick={() => removeEdge(edge.id)}
                            className="rounded p-1 text-slate-400 hover:bg-red-50 hover:text-red-600"
                          >
                            <X className="h-3.5 w-3.5" />
                          </button>
                        </div>
                      ))}
                    {definition.edges.filter((edge) => edge.source === selectedNode.id || edge.target === selectedNode.id).length === 0 && (
                      <div className="rounded-md border border-dashed border-slate-300 px-3 py-4 text-center text-xs text-slate-400">
                        暂无线
                      </div>
                    )}
                  </div>
                </div>
              </div>
            ) : (
              <div className="flex flex-1 items-center justify-center px-6 text-center text-sm text-slate-400">
                选择一个节点开始编辑
              </div>
            )}
          </aside>
        </div>
      </main>
    </div>
  );
}
