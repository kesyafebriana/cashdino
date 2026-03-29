"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

const NAV_ITEMS = [
  { label: "Campaigns", href: "/campaigns", icon: "⭐", match: "/campaigns" },
  { label: "Distributions", href: "/distributions", icon: "📋", match: "/distributions" },
];

export function Sidebar() {
  const pathname = usePathname();
  return (
    <aside className="w-52 shrink-0 border-r border-gray-200 bg-white">
      <div className="px-5 pb-1 pt-4 text-xs font-medium uppercase tracking-wide text-gray-400">
        Management
      </div>
      <nav className="flex flex-col py-1">
        {NAV_ITEMS.map((item) => {
          const active = pathname.startsWith(item.match);
          return (
            <Link
              key={item.label}
              href={item.href}
              className={`flex items-center gap-2.5 px-5 py-2.5 text-sm transition-colors ${
                active
                  ? "border-r-2 border-[#e65100] bg-orange-50 font-semibold text-[#e65100]"
                  : "text-gray-500 hover:bg-gray-50 hover:text-gray-700"
              }`}
            >
              <span className="text-base">{item.icon}</span>
              {item.label}
            </Link>
          );
        })}
      </nav>

      <div className="mx-5 my-3 border-t border-gray-200" />

      <div className="px-5 pb-1 text-xs font-medium uppercase tracking-wide text-gray-400">
        Tools
      </div>
      <nav className="flex flex-col py-1">
        <Link
          href="/campaigns"
          className="flex items-center gap-2.5 px-5 py-2.5 text-sm text-red-600 hover:bg-red-50"
        >
          <span className="text-base">🔄</span>
          Manual reset
        </Link>
      </nav>
    </aside>
  );
}
