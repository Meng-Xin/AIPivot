import { useEffect, useRef, useState, useCallback } from "react";
import {
  BookOpen,
  Plus,
  Search,
  Trash2,
  Edit3,
  Upload,
  FileText,
  Loader2,
  X,
  Database,
  Clock,
  AlertCircle,
  CheckCircle2,
  RefreshCw,
  File,
  MoreVertical,
} from "lucide-react";
import { useAuthStore } from "../store/auth";
import {
  listKnowledgeBases,
  createKnowledgeBase,
  updateKnowledgeBase,
  deleteKnowledgeBase,
  listDocuments,
  uploadDocument,
  deleteDocument,
} from "../lib/api";
import type { ShowKnowledgeBase, ShowDocument } from "../lib/api";

// ==================== KnowledgePage ====================

export default function KnowledgePage() {
  const token = useAuthStore((s) => s.token)!;

  const [kbList, setKbList] = useState<ShowKnowledgeBase[]>([]);
  const [kbTotal, setKbTotal] = useState(0);
  const [loadingKb, setLoadingKb] = useState(false);
  const [searchName, setSearchName] = useState("");
  const [activeKb, setActiveKb] = useState<ShowKnowledgeBase | null>(null);

  // 文档状态
  const [docs, setDocs] = useState<ShowDocument[]>([]);
  const [docsTotal, setDocsTotal] = useState(0);
  const [loadingDocs, setLoadingDocs] = useState(false);
  const [uploading, setUploading] = useState(false);

  // 弹窗状态
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingKb, setEditingKb] = useState<ShowKnowledgeBase | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<ShowKnowledgeBase | null>(
    null
  );
  const [deleteDocTarget, setDeleteDocTarget] = useState<ShowDocument | null>(
    null
  );

  const fileInputRef = useRef<HTMLInputElement>(null);

  // 加载知识库列表
  const loadKbList = useCallback(async () => {
    setLoadingKb(true);
    try {
      const res = await listKnowledgeBases(
        token,
        1,
        50,
        searchName || undefined
      );
      if (res.code === 0) {
        setKbList(res.data.list);
        setKbTotal(res.data.total);
      }
    } finally {
      setLoadingKb(false);
    }
  }, [token, searchName]);

  useEffect(() => {
    loadKbList();
  }, [loadKbList]);

  // 加载文档列表
  const loadDocs = useCallback(async () => {
    if (!activeKb) return;
    setLoadingDocs(true);
    try {
      const res = await listDocuments(token, activeKb.id);
      if (res.code === 0) {
        setDocs(res.data.list);
        setDocsTotal(res.data.total);
      }
    } finally {
      setLoadingDocs(false);
    }
  }, [token, activeKb]);

  useEffect(() => {
    if (activeKb) {
      loadDocs();
    } else {
      setDocs([]);
      setDocsTotal(0);
    }
  }, [activeKb, loadDocs]);

  // 创建知识库
  const handleCreate = async (data: {
    name: string;
    description?: string;
    model?: string;
    suggestedQuestions?: string[];
  }) => {
    const res = await createKnowledgeBase(token, data);
    if (res.code === 0) {
      setShowCreateModal(false);
      await loadKbList();
      setActiveKb(res.data);
    }
  };

  // 更新知识库
  const handleUpdate = async (data: {
    name?: string;
    description?: string;
    suggestedQuestions?: string[];
  }) => {
    if (!editingKb) return;
    const res = await updateKnowledgeBase(token, editingKb.id, data);
    if (res.code === 0) {
      setEditingKb(null);
      await loadKbList();
      // 刷新 activeKb 信息
      if (activeKb?.id === editingKb.id) {
        setActiveKb({ ...activeKb, ...data } as ShowKnowledgeBase);
      }
    }
  };

  // 删除知识库
  const handleDeleteKb = async () => {
    if (!deleteTarget) return;
    const res = await deleteKnowledgeBase(token, deleteTarget.id);
    if (res.code === 0) {
      setDeleteTarget(null);
      if (activeKb?.id === deleteTarget.id) setActiveKb(null);
      await loadKbList();
    }
  };

  // 上传文档
  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file || !activeKb) return;
    setUploading(true);
    try {
      const res = await uploadDocument(token, activeKb.id, file);
      if (res.code === 0) {
        await loadDocs();
        await loadKbList();
      }
    } catch (err) {
      console.error("上传失败:", err);
    } finally {
      setUploading(false);
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
  };

  // 删除文档
  const handleDeleteDoc = async () => {
    if (!deleteDocTarget || !activeKb) return;
    const res = await deleteDocument(token, activeKb.id, deleteDocTarget.id);
    if (res.code === 0) {
      setDeleteDocTarget(null);
      await loadDocs();
      await loadKbList();
    }
  };

  const formatTime = (ts: number) => {
    const d = new Date(ts);
    return d.toLocaleDateString("zh-CN", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  return (
    <div className="flex h-full bg-slate-50">
      {/* ========== 知识库列表侧边栏 ========== */}
      <aside className="flex w-80 flex-col border-r border-slate-200 bg-white">
        {/* 顶部 */}
        <div className="border-b border-slate-100 px-4 py-3">
          <div className="mb-3 flex items-center justify-between">
            <h2 className="text-sm font-semibold text-slate-800">知识库管理</h2>
            <button
              onClick={() => setShowCreateModal(true)}
              className="flex items-center gap-1.5 rounded-lg bg-indigo-600 px-3 py-1.5 text-xs font-medium text-white transition hover:bg-indigo-500"
            >
              <Plus className="h-3.5 w-3.5" />
              新建
            </button>
          </div>
          {/* 搜索 */}
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
            <input
              type="text"
              value={searchName}
              onChange={(e) => setSearchName(e.target.value)}
              placeholder="搜索知识库..."
              className="w-full rounded-lg border border-slate-200 bg-slate-50 py-2 pl-9 pr-3 text-sm outline-none transition placeholder:text-slate-400 focus:border-indigo-400 focus:bg-white focus:ring-2 focus:ring-indigo-100"
            />
          </div>
        </div>

        {/* 列表 */}
        <div className="flex-1 overflow-y-auto scrollbar-thin">
          {loadingKb ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-5 w-5 animate-spin text-slate-300" />
            </div>
          ) : kbList.length === 0 ? (
            <div className="px-4 py-12 text-center">
              <Database className="mx-auto mb-2 h-8 w-8 text-slate-200" />
              <p className="text-xs text-slate-400">
                {searchName ? "没有匹配的知识库" : "暂无知识库"}
              </p>
              {!searchName && (
                <button
                  onClick={() => setShowCreateModal(true)}
                  className="mt-3 text-xs font-medium text-indigo-600 hover:underline"
                >
                  创建第一个知识库
                </button>
              )}
            </div>
          ) : (
            <div className="p-2">
              {kbList.map((kb) => (
                <KbListItem
                  key={kb.id}
                  kb={kb}
                  active={activeKb?.id === kb.id}
                  onClick={() => setActiveKb(kb)}
                  onEdit={() => setEditingKb(kb)}
                  onDelete={() => setDeleteTarget(kb)}
                />
              ))}
              <p className="mt-2 text-center text-xs text-slate-400">
                共 {kbTotal} 个知识库
              </p>
            </div>
          )}
        </div>
      </aside>

      {/* ========== 主内容区 ========== */}
      <main className="flex flex-1 flex-col">
        {!activeKb ? (
          <KbEmptyState onCreate={() => setShowCreateModal(true)} />
        ) : (
          <>
            {/* 知识库详情头部 */}
            <KbDetailHeader
              kb={activeKb}
              onEdit={() => setEditingKb(activeKb)}
              onDelete={() => setDeleteTarget(activeKb)}
              onUpload={() => fileInputRef.current?.click()}
              uploading={uploading}
              formatTime={formatTime}
            />

            {/* 文档列表 */}
            <div className="flex-1 overflow-y-auto px-6 py-4 scrollbar-thin">
              <div className="mb-4 flex items-center justify-between">
                <h3 className="text-sm font-semibold text-slate-700">
                  文档列表
                  <span className="ml-2 text-xs font-normal text-slate-400">
                    ({docsTotal} 篇)
                  </span>
                </h3>
                <button
                  onClick={loadDocs}
                  className="flex items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-xs text-slate-500 transition hover:bg-slate-100"
                >
                  <RefreshCw className="h-3.5 w-3.5" />
                  刷新
                </button>
              </div>

              {loadingDocs ? (
                <div className="flex items-center justify-center py-16">
                  <Loader2 className="h-6 w-6 animate-spin text-slate-300" />
                </div>
              ) : docs.length === 0 ? (
                <DocEmptyState
                  onUpload={() => fileInputRef.current?.click()}
                />
              ) : (
                <div className="space-y-2">
                  {docs.map((doc) => (
                    <DocRow
                      key={doc.id}
                      doc={doc}
                      onDelete={() => setDeleteDocTarget(doc)}
                      formatTime={formatTime}
                      formatFileSize={formatFileSize}
                    />
                  ))}
                </div>
              )}
            </div>

            {/* 隐藏文件上传 input */}
            <input
              ref={fileInputRef}
              type="file"
              className="hidden"
              accept=".txt,.md,.pdf,.doc,.docx,.html,.htm,.csv"
              onChange={handleUpload}
            />
          </>
        )}
      </main>

      {/* ========== 弹窗 ========== */}
      {showCreateModal && (
        <KbFormModal
          title="新建知识库"
          onClose={() => setShowCreateModal(false)}
          onSubmit={handleCreate}
        />
      )}
      {editingKb && (
        <KbFormModal
          title="编辑知识库"
          initial={editingKb}
          onClose={() => setEditingKb(null)}
          onSubmit={handleUpdate}
        />
      )}
      {deleteTarget && (
        <ConfirmModal
          title="删除知识库"
          message={`确认删除知识库「${deleteTarget.name}」？该操作将同时删除所有文档和向量数据，且不可恢复。`}
          onClose={() => setDeleteTarget(null)}
          onConfirm={handleDeleteKb}
        />
      )}
      {deleteDocTarget && (
        <ConfirmModal
          title="删除文档"
          message={`确认删除文档「${deleteDocTarget.name}」？关联的切块和向量数据将一并删除。`}
          onClose={() => setDeleteDocTarget(null)}
          onConfirm={handleDeleteDoc}
        />
      )}
    </div>
  );
}

// ==================== 子组件 ====================

function KbListItem({
  kb,
  active,
  onClick,
  onEdit,
  onDelete,
}: {
  kb: ShowKnowledgeBase;
  active: boolean;
  onClick: () => void;
  onEdit: () => void;
  onDelete: () => void;
}) {
  const [showMenu, setShowMenu] = useState(false);

  return (
    <div
      onClick={onClick}
      className={`group relative cursor-pointer rounded-lg px-3 py-2.5 transition ${
        active
          ? "bg-indigo-50 text-indigo-700"
          : "text-slate-600 hover:bg-slate-50"
      }`}
    >
      <div className="flex items-start gap-3">
        <div
          className={`mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg ${
            active ? "bg-indigo-100" : "bg-slate-100"
          }`}
        >
          <BookOpen
            className={`h-4 w-4 ${
              active ? "text-indigo-500" : "text-slate-400"
            }`}
          />
        </div>
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-medium">{kb.name}</p>
          {kb.description && (
            <p className="mt-0.5 truncate text-xs text-slate-400">
              {kb.description}
            </p>
          )}
          <div className="mt-1 flex items-center gap-3 text-xs text-slate-400">
            <span>{kb.documentCount} 篇文档</span>
            <span>{kb.chunkCount} 切块</span>
          </div>
        </div>

        {/* 操作按钮 */}
        <div className="relative">
          <button
            onClick={(e) => {
              e.stopPropagation();
              setShowMenu(!showMenu);
            }}
            className="rounded p-1 text-slate-400 opacity-0 transition hover:bg-slate-200 group-hover:opacity-100"
          >
            <MoreVertical className="h-4 w-4" />
          </button>
          {showMenu && (
            <>
              <div
                className="fixed inset-0 z-10"
                onClick={(e) => {
                  e.stopPropagation();
                  setShowMenu(false);
                }}
              />
              <div className="absolute right-0 top-full z-20 mt-1 w-28 rounded-lg border border-slate-200 bg-white py-1 shadow-lg">
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    setShowMenu(false);
                    onEdit();
                  }}
                  className="flex w-full items-center gap-2 px-3 py-1.5 text-xs text-slate-600 hover:bg-slate-50"
                >
                  <Edit3 className="h-3.5 w-3.5" />
                  编辑
                </button>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    setShowMenu(false);
                    onDelete();
                  }}
                  className="flex w-full items-center gap-2 px-3 py-1.5 text-xs text-red-500 hover:bg-red-50"
                >
                  <Trash2 className="h-3.5 w-3.5" />
                  删除
                </button>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

function KbDetailHeader({
  kb,
  onEdit,
  onDelete,
  onUpload,
  uploading,
  formatTime,
}: {
  kb: ShowKnowledgeBase;
  onEdit: () => void;
  onDelete: () => void;
  onUpload: () => void;
  uploading: boolean;
  formatTime: (ts: number) => string;
}) {
  return (
    <div className="border-b border-slate-200 bg-white px-6 py-4">
      <div className="flex items-start justify-between">
        <div>
          <h2 className="text-lg font-semibold text-slate-800">{kb.name}</h2>
          {kb.description && (
            <p className="mt-1 text-sm text-slate-500">{kb.description}</p>
          )}
          <div className="mt-2 flex flex-wrap items-center gap-3 text-xs text-slate-400">
            <span className="flex items-center gap-1">
              <FileText className="h-3.5 w-3.5" />
              {kb.documentCount} 篇文档
            </span>
            <span className="flex items-center gap-1">
              <Database className="h-3.5 w-3.5" />
              {kb.chunkCount} 切块
            </span>
            <StatusBadge status={kb.status} />
            <span className="flex items-center gap-1">
              <Clock className="h-3.5 w-3.5" />
              {formatTime(kb.createdAt)}
            </span>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={onUpload}
            disabled={uploading}
            className="flex items-center gap-1.5 rounded-lg bg-indigo-600 px-3.5 py-2 text-xs font-medium text-white transition hover:bg-indigo-500 disabled:opacity-60"
          >
            {uploading ? (
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
            ) : (
              <Upload className="h-3.5 w-3.5" />
            )}
            {uploading ? "上传中..." : "上传文档"}
          </button>
          <button
            onClick={onEdit}
            className="rounded-lg p-2 text-slate-400 transition hover:bg-slate-100 hover:text-slate-600"
            title="编辑"
          >
            <Edit3 className="h-4 w-4" />
          </button>
          <button
            onClick={onDelete}
            className="rounded-lg p-2 text-slate-400 transition hover:bg-red-50 hover:text-red-500"
            title="删除"
          >
            <Trash2 className="h-4 w-4" />
          </button>
        </div>
      </div>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const config: Record<string, { color: string; icon: React.ReactNode }> = {
    active: {
      color: "bg-green-50 text-green-600",
      icon: <CheckCircle2 className="h-3 w-3" />,
    },
    processing: {
      color: "bg-amber-50 text-amber-600",
      icon: <RefreshCw className="h-3 w-3 animate-spin" />,
    },
    error: {
      color: "bg-red-50 text-red-500",
      icon: <AlertCircle className="h-3 w-3" />,
    },
  };
  const c = config[status] ?? {
    color: "bg-slate-100 text-slate-500",
    icon: null,
  };
  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ${c.color}`}
    >
      {c.icon}
      {status}
    </span>
  );
}

function DocStatusBadge({ status }: { status: string }) {
  const map: Record<string, { color: string; label: string }> = {
    pending: { color: "bg-slate-100 text-slate-500", label: "等待处理" },
    processing: { color: "bg-amber-50 text-amber-600", label: "处理中" },
    completed: { color: "bg-green-50 text-green-600", label: "已完成" },
    failed: { color: "bg-red-50 text-red-500", label: "处理失败" },
  };
  const c = map[status] ?? {
    color: "bg-slate-100 text-slate-500",
    label: status,
  };
  return (
    <span
      className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${c.color}`}
    >
      {c.label}
    </span>
  );
}

function DocRow({
  doc,
  onDelete,
  formatTime,
  formatFileSize,
}: {
  doc: ShowDocument;
  onDelete: () => void;
  formatTime: (ts: number) => string;
  formatFileSize: (bytes: number) => string;
}) {
  return (
    <div className="group flex items-center gap-4 rounded-lg border border-slate-100 bg-white px-4 py-3 transition hover:border-slate-200 hover:shadow-sm">
      <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-slate-100">
        <File className="h-5 w-5 text-slate-400" />
      </div>
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium text-slate-700">
          {doc.name}
        </p>
        <div className="mt-0.5 flex items-center gap-3 text-xs text-slate-400">
          <span>{formatFileSize(doc.fileSize)}</span>
          <span>{doc.contentType}</span>
          {doc.chunkCount > 0 && <span>{doc.chunkCount} 切块</span>}
          <span>{formatTime(doc.createdAt)}</span>
        </div>
        {doc.errorMsg && (
          <p className="mt-1 text-xs text-red-400">{doc.errorMsg}</p>
        )}
      </div>
      <DocStatusBadge status={doc.status} />
      <button
        onClick={onDelete}
        className="rounded-lg p-1.5 text-slate-400 opacity-0 transition hover:bg-red-50 hover:text-red-500 group-hover:opacity-100"
        title="删除文档"
      >
        <Trash2 className="h-4 w-4" />
      </button>
    </div>
  );
}

function KbEmptyState({ onCreate }: { onCreate: () => void }) {
  return (
    <div className="flex flex-1 flex-col items-center justify-center">
      <div className="mb-6 flex h-20 w-20 items-center justify-center rounded-2xl bg-indigo-50">
        <BookOpen className="h-10 w-10 text-indigo-400" />
      </div>
      <h2 className="mb-2 text-lg font-semibold text-slate-700">
        知识库管理
      </h2>
      <p className="mb-6 max-w-sm text-center text-sm text-slate-400">
        选择左侧知识库查看详情和文档，或创建新知识库开始构建你的知识体系。
      </p>
      <button
        onClick={onCreate}
        className="flex items-center gap-2 rounded-xl bg-indigo-600 px-5 py-2.5 text-sm font-medium text-white transition hover:bg-indigo-500"
      >
        <Plus className="h-4 w-4" />
        新建知识库
      </button>
    </div>
  );
}

function DocEmptyState({ onUpload }: { onUpload: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-slate-200 py-16">
      <Upload className="mb-3 h-10 w-10 text-slate-300" />
      <p className="mb-1 text-sm font-medium text-slate-600">暂无文档</p>
      <p className="mb-4 text-xs text-slate-400">
        上传文档后系统将自动进行切块和向量化处理
      </p>
      <button
        onClick={onUpload}
        className="flex items-center gap-2 rounded-lg bg-indigo-600 px-4 py-2 text-xs font-medium text-white transition hover:bg-indigo-500"
      >
        <Upload className="h-3.5 w-3.5" />
        上传文档
      </button>
      <p className="mt-3 text-xs text-slate-400">
        支持 TXT, MD, PDF, DOC, DOCX, HTML, CSV
      </p>
    </div>
  );
}

// ==================== 表单弹窗 ====================

const MAX_SUGGESTED = 6;
const MAX_SUGGESTED_LEN = 100;

function KbFormModal({
  title,
  initial,
  onClose,
  onSubmit,
}: {
  title: string;
  initial?: ShowKnowledgeBase;
  onClose: () => void;
  onSubmit: (data: {
    name: string;
    description?: string;
    model?: string;
    suggestedQuestions?: string[];
  }) => Promise<void>;
}) {
  const [name, setName] = useState(initial?.name ?? "");
  const [description, setDescription] = useState(initial?.description ?? "");
  const [model, setModel] = useState("text-embedding-3-small");
  // 编辑态用 initial.suggestedQuestions 初始化；新建态显式给空数组，提交时空数组也会落库
  const [questions, setQuestions] = useState<string[]>(
    initial?.suggestedQuestions ?? []
  );
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const addQuestion = () => {
    if (questions.length >= MAX_SUGGESTED) return;
    setQuestions([...questions, ""]);
  };
  const updateQuestion = (i: number, v: string) => {
    setQuestions(questions.map((q, idx) => (idx === i ? v : q)));
  };
  const removeQuestion = (i: number) => {
    setQuestions(questions.filter((_, idx) => idx !== i));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      setError("知识库名称不能为空");
      return;
    }
    // 清洗：去首尾空白、过滤空串、超长截断
    const cleaned = questions
      .map((q) => q.trim())
      .filter((q) => q.length > 0)
      .map((q) => (q.length > MAX_SUGGESTED_LEN ? q.slice(0, MAX_SUGGESTED_LEN) : q));

    setLoading(true);
    setError("");
    try {
      await onSubmit({
        name: name.trim(),
        description: description.trim() || undefined,
        model: initial ? undefined : model,
        // 显式传切片（含空数组），便于"清空"语义
        suggestedQuestions: cleaned,
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : "操作失败");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 backdrop-blur-sm">
      <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-xl max-h-[90vh] overflow-y-auto">
        <div className="mb-5 flex items-center justify-between">
          <h3 className="text-lg font-semibold text-slate-800">{title}</h3>
          <button
            onClick={onClose}
            className="rounded-lg p-1 text-slate-400 hover:bg-slate-100 hover:text-slate-600"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              知识库名称 <span className="text-red-400">*</span>
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="例如：产品文档库"
              className="w-full rounded-lg border border-slate-200 px-3 py-2 text-sm outline-none transition focus:border-indigo-400 focus:ring-2 focus:ring-indigo-100"
              autoFocus
            />
          </div>

          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              描述（可选）
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="简要描述知识库的用途和范围"
              rows={3}
              className="w-full resize-none rounded-lg border border-slate-200 px-3 py-2 text-sm outline-none transition focus:border-indigo-400 focus:ring-2 focus:ring-indigo-100"
            />
          </div>

          {!initial && (
            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-700">
                Embedding 模型
              </label>
              <select
                value={model}
                onChange={(e) => setModel(e.target.value)}
                className="w-full appearance-none rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm outline-none transition focus:border-indigo-400 focus:ring-2 focus:ring-indigo-100"
              >
                <option value="text-embedding-3-small">
                  text-embedding-3-small (1536d)
                </option>
                <option value="text-embedding-3-large">
                  text-embedding-3-large (3072d)
                </option>
                <option value="text-embedding-ada-002">
                  text-embedding-ada-002 (1536d)
                </option>
              </select>
              <p className="mt-1 text-xs text-slate-400">
                创建后不可更改，所有文档将使用此模型生成向量
              </p>
            </div>
          )}

          {/* 引导问答 / 快捷回复 */}
          <div>
            <div className="mb-1.5 flex items-center justify-between">
              <label className="block text-sm font-medium text-slate-700">
                引导问答（可选）
              </label>
              <span className="text-xs text-slate-400">
                {questions.length}/{MAX_SUGGESTED} 条
              </span>
            </div>
            <p className="mb-2 text-xs text-slate-400">
              Widget 首屏展示为快捷回复按钮，最多 {MAX_SUGGESTED} 条，每条 ≤ {MAX_SUGGESTED_LEN} 字
            </p>
            <div className="space-y-2">
              {questions.map((q, i) => (
                <div key={i} className="flex items-center gap-2">
                  <input
                    type="text"
                    value={q}
                    maxLength={MAX_SUGGESTED_LEN}
                    onChange={(e) => updateQuestion(i, e.target.value)}
                    placeholder={`快捷回复 ${i + 1}`}
                    className="flex-1 rounded-lg border border-slate-200 px-3 py-1.5 text-sm outline-none transition focus:border-indigo-400 focus:ring-2 focus:ring-indigo-100"
                  />
                  <button
                    type="button"
                    onClick={() => removeQuestion(i)}
                    className="rounded-lg p-1.5 text-slate-400 hover:bg-red-50 hover:text-red-500"
                    title="删除"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              ))}
            </div>
            {questions.length < MAX_SUGGESTED && (
              <button
                type="button"
                onClick={addQuestion}
                className="mt-2 flex items-center gap-1 rounded-lg border border-dashed border-slate-300 px-3 py-1.5 text-xs font-medium text-slate-500 transition hover:border-indigo-400 hover:text-indigo-600"
              >
                <Plus className="h-3.5 w-3.5" />
                添加问题
              </button>
            )}
          </div>

          {error && (
            <p className="rounded-lg bg-red-50 px-3 py-2 text-xs text-red-500">
              {error}
            </p>
          )}

          <div className="flex justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded-lg px-4 py-2 text-sm font-medium text-slate-600 transition hover:bg-slate-100"
            >
              取消
            </button>
            <button
              type="submit"
              disabled={loading}
              className="flex items-center gap-2 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-indigo-500 disabled:opacity-60"
            >
              {loading && <Loader2 className="h-4 w-4 animate-spin" />}
              {initial ? "保存" : "创建"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function ConfirmModal({
  title,
  message,
  onClose,
  onConfirm,
}: {
  title: string;
  message: string;
  onClose: () => void;
  onConfirm: () => Promise<void>;
}) {
  const [loading, setLoading] = useState(false);

  const handleConfirm = async () => {
    setLoading(true);
    try {
      await onConfirm();
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 backdrop-blur-sm">
      <div className="w-full max-w-sm rounded-2xl bg-white p-6 shadow-xl">
        <div className="mb-4 flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-full bg-red-50">
            <AlertCircle className="h-5 w-5 text-red-500" />
          </div>
          <h3 className="text-lg font-semibold text-slate-800">{title}</h3>
        </div>
        <p className="mb-6 text-sm text-slate-500">{message}</p>
        <div className="flex justify-end gap-3">
          <button
            onClick={onClose}
            className="rounded-lg px-4 py-2 text-sm font-medium text-slate-600 transition hover:bg-slate-100"
          >
            取消
          </button>
          <button
            onClick={handleConfirm}
            disabled={loading}
            className="flex items-center gap-2 rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-red-500 disabled:opacity-60"
          >
            {loading && <Loader2 className="h-4 w-4 animate-spin" />}
            确认删除
          </button>
        </div>
      </div>
    </div>
  );
}
