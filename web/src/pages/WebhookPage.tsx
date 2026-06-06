import { useState, useEffect, useCallback } from "react";
import {
  Webhook,
  Plus,
  Pencil,
  Trash2,
  Copy,
  Check,
  RefreshCw,
  Globe,
  Zap,
  AlertCircle,
  ChevronDown,
  ChevronUp,
  ToggleLeft,
  ToggleRight,
  X,
  ExternalLink,
  Clock,
} from "lucide-react";
import {
  listWebhooks,
  createWebhook,
  updateWebhook,
  deleteWebhook,
  ShowWebhook,
} from "../lib/api";
import { useAuthStore } from "../store/auth";

// ——— 常量 ———

const ALL_EVENTS = [
  { value: "conversation.created", label: "会话创建" },
  { value: "conversation.closed", label: "会话关闭" },
  { value: "conversation.escalated", label: "转人工" },
  { value: "message.created", label: "新消息" },
  { value: "message.ai_replied", label: "AI 回复" },
];

const CHANNEL_TYPES = [
  { value: "webhook", label: "通用 Webhook" },
  { value: "wechat", label: "微信" },
  { value: "feishu", label: "飞书" },
];

// ——— 辅助函数 ———

function formatTime(ms?: number): string {
  if (!ms) return "—";
  const d = new Date(ms);
  return d.toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function statusBadge(status: string) {
  if (status === "active") {
    return (
      <span className="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-700 ring-1 ring-inset ring-emerald-600/20">
        <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
        启用
      </span>
    );
  }
  return (
    <span className="inline-flex items-center gap-1 rounded-full bg-slate-100 px-2 py-0.5 text-xs font-medium text-slate-500 ring-1 ring-inset ring-slate-200">
      <span className="h-1.5 w-1.5 rounded-full bg-slate-400" />
      停用
    </span>
  );
}

function channelLabel(type: string): string {
  return CHANNEL_TYPES.find((c) => c.value === type)?.label ?? type;
}

// ——— 表单初始值 ———

interface WebhookForm {
  name: string;
  url: string;
  secret: string;
  events: string[];
  channelType: string;
  retryCount: number;
  timeoutMs: number;
}

const defaultForm = (): WebhookForm => ({
  name: "",
  url: "",
  secret: "",
  events: ["message.created"],
  channelType: "webhook",
  retryCount: 3,
  timeoutMs: 5000,
});

// ——— 复制按钮 ———

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);
  const copy = () => {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };
  return (
    <button
      onClick={copy}
      className="ml-1 rounded p-1 text-slate-400 transition hover:bg-slate-100 hover:text-slate-700"
      title="复制"
    >
      {copied ? (
        <Check className="h-3.5 w-3.5 text-emerald-500" />
      ) : (
        <Copy className="h-3.5 w-3.5" />
      )}
    </button>
  );
}

// ——— 展开行（Webhook 详情） ———

function WebhookDetail({
  wh,
  onEdit,
  onDelete,
  onToggle,
}: {
  wh: ShowWebhook;
  onEdit: () => void;
  onDelete: () => void;
  onToggle: () => void;
}) {
  const inboundUrl = `${window.location.origin}/api/v1/open/webhook/${wh.uuid}/inbound`;

  return (
    <div className="border-t border-slate-100 bg-slate-50/70 px-6 py-4">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {/* 入站 URL */}
        <div className="col-span-full">
          <p className="mb-1 text-xs font-medium text-slate-500">入站接收 URL</p>
          <div className="flex items-center gap-1 rounded-md border border-slate-200 bg-white px-3 py-1.5">
            <code className="flex-1 truncate font-mono text-xs text-slate-700">
              {inboundUrl}
            </code>
            <CopyButton text={inboundUrl} />
            <a
              href={inboundUrl}
              target="_blank"
              rel="noreferrer"
              className="ml-0.5 rounded p-1 text-slate-400 hover:text-slate-700"
            >
              <ExternalLink className="h-3.5 w-3.5" />
            </a>
          </div>
        </div>

        {/* 订阅事件 */}
        <div>
          <p className="mb-1.5 text-xs font-medium text-slate-500">订阅事件</p>
          <div className="flex flex-wrap gap-1.5">
            {wh.events.length > 0 ? (
              wh.events.map((e) => (
                <span
                  key={e}
                  className="rounded-full bg-indigo-50 px-2 py-0.5 text-xs font-medium text-indigo-700 ring-1 ring-inset ring-indigo-200"
                >
                  {ALL_EVENTS.find((ev) => ev.value === e)?.label ?? e}
                </span>
              ))
            ) : (
              <span className="text-xs text-slate-400">未订阅任何事件</span>
            )}
          </div>
        </div>

        {/* 配置参数 */}
        <div>
          <p className="mb-1.5 text-xs font-medium text-slate-500">配置参数</p>
          <div className="space-y-1 text-xs text-slate-600">
            <div className="flex gap-2">
              <span className="text-slate-400">重试次数</span>
              <span>{wh.retryCount} 次</span>
            </div>
            <div className="flex gap-2">
              <span className="text-slate-400">超时时长</span>
              <span>{wh.timeoutMs} ms</span>
            </div>
          </div>
        </div>

        {/* 最近状态 */}
        <div>
          <p className="mb-1.5 text-xs font-medium text-slate-500">最近状态</p>
          <div className="space-y-1 text-xs text-slate-600">
            <div className="flex gap-2">
              <span className="text-slate-400">最近触发</span>
              <span>{formatTime(wh.lastTrigger)}</span>
            </div>
            {wh.lastError && (
              <div className="flex items-start gap-1.5 rounded-md bg-red-50 px-2 py-1.5 text-red-600">
                <AlertCircle className="mt-0.5 h-3 w-3 shrink-0" />
                <span className="break-all">{wh.lastError}</span>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* 操作按钮 */}
      <div className="mt-4 flex gap-2 border-t border-slate-100 pt-3">
        <button
          onClick={onToggle}
          className="flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium text-slate-600 ring-1 ring-slate-200 transition hover:bg-slate-100"
        >
          {wh.status === "active" ? (
            <>
              <ToggleRight className="h-3.5 w-3.5 text-emerald-500" /> 停用
            </>
          ) : (
            <>
              <ToggleLeft className="h-3.5 w-3.5" /> 启用
            </>
          )}
        </button>
        <button
          onClick={onEdit}
          className="flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium text-indigo-600 ring-1 ring-indigo-200 transition hover:bg-indigo-50"
        >
          <Pencil className="h-3.5 w-3.5" /> 编辑
        </button>
        <button
          onClick={onDelete}
          className="flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium text-red-600 ring-1 ring-red-200 transition hover:bg-red-50"
        >
          <Trash2 className="h-3.5 w-3.5" /> 删除
        </button>
      </div>
    </div>
  );
}

// ——— 创建/编辑弹窗 ———

function WebhookModal({
  initial,
  onClose,
  onSave,
}: {
  initial?: ShowWebhook;
  onClose: () => void;
  onSave: (form: WebhookForm) => Promise<void>;
}) {
  const [form, setForm] = useState<WebhookForm>(() => {
    if (initial) {
      return {
        name: initial.name,
        url: initial.url,
        secret: "",
        events: initial.events,
        channelType: initial.channelType,
        retryCount: initial.retryCount,
        timeoutMs: initial.timeoutMs,
      };
    }
    return defaultForm();
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const toggleEvent = (ev: string) => {
    setForm((f) => ({
      ...f,
      events: f.events.includes(ev)
        ? f.events.filter((e) => e !== ev)
        : [...f.events, ev],
    }));
  };

  const submit = async () => {
    if (!form.name.trim()) return setError("请填写名称");
    if (!form.url.trim()) return setError("请填写回调 URL");
    setError("");
    setSaving(true);
    try {
      await onSave(form);
      onClose();
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "保存失败");
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm">
      <div className="w-full max-w-lg overflow-hidden rounded-2xl bg-white shadow-2xl">
        {/* 头部 */}
        <div className="flex items-center justify-between border-b border-slate-100 px-6 py-4">
          <h2 className="text-base font-semibold text-slate-800">
            {initial ? "编辑 Webhook" : "新建 Webhook"}
          </h2>
          <button
            onClick={onClose}
            className="rounded-lg p-1.5 text-slate-400 hover:bg-slate-100 hover:text-slate-600"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        {/* 表单体 */}
        <div className="max-h-[60vh] overflow-y-auto px-6 py-5">
          <div className="space-y-4">
            {/* 名称 */}
            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-700">
                名称 <span className="text-red-500">*</span>
              </label>
              <input
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                placeholder="如：生产环境推送"
                className="w-full rounded-lg border border-slate-200 px-3 py-2 text-sm text-slate-800 placeholder-slate-400 outline-none transition focus:border-indigo-400 focus:ring-2 focus:ring-indigo-100"
              />
            </div>

            {/* 回调 URL */}
            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-700">
                回调 URL <span className="text-red-500">*</span>
              </label>
              <input
                value={form.url}
                onChange={(e) => setForm({ ...form, url: e.target.value })}
                placeholder="https://your-server.com/webhook"
                className="w-full rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm text-slate-800 placeholder-slate-400 outline-none transition focus:border-indigo-400 focus:ring-2 focus:ring-indigo-100"
              />
            </div>

            {/* 签名密钥 */}
            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-700">
                签名密钥
                <span className="ml-1.5 text-xs font-normal text-slate-400">
                  （HMAC-SHA256，留空不签名）
                </span>
              </label>
              <input
                value={form.secret}
                onChange={(e) => setForm({ ...form, secret: e.target.value })}
                placeholder="留空或输入密钥字符串"
                type="password"
                className="w-full rounded-lg border border-slate-200 px-3 py-2 text-sm text-slate-800 placeholder-slate-400 outline-none transition focus:border-indigo-400 focus:ring-2 focus:ring-indigo-100"
              />
            </div>

            {/* 渠道类型 */}
            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-700">
                渠道类型
              </label>
              <div className="flex flex-wrap gap-2">
                {CHANNEL_TYPES.map((ct) => (
                  <button
                    key={ct.value}
                    type="button"
                    onClick={() => setForm({ ...form, channelType: ct.value })}
                    className={`rounded-lg px-3 py-1.5 text-sm font-medium transition ${
                      form.channelType === ct.value
                        ? "bg-indigo-600 text-white"
                        : "bg-slate-100 text-slate-600 hover:bg-slate-200"
                    }`}
                  >
                    {ct.label}
                  </button>
                ))}
              </div>
            </div>

            {/* 订阅事件 */}
            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-700">
                订阅事件
              </label>
              <div className="rounded-lg border border-slate-200 divide-y divide-slate-100">
                {ALL_EVENTS.map((ev) => (
                  <label
                    key={ev.value}
                    className="flex cursor-pointer items-center gap-3 px-3 py-2.5 hover:bg-slate-50"
                  >
                    <input
                      type="checkbox"
                      checked={form.events.includes(ev.value)}
                      onChange={() => toggleEvent(ev.value)}
                      className="h-4 w-4 rounded border-slate-300 text-indigo-600 focus:ring-indigo-500"
                    />
                    <span className="text-sm text-slate-700">{ev.label}</span>
                    <code className="ml-auto text-xs text-slate-400">
                      {ev.value}
                    </code>
                  </label>
                ))}
              </div>
            </div>

            {/* 高级参数 */}
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="mb-1.5 block text-sm font-medium text-slate-700">
                  重试次数
                </label>
                <input
                  type="number"
                  min={0}
                  max={10}
                  value={form.retryCount}
                  onChange={(e) =>
                    setForm({ ...form, retryCount: Number(e.target.value) })
                  }
                  className="w-full rounded-lg border border-slate-200 px-3 py-2 text-sm text-slate-800 outline-none focus:border-indigo-400 focus:ring-2 focus:ring-indigo-100"
                />
              </div>
              <div>
                <label className="mb-1.5 block text-sm font-medium text-slate-700">
                  超时 (ms)
                </label>
                <input
                  type="number"
                  min={1000}
                  step={1000}
                  value={form.timeoutMs}
                  onChange={(e) =>
                    setForm({ ...form, timeoutMs: Number(e.target.value) })
                  }
                  className="w-full rounded-lg border border-slate-200 px-3 py-2 text-sm text-slate-800 outline-none focus:border-indigo-400 focus:ring-2 focus:ring-indigo-100"
                />
              </div>
            </div>
          </div>
        </div>

        {/* 底部 */}
        <div className="flex items-center justify-between border-t border-slate-100 px-6 py-4">
          {error ? (
            <p className="flex items-center gap-1.5 text-sm text-red-500">
              <AlertCircle className="h-4 w-4" />
              {error}
            </p>
          ) : (
            <span />
          )}
          <div className="flex gap-2">
            <button
              onClick={onClose}
              className="rounded-lg px-4 py-2 text-sm font-medium text-slate-600 ring-1 ring-slate-200 transition hover:bg-slate-50"
            >
              取消
            </button>
            <button
              onClick={submit}
              disabled={saving}
              className="flex items-center gap-1.5 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-indigo-700 disabled:opacity-60"
            >
              {saving && <RefreshCw className="h-3.5 w-3.5 animate-spin" />}
              {initial ? "保存" : "创建"}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

// ——— 删除确认弹窗 ———

function DeleteConfirm({
  name,
  onClose,
  onConfirm,
}: {
  name: string;
  onClose: () => void;
  onConfirm: () => Promise<void>;
}) {
  const [deleting, setDeleting] = useState(false);
  const confirm = async () => {
    setDeleting(true);
    try {
      await onConfirm();
      onClose();
    } finally {
      setDeleting(false);
    }
  };
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm">
      <div className="w-full max-w-sm rounded-2xl bg-white p-6 shadow-2xl">
        <div className="mb-4 flex h-11 w-11 items-center justify-center rounded-full bg-red-100">
          <Trash2 className="h-5 w-5 text-red-600" />
        </div>
        <h3 className="mb-1 text-base font-semibold text-slate-800">
          删除 Webhook
        </h3>
        <p className="mb-5 text-sm text-slate-500">
          确定要删除 <span className="font-medium text-slate-700">"{name}"</span>{" "}
          吗？此操作不可撤销。
        </p>
        <div className="flex gap-3">
          <button
            onClick={onClose}
            className="flex-1 rounded-lg py-2 text-sm font-medium text-slate-600 ring-1 ring-slate-200 transition hover:bg-slate-50"
          >
            取消
          </button>
          <button
            onClick={confirm}
            disabled={deleting}
            className="flex flex-1 items-center justify-center gap-1.5 rounded-lg bg-red-600 py-2 text-sm font-medium text-white transition hover:bg-red-700 disabled:opacity-60"
          >
            {deleting && <RefreshCw className="h-3.5 w-3.5 animate-spin" />}
            删除
          </button>
        </div>
      </div>
    </div>
  );
}

// ——— 主页面 ———

export default function WebhookPage() {
  const token = useAuthStore((s) => s.token)!;
  const [webhooks, setWebhooks] = useState<ShowWebhook[]>([]);
  const [loading, setLoading] = useState(true);
  const [expandedId, setExpandedId] = useState<number | null>(null);
  const [modalMode, setModalMode] = useState<"create" | "edit" | null>(null);
  const [editTarget, setEditTarget] = useState<ShowWebhook | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<ShowWebhook | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await listWebhooks(token);
      if (res.code === 0) setWebhooks(res.data ?? []);
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    load();
  }, [load]);

  const handleCreate = async (form: WebhookForm) => {
    await createWebhook(token, {
      name: form.name,
      url: form.url,
      secret: form.secret || undefined,
      events: form.events,
      channelType: form.channelType,
      retryCount: form.retryCount,
      timeoutMs: form.timeoutMs,
    });
    await load();
  };

  const handleEdit = async (form: WebhookForm) => {
    if (!editTarget) return;
    await updateWebhook(token, editTarget.id, {
      name: form.name,
      url: form.url,
      secret: form.secret || undefined,
      events: form.events,
      retryCount: form.retryCount,
      timeoutMs: form.timeoutMs,
    });
    await load();
  };

  const handleToggle = async (wh: ShowWebhook) => {
    await updateWebhook(token, wh.id, {
      status: wh.status === "active" ? "disabled" : "active",
    });
    await load();
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    await deleteWebhook(token, deleteTarget.id);
    await load();
  };

  return (
    <div className="flex h-full flex-col bg-slate-50">
      {/* 页头 */}
      <div className="flex items-center justify-between border-b border-slate-200 bg-white px-6 py-4">
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-violet-100">
            <Webhook className="h-5 w-5 text-violet-600" />
          </div>
          <div>
            <h1 className="text-base font-semibold text-slate-800">
              Webhook 管理
            </h1>
            <p className="text-xs text-slate-500">
              配置出站事件推送和第三方平台集成
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={load}
            className="flex items-center gap-1.5 rounded-lg px-3 py-2 text-sm text-slate-500 ring-1 ring-slate-200 transition hover:bg-slate-50 hover:text-slate-700"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
            刷新
          </button>
          <button
            onClick={() => {
              setEditTarget(null);
              setModalMode("create");
            }}
            className="flex items-center gap-1.5 rounded-lg bg-indigo-600 px-3 py-2 text-sm font-medium text-white transition hover:bg-indigo-700"
          >
            <Plus className="h-4 w-4" />
            新建 Webhook
          </button>
        </div>
      </div>

      {/* 内容 */}
      <div className="flex-1 overflow-y-auto p-6">
        {loading && webhooks.length === 0 ? (
          <div className="flex h-40 items-center justify-center">
            <RefreshCw className="h-6 w-6 animate-spin text-slate-400" />
          </div>
        ) : webhooks.length === 0 ? (
          /* 空状态 */
          <div className="flex flex-col items-center justify-center rounded-2xl border border-dashed border-slate-300 bg-white py-20 text-center">
            <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-violet-50">
              <Webhook className="h-8 w-8 text-violet-400" />
            </div>
            <p className="mb-1 text-base font-medium text-slate-700">
              还没有 Webhook
            </p>
            <p className="mb-6 text-sm text-slate-400">
              创建 Webhook 以接收实时事件推送，或集成第三方平台
            </p>
            <button
              onClick={() => {
                setEditTarget(null);
                setModalMode("create");
              }}
              className="flex items-center gap-1.5 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-indigo-700"
            >
              <Plus className="h-4 w-4" />
              新建 Webhook
            </button>
          </div>
        ) : (
          /* Webhook 列表 */
          <div className="space-y-3">
            {/* 统计条 */}
            <div className="flex items-center gap-6 rounded-xl border border-slate-200 bg-white px-5 py-3">
              <Stat
                icon={<Zap className="h-4 w-4 text-indigo-500" />}
                label="总计"
                value={webhooks.length}
              />
              <Stat
                icon={
                  <span className="h-2 w-2 rounded-full bg-emerald-500 inline-block" />
                }
                label="启用"
                value={webhooks.filter((w) => w.status === "active").length}
              />
              <Stat
                icon={
                  <span className="h-2 w-2 rounded-full bg-slate-400 inline-block" />
                }
                label="停用"
                value={webhooks.filter((w) => w.status !== "active").length}
              />
              <Stat
                icon={<Globe className="h-4 w-4 text-violet-500" />}
                label="渠道类型"
                value={new Set(webhooks.map((w) => w.channelType)).size}
              />
            </div>

            {/* 列表表头 */}
            <div className="grid grid-cols-[2fr_3fr_1fr_1fr_auto] items-center gap-4 rounded-xl border border-slate-200 bg-white px-5 py-2.5 text-xs font-medium text-slate-500">
              <span>名称 / 渠道</span>
              <span>回调 URL</span>
              <span>状态</span>
              <span>最近触发</span>
              <span className="w-6" />
            </div>

            {/* 列表行 */}
            {webhooks.map((wh) => (
              <div
                key={wh.id}
                className="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm transition hover:shadow-md"
              >
                {/* 主行 */}
                <button
                  className="grid w-full grid-cols-[2fr_3fr_1fr_1fr_auto] items-center gap-4 px-5 py-3.5 text-left"
                  onClick={() =>
                    setExpandedId(expandedId === wh.id ? null : wh.id)
                  }
                >
                  {/* 名称 + 渠道 */}
                  <div>
                    <p className="font-medium text-slate-800">{wh.name}</p>
                    <p className="mt-0.5 text-xs text-slate-400">
                      {channelLabel(wh.channelType)}
                    </p>
                  </div>

                  {/* URL */}
                  <div className="flex items-center gap-1 min-w-0">
                    <Globe className="h-3.5 w-3.5 shrink-0 text-slate-400" />
                    <span className="truncate font-mono text-xs text-slate-600">
                      {wh.url}
                    </span>
                  </div>

                  {/* 状态 */}
                  <div>{statusBadge(wh.status)}</div>

                  {/* 最近触发 */}
                  <div className="flex items-center gap-1 text-xs text-slate-500">
                    <Clock className="h-3 w-3" />
                    {formatTime(wh.lastTrigger)}
                  </div>

                  {/* 展开箭头 */}
                  <div className="text-slate-400">
                    {expandedId === wh.id ? (
                      <ChevronUp className="h-4 w-4" />
                    ) : (
                      <ChevronDown className="h-4 w-4" />
                    )}
                  </div>
                </button>

                {/* 展开详情 */}
                {expandedId === wh.id && (
                  <WebhookDetail
                    wh={wh}
                    onEdit={() => {
                      setEditTarget(wh);
                      setModalMode("edit");
                    }}
                    onDelete={() => setDeleteTarget(wh)}
                    onToggle={() => handleToggle(wh)}
                  />
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* 弹窗 */}
      {modalMode === "create" && (
        <WebhookModal onClose={() => setModalMode(null)} onSave={handleCreate} />
      )}
      {modalMode === "edit" && editTarget && (
        <WebhookModal
          initial={editTarget}
          onClose={() => {
            setModalMode(null);
            setEditTarget(null);
          }}
          onSave={handleEdit}
        />
      )}
      {deleteTarget && (
        <DeleteConfirm
          name={deleteTarget.name}
          onClose={() => setDeleteTarget(null)}
          onConfirm={handleDelete}
        />
      )}
    </div>
  );
}

// ——— 统计卡片 ———
function Stat({
  icon,
  label,
  value,
}: {
  icon: React.ReactNode;
  label: string;
  value: number;
}) {
  return (
    <div className="flex items-center gap-2">
      {icon}
      <span className="text-xs text-slate-500">{label}</span>
      <span className="text-sm font-semibold text-slate-800">{value}</span>
    </div>
  );
}
