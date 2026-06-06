import { useState } from "react";
import { MessageSquare, BookOpen, Bot, BarChart2, Webhook, Settings, Wrench } from "lucide-react";
import { useAuthStore } from "./store/auth";
import LoginPage from "./pages/LoginPage";
import ChatPage from "./pages/ChatPage";
import KnowledgePage from "./pages/KnowledgePage";
import AnalyticsPage from "./pages/AnalyticsPage";
import WebhookPage from "./pages/WebhookPage";
import AdminPage from "./pages/AdminPage";
import SkillPage from "./pages/SkillPage";

type NavTab = "chat" | "knowledge" | "analytics" | "webhook" | "skill" | "admin";

export default function App() {
  const { token, user } = useAuthStore();
  const [tab, setTab] = useState<NavTab>("chat");

  if (!token) return <LoginPage />;

  return (
    <div className="flex h-full">
      {/* 导航侧栏 */}
      <nav aria-label="main-nav" className="flex w-14 flex-col items-center border-r border-slate-200 bg-slate-900 py-4 print:hidden">
        <div className="mb-6 flex h-9 w-9 items-center justify-center rounded-lg bg-indigo-600">
          <Bot className="h-5 w-5 text-white" />
        </div>
        <NavButton
          icon={<MessageSquare className="h-5 w-5" />}
          label="对话"
          active={tab === "chat"}
          onClick={() => setTab("chat")}
        />
        <NavButton
          icon={<BookOpen className="h-5 w-5" />}
          label="知识库"
          active={tab === "knowledge"}
          onClick={() => setTab("knowledge")}
        />
        <NavButton
          icon={<BarChart2 className="h-5 w-5" />}
          label="分析"
          active={tab === "analytics"}
          onClick={() => setTab("analytics")}
        />
        <NavButton
          icon={<Webhook className="h-5 w-5" />}
          label="渠道"
          active={tab === "webhook"}
          onClick={() => setTab("webhook")}
        />
        {user?.role === "admin" && (
          <>
            <NavButton
              icon={<Wrench className="h-5 w-5" />}
              label="Skill"
              active={tab === "skill"}
              onClick={() => setTab("skill")}
            />
            <NavButton
              icon={<Settings className="h-5 w-5" />}
              label="管理"
              active={tab === "admin"}
              onClick={() => setTab("admin")}
            />
          </>
        )}
      </nav>

      {/* 页面内容 */}
      <div className="flex-1 overflow-hidden">
        {tab === "chat" && <ChatPage />}
        {tab === "knowledge" && <KnowledgePage />}
        {tab === "analytics" && <AnalyticsPage />}
        {tab === "webhook" && <WebhookPage />}
        {tab === "skill" && <SkillPage />}
        {tab === "admin" && <AdminPage />}
      </div>
    </div>
  );
}

function NavButton({
  icon,
  label,
  active,
  onClick,
}: {
  icon: React.ReactNode;
  label: string;
  active: boolean;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      title={label}
      className={`mb-2 flex h-10 w-10 items-center justify-center rounded-lg transition ${
        active
          ? "bg-indigo-600 text-white"
          : "text-slate-400 hover:bg-slate-800 hover:text-white"
      }`}
    >
      {icon}
    </button>
  );
}
