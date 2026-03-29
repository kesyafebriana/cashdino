"use client";

import { useAuth } from "@/contexts/AuthContext";

export function AdminNavbar() {
  const { email, logout } = useAuth();

  return (
    <header className="flex items-center justify-between border-b border-gray-200 bg-white px-5 py-3">
      <div className="flex items-center gap-3">
        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-[#e65100] text-xs font-bold text-white">
          CD
        </div>
        <span className="text-lg font-bold text-gray-900">
          CashDino Admin
        </span>
      </div>
      <div className="flex items-center gap-4">
        <span className="text-sm text-gray-500">{email}</span>
        <button
          onClick={logout}
          className="text-sm text-gray-400 hover:text-gray-600"
        >
          Logout
        </button>
      </div>
    </header>
  );
}
