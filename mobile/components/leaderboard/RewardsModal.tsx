import { View, Text, Modal, Pressable, ScrollView } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { GemStar } from "@/components/GemStar";
import type { RewardsSummaryRow } from "@/services/api";

interface RewardsModalProps {
  visible: boolean;
  onClose: () => void;
  rewardsSummary: RewardsSummaryRow[];
}

const RANK_STYLE: Record<
  number,
  { emoji: string; bg: string; border: string; titleColor: string }
> = {
  1: { emoji: "\u{1F947}", bg: "#fffde7", border: "#fff176", titleColor: "#f57f17" },
  2: { emoji: "\u{1F948}", bg: "#f5f5f5", border: "#e0e0e0", titleColor: "#1a1a1a" },
  3: { emoji: "\u{1F949}", bg: "#fff3e0", border: "#ffe0b2", titleColor: "#e65100" },
};

function RankCard({
  rank,
  rewards,
}: {
  rank: number;
  rewards: RewardsSummaryRow["rewards"];
}) {
  const style = RANK_STYLE[rank];
  const bg = style?.bg ?? "#f5f5f5";
  const border = style?.border ?? "#e0e0e0";
  const titleColor = style?.titleColor ?? "#1a1a1a";

  return (
    <View
      className="rounded-2xl"
      style={{
        backgroundColor: bg,
        borderWidth: style ? 1 : 0.5,
        borderColor: border,
        padding: 14,
        marginBottom: 8,
      }}
    >
      {/* Rank header */}
      <View className="flex-row items-center" style={{ gap: 10, marginBottom: 10 }}>
        {style ? (
          <View
            className="rounded-full items-center justify-center"
            style={{
              width: 32,
              height: 32,
              backgroundColor: rank === 1 ? "#ffd600" : rank === 3 ? "#ffcc80" : "#e0e0e0",
            }}
          >
            <Text style={{ fontSize: 16 }}>{style.emoji}</Text>
          </View>
        ) : (
          <View
            className="rounded-full items-center justify-center"
            style={{ width: 32, height: 32, backgroundColor: "#e0e0e0" }}
          >
            <Text className="font-bold" style={{ fontSize: 13, color: "#666" }}>
              {rank}
            </Text>
          </View>
        )}
        <Text className="font-bold" style={{ fontSize: 15, color: titleColor }}>
          Rank #{rank}
        </Text>
      </View>

      {/* Rewards list */}
      <View style={{ gap: 6 }}>
        {rewards.map((reward, i) => (
          <View
            key={i}
            className="flex-row items-center rounded-xl bg-white"
            style={{ padding: 8, paddingHorizontal: 10, gap: 10 }}
          >
            <View
              className="rounded-md items-center justify-center"
              style={{
                width: 32,
                height: 32,
                backgroundColor: reward.type === "gem" ? "#e8f5e9" : "#fff3e0",
              }}
            >
              {reward.type === "gems" ? (
                <GemStar size={16} />
              ) : (
                <Text style={{ fontSize: 15 }}>🎫</Text>
              )}
            </View>
            <View className="flex-1">
              <Text className="font-semibold" style={{ fontSize: 13, color: "#1a1a1a" }}>
                {reward.name}
              </Text>
              <Text style={{ fontSize: 10, color: "#888" }}>
                {reward.type === "gems" ? "Auto-credited" : "Claim via email"}
              </Text>
            </View>
          </View>
        ))}
      </View>
    </View>
  );
}

export function RewardsModal({
  visible,
  onClose,
  rewardsSummary,
}: RewardsModalProps) {
  // Expand rank ranges into individual rank cards
  const rankCards: { rank: number; rewards: RewardsSummaryRow["rewards"] }[] = [];
  for (const row of rewardsSummary) {
    for (let r = row.rank_from; r <= row.rank_to; r++) {
      rankCards.push({ rank: r, rewards: row.rewards });
    }
  }

  const maxRewardedRank =
    rewardsSummary.length > 0
      ? Math.max(...rewardsSummary.map((r) => r.rank_to))
      : 0;

  return (
    <Modal
      visible={visible}
      transparent
      animationType="slide"
      statusBarTranslucent
      onRequestClose={onClose}
    >
      <View className="flex-1" style={{ backgroundColor: "rgba(0,0,0,0.45)" }}>
        {/* Tap overlay to close */}
        <Pressable className="flex-1" onPress={onClose} />

        {/* Bottom sheet */}
        <View
          style={{
            backgroundColor: "#fff",
            borderTopLeftRadius: 24,
            borderTopRightRadius: 24,
            maxHeight: "80%",
          }}
        >
          {/* Drag handle */}
          <View className="items-center" style={{ paddingTop: 10, paddingBottom: 4 }}>
            <View
              className="rounded-full"
              style={{ width: 36, height: 4, backgroundColor: "#ddd" }}
            />
          </View>

          {/* Header */}
          <View
            className="flex-row justify-between items-center"
            style={{ paddingHorizontal: 20, paddingTop: 8, paddingBottom: 16 }}
          >
            <Text className="font-bold" style={{ fontSize: 18, color: "#1a1a1a" }}>
              This week's rewards
            </Text>
            <Pressable
              onPress={onClose}
              className="items-center justify-center rounded-full"
              style={{ width: 28, height: 28, backgroundColor: "#f5f5f5" }}
            >
              <Ionicons name="close" size={14} color="#888" />
            </Pressable>
          </View>

          {/* Content */}
          <ScrollView
            style={{ paddingHorizontal: 16 }}
            contentContainerStyle={{ paddingBottom: 32 }}
            showsVerticalScrollIndicator={false}
          >
            {rankCards.map((card) => (
              <RankCard key={card.rank} rank={card.rank} rewards={card.rewards} />
            ))}

            {maxRewardedRank > 0 && (
              <View className="items-center" style={{ paddingTop: 8 }}>
                <Text style={{ fontSize: 12, color: "#888" }}>
                  Rank #{maxRewardedRank + 1} and below — no rewards this week
                </Text>
                <Text
                  className="font-medium"
                  style={{ fontSize: 12, color: "#e65100", marginTop: 4 }}
                >
                  Keep earning gems to climb the ranks!
                </Text>
              </View>
            )}
          </ScrollView>
        </View>
      </View>
    </Modal>
  );
}
