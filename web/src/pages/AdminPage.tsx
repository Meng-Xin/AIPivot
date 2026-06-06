import { useState, useEffect, useCallback } from "react";
import {
  Building2,
  Users,
  Key,
  RefreshCw,
  Plus,
  Pencil,
  Trash2,
  ShieldCheck,
  ShieldOff,
  Copy,
  Check,
  Eye,
  EyeOff,
  X,
  AlertCircle,
  Crown,
} from "lucide-react";
import {
  getAdminTenant,
  updateAdminTenant,
  listAdminUsers,
  createAdminUser,
  updateAdminUser,
  deleteAdminUser,
  listApiKeys,
  createApiKey,
  revokeApiKey,
  ShowTenant,
  ShowAdminUser,
  ShowApiKey,
  CreateApiKeyData,
} from "../lib/api";
import { useAuthStore } from "../store/auth";

// ——— helpers ———

const formatDate = (ts: number) =>
  ts ? new Date(ts * 1000).toLocaleDateString("zh-CN") : "—";

const PLAN_LABELS: Record<string, string> = {
  free: "免费版",
  pro: "专业版",
  enterprise: "企业版",
};

const PLAN_COLORS: Record<string, string> = {
  free: "bg-slate-100 text-slate-600",
  pro: "bg-blue-100 text-blue-700",
  enterprise: "bg-amber-100 text-amber-700",
};

// ——— 子组件：弹窗容器 ———

function Modal({
  title,
  onClose,
  children,
}: {
  title: string;
  onClose: () => void;
  children: React.ReactNode;
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm">
      <div className="w-full max-w-md bg-white rounded-2xl shadow-2xl overflow-hidden">
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-100">
          <h3 className="text-base font-semibold text-slate-800">{title}</h3>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg hover:bg-slate-100 text-slate-400 hover:text-slate-600 transition-colors"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
        <div className="px-6 py-5">{children}</div>
      </div>
    </div>
  );
}

// ——— Tab：租户设置 ———

function TenantTab({ token }: { token: string }) {
  const [tenant, setTenant] = useState<ShowTenant | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [editName, setEditName] = useState("");
  const [editPlan, setEditPlan] = useState("free");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    const res = await getAdminTenant(token);
    if (res.code === 0 && res.data) {
      setTenant(res.data);
      setEditName(res.data.name);
      setEditPlan(res.data.plan);
    }
    setLoading(false);
  }, [token]);

  useEffect(() => {
    load();
  }, [load]);

  const handleSave = async () => {
    setError("");
    setSuccess(false);
    setSaving(true);
    const res = await updateAdminTenant(token, {
      name: editName,
      plan: editPlan,
    });
    setSaving(false);
    if (res.code === 0) {
      setTenant(res.data);
      setSuccess(true);
      setTimeout(() => setSuccess(false), 2000);
    } else {
      setError(res.msg || "更新失败");
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center py-20">
        <RefreshCw className="h-6 w-6 text-slate-300 animate-spin" />
      </div>
    );
  }

  if (!tenant) return null;

  return (
    <div className="max-w-xl space-y-6">
      {/* 只读信息 */}
      <div className="bg-slate-50 rounded-xl p-5 space-y-3">
        <div className="flex items-center gap-2 text-xs font-medium text-slate-400 uppercase tracking-wider">
          <Building2 className="h-3.5 w-3.5" />
          租户信息
        </div>
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <p className="text-slate-400 text-xs mb-0.5">标识符 (Slug)</p>
            <p className="font-mono font-medium text-slate-700">{tenant.slug}</p>
          </div>
          <div>
            <p className="text-slate-400 text-xs mb-0.5">状态</p>
            <span
              className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${
                tenant.status === "active"
                  ? "bg-emerald-100 text-emerald-700"
                  : "bg-red-100 text-red-600"
              }`}
            >
              {tenant.status === "active" ? "正常" : "已暂停"}
            </span>
          </div>
          <div>
            <p className="text-slate-400 text-xs mb-0.5">创建时间</p>
            <p className="text-slate-700">{formatDate(tenant.createdAt)}</p>
          </div>
          <div>
            <p className="text-slate-400 text-xs mb-0.5">当前套餐</p>
            <span
              className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${
                PLAN_COLORS[tenant.plan] ?? "bg-slate-100 text-slate-600"
              }`}
            >
              {tenant.plan === "enterprise" && (
                <Crown className="h-3 w-3" />
              )}
              {PLAN_LABELS[tenant.plan] ?? tenant.plan}
            </span>
          </div>
        </div>
      </div>

      {/* 可编辑字段 */}
      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1.5">
            租户名称
          </label>
          <input
            className="w-full px-3 py-2 text-sm border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500/30 focus:border-blue-400"
            value={editName}
            onChange={(e) => setEditName(e.target.value)}
            placeholder="输入租户名称"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1.5">
            订阅计划
          </label>
          <select
            className="w-full px-3 py-2 text-sm border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500/30 focus:border-blue-400 bg-white"
            value={editPlan}
            onChange={(e) => setEditPlan(e.target.value)}
          >
            <option value="free">免费版</option>
            <option value="pro">专业版</option>
            <option value="enterprise">企业版</option>
          </select>
        </div>

        {error && (
          <div className="flex items-center gap-2 text-sm text-red-600 bg-red-50 px-3 py-2 rounded-lg">
            <AlertCircle className="h-4 w-4 shrink-0" />
            {error}
          </div>
        )}

        <button
          onClick={handleSave}
          disabled={saving}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 disabled:opacity-60 transition-colors"
        >
          {saving ? (
            <RefreshCw className="h-4 w-4 animate-spin" />
          ) : success ? (
            <Check className="h-4 w-4" />
          ) : null}
          {success ? "已保存" : "保存更改"}
        </button>
      </div>
    </div>
  );
}

// ——— Tab：用户管理 ———

type UserModalMode = "create" | "edit";

function UsersTab({ token }: { token: string }) {
  const { user: me } = useAuthStore();
  const [users, setUsers] = useState<ShowAdminUser[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);

  const [modal, setModal] = useState<UserModalMode | null>(null);
  const [editTarget, setEditTarget] = useState<ShowAdminUser | null>(null);

  const [formEmail, setFormEmail] = useState("");
  const [formNickName, setFormNickName] = useState("");
  const [formPassword, setFormPassword] = useState("");
  const [formRole, setFormRole] = useState("member");
  const [showPw, setShowPw] = useState(false);
  const [formError, setFormError] = useState("");
  const [saving, setSaving] = useState(false);

  const [deleteTarget, setDeleteTarget] = useState<ShowAdminUser | null>(null);
  const [deleting, setDeleting] = useState(false);

  const PAGE_SIZE = 20;

  const load = useCallback(async () => {
    setLoading(true);
    const res = await listAdminUsers(token, page, PAGE_SIZE);
    if (res.code === 0 && res.data) {
      setUsers(res.data.list);
      setTotal(res.data.total);
    }
    setLoading(false);
  }, [token, page]);

  useEffect(() => {
    load();
  }, [load]);

  const openCreate = () => {
    setFormEmail("");
    setFormNickName("");
    setFormPassword("");
    setFormRole("member");
    setFormError("");
    setModal("create");
  };

  const openEdit = (u: ShowAdminUser) => {
    setEditTarget(u);
    setFormRole(u.role);
    setFormError("");
    setModal("edit");
  };

  const handleSubmit = async () => {
    setFormError("");
    setSaving(true);
    if (modal === "create") {
      if (!formEmail || !formPassword || !formNickName) {
        setFormError("请填写所有必填项");
        setSaving(false);
        return;
      }
      const res = await createAdminUser(token, {
        email: formEmail,
        nickName: formNickName,
        password: formPassword,
        role: formRole,
      });
      if (res.code === 0) {
        setModal(null);
        load();
      } else {
        setFormError(res.msg || "创建失败");
      }
    } else if (modal === "edit" && editTarget) {
      const res = await updateAdminUser(token, editTarget.id, {
        role: formRole,
      });
      if (res.code === 0) {
        setModal(null);
        load();
      } else {
        setFormError(res.msg || "更新失败");
      }
    }
    setSaving(false);
  };

  const toggleStatus = async (u: ShowAdminUser) => {
    const next = u.status === "active" ? "disabled" : "active";
    const res = await updateAdminUser(token, u.id, { status: next });
    if (res.code === 0) load();
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    setDeleting(true);
    const res = await deleteAdminUser(token, deleteTarget.id);
    setDeleting(false);
    if (res.code === 0) {
      setDeleteTarget(null);
      load();
    }
  };

  const totalPages = Math.ceil(total / PAGE_SIZE);

  return (
    <div className="space-y-4">
      {/* 工具栏 */}
      <div className="flex items-center justify-between">
        <p className="text-sm text-slate-500">
          共 <span className="font-semibold text-slate-700">{total}</span> 位成员
        </p>
        <div className="flex items-center gap-2">
          <button
            onClick={load}
            className="p-2 rounded-lg hover:bg-slate-100 text-slate-400 transition-colors"
            title="刷新"
          >
            <RefreshCw className="h-4 w-4" />
          </button>
          <button
            onClick={openCreate}
            className="flex items-center gap-1.5 px-3 py-1.5 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
          >
            <Plus className="h-4 w-4" />
            邀请用户
          </button>
        </div>
      </div>

      {/* 用户列表 */}
      {loading ? (
        <div className="flex justify-center py-16">
          <RefreshCw className="h-6 w-6 text-slate-300 animate-spin" />
        </div>
      ) : users.length === 0 ? (
        <div className="text-center py-16 text-slate-400">
          <Users className="h-10 w-10 mx-auto mb-3 opacity-30" />
          <p className="text-sm">还没有用户</p>
        </div>
      ) : (
        <div className="bg-white rounded-xl border border-slate-100 overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-slate-50 text-slate-500 text-xs uppercase tracking-wider">
                <th className="text-left px-4 py-3 font-medium">用户</th>
                <th className="text-left px-4 py-3 font-medium">角色</th>
                <th className="text-left px-4 py-3 font-medium">状态</th>
                <th className="text-left px-4 py-3 font-medium">最近登录</th>
                <th className="text-left px-4 py-3 font-medium">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-50">
              {users.map((u) => (
                <tr key={u.id} className="hover:bg-slate-50/50 transition-colors">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-3">
                      <div className="h-8 w-8 rounded-full bg-gradient-to-br from-blue-400 to-indigo-500 flex items-center justify-center text-white text-xs font-bold shrink-0">
                        {((u.nickName || u.email || "?")[0] ?? "?").toUpperCase()}
                      </div>
                      <div>
                        <p className="font-medium text-slate-800">
                          {u.nickName || "—"}
                          {me && u.id === me.id && (
                            <span className="ml-1.5 text-xs text-slate-400">(我)</span>
                          )}
                        </p>
                        <p className="text-xs text-slate-400">{u.email}</p>
                      </div>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${
                        u.role === "admin"
                          ? "bg-purple-100 text-purple-700"
                          : "bg-slate-100 text-slate-600"
                      }`}
                    >
                      {u.role === "admin" && <Crown className="h-3 w-3" />}
                      {u.role === "admin" ? "管理员" : "成员"}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${
                        u.status === "active"
                          ? "bg-emerald-100 text-emerald-700"
                          : "bg-red-100 text-red-600"
                      }`}
                    >
                      {u.status === "active" ? "正常" : "已禁用"}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-slate-500 text-xs">
                    {formatDate(u.lastLogin ?? 0)}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-1">
                      <button
                        onClick={() => openEdit(u)}
                        className="p-1.5 rounded-md hover:bg-slate-100 text-slate-400 hover:text-slate-600 transition-colors"
                        title="编辑角色"
                      >
                        <Pencil className="h-3.5 w-3.5" />
                      </button>
                      <button
                        onClick={() => toggleStatus(u)}
                        disabled={me?.id === u.id}
                        className="p-1.5 rounded-md hover:bg-slate-100 text-slate-400 hover:text-slate-600 transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
                        title={u.status === "active" ? "禁用" : "启用"}
                      >
                        {u.status === "active" ? (
                          <ShieldOff className="h-3.5 w-3.5" />
                        ) : (
                          <ShieldCheck className="h-3.5 w-3.5" />
                        )}
                      </button>
                      <button
                        onClick={() => setDeleteTarget(u)}
                        disabled={me?.id === u.id}
                        className="p-1.5 rounded-md hover:bg-red-50 text-slate-400 hover:text-red-500 transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
                        title="删除"
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* 分页 */}
      {totalPages > 1 && (
        <div className="flex items-center justify-end gap-2 text-sm">
          <button
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page === 1}
            className="px-3 py-1.5 rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-50 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
          >
            上一页
          </button>
          <span className="text-slate-500">
            {page} / {totalPages}
          </span>
          <button
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            disabled={page === totalPages}
            className="px-3 py-1.5 rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-50 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
          >
            下一页
          </button>
        </div>
      )}

      {/* 创建/编辑弹窗 */}
      {modal && (
        <Modal
          title={modal === "create" ? "邀请用户" : "编辑用户"}
          onClose={() => setModal(null)}
        >
          <div className="space-y-4">
            {modal === "create" && (
              <>
                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-1.5">
                    邮箱 <span className="text-red-500">*</span>
                  </label>
                  <input
                    type="email"
                    className="w-full px-3 py-2 text-sm border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500/30 focus:border-blue-400"
                    value={formEmail}
                    onChange={(e) => setFormEmail(e.target.value)}
                    placeholder="user@example.com"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-1.5">
                    昵称 <span className="text-red-500">*</span>
                  </label>
                  <input
                    className="w-full px-3 py-2 text-sm border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500/30 focus:border-blue-400"
                    value={formNickName}
                    onChange={(e) => setFormNickName(e.target.value)}
                    placeholder="张三"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-1.5">
                    初始密码 <span className="text-red-500">*</span>
                  </label>
                  <div className="relative">
                    <input
                      type={showPw ? "text" : "password"}
                      className="w-full px-3 py-2 pr-10 text-sm border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500/30 focus:border-blue-400"
                      value={formPassword}
                      onChange={(e) => setFormPassword(e.target.value)}
                      placeholder="至少 8 位"
                    />
                    <button
                      type="button"
                      onClick={() => setShowPw((v) => !v)}
                      className="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600"
                    >
                      {showPw ? (
                        <EyeOff className="h-4 w-4" />
                      ) : (
                        <Eye className="h-4 w-4" />
                      )}
                    </button>
                  </div>
                </div>
              </>
            )}

            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1.5">
                角色
              </label>
              <select
                className="w-full px-3 py-2 text-sm border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500/30 focus:border-blue-400 bg-white"
                value={formRole}
                onChange={(e) => setFormRole(e.target.value)}
              >
                <option value="member">成员</option>
                <option value="admin">管理员</option>
              </select>
            </div>

            {formError && (
              <div className="flex items-center gap-2 text-sm text-red-600 bg-red-50 px-3 py-2 rounded-lg">
                <AlertCircle className="h-4 w-4 shrink-0" />
                {formError}
              </div>
            )}

            <div className="flex justify-end gap-2 pt-1">
              <button
                onClick={() => setModal(null)}
                className="px-4 py-2 text-sm border border-slate-200 rounded-lg hover:bg-slate-50 text-slate-600 transition-colors"
              >
                取消
              </button>
              <button
                onClick={handleSubmit}
                disabled={saving}
                className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 disabled:opacity-60 transition-colors"
              >
                {saving && <RefreshCw className="h-4 w-4 animate-spin" />}
                {modal === "create" ? "创建" : "保存"}
              </button>
            </div>
          </div>
        </Modal>
      )}

      {/* 删除确认 */}
      {deleteTarget && (
        <Modal title="删除用户" onClose={() => setDeleteTarget(null)}>
          <div className="space-y-4">
            <p className="text-sm text-slate-600">
              确定要删除用户{" "}
              <span className="font-semibold text-slate-800">
                {deleteTarget.nickName || deleteTarget.email}
              </span>{" "}
              吗？此操作不可恢复。
            </p>
            <div className="flex justify-end gap-2">
              <button
                onClick={() => setDeleteTarget(null)}
                className="px-4 py-2 text-sm border border-slate-200 rounded-lg hover:bg-slate-50 text-slate-600 transition-colors"
              >
                取消
              </button>
              <button
                onClick={handleDelete}
                disabled={deleting}
                className="flex items-center gap-2 px-4 py-2 bg-red-600 text-white text-sm font-medium rounded-lg hover:bg-red-700 disabled:opacity-60 transition-colors"
              >
                {deleting && <RefreshCw className="h-4 w-4 animate-spin" />}
                确认删除
              </button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  );
}

// ——— Tab：API Key 管理 ———

function ApiKeysTab({ token }: { token: string }) {
  const [keys, setKeys] = useState<ShowApiKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [newKeyName, setNewKeyName] = useState("");
  const [creating, setCreating] = useState(false);
  const [createError, setCreateError] = useState("");
  const [newKeyData, setNewKeyData] = useState<CreateApiKeyData | null>(null);
  const [copied, setCopied] = useState(false);
  const [revoking, setRevoking] = useState<number | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    const res = await listApiKeys(token);
    if (res.code === 0 && res.data) {
      setKeys(res.data);
    }
    setLoading(false);
  }, [token]);

  useEffect(() => {
    load();
  }, [load]);

  const handleCreate = async () => {
    if (!newKeyName.trim()) {
      setCreateError("请输入密钥名称");
      return;
    }
    setCreating(true);
    setCreateError("");
    const res = await createApiKey(token, { name: newKeyName });
    setCreating(false);
    if (res.code === 0 && res.data) {
      setNewKeyData(res.data);
      setNewKeyName("");
      setShowCreate(false);
      load();
    } else {
      setCreateError(res.msg || "创建失败");
    }
  };

  const handleCopy = async (text: string) => {
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleRevoke = async (id: number) => {
    setRevoking(id);
    await revokeApiKey(token, id);
    setRevoking(null);
    load();
  };

  return (
    <div className="space-y-4">
      {/* 新密钥展示（仅创建后一次性展示） */}
      {newKeyData && (
        <div className="bg-emerald-50 border border-emerald-200 rounded-xl p-4 space-y-3">
          <div className="flex items-start justify-between">
            <div>
              <p className="text-sm font-semibold text-emerald-800">
                🎉 密钥已创建：{newKeyData.name}
              </p>
              <p className="text-xs text-emerald-600 mt-0.5">
                请立即复制并妥善保存，此后将不再显示完整密钥。
              </p>
            </div>
            <button
              onClick={() => setNewKeyData(null)}
              className="text-emerald-400 hover:text-emerald-600"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
          <div className="flex items-center gap-2 bg-white rounded-lg border border-emerald-200 px-3 py-2">
            <code className="flex-1 text-sm font-mono text-slate-700 truncate">
              {newKeyData.key}
            </code>
            <button
              onClick={() => handleCopy(newKeyData.key)}
              className="shrink-0 text-emerald-600 hover:text-emerald-800"
              title="复制"
            >
              {copied ? (
                <Check className="h-4 w-4" />
              ) : (
                <Copy className="h-4 w-4" />
              )}
            </button>
          </div>
        </div>
      )}

      {/* 工具栏 */}
      <div className="flex items-center justify-between">
        <p className="text-sm text-slate-500">
          共 <span className="font-semibold text-slate-700">{keys.length}</span> 个密钥
        </p>
        <div className="flex items-center gap-2">
          <button
            onClick={load}
            className="p-2 rounded-lg hover:bg-slate-100 text-slate-400 transition-colors"
            title="刷新"
          >
            <RefreshCw className="h-4 w-4" />
          </button>
          <button
            onClick={() => setShowCreate(true)}
            className="flex items-center gap-1.5 px-3 py-1.5 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
          >
            <Plus className="h-4 w-4" />
            新建密钥
          </button>
        </div>
      </div>

      {/* 密钥列表 */}
      {loading ? (
        <div className="flex justify-center py-16">
          <RefreshCw className="h-6 w-6 text-slate-300 animate-spin" />
        </div>
      ) : keys.length === 0 ? (
        <div className="text-center py-16 text-slate-400">
          <Key className="h-10 w-10 mx-auto mb-3 opacity-30" />
          <p className="text-sm">还没有 API 密钥</p>
        </div>
      ) : (
        <div className="bg-white rounded-xl border border-slate-100 overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-slate-50 text-slate-500 text-xs uppercase tracking-wider">
                <th className="text-left px-4 py-3 font-medium">名称</th>
                <th className="text-left px-4 py-3 font-medium">前缀</th>
                <th className="text-left px-4 py-3 font-medium">状态</th>
                <th className="text-left px-4 py-3 font-medium">最近使用</th>
                <th className="text-left px-4 py-3 font-medium">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-50">
              {keys.map((k) => (
                <tr key={k.id} className="hover:bg-slate-50/50 transition-colors">
                  <td className="px-4 py-3 font-medium text-slate-800">{k.name}</td>
                  <td className="px-4 py-3">
                    <code className="text-xs font-mono bg-slate-100 text-slate-600 px-2 py-0.5 rounded">
                      {k.keyPrefix}…
                    </code>
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${
                        k.status === "active"
                          ? "bg-emerald-100 text-emerald-700"
                          : "bg-red-100 text-red-600"
                      }`}
                    >
                      {k.status === "active" ? "有效" : "已吊销"}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-xs text-slate-500">
                    {formatDate(k.lastUsed ?? 0)}
                  </td>
                  <td className="px-4 py-3">
                    {k.status === "active" && (
                      <button
                        onClick={() => handleRevoke(k.id)}
                        disabled={revoking === k.id}
                        className="flex items-center gap-1 text-xs text-red-500 hover:text-red-700 disabled:opacity-50 transition-colors"
                      >
                        {revoking === k.id ? (
                          <RefreshCw className="h-3.5 w-3.5 animate-spin" />
                        ) : (
                          <ShieldOff className="h-3.5 w-3.5" />
                        )}
                        吊销
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* 新建密钥弹窗 */}
      {showCreate && (
        <Modal title="新建 API 密钥" onClose={() => setShowCreate(false)}>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1.5">
                密钥名称
              </label>
              <input
                className="w-full px-3 py-2 text-sm border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500/30 focus:border-blue-400"
                value={newKeyName}
                onChange={(e) => setNewKeyName(e.target.value)}
                placeholder="如：生产环境"
                onKeyDown={(e) => e.key === "Enter" && handleCreate()}
              />
            </div>
            {createError && (
              <div className="flex items-center gap-2 text-sm text-red-600 bg-red-50 px-3 py-2 rounded-lg">
                <AlertCircle className="h-4 w-4 shrink-0" />
                {createError}
              </div>
            )}
            <div className="flex justify-end gap-2 pt-1">
              <button
                onClick={() => setShowCreate(false)}
                className="px-4 py-2 text-sm border border-slate-200 rounded-lg hover:bg-slate-50 text-slate-600 transition-colors"
              >
                取消
              </button>
              <button
                onClick={handleCreate}
                disabled={creating}
                className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 disabled:opacity-60 transition-colors"
              >
                {creating && <RefreshCw className="h-4 w-4 animate-spin" />}
                创建
              </button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  );
}

// ——— 主页面 ———

type AdminTab = "tenant" | "users" | "apikeys";

export default function AdminPage() {
  const { token } = useAuthStore();
  const [activeTab, setActiveTab] = useState<AdminTab>("tenant");

  if (!token) return null;

  const tabs: { id: AdminTab; label: string; icon: React.ReactNode }[] = [
    { id: "tenant", label: "租户设置", icon: <Building2 className="h-4 w-4" /> },
    { id: "users", label: "用户管理", icon: <Users className="h-4 w-4" /> },
    { id: "apikeys", label: "API 密钥", icon: <Key className="h-4 w-4" /> },
  ];

  return (
    <div className="flex flex-col h-full bg-slate-50">
      {/* 页头 */}
      <div className="bg-white border-b border-slate-100 px-6 py-4 shrink-0">
        <h1 className="text-lg font-semibold text-slate-800">管理后台</h1>
        <p className="text-sm text-slate-400 mt-0.5">
          租户设置 · 用户管理 · API 密钥
        </p>
      </div>

      <div className="flex flex-1 min-h-0">
        {/* 侧边 Tab 导航 */}
        <div className="w-48 bg-white border-r border-slate-100 shrink-0 py-4">
          <nav className="space-y-0.5 px-2">
            {tabs.map((t) => (
              <button
                key={t.id}
                onClick={() => setActiveTab(t.id)}
                className={`w-full flex items-center gap-2.5 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                  activeTab === t.id
                    ? "bg-blue-50 text-blue-700"
                    : "text-slate-600 hover:bg-slate-50 hover:text-slate-800"
                }`}
              >
                <span
                  className={
                    activeTab === t.id ? "text-blue-600" : "text-slate-400"
                  }
                >
                  {t.icon}
                </span>
                {t.label}
              </button>
            ))}
          </nav>
        </div>

        {/* 内容区 */}
        <div className="flex-1 overflow-y-auto p-6">
          {activeTab === "tenant" && <TenantTab token={token} />}
          {activeTab === "users" && <UsersTab token={token} />}
          {activeTab === "apikeys" && <ApiKeysTab token={token} />}
        </div>
      </div>
    </div>
  );
}
