"use client";

import { Suspense, useState, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import {
  createCampaign,
  updateCampaign,
  deleteCampaign,
  fetchCampaign,
  type CreateRewardTypeInput,
  type CreateCampaignRuleInput,
  type AdminCampaignDetail,
} from "@/services/api";

const EMPTY_REWARD: CreateRewardTypeInput = {
  name: "",
  type: "gems",
  value: 0,
  image: null,
  stock: 1,
};

const EMPTY_RULE: CreateCampaignRuleInput = {
  rank_from: 1,
  rank_to: 1,
  reward_type_indexes: [],
};

function computeEndDate(startDate: string): string {
  if (!startDate) return "";
  const d = new Date(startDate + "T00:00:00Z");
  const end = new Date(d.getTime() + 6 * 24 * 60 * 60 * 1000);
  const months = ["Jan","Feb","Mar","Apr","May","Jun","Jul","Aug","Sep","Oct","Nov","Dec"];
  return `${months[end.getUTCMonth()]} ${end.getUTCDate()}, 23:59 UTC (Sunday)`;
}

function getNextMonday(): string {
  const now = new Date();
  const day = now.getUTCDay();
  const daysUntilMonday = day === 0 ? 1 : 8 - day;
  const next = new Date(now.getTime() + daysUntilMonday * 24 * 60 * 60 * 1000);
  return next.toISOString().slice(0, 10);
}

function rewardBadgeStyle(type: string) {
  if (type === "gems") return "bg-green-50 text-green-700 border-green-300";
  if (type === "gift_card") return "bg-orange-50 text-[#e65100] border-orange-300";
  return "bg-pink-50 text-red-700 border-red-300";
}

function buildRulesFromDetail(
  detail: AdminCampaignDetail
): CreateCampaignRuleInput[] {
  return detail.rules.map((rule) => {
    const indexes: number[] = [];
    for (const rt of rule.reward_types) {
      const idx = detail.reward_types.findIndex((t) => t.name === rt.name);
      if (idx >= 0 && !indexes.includes(idx)) indexes.push(idx);
    }
    return {
      rank_from: rule.rank_from,
      rank_to: rule.rank_to,
      reward_type_indexes: indexes,
    };
  });
}

export default function CampaignFormPageWrapper() {
  return (
    <Suspense fallback={<div className="py-12 text-center text-sm text-gray-400">Loading...</div>}>
      <CampaignFormPage />
    </Suspense>
  );
}

function CampaignFormPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const editId = searchParams.get("edit");
  const isEdit = !!editId;

  const [loading, setLoading] = useState(isEdit);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [challengeId, setChallengeId] = useState("");

  const [name, setName] = useState("");
  const [startDate, setStartDate] = useState(getNextMonday());
  const [bannerImage, setBannerImage] = useState("");
  const [emailSubject, setEmailSubject] = useState(
    "Congratulations! You won a reward from CashDino!"
  );
  const [emailBody, setEmailBody] = useState(
    "Hi {{username}},\n\nCongratulations! You placed #{{rank}} in this week's challenge and won: {{reward_type}} ({{reward_value}}).\n\n{{reward_image}}\n\nReply to this email to claim your reward."
  );
  const [rewardTypes, setRewardTypes] = useState<CreateRewardTypeInput[]>([
    { ...EMPTY_REWARD },
  ]);
  const [rules, setRules] = useState<CreateCampaignRuleInput[]>([
    { ...EMPTY_RULE },
  ]);

  const [deleting, setDeleting] = useState(false);

  // Load existing campaign for edit mode
  useEffect(() => {
    if (!editId) return;
    fetchCampaign(editId)
      .then((c) => {
        setName(c.name);
        setBannerImage(c.banner_image);
        setChallengeId(c.challenge_id);
        setEmailSubject(c.non_gem_claim_email_subject);
        setEmailBody(c.non_gem_claim_email_body);
        setRewardTypes(
          c.reward_types.map((rt) => ({
            name: rt.name,
            type: rt.type,
            value: rt.value,
            image: rt.image,
            stock: rt.stock,
          }))
        );
        setRules(buildRulesFromDetail(c));
      })
      .catch(() => setError("Failed to load campaign"))
      .finally(() => setLoading(false));
  }, [editId]);

  // Reward type helpers
  const updateRT = (i: number, field: string, val: string | number | null) =>
    setRewardTypes(rewardTypes.map((rt, idx) => (idx === i ? { ...rt, [field]: val } : rt)));
  const removeRT = (i: number) => {
    setRewardTypes(rewardTypes.filter((_, idx) => idx !== i));
    setRules(rules.map((r) => ({
      ...r,
      reward_type_indexes: r.reward_type_indexes
        .filter((ri) => ri !== i)
        .map((ri) => (ri > i ? ri - 1 : ri)),
    })));
  };

  // Rule helpers
  const updateRule = (i: number, field: string, val: number | number[]) =>
    setRules(rules.map((r, idx) => (idx === i ? { ...r, [field]: val } : r)));
  const toggleRewardInRule = (ruleIdx: number, rtIdx: number) => {
    const curr = rules[ruleIdx].reward_type_indexes;
    updateRule(
      ruleIdx,
      "reward_type_indexes",
      curr.includes(rtIdx) ? curr.filter((i) => i !== rtIdx) : [...curr, rtIdx]
    );
  };

  const handleSave = async () => {
    setSaving(true);
    setError(null);
    try {
      const payload = {
        challenge_id: challengeId || "",
        start_date: startDate,
        name,
        banner_image: bannerImage,
        reward_types: rewardTypes,
        rules,
        non_gem_claim_email_subject: emailSubject,
        non_gem_claim_email_body: emailBody,
      };
      if (isEdit) {
        await updateCampaign(editId!, payload);
      } else {
        await createCampaign(payload);
      }
      router.push("/campaigns");
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to save");
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return <div className="py-12 text-center text-sm text-gray-400">Loading campaign...</div>;
  }

  return (
    <div>
      {/* Breadcrumb */}
      <div className="mb-6 flex items-center gap-1.5 text-[13px]">
        <Link href="/campaigns" className="text-gray-400 hover:text-gray-600">Campaigns</Link>
        <span className="text-gray-300">›</span>
        <span className="font-medium text-gray-900">{isEdit ? "Edit campaign" : "New campaign"}</span>
      </div>

      <h1 className="mb-7 text-[26px] font-bold text-gray-900">
        {isEdit ? "Edit campaign" : "Create new campaign"}
      </h1>

      <div className="grid max-w-[1100px] grid-cols-2 gap-6">
        {/* LEFT COLUMN */}
        <div className="flex flex-col gap-6">
          {/* Basic info */}
          <Card title="Basic info">
            <Field label="Campaign name">
              <input value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g. Week 14 — Easter" className="input" />
            </Field>
            <div className="mt-3.5 grid grid-cols-2 gap-3.5">
              <Field label="Start date (Monday only)">
                <input
                  type="date"
                  value={startDate}
                  onChange={(e) => setStartDate(e.target.value)}
                  min={getNextMonday()}
                  className="input"
                />
              </Field>
              <Field label="End date (auto-calculated)">
                <div className="rounded-lg border border-gray-200 bg-gray-50 px-3 py-2 text-sm text-gray-500">
                  {startDate ? computeEndDate(startDate) : "Pick a start date"}
                </div>
              </Field>
            </div>
            <Field label="Banner image URL" className="mt-3.5">
              <input value={bannerImage} onChange={(e) => setBannerImage(e.target.value)} placeholder="https://cdn.example.com/banner.png" className="input" />
            </Field>
          </Card>

          {/* Reward types table */}
          <Card
            title="Reward types"
            action={<button onClick={() => setRewardTypes([...rewardTypes, { ...EMPTY_REWARD }])} className="btn-secondary text-xs">+ Add type</button>}
          >
            <div className="mb-2 grid grid-cols-[2fr_100px_90px_60px_1.5fr_28px] gap-2 text-[11px] font-semibold uppercase tracking-wide text-gray-400">
              <span>Name</span><span>Type</span><span>Value</span><span>Stock</span><span>Image URL</span><span />
            </div>
            {rewardTypes.map((rt, i) => (
              <div key={i} className="mb-2 grid grid-cols-[2fr_100px_90px_60px_1.5fr_28px] items-center gap-2">
                <input value={rt.name} onChange={(e) => updateRT(i, "name", e.target.value)} placeholder="Name" className="input" />
                <select value={rt.type} onChange={(e) => updateRT(i, "type", e.target.value)} className="input text-xs">
                  <option value="gems">gems</option>
                  <option value="gift_card">gift_card</option>
                  <option value="cash">cash</option>
                  <option value="other">other</option>
                </select>
                <input type="number" value={rt.value || ""} onChange={(e) => updateRT(i, "value", Number(e.target.value))} className="input" />
                <input type="number" value={rt.stock || ""} onChange={(e) => updateRT(i, "stock", Number(e.target.value))} className="input" />
                <input value={rt.image ?? ""} onChange={(e) => updateRT(i, "image", e.target.value || null)} placeholder="optional" className="input text-xs" />
                <button onClick={() => removeRT(i)} className="flex items-center justify-center text-gray-400 hover:text-red-500">✕</button>
              </div>
            ))}
          </Card>
        </div>

        {/* RIGHT COLUMN */}
        <div className="flex flex-col gap-6">
          {/* Rank rules */}
          <Card
            title="Rank rules"
            action={<button onClick={() => setRules([...rules, { ...EMPTY_RULE }])} className="btn-secondary text-xs">+ Add rule</button>}
          >
            <div className="flex flex-col gap-2.5">
              {rules.map((rule, ri) => {
                const userCount = rule.rank_to - rule.rank_from + 1;
                return (
                  <div key={ri} className="rounded-xl border border-gray-200 p-4">
                    <div className="mb-3 flex items-center gap-2.5">
                      <span className="text-xs font-medium text-gray-500">Rank</span>
                      <input type="number" value={rule.rank_from || ""} onChange={(e) => updateRule(ri, "rank_from", Number(e.target.value))} className="input w-[60px] text-center" />
                      <span className="text-xs text-gray-400">to</span>
                      <input type="number" value={rule.rank_to || ""} onChange={(e) => updateRule(ri, "rank_to", Number(e.target.value))} className="input w-[60px] text-center" />
                      {userCount > 1 && <span className="text-[11px] text-gray-400">{userCount} users</span>}
                      {rules.length > 1 && (
                        <button onClick={() => setRules(rules.filter((_, idx) => idx !== ri))} className="ml-auto text-gray-400 hover:text-red-500">✕</button>
                      )}
                    </div>
                    <div className="mb-1.5 text-[11px] text-gray-400">Assigned rewards:</div>
                    <div className="flex flex-wrap gap-1.5">
                      {rewardTypes.map((rt, rtIdx) => {
                        const selected = rule.reward_type_indexes.includes(rtIdx);
                        return (
                          <button
                            key={rtIdx}
                            onClick={() => toggleRewardInRule(ri, rtIdx)}
                            className={`rounded-md border px-3 py-1 text-xs font-medium transition-colors ${
                              selected ? rewardBadgeStyle(rt.type) : "border-dashed border-gray-300 text-gray-400 hover:bg-gray-50"
                            }`}
                          >
                            {selected ? rt.name || `Type ${rtIdx + 1}` : `+ ${rt.name || `Type ${rtIdx + 1}`}`}
                          </button>
                        );
                      })}
                    </div>
                  </div>
                );
              })}
            </div>
          </Card>

          {/* Email template */}
          <Card title="Claim email template" subtitle="Sent to users who win non-gem rewards">
            <Field label="Subject">
              <input value={emailSubject} onChange={(e) => setEmailSubject(e.target.value)} className="input" />
            </Field>
            <Field label="Body" className="mt-3.5">
              <textarea value={emailBody} onChange={(e) => setEmailBody(e.target.value)} rows={5} className="input resize-y leading-relaxed" />
            </Field>
            <div className="mt-2 flex flex-wrap gap-1">
              <span className="text-[11px] text-gray-400">Placeholders:</span>
              {["username", "rank", "reward_type", "reward_value", "reward_image"].map((p) => (
                <code key={p} className="rounded bg-gray-100 px-1.5 py-0.5 text-[11px] text-gray-500">{`{{${p}}}`}</code>
              ))}
            </div>
          </Card>
        </div>
      </div>

      {/* Error */}
      {error && (
        <p className="mt-4 max-w-[1100px] rounded-lg bg-red-50 px-4 py-2 text-sm text-red-600">{error}</p>
      )}

      {/* Actions */}
      <div className="mt-6 flex max-w-[1100px] items-center gap-3">
        {isEdit && (
          <button
            onClick={async () => {
              if (!confirm("Delete this campaign? This cannot be undone.")) return;
              setDeleting(true);
              try {
                await deleteCampaign(editId!);
                router.push("/campaigns");
              } catch (err: unknown) {
                setError(err instanceof Error ? err.message : "Failed to delete");
                setDeleting(false);
              }
            }}
            disabled={deleting}
            className="rounded-lg border border-red-300 px-6 py-2.5 text-sm font-medium text-red-600 hover:bg-red-50 disabled:opacity-50 transition-colors"
          >
            {deleting ? "Deleting..." : "Delete campaign"}
          </button>
        )}
        <div className="flex-1" />
        <Link href="/campaigns" className="rounded-lg border border-gray-300 px-6 py-2.5 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors">
          Cancel
        </Link>
        <button
          onClick={handleSave}
          disabled={saving}
          className="rounded-lg bg-[#e65100] px-6 py-2.5 text-sm font-semibold text-white hover:bg-[#bf360c] disabled:opacity-50 transition-colors"
        >
          {saving ? "Saving..." : isEdit ? "Update campaign" : "Save campaign"}
        </button>
      </div>
    </div>
  );
}

function Card({
  title,
  subtitle,
  action,
  children,
}: {
  title: string;
  subtitle?: string;
  action?: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <div className="rounded-xl border border-gray-200 bg-white p-[22px]">
      <div className="mb-[18px] flex items-center justify-between">
        <div>
          <div className="text-[15px] font-semibold text-gray-900">{title}</div>
          {subtitle && <div className="mt-0.5 text-xs text-gray-500">{subtitle}</div>}
        </div>
        {action}
      </div>
      {children}
    </div>
  );
}

function Field({
  label,
  className,
  children,
}: {
  label: string;
  className?: string;
  children: React.ReactNode;
}) {
  return (
    <div className={className}>
      <label className="mb-1 block text-xs font-medium text-gray-500">{label}</label>
      {children}
    </div>
  );
}
