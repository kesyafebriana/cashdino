"use client";

import { useState, useEffect, use } from "react";
import Link from "next/link";
import {
  fetchCampaign,
  fetchDistributions,
  type AdminCampaignDetail,
  type AdminCampaignRuleDetail,
  type AdminDistributionRow,
  type RewardType,
} from "@/services/api";
import { StatusBadge } from "@/components/StatusBadge";

function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  const months = [
    "Jan","Feb","Mar","Apr","May","Jun",
    "Jul","Aug","Sep","Oct","Nov","Dec",
  ];
  return `${months[d.getUTCMonth()]} ${d.getUTCDate()}`;
}

function formatRankLabel(from: number, to: number): string {
  if (from === to) return `Rank #${from}`;
  return `Rank #${from} – #${to}`;
}

function rewardIcon(type: string) {
  if (type === "gems") return "💎";
  if (type === "gift_card") return "🎫";
  return "🎁";
}

function rewardIconBg(type: string) {
  if (type === "gems") return "bg-green-50";
  if (type === "gift_card") return "bg-orange-50";
  return "bg-pink-50";
}

function rewardDotColor(type: string) {
  if (type === "gems") return "bg-green-500";
  if (type === "gift_card") return "bg-orange-400";
  return "bg-pink-400";
}

// --- Sub-components ---

function RewardTypesTable({ types }: { types: RewardType[] }) {
  return (
    <div>
      <h2 className="mb-2.5 text-sm font-semibold text-gray-900">
        Reward types
      </h2>
      <div className="overflow-hidden rounded-xl border border-gray-200 bg-white">
        <div className="grid grid-cols-[2fr_1fr_0.7fr_0.7fr] gap-2 border-b border-gray-200 bg-gray-50 px-3.5 py-2 text-[11px] font-semibold uppercase tracking-wide text-gray-400">
          <span>Name</span>
          <span>Type</span>
          <span>Value</span>
          <span>Stock</span>
        </div>
        {types.map((rt, i) => (
          <div
            key={rt.id}
            className={`grid grid-cols-[2fr_1fr_0.7fr_0.7fr] items-center gap-2 px-3.5 py-2.5 ${
              i < types.length - 1 ? "border-b border-gray-100" : ""
            }`}
          >
            <div className="flex items-center gap-2">
              <div
                className={`flex h-6 w-6 items-center justify-center rounded text-xs ${rewardIconBg(rt.type)}`}
              >
                {rewardIcon(rt.type)}
              </div>
              <span className="text-[13px] font-medium text-gray-900">
                {rt.name}
              </span>
            </div>
            <span className="text-xs text-gray-500">{rt.type}</span>
            <span className="text-[13px] text-gray-900">
              {rt.type === "gems"
                ? rt.value.toLocaleString()
                : `$${rt.value}`}
            </span>
            <span className="text-[13px] text-gray-900">{rt.stock}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

function RankRulesCards({ rules }: { rules: AdminCampaignRuleDetail[] }) {
  return (
    <div>
      <h2 className="mb-2.5 text-sm font-semibold text-gray-900">
        Rank rules
      </h2>
      <div className="flex flex-col gap-2">
        {rules.map((rule, i) => {
          const userCount = rule.rank_to - rule.rank_from + 1;
          return (
            <div
              key={i}
              className="rounded-xl border border-gray-200 bg-white p-3.5"
            >
              <div className="mb-2.5 flex items-center gap-2">
                <span className="rounded-md bg-[#e65100] px-3 py-0.5 text-xs font-bold text-white">
                  {formatRankLabel(rule.rank_from, rule.rank_to)}
                </span>
                {userCount > 1 && (
                  <span className="text-[11px] text-gray-400">
                    {userCount} users
                  </span>
                )}
              </div>
              <div className="flex flex-col gap-1">
                {rule.reward_types.map((rt, j) => (
                  <div
                    key={j}
                    className="flex items-center gap-1.5 text-xs text-gray-900"
                  >
                    <div
                      className={`h-1.5 w-1.5 rounded-full ${rewardDotColor(rt.type)}`}
                    />
                    <span className="font-medium">{rt.name}</span>
                    {rt.type !== "gems" && (
                      <span className="text-gray-900">
                        — ${rt.value}
                      </span>
                    )}
                    <span className="text-gray-400">
                      ({rt.type === "gems" ? "auto-credit" : "email"})
                    </span>
                  </div>
                ))}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

function EmailTemplateCard({
  subject,
  body,
}: {
  subject: string;
  body: string;
}) {
  return (
    <div>
      <h2 className="mb-2.5 text-sm font-semibold text-gray-900">
        Claim email template
      </h2>
      <div className="rounded-xl border border-gray-200 bg-white p-4">
        <div className="mb-3">
          <div className="mb-1 text-[11px] font-semibold uppercase tracking-wide text-gray-400">
            Subject
          </div>
          <div className="rounded-md bg-gray-50 px-3 py-2 text-[13px] text-gray-900">
            {subject || "—"}
          </div>
        </div>
        <div>
          <div className="mb-1 text-[11px] font-semibold uppercase tracking-wide text-gray-400">
            Body
          </div>
          <div className="whitespace-pre-wrap rounded-md bg-gray-50 px-3 py-2.5 text-[13px] leading-relaxed text-gray-900">
            {body || "—"}
          </div>
        </div>
      </div>
    </div>
  );
}

function DistributionInfo({
  campaign,
  distributions,
}: {
  campaign: AdminCampaignDetail;
  distributions: AdminDistributionRow[];
}) {
  if (campaign.status === "active") {
    return (
      <div className="mt-5 flex items-center gap-2.5 rounded-xl bg-blue-50 px-4 py-3.5">
        <span className="text-blue-700">ℹ️</span>
        <span className="text-[13px] text-blue-700">
          Rewards will be automatically distributed when the challenge ends.
        </span>
      </div>
    );
  }

  if (distributions.length === 0) {
    return (
      <div className="mt-5 rounded-xl bg-gray-50 px-4 py-3 text-sm text-gray-500">
        No distributions found.
      </div>
    );
  }

  return (
    <div className="mt-5">
      <h2 className="mb-2.5 text-sm font-semibold text-gray-900">
        Distributions
      </h2>
      <div className="overflow-hidden rounded-xl border border-gray-200 bg-white">
        <div className="grid grid-cols-[60px_1fr_1fr_100px_100px] gap-3 border-b border-gray-200 bg-gray-50 px-3.5 py-2 text-[11px] font-semibold uppercase tracking-wide text-gray-400">
          <div>Rank</div>
          <div>User</div>
          <div>Reward</div>
          <div>Status</div>
          <div>Delivered</div>
        </div>
        {distributions.map((d) => (
          <div
            key={d.id}
            className="grid grid-cols-[60px_1fr_1fr_100px_100px] items-center gap-3 border-b border-gray-50 px-3.5 py-2.5"
          >
            <div className="text-sm font-semibold text-gray-700">
              #{d.final_rank}
            </div>
            <div>
              <div className="text-sm font-medium">{d.display_name}</div>
              <div className="text-xs text-gray-400">{d.masked_email}</div>
            </div>
            <div className="text-sm text-gray-600">{d.reward_name}</div>
            <StatusBadge status={d.status} />
            <div className="text-xs text-gray-400">
              {d.delivered_at
                ? new Date(d.delivered_at).toLocaleDateString()
                : "—"}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// --- Main page ---

export default function CampaignDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const [campaign, setCampaign] = useState<AdminCampaignDetail | null>(null);
  const [distributions, setDistributions] = useState<AdminDistributionRow[]>(
    []
  );
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([fetchCampaign(id), fetchDistributions(id).catch(() => [])])
      .then(([c, d]) => {
        setCampaign(c);
        setDistributions(d ?? []);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) {
    return (
      <div className="py-12 text-center text-sm text-gray-400">
        Loading campaign...
      </div>
    );
  }

  if (!campaign) {
    return (
      <div className="py-12 text-center text-sm text-gray-400">
        Campaign not found.
      </div>
    );
  }

  return (
    <div>
      {/* Breadcrumb */}
      <div className="mb-5 flex items-center gap-1.5 text-[13px]">
        <Link
          href="/campaigns"
          className="text-gray-400 hover:text-gray-600"
        >
          Campaigns
        </Link>
        <span className="text-gray-300">›</span>
        <span className="font-medium text-gray-900">{campaign.name}</span>
      </div>

      {/* Header */}
      <div className="mb-6 flex items-start justify-between">
        <div>
          <div className="mb-1 flex items-center gap-2.5">
            <h1 className="text-[22px] font-bold text-gray-900">
              {campaign.name}
            </h1>
            <StatusBadge status={campaign.status} />
          </div>
          <p className="text-[13px] text-gray-500">
            Linked to challenge {campaign.challenge_id.slice(0, 8)}...
          </p>
        </div>
        {campaign.status === "draft" && (
          <Link
            href={`/campaigns/new?edit=${campaign.id}`}
            className="rounded-lg border border-gray-300 px-4 py-2 text-[13px] font-medium text-gray-700 hover:bg-gray-50 transition-colors"
          >
            Edit campaign
          </Link>
        )}
      </div>

      {/* Three column layout */}
      <div className="grid grid-cols-3 gap-5">
        <RewardTypesTable types={campaign.reward_types} />
        <RankRulesCards rules={campaign.rules} />
        <EmailTemplateCard
          subject={campaign.non_gem_claim_email_subject}
          body={campaign.non_gem_claim_email_body}
        />
      </div>

      {/* Distribution info */}
      <DistributionInfo
        campaign={campaign}
        distributions={distributions}
      />
    </div>
  );
}
