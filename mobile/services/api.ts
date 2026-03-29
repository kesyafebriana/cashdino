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

export async function fetchBanner(userId: string): Promise<BannerResponse> {
  const res = await fetch(
    `${API_BASE_URL}/api/challenge/banner?user_id=${userId}`
  );
  if (!res.ok) {
    throw new Error(`Failed to fetch banner: ${res.status}`);
  }
  return res.json();
}
