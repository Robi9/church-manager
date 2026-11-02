"use client";

import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from "react";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";

interface AuthContextType {
  login: () => void;
  logout: () => Promise<void>;
  isAuthenticated: boolean;
}
const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isAuthenticated, setAuthenticated] = useState(false);
  const [loaded, setLoaded] = useState(false);
  const router = useRouter();

  useEffect(() => {
    let active = true;
    // Remove tokens left behind by the previous authentication mechanism.
    localStorage.removeItem("church_manager_token");
    fetch("/api/auth/me", { cache: "no-store" })
      .then((res) => { if (active) setAuthenticated(res.ok); })
      .catch(() => { if (active) setAuthenticated(false); })
      .finally(() => { if (active) setLoaded(true); });
    return () => { active = false; };
  }, []);

  const login = useCallback(() => {
    setAuthenticated(true);
    router.push("/members");
  }, [router]);

  const logout = useCallback(async () => {
    try {
      await api("/auth/logout-all", { method: "POST" });
      setAuthenticated(false);
      router.push("/login");
    } catch {
      window.alert("Não foi possível encerrar a sessão no servidor. Tente sair novamente.");
    }
  }, [router]);

  if (!loaded) return null;
  return <AuthContext.Provider value={{ login, logout, isAuthenticated }}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
