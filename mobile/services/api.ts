import { API_BASE_URL } from "@/constants/api";

export async function fetchHealth(): Promise<{ status: string }> {
  const res = await fetch(`${API_BASE_URL}/api/health`);
  return res.json();
}

export interface UserInfo {
  id: string;
  username: string;
  email: string;
}

export async function fetchUsers(usernames?: string[]): Promise<UserInfo[]> {
  const params = usernames
    ? `usernames=${usernames.join(",")}`
    : "limit=4";
  const res = await fetch(`${API_BASE_URL}/api/users?${params}`);
  if (!res.ok) {
    throw new Error(`Failed to fetch users: ${res.status}`);
  }
  return res.json();
}

export interface BannerResponse {
  challenge_id: string;
  end_time: string;
  total_gems: number;
  weekly_gems: number;
  rank_display: string;
  gap_to_next: number | null;
  display_name: string;
}

export interface EarnGemsResponse {
  user_id: string;
  weekly_gems: number;
}

export async function earnGems(
  userId: string,
  amount: number,
  source: string = "gameplay",
  gameName?: string
): Promise<EarnGemsResponse> {
  const res = await fetch(`${API_BASE_URL}/api/gems/earn`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      user_id: userId,
      source,
      amount,
      ...(gameName ? { game_name: gameName } : {}),
    }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error || `Failed to earn gems: ${res.status}`);
  }
  return res.json();
}

export interface CheckinResponse {
  gems_earned: number;
  current_streak: number;
  weekly_gems: number;
}

export async function checkin(userId: string): Promise<CheckinResponse> {
  const res = await fetch(`${API_BASE_URL}/api/checkin`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ user_id: userId }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error || `Failed to check in: ${res.status}`);
  }
  return res.json();
}

// --- Leaderboard types ---

export interface ChallengeInfo {
  id: string;
  start_time: string;
  end_time: string;
  status: string;
}

export interface CurrentLeaderboardRow {
  rank: number;
  display_name: string;
  weekly_gems: number;
}

export interface CurrentUserInfo {
  rank: number | null;
  rank_display: string;
  weekly_gems: number;
  gap_to_next: number | null;
  display_name: string;
}

export interface RewardInfo {
  name: string;
  image: string | null;
  value: number;
  type: string;
}

export interface RewardsSummaryRow {
  rank_from: number;
  rank_to: number;
  rewards: RewardInfo[];
}

export interface CampaignSummary {
  banner_image: string;
  rewards_summary: RewardsSummaryRow[];
}

export interface CurrentLeaderboardResponse {
  challenge: ChallengeInfo;
  leaderboard: CurrentLeaderboardRow[];
  current_user: CurrentUserInfo;
  campaign: CampaignSummary | null;
}

export async function fetchCurrentLeaderboard(
  userId: string
): Promise<CurrentLeaderboardResponse> {
  const res = await fetch(
    `${API_BASE_URL}/api/leaderboard/current?user_id=${userId}`
  );
  if (!res.ok) {
    throw new Error(`Failed to fetch leaderboard: ${res.status}`);
  }
  return res.json();
}

// --- Last week leaderboard types ---

export interface LastWeekChallengeInfo {
  id: string;
  start_time: string;
  end_time: string;
}

export interface LastWeekRow {
  rank: number;
  display_name: string;
  final_gems: number;
  rewards: RewardInfo[];
}

export interface LastWeekUserInfo {
  rank: number | null;
  rank_display: string;
  final_gems: number;
  rewards: RewardInfo[];
}

export interface LastWeekResponse {
  challenge: LastWeekChallengeInfo | null;
  leaderboard: LastWeekRow[];
  current_user: LastWeekUserInfo | null;
}

export async function fetchLastWeekLeaderboard(
  userId: string
): Promise<LastWeekResponse> {
  const res = await fetch(
    `${API_BASE_URL}/api/leaderboard/last-week?user_id=${userId}`
  );
  if (!res.ok) {
    throw new Error(`Failed to fetch last week leaderboard: ${res.status}`);
  }
  return res.json();
}

export async function fetchBanner(userId: string): Promise<BannerResponse> {
  const res = await fetch(
    `${API_BASE_URL}/api/challenge/banner?user_id=${userId}`
  );
  if (!res.ok) {
    throw new Error(`Failed to fetch banner: ${res.status}`);
  }
  return res.json();
}
