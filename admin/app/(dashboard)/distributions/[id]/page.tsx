"use client";

import { useState, useEffect, use } from "react";
import Link from "next/link";
import {
  fetchCampaign,
  fetchDistributions,
  retryDistribution,
  markDistributionDelivered,
  type AdminCampaignDetail,
  type AdminDistributionRow,
} from "@/services/api";
import { StatusBadge } from "@/components/StatusBadge";

type Filter = "all" | "delivered" | "failed";

function formatDateTime(dateStr: string | null): string {
  if (!dateStr) return "—";
  const d = new Date(dateStr);
  const months = ["Jan","Feb","Mar","Apr","May","Jun","Jul","Aug","Sep","Oct","Nov","Dec"];
  return `${months[d.getUTCMonth()]} ${d.getUTCDate()}, ${String(d.getUTCHours()).padStart(2, "0")}:${String(d.getUTCMinutes()).padStart(2, "0")}`;
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  const months = ["Jan","Feb","Mar","Apr","May","Jun","Jul","Aug","Sep","Oct","Nov","Dec"];
  return `${months[d.getUTCMonth()]} ${d.getUTCDate()}`;
}

export default function DistributionDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const [campaign, setCampaign] = useState<AdminCampaignDetail | null>(null);
  const [distributions, setDistributions] = useState<AdminDistributionRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<Filter>("all");
  const [actioningId, setActioningId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const loadData = () => {
    Promise.all([fetchCampaign(id), fetchDistributions(id).catch(() => [])])
      .then(([c, d]) => {
        setCampaign(c);
        setDistributions(d ?? []);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  };

  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(loadData, [id]);

  const reloadData = async () => {
    const [c, d] = await Promise.all([
      fetchCampaign(id),
      fetchDistributions(id).catch(() => []),
    ]);
    setCampaign(c);
    setDistributions(d ?? []);
  };

  const handleRetry = async (distId: string) => {
    setActioningId(distId);
    setError(null);
    try {
      await retryDistribution(distId);
      await reloadData();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Retry failed");
      await reloadData();
    } finally {
      setActioningId(null);
    }
  };

  const handleMarkDelivered = async (distId: string) => {
    setActioningId(distId);
    setError(null);
    try {
      await markDistributionDelivered(distId);
      await reloadData();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Action failed");
    } finally {
      setActioningId(null);
    }
  };

  if (loading) {
    return <div className="py-12 text-center text-sm text-gray-400">Loading...</div>;
  }
  if (!campaign) {
    return <div className="py-12 text-center text-sm text-gray-400">Campaign not found.</div>;
  }

  const delivered = distributions.filter((d) => d.status === "delivered").length;
  const failed = distributions.filter((d) => d.status === "failed").length;
  const uniqueUsers = new Set(distributions.map((d) => d.user_id)).size;

  const filtered =
    filter === "all"
      ? distributions
      : distributions.filter((d) => d.status === filter);

  // Track which ranks are first occurrence to show rank number
  const seenRanks = new Set<number>();

  return (
    <div>
      {/* Breadcrumb */}
      <div className="mb-6 flex items-center gap-1.5 text-[13px]">
        <Link href="/distributions" className="text-gray-400 hover:text-gray-600">Distributions</Link>
        <span className="text-gray-300">›</span>
        <span className="font-medium text-gray-900">{campaign.name}</span>
      </div>

      {/* Header */}
      <div className="mb-6">
        <div className="mb-1 flex items-center gap-3">
          <h1 className="text-[26px] font-bold text-gray-900">Distribution report</h1>
          <StatusBadge status={campaign.status} />
        </div>
        <p className="text-sm text-gray-500">{campaign.name}</p>
      </div>

      {/* Summary cards */}
      <div className="mb-7 grid grid-cols-4 gap-4">
        <SummaryCard label="Total rewards" value={distributions.length} />
        <SummaryCard label="Delivered" value={delivered} color="text-green-700" />
        <SummaryCard label="Failed" value={failed} color="text-red-600" />
        <SummaryCard label="Users rewarded" value={uniqueUsers} />
      </div>

      {/* Error */}
      {error && (
        <p className="mb-4 rounded-lg bg-red-50 px-4 py-2 text-sm text-red-600">{error}</p>
      )}

      {/* Filters */}
      <div className="mb-5 flex gap-2">
        {([
          { key: "all" as Filter, label: `All (${distributions.length})` },
          { key: "delivered" as Filter, label: `Delivered (${delivered})` },
          { key: "failed" as Filter, label: `Failed (${failed})`, danger: true },
        ]).map((f) => (
          <button
            key={f.key}
            onClick={() => setFilter(f.key)}
            className={`rounded-full px-4 py-1.5 text-xs font-semibold transition-colors ${
              filter === f.key
                ? "bg-gray-900 text-white"
                : f.danger
                  ? "border border-red-200 text-red-600 hover:bg-red-50"
                  : "border border-gray-300 text-gray-500 hover:bg-gray-50"
            }`}
          >
            {f.label}
          </button>
        ))}
      </div>

      {/* Table */}
      <div className="overflow-hidden rounded-xl border border-gray-200 bg-white">
        <div className="grid grid-cols-[60px_1.5fr_1fr_1.5fr_0.8fr_0.8fr_1.2fr_100px] gap-3 border-b border-gray-200 bg-gray-50 px-5 py-3 text-[11px] font-semibold uppercase tracking-wide text-gray-400">
          <span>Rank</span>
          <span>User</span>
          <span>Email</span>
          <span>Reward</span>
          <span>Type</span>
          <span>Value</span>
          <span>Delivered at</span>
          <span>Status</span>
        </div>

        {filtered.length === 0 ? (
          <div className="px-5 py-8 text-center text-sm text-gray-400">
            No distributions found.
          </div>
        ) : (
          filtered.map((d) => {
            const isFirstForRank = !seenRanks.has(d.final_rank);
            if (isFirstForRank) seenRanks.add(d.final_rank);
            const isFailed = d.status === "failed";

            return (
              <div
                key={d.id}
                className={`grid grid-cols-[60px_1.5fr_1fr_1.5fr_0.8fr_0.8fr_1.2fr_100px] items-center gap-3 border-b border-gray-100 px-5 py-3 text-[13px] ${
                  isFailed ? "bg-red-50/50" : ""
                }`}
              >
                <span className={isFirstForRank ? "font-semibold text-gray-900" : "text-gray-300"}>
                  {isFirstForRank ? `#${d.final_rank}` : ""}
                </span>
                <span className={isFirstForRank ? "font-medium text-gray-900" : "text-gray-500"}>
                  {d.display_name}
                </span>
                <span className="text-xs text-gray-400">{d.masked_email}</span>
                <span className={isFailed ? "font-medium text-red-600" : ""}>
                  {d.reward_name}
                </span>
                <span className="text-gray-500">{d.reward_type}</span>
                <span>
                  {d.reward_type === "gems"
                    ? d.reward_value.toLocaleString()
                    : `$${d.reward_value}`}
                </span>
                <span className={`text-xs ${isFailed ? "text-red-600" : "text-gray-400"}`}>
                  {isFailed
                    ? "Failed"
                    : formatDateTime(d.delivered_at)}
                </span>
                <span>
                  {isFailed ? (
                    <div className="flex gap-1.5">
                      <button
                        onClick={() => handleRetry(d.id)}
                        disabled={actioningId === d.id}
                        className="rounded-md bg-red-50 px-2.5 py-1 text-[11px] font-semibold text-red-600 hover:bg-red-100 disabled:opacity-50 transition-colors"
                      >
                        {actioningId === d.id ? "..." : "Retry"}
                      </button>
                      <button
                        onClick={() => handleMarkDelivered(d.id)}
                        disabled={actioningId === d.id}
                        className="rounded-md bg-green-50 px-2.5 py-1 text-[11px] font-semibold text-green-700 hover:bg-green-100 disabled:opacity-50 transition-colors"
                        title="Mark as manually delivered"
                      >
                        ✓
                      </button>
                    </div>
                  ) : (
                    <StatusBadge status={d.status} />
                  )}
                </span>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}

function SummaryCard({
  label,
  value,
  color,
}: {
  label: string;
  value: number;
  color?: string;
}) {
  return (
    <div className="rounded-lg bg-gray-100 px-5 py-4">
      <div className="mb-1.5 text-xs text-gray-500">{label}</div>
      <div className={`text-[28px] font-medium ${color ?? "text-gray-900"}`}>
        {value}
      </div>
    </div>
  );
}
