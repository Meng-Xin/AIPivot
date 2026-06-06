import { useCallback, useEffect, useMemo, useState } from "react";
import {
  AlertCircle,
  Check,
  Clock,
  Code2,
  Copy,
  Globe,
  Pencil,
  Plus,
  RefreshCw,
  Trash2,
  Wrench,
  X,
  ToggleLeft,
  ToggleRight,
} from "lucide-react";
import {
  createSkill,
  deleteSkill,
  listSkills,
  ShowSkill,
  updateSkill,
} from "../lib/api";
import { useAuthStore } from "../store/auth";

type SkillForm = {
  name: string;
  description: string;
  parameters: string;
  endpoint: string;
  method: string;
  headers: string;
  timeoutMs: number;
  enabled: boolean;
};

const DEFAULT_SCHEMA = JSON.stringify(
  {
    type: "object",
    properties: {
      query: { type: "string", description: "查询内容" },
    },
    required: ["query"],
  },
  null,
  2
);

const DEFAULT_HEADERS = JSON.stringify({}, null, 2);

function defaultForm(): SkillForm {
  return {
    name: "",
    description: "",
    parameters: DEFAULT_SCHEMA,
    endpoint: "",
    method: "POST",
    headers: DEFAULT_HEADERS,
    timeoutMs: 5000,
    enabled: true,
  };
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

function prettyJson(value: string) {
  try {
    return JSON.stringify(JSON.parse(value || "{}"), null, 2);
  } catch {
    return value;
  }
}

function validateJson(value: string, label: string) {
  try {
    JSON.parse(value || "{}");
    return "";
  } catch {
    return `${label} 不是有效 JSON`;
  }
}

function enabledBadge(enabled: boolean) {
  return enabled ? (
    <span className="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-700 ring-1 ring-inset ring-emerald-600/20">
      <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
      启用
    </span>
  ) : (
    <span className="inline-flex items-center gap-1 rounded-full bg-slate-100 px-2 py-0.5 text-xs font-medium text-slate-500 ring-1 ring-inset ring-slate-200">
      <span className="h-1.5 w-1.5 rounded-full bg-slate-400" />
      停用
    </span>
  );
}

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);

  return (
    <button
      type="button"
      title="复制"
      onClick={() => {
        navigator.clipboard.writeText(text).then(() => {
          setCopied(true);
          window.setTimeout(() => setCopied(false), 1600);
        });
      }}
      className="rounded p-1 text-slate-400 transition hover:bg-slate-100 hover:text-slate-700"
    >
      {copied ? <Check className="h-3.5 w-3.5 text-emerald-500" /> : <Copy className="h-3.5 w-3.5" />}
    </button>
  );
}

function SkillModal({
  initial,
  onClose,
  onSubmit,
}: {
  initial: SkillForm;
  onClose: () => void;
  onSubmit: (form: SkillForm) => Promise<void>;
}) {
  const [form, setForm] = useState<SkillForm>(initial);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const set = <K extends keyof SkillForm>(key: K, value: SkillForm[K]) =>
    setForm((prev) => ({ ...prev, [key]: value }));

  const submit = async () => {
    const schemaError = validateJson(form.parameters, "参数 Schema");
    const headersError = validateJson(form.headers, "请求头");
    if (schemaError || headersError) {
      setError(schemaError || headersError);
      return;
    }
    if (!form.name.trim() || !form.endpoint.trim()) {
      setError("名称和 Endpoint 不能为空");
      return;
    }

    setSaving(true);
    setError("");
    try {
      await onSubmit({
        ...form,
        name: form.name.trim(),
        endpoint: form.endpoint.trim(),
        parameters: prettyJson(form.parameters),
        headers: prettyJson(form.headers || "{}"),
      });
    } catch (e) {
      setError(e instanceof Error ? e.message : "保存失败");
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/45 px-4">
      <div className="flex max-h-[92vh] w-full max-w-3xl flex-col rounded-lg bg-white shadow-xl">
        <div className="flex items-center justify-between border-b border-slate-200 px-5 py-4">
          <div className="flex items-center gap-2">
            <Wrench className="h-5 w-5 text-indigo-600" />
            <h2 className="text-base font-semibold text-slate-900">自定义 Skill</h2>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="rounded p-1 text-slate-400 hover:bg-slate-100 hover:text-slate-700"
            title="关闭"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        <div className="flex-1 overflow-y-auto px-5 py-4">
          {error && (
            <div className="mb-4 flex items-start gap-2 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
              <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
              <span>{error}</span>
            </div>
          )}

          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <label className="block">
              <span className="mb-1 block text-sm font-medium text-slate-700">名称</span>
              <input
                value={form.name}
                onChange={(e) => set("name", e.target.value)}
                className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm outline-none transition focus:border-indigo-500 focus:ring-2 focus:ring-indigo-100"
                placeholder="query_order"
              />
            </label>

            <label className="block">
              <span className="mb-1 block text-sm font-medium text-slate-700">HTTP 方法</span>
              <select
                value={form.method}
                onChange={(e) => set("method", e.target.value)}
                className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm outline-none transition focus:border-indigo-500 focus:ring-2 focus:ring-indigo-100"
              >
                <option value="POST">POST</option>
                <option value="GET">GET</option>
              </select>
            </label>

            <label className="block md:col-span-2">
              <span className="mb-1 block text-sm font-medium text-slate-700">Endpoint</span>
              <input
                value={form.endpoint}
                onChange={(e) => set("endpoint", e.target.value)}
                className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm outline-none transition focus:border-indigo-500 focus:ring-2 focus:ring-indigo-100"
                placeholder="https://example.com/tools/query-order"
              />
            </label>

            <label className="block md:col-span-2">
              <span className="mb-1 block text-sm font-medium text-slate-700">描述</span>
              <textarea
                value={form.description}
                onChange={(e) => set("description", e.target.value)}
                rows={3}
                className="w-full resize-none rounded-md border border-slate-300 px-3 py-2 text-sm outline-none transition focus:border-indigo-500 focus:ring-2 focus:ring-indigo-100"
                placeholder="根据订单号查询订单状态"
              />
            </label>

            <label className="block">
              <span className="mb-1 block text-sm font-medium text-slate-700">超时毫秒</span>
              <input
                type="number"
                min={1000}
                max={30000}
                value={form.timeoutMs}
                onChange={(e) => set("timeoutMs", Number(e.target.value) || 5000)}
                className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm outline-none transition focus:border-indigo-500 focus:ring-2 focus:ring-indigo-100"
              />
            </label>

            <label className="flex items-end">
              <button
                type="button"
                onClick={() => set("enabled", !form.enabled)}
                className="inline-flex h-10 items-center gap-2 rounded-md border border-slate-300 px-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50"
              >
                {form.enabled ? <ToggleRight className="h-5 w-5 text-emerald-600" /> : <ToggleLeft className="h-5 w-5 text-slate-400" />}
                {form.enabled ? "启用" : "停用"}
              </button>
            </label>

            <label className="block md:col-span-2">
              <span className="mb-1 block text-sm font-medium text-slate-700">参数 Schema</span>
              <textarea
                value={form.parameters}
                onChange={(e) => set("parameters", e.target.value)}
                onBlur={() => set("parameters", prettyJson(form.parameters))}
                rows={10}
                spellCheck={false}
                className="w-full resize-y rounded-md border border-slate-300 bg-slate-950 px-3 py-2 font-mono text-xs leading-5 text-slate-100 outline-none transition focus:border-indigo-500 focus:ring-2 focus:ring-indigo-100"
              />
            </label>

            <label className="block md:col-span-2">
              <span className="mb-1 block text-sm font-medium text-slate-700">请求头</span>
              <textarea
                value={form.headers}
                onChange={(e) => set("headers", e.target.value)}
                onBlur={() => set("headers", prettyJson(form.headers || "{}"))}
                rows={5}
                spellCheck={false}
                className="w-full resize-y rounded-md border border-slate-300 bg-slate-950 px-3 py-2 font-mono text-xs leading-5 text-slate-100 outline-none transition focus:border-indigo-500 focus:ring-2 focus:ring-indigo-100"
              />
            </label>
          </div>
        </div>

        <div className="flex justify-end gap-2 border-t border-slate-200 px-5 py-4">
          <button
            type="button"
            onClick={onClose}
            className="rounded-md px-4 py-2 text-sm font-medium text-slate-600 hover:bg-slate-100"
          >
            取消
          </button>
          <button
            type="button"
            onClick={submit}
            disabled={saving}
            className="inline-flex items-center gap-2 rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-indigo-700 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {saving && <RefreshCw className="h-4 w-4 animate-spin" />}
            保存
          </button>
        </div>
      </div>
    </div>
  );
}

export default function SkillPage() {
  const { token } = useAuthStore();
  const [skills, setSkills] = useState<ShowSkill[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [editing, setEditing] = useState<ShowSkill | null>(null);
  const [creating, setCreating] = useState(false);
  const [deleting, setDeleting] = useState<ShowSkill | null>(null);

  const refresh = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    setError("");
    try {
      const resp = await listSkills(token);
      setSkills(resp.data ?? []);
    } catch (e) {
      setError(e instanceof Error ? e.message : "加载失败");
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const stats = useMemo(() => {
    const enabled = skills.filter((s) => s.enabled).length;
    const post = skills.filter((s) => s.method === "POST").length;
    return { total: skills.length, enabled, disabled: skills.length - enabled, post };
  }, [skills]);

  const toForm = (skill: ShowSkill): SkillForm => ({
    name: skill.name,
    description: skill.description,
    parameters: prettyJson(skill.parameters || "{}"),
    endpoint: skill.endpoint,
    method: skill.method || "POST",
    headers: prettyJson(skill.headers || "{}"),
    timeoutMs: skill.timeoutMs || 5000,
    enabled: skill.enabled,
  });

  const submitCreate = async (form: SkillForm) => {
    if (!token) return;
    await createSkill(token, form);
    setCreating(false);
    await refresh();
  };

  const submitUpdate = async (form: SkillForm) => {
    if (!token || !editing) return;
    await updateSkill(token, editing.id, form);
    setEditing(null);
    await refresh();
  };

  const toggleSkill = async (skill: ShowSkill) => {
    if (!token) return;
    await updateSkill(token, skill.id, {
      name: skill.name,
      description: skill.description,
      parameters: skill.parameters,
      endpoint: skill.endpoint,
      method: skill.method,
      headers: skill.headers,
      timeoutMs: skill.timeoutMs,
      enabled: !skill.enabled,
    });
    await refresh();
  };

  const confirmDelete = async () => {
    if (!token || !deleting) return;
    await deleteSkill(token, deleting.id);
    setDeleting(null);
    await refresh();
  };

  return (
    <div className="flex h-full flex-col bg-slate-50">
      <header className="border-b border-slate-200 bg-white px-6 py-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-indigo-50 text-indigo-600">
              <Wrench className="h-5 w-5" />
            </div>
            <div>
              <h1 className="text-lg font-semibold text-slate-900">Skill</h1>
              <p className="text-sm text-slate-500">HTTP Function Calling 工具</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={refresh}
              className="inline-flex h-9 items-center gap-2 rounded-md border border-slate-300 px-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50"
            >
              <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
              刷新
            </button>
            <button
              type="button"
              onClick={() => setCreating(true)}
              className="inline-flex h-9 items-center gap-2 rounded-md bg-indigo-600 px-3 text-sm font-medium text-white transition hover:bg-indigo-700"
            >
              <Plus className="h-4 w-4" />
              新建
            </button>
          </div>
        </div>

        <div className="mt-4 grid grid-cols-2 gap-3 md:grid-cols-4">
          {[
            ["总计", stats.total],
            ["启用", stats.enabled],
            ["停用", stats.disabled],
            ["POST", stats.post],
          ].map(([label, value]) => (
            <div key={label} className="rounded-md border border-slate-200 bg-slate-50 px-3 py-2">
              <div className="text-xs text-slate-500">{label}</div>
              <div className="mt-0.5 text-lg font-semibold text-slate-900">{value}</div>
            </div>
          ))}
        </div>
      </header>

      <main className="flex-1 overflow-y-auto px-6 py-5">
        {error && (
          <div className="mb-4 flex items-start gap-2 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
            <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
            <span>{error}</span>
          </div>
        )}

        {skills.length === 0 && !loading ? (
          <div className="flex min-h-[360px] flex-col items-center justify-center rounded-lg border border-dashed border-slate-300 bg-white px-6 text-center">
            <Wrench className="h-10 w-10 text-slate-300" />
            <h2 className="mt-3 text-base font-semibold text-slate-900">暂无 Skill</h2>
            <button
              type="button"
              onClick={() => setCreating(true)}
              className="mt-4 inline-flex items-center gap-2 rounded-md bg-indigo-600 px-3 py-2 text-sm font-medium text-white hover:bg-indigo-700"
            >
              <Plus className="h-4 w-4" />
              新建 Skill
            </button>
          </div>
        ) : (
          <div className="overflow-hidden rounded-lg border border-slate-200 bg-white">
            <div className="grid grid-cols-[minmax(180px,1.1fr)_minmax(220px,1.5fr)_120px_120px_120px] border-b border-slate-200 bg-slate-50 px-4 py-2 text-xs font-medium uppercase tracking-wide text-slate-500">
              <div>名称</div>
              <div>Endpoint</div>
              <div>方法</div>
              <div>状态</div>
              <div className="text-right">操作</div>
            </div>

            {skills.map((skill) => (
              <div key={skill.id} className="border-b border-slate-100 last:border-b-0">
                <div className="grid grid-cols-[minmax(180px,1.1fr)_minmax(220px,1.5fr)_120px_120px_120px] items-center px-4 py-3 text-sm">
                  <div className="min-w-0">
                    <div className="flex items-center gap-2">
                      <Code2 className="h-4 w-4 shrink-0 text-indigo-500" />
                      <span className="truncate font-medium text-slate-900">{skill.name}</span>
                      <CopyButton text={skill.name} />
                    </div>
                    <div className="mt-1 line-clamp-1 text-xs text-slate-500">{skill.description || "-"}</div>
                  </div>

                  <div className="min-w-0">
                    <div className="flex items-center gap-2 text-slate-700">
                      <Globe className="h-4 w-4 shrink-0 text-slate-400" />
                      <span className="truncate">{skill.endpoint}</span>
                    </div>
                    <div className="mt-1 flex items-center gap-1 text-xs text-slate-400">
                      <Clock className="h-3 w-3" />
                      {skill.timeoutMs} ms · {formatTime(skill.updatedAt)}
                    </div>
                  </div>

                  <div>
                    <span className="rounded bg-slate-100 px-2 py-1 font-mono text-xs font-medium text-slate-700">
                      {skill.method}
                    </span>
                  </div>
                  <div>{enabledBadge(skill.enabled)}</div>
                  <div className="flex justify-end gap-1">
                    <button
                      type="button"
                      title={skill.enabled ? "停用" : "启用"}
                      onClick={() => toggleSkill(skill)}
                      className="rounded p-1.5 text-slate-500 hover:bg-slate-100 hover:text-slate-800"
                    >
                      {skill.enabled ? <ToggleRight className="h-4 w-4 text-emerald-600" /> : <ToggleLeft className="h-4 w-4" />}
                    </button>
                    <button
                      type="button"
                      title="编辑"
                      onClick={() => setEditing(skill)}
                      className="rounded p-1.5 text-slate-500 hover:bg-slate-100 hover:text-slate-800"
                    >
                      <Pencil className="h-4 w-4" />
                    </button>
                    <button
                      type="button"
                      title="删除"
                      onClick={() => setDeleting(skill)}
                      className="rounded p-1.5 text-slate-500 hover:bg-red-50 hover:text-red-600"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>
                </div>

                <div className="border-t border-slate-100 bg-slate-50 px-4 py-3">
                  <div className="grid gap-3 lg:grid-cols-2">
                    <div>
                      <div className="mb-1 text-xs font-medium text-slate-500">参数 Schema</div>
                      <pre className="max-h-40 overflow-auto rounded-md bg-slate-950 p-3 text-xs leading-5 text-slate-100">
                        {prettyJson(skill.parameters || "{}")}
                      </pre>
                    </div>
                    <div>
                      <div className="mb-1 text-xs font-medium text-slate-500">请求头</div>
                      <pre className="max-h-40 overflow-auto rounded-md bg-slate-950 p-3 text-xs leading-5 text-slate-100">
                        {prettyJson(skill.headers || "{}")}
                      </pre>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </main>

      {creating && (
        <SkillModal
          initial={defaultForm()}
          onClose={() => setCreating(false)}
          onSubmit={submitCreate}
        />
      )}

      {editing && (
        <SkillModal
          initial={toForm(editing)}
          onClose={() => setEditing(null)}
          onSubmit={submitUpdate}
        />
      )}

      {deleting && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/45 px-4">
          <div className="w-full max-w-sm rounded-lg bg-white p-5 shadow-xl">
            <h2 className="text-base font-semibold text-slate-900">删除 Skill</h2>
            <p className="mt-2 text-sm text-slate-600">
              确认删除 <span className="font-medium text-slate-900">{deleting.name}</span>？
            </p>
            <div className="mt-5 flex justify-end gap-2">
              <button
                type="button"
                onClick={() => setDeleting(null)}
                className="rounded-md px-3 py-2 text-sm font-medium text-slate-600 hover:bg-slate-100"
              >
                取消
              </button>
              <button
                type="button"
                onClick={confirmDelete}
                className="rounded-md bg-red-600 px-3 py-2 text-sm font-medium text-white hover:bg-red-700"
              >
                删除
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
