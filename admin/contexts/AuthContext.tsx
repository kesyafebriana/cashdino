"use client";

import {
  createContext,
  useContext,
  useState,
  useEffect,
  type ReactNode,
} from "react";
import { useRouter, usePathname } from "next/navigation";

const ADMIN_EMAIL = process.env.NEXT_PUBLIC_ADMIN_EMAIL!;
const ADMIN_PASSWORD = process.env.NEXT_PUBLIC_ADMIN_PASSWORD!;
const STORAGE_KEY = process.env.NEXT_PUBLIC_AUTH_STORAGE_KEY!;

interface AuthContextValue {
  email: string | null;
  login: (email: string, password: string) => string | null;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue>({
  email: null,
  login: () => "Not initialized",
  logout: () => {},
});

export function useAuth() {
  return useContext(AuthContext);
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [email, setEmail] = useState<string | null>(null);
  const [ready, setReady] = useState(false);
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) setEmail(stored);
    setReady(true);
  }, []);

  useEffect(() => {
    if (!ready) return;
    if (!email && pathname !== "/login") {
      router.replace("/login");
    }
    if (email && pathname === "/login") {
      router.replace("/campaigns");
    }
  }, [email, pathname, ready, router]);

  const login = (inputEmail: string, inputPassword: string): string | null => {
    if (inputEmail === ADMIN_EMAIL && inputPassword === ADMIN_PASSWORD) {
      setEmail(inputEmail);
      localStorage.setItem(STORAGE_KEY, inputEmail);
      return null;
    }
    return "Invalid email or password";
  };

  const logout = () => {
    setEmail(null);
    localStorage.removeItem(STORAGE_KEY);
    router.replace("/login");
  };

  if (!ready) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-sm text-gray-400">Loading...</div>
      </div>
    );
  }

  return (
    <AuthContext.Provider value={{ email, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}
