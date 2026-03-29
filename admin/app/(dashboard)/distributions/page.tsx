"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { fetchCampaigns, type AdminCampaignListItem } from "@/services/api";
import { StatusBadge } from "@/components/StatusBadge";

function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  const months = ["Jan","Feb","Mar","Apr","May","Jun","Jul","Aug","Sep","Oct","Nov","Dec"];
  return `${months[d.getUTCMonth()]} ${d.getUTCDate()}`;
}

export default function DistributionsPage() {
  const [campaigns, setCampaigns] = useState<AdminCampaignListItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchCampaigns()
      .then((data) => setCampaigns(data.filter((c) => c.status === "completed")))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  return (
    <div>
      <div className="mb-5">
        <h1 className="text-xl font-bold text-gray-900">Distributions</h1>
        <p className="mt-0.5 text-sm text-gray-500">
          View reward distribution reports for completed campaigns
        </p>
      </div>

      {loading ? (
        <div className="py-12 text-center text-sm text-gray-400">Loading...</div>
      ) : campaigns.length === 0 ? (
        <div className="py-12 text-center text-sm text-gray-400">
          No completed campaigns with distributions yet.
        </div>
      ) : (
        <div className="rounded-xl border border-gray-200 bg-white">
          <div className="grid grid-cols-[1fr_160px_120px_100px_32px] gap-4 border-b border-gray-100 px-4 py-3 text-xs font-medium uppercase tracking-wide text-gray-400">
            <div>Campaign</div>
            <div>Period</div>
            <div>Distributed</div>
            <div>Status</div>
            <div />
          </div>
          {campaigns.map((c) => (
            <Link
              key={c.id}
              href={`/distributions/${c.id}`}
              className="grid grid-cols-[1fr_160px_120px_100px_32px] items-center gap-4 border-b border-gray-50 px-4 py-3.5 transition-colors hover:bg-gray-50"
            >
              <div className="text-sm font-semibold text-gray-900">{c.name}</div>
              <div className="text-sm text-gray-600">
                {formatDate(c.challenge_start)} – {formatDate(c.challenge_end)}
              </div>
              <div className="text-sm text-gray-600">
                {c.distributed_count} / {c.total_stock}
              </div>
              <StatusBadge status={c.status} />
              <div className="text-gray-400">›</div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
