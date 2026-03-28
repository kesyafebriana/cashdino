import { API_BASE_URL } from "@/constants/api";

export async function fetchHealth(): Promise<{ status: string }> {
  const res = await fetch(`${API_BASE_URL}/api/health`);
  return res.json();
}
