"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { fetchCampaigns, type AdminCampaignListItem } from "@/services/api";
import { StatusBadge } from "@/components/StatusBadge";
import { RewardTypeBadge } from "@/components/RewardTypeBadge";

function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  const months = [
    "Jan", "Feb", "Mar", "Apr", "May", "Jun",
    "Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
  ];
  return `${months[d.getUTCMonth()]} ${d.getUTCDate()}`;
}

function formatPeriod(start: string, end: string): string {
  return `${formatDate(start)} – ${formatDate(end)}`;
}

const FILTERS = ["all", "active", "completed", "draft"] as const;
type Filter = (typeof FILTERS)[number];

export default function CampaignsPage() {
  const [campaigns, setCampaigns] = useState<AdminCampaignListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<Filter>("all");

  useEffect(() => {
    fetchCampaigns()
      .then(setCampaigns)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const filtered =
    filter === "all"
      ? campaigns
      : campaigns.filter((c) => c.status === filter);

  const counts: Record<Filter, number> = {
    all: campaigns.length,
    active: campaigns.filter((c) => c.status === "active").length,
    completed: campaigns.filter((c) => c.status === "completed").length,
    draft: campaigns.filter((c) => c.status === "draft").length,
  };

  return (
    <div>
      {/* Header */}
      <div className="mb-5 flex items-start justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Reward campaigns</h1>
          <p className="mt-0.5 text-sm text-gray-500">
            Manage weekly challenge reward campaigns
          </p>
        </div>
        <Link
          href="/campaigns/new"
          className="rounded-lg bg-[#e65100] px-4 py-2 text-sm font-semibold text-white hover:bg-[#bf360c] transition-colors"
        >
          + New campaign
        </Link>
      </div>

      {/* Filter pills */}
      <div className="mb-4 flex gap-2">
        {FILTERS.map((f) => (
          <button
            key={f}
            onClick={() => setFilter(f)}
            className={`rounded-full px-3.5 py-1 text-xs font-semibold capitalize transition-colors ${
              filter === f
                ? "bg-gray-900 text-white"
                : "border border-gray-300 text-gray-500 hover:bg-gray-50"
            }`}
          >
            {f} ({counts[f]})
          </button>
        ))}
      </div>

      {/* Table */}
      {loading ? (
        <div className="py-12 text-center text-sm text-gray-400">
          Loading campaigns...
        </div>
      ) : (
        <div className="rounded-xl border border-gray-200 bg-white">
          {/* Table header */}
          <div className="grid grid-cols-[1fr_160px_140px_120px_100px_32px] gap-4 border-b border-gray-100 px-4 py-3 text-xs font-medium uppercase tracking-wide text-gray-400">
            <div>Campaign</div>
            <div>Challenge period</div>
            <div>Reward types</div>
            <div>Total rewards</div>
            <div>Status</div>
            <div />
          </div>

          {filtered.length === 0 ? (
            <div className="px-4 py-8 text-center text-sm text-gray-400">
              No campaigns found.
            </div>
          ) : (
            filtered.map((campaign) => (
              <CampaignRow key={campaign.id} campaign={campaign} />
            ))
          )}
        </div>
      )}
    </div>
  );
}

function CampaignRow({ campaign }: { campaign: AdminCampaignListItem }) {
  const isActive = campaign.status === "active";
  const isDraft = campaign.status === "draft";

  return (
    <Link
      href={`/campaigns/${campaign.id}`}
      className={`grid grid-cols-[1fr_160px_140px_120px_100px_32px] items-center gap-4 border-b border-gray-50 px-4 py-3.5 transition-colors hover:bg-gray-50 ${
        !isActive && !isDraft ? "opacity-70" : ""
      } ${isDraft ? "opacity-50" : ""}`}
    >
      {/* Campaign name */}
      <div className="flex items-center gap-3">
        <div
          className={`flex h-12 w-12 shrink-0 items-center justify-center rounded-xl text-xl ${
            isActive
              ? "bg-gradient-to-br from-orange-500 to-orange-300"
              : isDraft
                ? "bg-gray-100"
                : "bg-gray-100"
          }`}
        >
          {isDraft ? "📝" : "🏆"}
        </div>
        <div>
          <div className="text-sm font-semibold text-gray-900">
            {campaign.name}
          </div>
          <div className="text-xs text-gray-400">
            {campaign.reward_types_count > 0
              ? `${campaign.reward_types_count} reward types`
              : "Not configured"}
          </div>
        </div>
      </div>

      {/* Period */}
      <div className="text-sm text-gray-600">
        {formatPeriod(campaign.challenge_start, campaign.challenge_end)}
      </div>

      {/* Reward types */}
      <div className="flex flex-wrap gap-1">
        {campaign.reward_types_count > 0 ? (
          <>
            <RewardTypeBadge type="gems" />
            <RewardTypeBadge type="gift_card" />
          </>
        ) : (
          <span className="text-sm text-gray-400">—</span>
        )}
      </div>

      {/* Total rewards */}
      <div className="text-sm text-gray-600">
        {campaign.total_stock > 0
          ? campaign.status === "completed"
            ? `${campaign.distributed_count} / ${campaign.total_stock} distributed`
            : campaign.total_stock
          : "—"}
      </div>

      {/* Status */}
      <StatusBadge status={campaign.status} />

      {/* Chevron */}
      <div className="text-gray-400">›</div>
    </Link>
  );
}
