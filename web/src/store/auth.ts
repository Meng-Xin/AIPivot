import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { ShowUser } from "../lib/api";

interface AuthState {
  token: string | null;
  user: ShowUser | null;
  setAuth: (token: string, user: ShowUser) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      setAuth: (token, user) => set({ token, user }),
      logout: () => set({ token: null, user: null }),
    }),
    { name: "aipivot-auth" }
  )
);
