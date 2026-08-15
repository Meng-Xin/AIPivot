import { useEffect, useState, useCallback } from "react";
import {
  MessageSquare,
  Users,
  Zap,
  DollarSign,
  CheckCircle,
  AlertTriangle,
  RefreshCw,
  Download,
  Printer,
  ThumbsUp,
} from "lucide-react";
import { useAuthStore } from "../store/auth";
import {
  getAnalyticsOverview,
  getAnalyticsDaily,
  exportAnalyticsCsv,
  type AnalyticsOverviewData,
  type DailyStatPoint,
} from "../lib/api";

// ==================== KPI Card ====================

function KpiCard({
  label,
  value,
  sub,
  icon,
  color,
}: {
  label: string;
  value: string;
  sub?: string;
  icon: React.ReactNode;
  color: string;
}) {
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
      <div className="flex items-start justify-between">
        <div>
          <p className="text-sm font-medium text-slate-500">{label}</p>
          <p className="mt-1 text-2xl font-bold text-slate-800">{value}</p>
          {sub && <p className="mt-1 text-xs text-slate-400">{sub}</p>}
        </div>
        <div className={`rounded-lg p-2 ${color}`}>{icon}</div>
      </div>
    </div>
  );
}

// ==================== SVG Area Chart ====================

interface ChartSeries {
  label: string;
  color: string;
  fillColor: string;
  key: keyof DailyStatPoint;
}

function AreaChart({
  data,
  series,
  height = 180,
}: {
  data: DailyStatPoint[];
  series: ChartSeries[];
  height?: number;
}) {
  if (!data.length) return null;

  const paddingLeft = 48;
  const paddingRight = 12;
  const paddingTop = 12;
  const paddingBottom = 32;
  const w = 600;
  const h = height;
  const innerW = w - paddingLeft - paddingRight;
  const innerH = h - paddingTop - paddingBottom;

  const maxValues = series.map((s) =>
    Math.max(...data.map((d) => Number(d[s.key]) || 0))
  );
  const globalMax = Math.max(...maxValues, 1);

  function toX(i: number) {
    return paddingLeft + (i / Math.max(data.length - 1, 1)) * innerW;
  }
  function toY(val: number) {
    return paddingTop + innerH - (val / globalMax) * innerH;
  }

  function buildPath(key: keyof DailyStatPoint) {
    return data
      .map((d, i) => `${i === 0 ? "M" : "L"}${toX(i)},${toY(Number(d[key]) || 0)}`)
      .join(" ");
  }

  function buildArea(key: keyof DailyStatPoint) {
    const line = buildPath(key);
    const lastX = toX(data.length - 1);
    const baseY = paddingTop + innerH;
    return `${line} L${lastX},${baseY} L${paddingLeft},${baseY} Z`;
  }

  // Y-axis tick marks (4 ticks)
  const ticks = [0, 0.25, 0.5, 0.75, 1].map((t) => ({
    val: Math.round(globalMax * t),
    y: toY(globalMax * t),
  }));

  // X-axis: show every N-th label to avoid crowding
  const labelStep = Math.ceil(data.length / 7);

  return (
    <svg
      viewBox={`0 0 ${w} ${h}`}
      className="w-full"
      style={{ height }}
      preserveAspectRatio="none"
    >
      {/* Grid lines */}
      {ticks.map((t, i) => (
        <g key={`tick-${i}-${t.val}`}>
          <line
            x1={paddingLeft}
            y1={t.y}
            x2={w - paddingRight}
            y2={t.y}
            stroke="#e2e8f0"
            strokeWidth="1"
          />
          <text
            x={paddingLeft - 6}
            y={t.y + 4}
            textAnchor="end"
            fontSize="10"
            fill="#94a3b8"
          >
            {t.val >= 1000 ? `${(t.val / 1000).toFixed(1)}k` : t.val}
          </text>
        </g>
      ))}

      {/* X-axis labels */}
      {data.map((d, i) =>
        i % labelStep === 0 || i === data.length - 1 ? (
          <text
            key={d.date}
            x={toX(i)}
            y={h - 6}
            textAnchor="middle"
            fontSize="10"
            fill="#94a3b8"
          >
            {d.date.slice(5)}
          </text>
        ) : null
      )}

      {/* Area fills */}
      {series.map((s) => (
        <path
          key={`area-${s.key}`}
          d={buildArea(s.key)}
          fill={s.fillColor}
          opacity="0.15"
        />
      ))}

      {/* Lines */}
      {series.map((s) => (
        <path
          key={`line-${s.key}`}
          d={buildPath(s.key)}
          fill="none"
          stroke={s.color}
          strokeWidth="2"
          strokeLinejoin="round"
          strokeLinecap="round"
        />
      ))}

      {/* Dots on last point */}
      {series.map((s) => {
        const last = data[data.length - 1] as DailyStatPoint | undefined;
        if (!last) return null;
        return (
          <circle
            key={`dot-${s.key}`}
            cx={toX(data.length - 1)}
            cy={toY(Number(last[s.key]) || 0)}
            r="3"
            fill={s.color}
          />
        );
      })}
    </svg>
  );
}

// ==================== Distribution Bar ====================

function DistBar({
  label,
  value,
  total,
  color,
}: {
  label: string;
  value: number;
  total: number;
  color: string;
}) {
  const pct = total > 0 ? Math.round((value / total) * 100) : 0;
  return (
    <div className="mb-3">
      <div className="mb-1 flex justify-between text-sm">
        <span className="font-medium capitalize text-slate-700">{label}</span>
        <span className="text-slate-500">
          {value.toLocaleString()} ({pct}%)
        </span>
      </div>
      <div className="h-2 overflow-hidden rounded-full bg-slate-100">
        <div
          className={`h-full rounded-full transition-all ${color}`}
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  );
}

// ==================== Model Usage Table ====================

function ModelTable({ data }: { data: AnalyticsOverviewData["modelUsage"] }) {
  if (!data?.length) {
    return (
      <p className="py-8 text-center text-sm text-slate-400">暂无模型使用数据</p>
    );
  }

  const maxTokens = Math.max(...data.map((m) => m.tokenCount), 1);

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-slate-100 text-left text-xs font-medium uppercase tracking-wide text-slate-400">
            <th className="pb-2 pr-4">模型</th>
            <th className="pb-2 pr-4 text-right">消息数</th>
            <th className="pb-2 pr-4">Token 用量</th>
            <th className="pb-2 text-right">估算费用</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-50">
          {data.map((row) => (
            <tr key={row.model} className="group hover:bg-slate-50">
              <td className="py-3 pr-4 font-mono text-xs font-semibold text-indigo-600">
                {row.model || "unknown"}
              </td>
              <td className="py-3 pr-4 text-right text-slate-600">
                {row.messageCount.toLocaleString()}
              </td>
              <td className="py-3 pr-4">
                <div className="flex items-center gap-2">
                  <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-slate-100">
                    <div
                      className="h-full rounded-full bg-violet-500"
                      style={{ width: `${(row.tokenCount / maxTokens) * 100}%` }}
                    />
                  </div>
                  <span className="w-16 text-right text-xs text-slate-500">
                    {row.tokenCount >= 1000
                      ? `${(row.tokenCount / 1000).toFixed(1)}k`
                      : row.tokenCount}
                  </span>
                </div>
              </td>
              <td className="py-3 text-right text-slate-600">
                ${row.estimatedCost.toFixed(4)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

// ==================== Main Page ====================

const DAYS_OPTIONS = [7, 14, 30] as const;
type DaysOption = (typeof DAYS_OPTIONS)[number];

const STATUS_COLORS: Record<string, string> = {
  active: "bg-emerald-500",
  waiting_human: "bg-amber-400",
  closed: "bg-slate-400",
};

const CHANNEL_COLORS: Record<string, string> = {
  web: "bg-indigo-500",
  api: "bg-blue-500",
  webhook: "bg-violet-500",
};

export default function AnalyticsPage() {
  const token = useAuthStore((s) => s.token)!;
  const [days, setDays] = useState<DaysOption>(7);
  const [overview, setOverview] = useState<AnalyticsOverviewData | null>(null);
  const [daily, setDaily] = useState<DailyStatPoint[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [exporting, setExporting] = useState(false);

  const handleExportCsv = useCallback(async () => {
    if (exporting) return;
    setExporting(true);
    try {
      await exportAnalyticsCsv(token, days);
    } catch (e) {
      alert(`导出失败: ${String(e)}`);
    } finally {
      setExporting(false);
    }
  }, [token, days, exporting]);

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const [ov, dly] = await Promise.all([
        getAnalyticsOverview(token),
        getAnalyticsDaily(token, days),
      ]);
      if (ov.code === 0) setOverview(ov.data);
      if (dly.code === 0) setDaily(dly.data ?? []);
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  }, [token, days]);

  useEffect(() => {
    load();
  }, [load]);

  const dailySeries: ChartSeries[] = [
    {
      label: "会话数",
      color: "#6366f1",
      fillColor: "#6366f1",
      key: "conversationCount",
    },
    {
      label: "AI 消息数",
      color: "#10b981",
      fillColor: "#10b981",
      key: "messageCount",
    },
  ];

  const tokenSeries: ChartSeries[] = [
    {
      label: "Token 消耗",
      color: "#8b5cf6",
      fillColor: "#8b5cf6",
      key: "tokenCount",
    },
  ];

  const statusTotal = overview?.byStatus?.reduce((s, r) => s + r.count, 0) ?? 0;
  const channelTotal = overview?.byChannel?.reduce((s, r) => s + r.count, 0) ?? 0;

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center bg-slate-50">
        <div className="flex items-center gap-2 text-slate-500">
          <RefreshCw className="h-5 w-5 animate-spin" />
          <span>加载数据中…</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex h-full items-center justify-center bg-slate-50">
        <div className="text-center">
          <p className="text-sm text-red-500">{error}</p>
          <button
            onClick={load}
            className="mt-3 rounded-lg bg-indigo-600 px-4 py-2 text-sm text-white hover:bg-indigo-700"
          >
            重试
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-full flex-col overflow-auto bg-slate-50 print:overflow-visible print:bg-white">
      {/* 打印专用文件头 */}
      <div className="hidden print:block px-6 py-4 border-b border-slate-200">
        <h1 className="text-xl font-bold text-slate-800">AIPivot SLA 报表</h1>
        <p className="text-sm text-slate-500 mt-1">
          统计范围：最近 {days} 天 &nbsp;·&nbsp; 生成时间：{new Date().toLocaleString("zh-CN")}
        </p>
      </div>

      {/* Header（打印时隐藏操作区） */}
      <div className="flex items-center justify-between border-b border-slate-200 bg-white px-6 py-4 print:hidden">
        <h1 className="text-lg font-semibold text-slate-800">对话分析</h1>
        <div className="flex items-center gap-3">
          <div className="flex overflow-hidden rounded-lg border border-slate-200">
            {DAYS_OPTIONS.map((d) => (
              <button
                key={d}
                onClick={() => setDays(d)}
                className={`px-3 py-1.5 text-sm font-medium transition ${
                  days === d
                    ? "bg-indigo-600 text-white"
                    : "bg-white text-slate-600 hover:bg-slate-50"
                }`}
              >
                {d}天
              </button>
            ))}
          </div>
          <button
            onClick={handleExportCsv}
            disabled={exporting}
            className="flex items-center gap-1.5 rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-sm text-slate-600 hover:bg-slate-50 disabled:opacity-50"
          >
            {exporting ? (
              <RefreshCw className="h-3.5 w-3.5 animate-spin" />
            ) : (
              <Download className="h-3.5 w-3.5" />
            )}
            导出 CSV
          </button>
          <button
            onClick={() => window.print()}
            className="flex items-center gap-1.5 rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-sm text-slate-600 hover:bg-slate-50"
          >
            <Printer className="h-3.5 w-3.5" />
            打印 PDF
          </button>
          <button
            onClick={load}
            className="flex items-center gap-1.5 rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-sm text-slate-600 hover:bg-slate-50"
          >
            <RefreshCw className="h-3.5 w-3.5" />
            刷新
          </button>
        </div>
      </div>

      <div className="flex-1 p-6 space-y-6">
        {/* KPI Cards */}
        <div className="grid grid-cols-2 gap-4 xl:grid-cols-3">
          <KpiCard
            label="会话总数"
            value={(overview?.totalConversations ?? 0).toLocaleString()}
            sub={`活跃 ${overview?.activeConversations ?? 0} 个`}
            icon={<Users className="h-5 w-5 text-indigo-600" />}
            color="bg-indigo-50"
          />
          <KpiCard
            label="消息总数"
            value={(overview?.totalMessages ?? 0).toLocaleString()}
            icon={<MessageSquare className="h-5 w-5 text-emerald-600" />}
            color="bg-emerald-50"
          />
          <KpiCard
            label="Token 消耗"
            value={
              (overview?.totalTokens ?? 0) >= 1000
                ? `${((overview?.totalTokens ?? 0) / 1000).toFixed(1)}k`
                : String(overview?.totalTokens ?? 0)
            }
            icon={<Zap className="h-5 w-5 text-violet-600" />}
            color="bg-violet-50"
          />
          <KpiCard
            label="估算费用"
            value={`$${(overview?.estimatedCost ?? 0).toFixed(4)}`}
            sub="混合口径估算"
            icon={<DollarSign className="h-5 w-5 text-amber-600" />}
            color="bg-amber-50"
          />
          <KpiCard
            label="AI 解决率"
            value={`${overview?.aiResolveRate ?? 0}%`}
            sub="已关闭 / 会话总数"
            icon={<CheckCircle className="h-5 w-5 text-teal-600" />}
            color="bg-teal-50"
          />
          <KpiCard
            label="转人工率"
            value={`${overview?.escalationRate ?? 0}%`}
            sub="转人工 / 会话总数"
            icon={<AlertTriangle className="h-5 w-5 text-rose-600" />}
            color="bg-rose-50"
          />
          <KpiCard
            label="满意度"
            value={`${(overview?.satisfactionRate ?? 0).toFixed(1)}%`}
            sub={`已评分 ${overview?.ratedCount ?? 0} 条`}
            icon={<ThumbsUp className="h-5 w-5 text-sky-600" />}
            color="bg-sky-50"
          />
        </div>

        {/* Daily Trend Charts */}
        <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
          <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-sm font-semibold text-slate-700">会话 & 消息趋势</h2>
              <div className="flex gap-4">
                {dailySeries.map((s) => (
                  <span key={s.key} className="flex items-center gap-1 text-xs text-slate-500">
                    <span
                      className="inline-block h-2 w-3 rounded-full"
                      style={{ backgroundColor: s.color }}
                    />
                    {s.label}
                  </span>
                ))}
              </div>
            </div>
            <AreaChart data={daily} series={dailySeries} height={180} />
          </div>

          <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
            <h2 className="mb-4 text-sm font-semibold text-slate-700">Token 消耗趋势</h2>
            <AreaChart data={daily} series={tokenSeries} height={180} />
          </div>
        </div>

        {/* Bottom: Model Usage + Distribution */}
        <div className="grid grid-cols-1 gap-4 xl:grid-cols-3">
          {/* Model Usage Table */}
          <div className="xl:col-span-2 rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
            <h2 className="mb-4 text-sm font-semibold text-slate-700">模型用量明细</h2>
            <ModelTable data={overview?.modelUsage ?? []} />
          </div>

          {/* Status + Channel Distribution */}
          <div className="flex flex-col gap-4">
            <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
              <h2 className="mb-4 text-sm font-semibold text-slate-700">会话状态分布</h2>
              {overview?.byStatus?.length ? (
                overview.byStatus.map((r) => (
                  <DistBar
                    key={r.status}
                    label={
                      r.status === "active"
                        ? "进行中"
                        : r.status === "waiting_human"
                        ? "转人工"
                        : "已关闭"
                    }
                    value={r.count}
                    total={statusTotal}
                    color={STATUS_COLORS[r.status] ?? "bg-slate-400"}
                  />
                ))
              ) : (
                <p className="py-4 text-center text-sm text-slate-400">暂无数据</p>
              )}
            </div>

            <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
              <h2 className="mb-4 text-sm font-semibold text-slate-700">渠道来源分布</h2>
              {overview?.byChannel?.length ? (
                overview.byChannel.map((r) => (
                  <DistBar
                    key={r.channel}
                    label={r.channel}
                    value={r.count}
                    total={channelTotal}
                    color={CHANNEL_COLORS[r.channel] ?? "bg-slate-400"}
                  />
                ))
              ) : (
                <p className="py-4 text-center text-sm text-slate-400">暂无数据</p>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
