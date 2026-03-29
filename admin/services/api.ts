const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

// --- Types ---

export interface AdminCampaignListItem {
  id: string;
  challenge_id: string;
  name: string;
  banner_image: string;
  status: string;
  challenge_start: string;
  challenge_end: string;
  reward_types_count: number;
  total_stock: number;
  distributed_count: number;
}

export interface RewardType {
  id: string;
  campaign_id: string;
  name: string;
  type: string;
  value: number;
  image: string | null;
  stock: number;
}

export interface RewardInfo {
  name: string;
  image: string | null;
  value: number;
  type: string;
}

export interface AdminCampaignRuleDetail {
  rank_from: number;
  rank_to: number;
  reward_names: string[];
  reward_types: RewardInfo[];
}

export interface AdminCampaignDetail {
  id: string;
  challenge_id: string;
  name: string;
  banner_image: string;
  status: string;
  non_gem_claim_email_subject: string;
  non_gem_claim_email_body: string;
  reward_types: RewardType[];
  rules: AdminCampaignRuleDetail[];
}

export interface CreateRewardTypeInput {
  name: string;
  type: string;
  value: number;
  image: string | null;
  stock: number;
}

export interface CreateCampaignRuleInput {
  rank_from: number;
  rank_to: number;
  reward_type_indexes: number[];
}

export interface CreateCampaignRequest {
  challenge_id: string;
  start_date?: string; // YYYY-MM-DD, Monday only, future
  name: string;
  banner_image: string;
  reward_types: CreateRewardTypeInput[];
  rules: CreateCampaignRuleInput[];
  non_gem_claim_email_subject: string;
  non_gem_claim_email_body: string;
}

export interface AdminDistributionRow {
  id: string;
  user_id: string;
  display_name: string;
  masked_email: string;
  reward_name: string;
  reward_type: string;
  reward_value: number;
  reward_image: string | null;
  status: string;
  delivered_at: string | null;
  email_sent_at: string | null;
  final_rank: number;
}

export interface WeeklyResetResponse {
  status: string;
  results_archived: number;
  rewards_distributed: number;
  new_challenge_id: string;
}

// --- API functions ---

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`, options);
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error || `Request failed: ${res.status}`);
  }
  return res.json();
}

export interface WeeklyChallenge {
  id: string;
  start_time: string;
  end_time: string;
  status: string;
}

export function fetchChallenges(): Promise<WeeklyChallenge[]> {
  return apiFetch("/api/admin/challenges");
}

export function fetchCampaigns(): Promise<AdminCampaignListItem[]> {
  return apiFetch("/api/admin/campaigns");
}

export function fetchCampaign(id: string): Promise<AdminCampaignDetail> {
  return apiFetch(`/api/admin/campaigns/${id}`);
}

export function createCampaign(
  req: CreateCampaignRequest
): Promise<AdminCampaignDetail> {
  return apiFetch("/api/admin/campaigns", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
}

export function updateCampaign(
  id: string,
  req: CreateCampaignRequest
): Promise<AdminCampaignDetail> {
  return apiFetch(`/api/admin/campaigns/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
}

export function fetchDistributions(
  campaignId: string
): Promise<AdminDistributionRow[]> {
  return apiFetch(`/api/admin/campaigns/${campaignId}/distributions`);
}

export function retryDistribution(
  distributionId: string
): Promise<{ status: string }> {
  return apiFetch(`/api/admin/distributions/${distributionId}/retry`, {
    method: "POST",
  });
}

export function markDistributionDelivered(
  distributionId: string
): Promise<{ status: string }> {
  return apiFetch(`/api/admin/distributions/${distributionId}/deliver`, {
    method: "POST",
  });
}

export function deleteCampaign(id: string): Promise<{ status: string }> {
  return apiFetch(`/api/admin/campaigns/${id}`, { method: "DELETE" });
}

export function resetWeek(): Promise<WeeklyResetResponse> {
  return apiFetch("/api/admin/reset-week", { method: "POST" });
}
