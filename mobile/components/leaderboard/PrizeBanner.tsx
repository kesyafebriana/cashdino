import { View, Text, Pressable } from "react-native";
import type { RewardsSummaryRow } from "@/services/api";

interface PrizeBannerProps {
  rewardsSummary: RewardsSummaryRow[];
  onSeeAll?: () => void;
}

function buildSubtitle(summary: RewardsSummaryRow[]): string {
  if (summary.length === 0) return "Earn gems to win rewards";

  const maxRank = Math.max(...summary.map((r) => r.rank_to));

  const rewardTypes = new Set<string>();
  for (const row of summary) {
    for (const reward of row.rewards) {
      rewardTypes.add(reward.type);
    }
  }

  const labels: string[] = [];
  if (rewardTypes.has("gem")) labels.push("gems");
  if (rewardTypes.has("gift_card")) labels.push("gift cards");
  if (rewardTypes.has("cash")) labels.push("cash");
  if (rewardTypes.has("other")) labels.push("rewards");

  const rewardText = labels.length > 0 ? labels.join(" + ") : "rewards";
  return `Top ${maxRank} players earn ${rewardText}`;
}

export function PrizeBanner({ rewardsSummary, onSeeAll }: PrizeBannerProps) {
  const rewardCount = rewardsSummary.reduce(
    (sum, row) => sum + row.rewards.length,
    0
  );
  const title =
    rewardCount > 0 ? "Win amazing rewards!" : "Rewards coming soon";

  return (
    <Pressable
      onPress={onSeeAll}
      className="mx-4 rounded-2xl overflow-hidden"
      style={{ marginTop: 12 }}
    >
      <View
        className="flex-row items-center"
        style={{ backgroundColor: "#ff8f00", padding: 14, paddingHorizontal: 16, gap: 12 }}
      >
        <Text style={{ fontSize: 28 }}>🏆</Text>
        <View className="flex-1">
          <Text className="font-bold text-white" style={{ fontSize: 14 }}>
            {title}
          </Text>
          <Text
            style={{ fontSize: 11, color: "rgba(255,255,255,0.85)", marginTop: 2 }}
          >
            {buildSubtitle(rewardsSummary)}
          </Text>
        </View>
        <Text
          style={{
            fontSize: 11,
            color: "rgba(255,255,255,0.85)",
            textDecorationLine: "underline",
          }}
        >
          See all →
        </Text>
      </View>
    </Pressable>
  );
}
