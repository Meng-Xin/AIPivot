import { useAuthStore } from "./store/auth";
import LoginPage from "./pages/LoginPage";
import ChatPage from "./pages/ChatPage";

export default function App() {
  const token = useAuthStore((s) => s.token);
  return token ? <ChatPage /> : <LoginPage />;
}
