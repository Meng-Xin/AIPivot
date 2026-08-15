import { useCallback, useEffect, useRef, useState } from "react";
import {
  AlertCircle,
  CheckCircle2,
  ChevronDown,
  ChevronRight,
  Clock,
  History,
  MinusCircle,
  Play,
  Square,
  XCircle,
} from "lucide-react";
import {
  FlowRunInfo,
  listFlowRuns,
  runFlowStream,
  ShowFlow,
} from "../lib/api";

type TimelineEntry = {
  nodeId: string;
  nodeType: string;
  label: string;
  status: "running" | "success" | "skipped" | "failed";
  durationMs?: number;
  content: string;
  summary?: Record<string, unknown>;
};

function statusIcon(status: string) {
  const cls = "h-4 w-4 shrink-0";
  switch (status) {
    case "success":
      return <CheckCircle2 className={`${cls} text-emerald-500`} />;
    case "failed":
      return <XCircle className={`${cls} text-red-500`} />;
    case "skipped":
      return <MinusCircle className={`${cls} text-amber-500`} />;
    default:
      return <Clock className={`${cls} animate-spin text-indigo-500`} />;
  }
}

function runStatusBadge(status: string) {
  const map: Record<string, { label: string; cls: string }> = {
    running: { label: "执行中", cls: "bg-indigo-50 text-indigo-700 ring-indigo-600/20" },
    success: { label: "成功", cls: "bg-emerald-50 text-emerald-700 ring-emerald-600/20" },
    failed: { label: "失败", cls: "bg-red-50 text-red-700 ring-red-600/20" },
    timeout: { label: "超时", cls: "bg-amber-50 text-amber-700 ring-amber-600/20" },
  };
  const meta = map[status] ?? { label: status, cls: "bg-slate-100 text-slate-600 ring-slate-200" };
  return (
    <span className={`rounded-full px-2 py-0.5 text-xs font-medium ring-1 ring-inset ${meta.cls}`}>
      {meta.label}
    </span>
  );
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

// fail-soft 解析 JSON 字符串（历史记录里的 input/nodeResults 可能被人为改坏）
function safeParse<T>(raw: string, fallback: T): T {
  try {
    return JSON.parse(raw) as T;
  } catch {
    return fallback;
  }
}

interface FlowRunPanelProps {
  token: string;
  flow: ShowFlow;
  onNodeActive?: (nodeId: string | null) => void;
}

/**
 * Flow 试运行面板：测试输入 + 实时执行流 + 执行历史。
 * 状态为页面级瞬态（useState/useRef），不进 zustand。
 */
export default function FlowRunPanel({ token, flow, onNodeActive }: FlowRunPanelProps) {
  const [message, setMessage] = useState("你好，请帮我查一下订单状态");
  const [running, setRunning] = useState(false);
  const [timeline, setTimeline] = useState<TimelineEntry[]>([]);
  const [runError, setRunError] = useState("");
  const [runs, setRuns] = useState<FlowRunInfo[]>([]);
  const [expandedRunId, setExpandedRunId] = useState<number | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const refreshRuns = useCallback(async () => {
    if (!token) return;
    try {
      const resp = await listFlowRuns(token, flow.id);
      setRuns(resp.data ?? []);
    } catch {
      // 历史加载失败不打断面板，静默保留已有数据
    }
  }, [token, flow.id]);

  useEffect(() => {
    setTimeline([]);
    setRunError("");
    refreshRuns();
    return () => abortRef.current?.abort();
  }, [refreshRuns]);

  const patchEntry = (nodeId: string, patch: Partial<TimelineEntry>) => {
    setTimeline((prev) =>
      prev.map((entry) => (entry.nodeId === nodeId ? { ...entry, ...patch } : entry))
    );
  };

  const handleRun = async () => {
    if (running || !message.trim()) return;
    setRunning(true);
    setTimeline([]);
    setRunError("");

    const controller = new AbortController();
    abortRef.current = controller;

    try {
      await runFlowStream(
        token,
        flow.id,
        message.trim(),
        undefined,
        {
          onNodeStart: (data) => {
            onNodeActive?.(data.nodeId);
            setTimeline((prev) => [
              ...prev,
              {
                nodeId: data.nodeId,
                nodeType: data.nodeType,
                label: data.label || data.nodeId,
                status: "running",
                content: "",
              },
            ]);
          },
          onDelta: (data) => {
            setTimeline((prev) =>
              prev.map((entry) =>
                entry.nodeId === data.nodeId
                  ? { ...entry, content: entry.content + data.content }
                  : entry
              )
            );
          },
          onNodeEnd: (data) => {
            patchEntry(data.nodeId, {
              status: (data.status as TimelineEntry["status"]) ?? "success",
              durationMs: data.durationMs,
              summary: data.summary,
            });
          },
          onError: (data) => {
            setRunError(data.msg || "试运行失败");
          },
        },
        controller.signal
      );
    } catch (e) {
      if (!(e instanceof DOMException && e.name === "AbortError")) {
        setRunError(e instanceof Error ? e.message : "试运行失败");
      }
    } finally {
      setRunning(false);
      onNodeActive?.(null);
      abortRef.current = null;
      refreshRuns();
    }
  };

  const handleStop = () => {
    abortRef.current?.abort();
    setRunning(false);
    onNodeActive?.(null);
  };

  return (
    <div className="flex flex-1 flex-col overflow-hidden">
      <div className="border-b border-slate-200 px-4 py-3">
        <label className="block">
          <span className="mb-1 block text-xs font-medium text-slate-500">测试消息</span>
          <textarea
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            rows={3}
            maxLength={2000}
            placeholder="输入测试消息（≤ 2000 字符），模拟 conversation.message 触发"
            className="w-full resize-y rounded-md border border-slate-300 px-3 py-2 text-sm outline-none focus:border-indigo-500 focus:ring-2 focus:ring-indigo-100"
          />
        </label>
        <div className="mt-2 flex items-center gap-2">
          <button
            type="button"
            onClick={handleRun}
            disabled={running || !message.trim() || flow.status === "archived"}
            className="inline-flex h-8 items-center gap-1.5 rounded-md bg-indigo-600 px-3 text-sm font-medium text-white hover:bg-indigo-700 disabled:cursor-not-allowed disabled:opacity-60"
          >
            <Play className="h-4 w-4" />
            运行
          </button>
          {running && (
            <button
              type="button"
              onClick={handleStop}
              className="inline-flex h-8 items-center gap-1.5 rounded-md border border-slate-300 px-3 text-sm font-medium text-slate-600 hover:bg-slate-50"
            >
              <Square className="h-3.5 w-3.5" />
              中止
            </button>
          )}
          {flow.status === "archived" && (
            <span className="text-xs text-slate-400">已归档流程不可运行</span>
          )}
        </div>
        {runError && (
          <div className="mt-2 flex items-start gap-1.5 rounded-md bg-red-50 px-2 py-1.5 text-xs text-red-600">
            <AlertCircle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
            <span className="break-all">{runError}</span>
          </div>
        )}
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto px-4 py-3">
        <div className="mb-2 text-xs font-medium text-slate-500">执行流</div>
        {timeline.length === 0 ? (
          <div className="rounded-md border border-dashed border-slate-300 px-3 py-6 text-center text-xs text-slate-400">
            点击「运行」查看实时执行过程
          </div>
        ) : (
          <div className="space-y-2">
            {timeline.map((entry) => (
              <div key={entry.nodeId} className="rounded-md border border-slate-200 px-3 py-2">
                <div className="flex items-center gap-2">
                  {statusIcon(entry.status)}
                  <span className="truncate text-sm font-medium text-slate-900">{entry.label}</span>
                  <span className="rounded bg-slate-100 px-1.5 py-0.5 text-[10px] text-slate-500">
                    {entry.nodeType}
                  </span>
                  {entry.durationMs !== undefined && (
                    <span className="ml-auto text-xs text-slate-400">{entry.durationMs}ms</span>
                  )}
                </div>
                {entry.content && (
                  <pre className="mt-1.5 max-h-40 overflow-y-auto whitespace-pre-wrap break-all rounded bg-slate-50 px-2 py-1.5 font-mono text-xs leading-5 text-slate-700">
                    {entry.content}
                  </pre>
                )}
                {entry.summary && Object.keys(entry.summary).length > 0 && (
                  <div className="mt-1.5 flex flex-wrap gap-1">
                    {Object.entries(entry.summary).map(([key, value]) => (
                      <span
                        key={key}
                        className="rounded bg-slate-100 px-1.5 py-0.5 font-mono text-[10px] text-slate-500"
                      >
                        {key}: {String(value)}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}

        <div className="mb-2 mt-5 flex items-center justify-between text-xs font-medium text-slate-500">
          <span className="inline-flex items-center gap-1">
            <History className="h-3.5 w-3.5" />
            执行历史
          </span>
          <button
            type="button"
            onClick={refreshRuns}
            className="rounded px-1.5 py-0.5 text-xs text-slate-400 hover:bg-slate-100 hover:text-slate-600"
          >
            刷新
          </button>
        </div>
        {runs.length === 0 ? (
          <div className="rounded-md border border-dashed border-slate-300 px-3 py-4 text-center text-xs text-slate-400">
            暂无执行记录
          </div>
        ) : (
          <div className="space-y-1.5">
            {runs.map((run) => {
              const expanded = expandedRunId === run.id;
              const nodeResults = safeParse<Array<Record<string, unknown>>>(
                run.nodeResults,
                []
              );
              return (
                <div key={run.id} className="rounded-md border border-slate-200">
                  <button
                    type="button"
                    onClick={() => setExpandedRunId(expanded ? null : run.id)}
                    className="flex w-full items-center gap-2 px-2.5 py-2 text-left hover:bg-slate-50"
                  >
                    {expanded ? (
                      <ChevronDown className="h-3.5 w-3.5 text-slate-400" />
                    ) : (
                      <ChevronRight className="h-3.5 w-3.5 text-slate-400" />
                    )}
                    {runStatusBadge(run.status)}
                    <span className="text-xs text-slate-500">{run.totalMs}ms</span>
                    <span className="text-xs text-slate-400">{run.tokenCount} tok</span>
                    <span className="ml-auto text-xs text-slate-400">
                      {formatTime(run.createdAt)}
                    </span>
                  </button>
                  {expanded && (
                    <div className="space-y-2 border-t border-slate-100 px-3 py-2">
                      {run.error && (
                        <div className="rounded bg-red-50 px-2 py-1.5 text-xs text-red-600">
                          {run.error}
                        </div>
                      )}
                      <div>
                        <div className="mb-1 text-[10px] font-medium text-slate-400">输出</div>
                        <pre className="max-h-32 overflow-y-auto whitespace-pre-wrap break-all rounded bg-slate-50 px-2 py-1.5 font-mono text-xs leading-5 text-slate-700">
                          {run.output || "-"}
                        </pre>
                      </div>
                      <div>
                        <div className="mb-1 text-[10px] font-medium text-slate-400">
                          节点结果（{nodeResults.length}）
                        </div>
                        <div className="space-y-1">
                          {nodeResults.map((node, index) => (
                            <div
                              key={`${node.nodeId}-${index}`}
                              className="flex flex-wrap items-center gap-1.5 text-xs text-slate-600"
                            >
                              {statusIcon(String(node.status ?? "success"))}
                              <span className="font-medium">{String(node.label ?? node.nodeId)}</span>
                              <span className="text-slate-400">{String(node.durationMs ?? 0)}ms</span>
                              {node.warning ? (
                                <span className="text-amber-600">{String(node.warning)}</span>
                              ) : null}
                            </div>
                          ))}
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
