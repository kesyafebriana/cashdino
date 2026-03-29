"use client";

import { useAuth } from "@/contexts/AuthContext";
import { AdminNavbar } from "@/components/AdminNavbar";
import { Sidebar } from "@/components/Sidebar";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { email } = useAuth();

  if (!email) return null;

  return (
    <div className="flex min-h-screen flex-col">
      <AdminNavbar />
      <div className="flex flex-1">
        <Sidebar />
        <main className="flex-1 p-6">{children}</main>
      </div>
    </div>
  );
}
