import { View, Text, Pressable } from "react-native";
import { GemStar } from "@/components/GemStar";
import type { RewardInfo } from "@/services/api";

interface LastWeekRowProps {
  rank: number;
  displayName: string;
  finalGems: number;
  rewards: RewardInfo[];
  isCurrentUser?: boolean;
  tooltipOpen?: boolean;
  onToggleTooltip?: () => void;
}

const MEDAL: Record<number, { emoji: string; bg: string }> = {
  1: { emoji: "\u{1F947}", bg: "#fffde7" },
  2: { emoji: "\u{1F948}", bg: "#f5f5f5" },
  3: { emoji: "\u{1F949}", bg: "#fff3e0" },
};

function RewardTooltip({
  rank,
  rewards,
}: {
  rank: number;
  rewards: RewardInfo[];
}) {
  return (
    <View
      className="rounded-2xl"
      style={{
        backgroundColor: "#fff",
        padding: 14,
        marginTop: 6,
        marginBottom: 4,
        borderWidth: 0.5,
        borderColor: "#e0e0e0",
        shadowColor: "#000",
        shadowOffset: { width: 0, height: 4 },
        shadowOpacity: 0.12,
        shadowRadius: 20,
        elevation: 8,
      }}
    >
      {/* Arrow */}
      <View
        style={{
          position: "absolute",
          top: -6,
          right: 24,
          width: 12,
          height: 12,
          backgroundColor: "#fff",
          borderLeftWidth: 0.5,
          borderTopWidth: 0.5,
          borderColor: "#e0e0e0",
          transform: [{ rotate: "45deg" }],
        }}
      />

      <Text
        className="font-bold"
        style={{ fontSize: 13, color: "#e65100", marginBottom: 10 }}
      >
        🏆 Rank #{rank} rewards
      </Text>

      <View style={{ gap: 8 }}>
        {rewards.map((reward, i) => (
          <View key={i} className="flex-row items-center" style={{ gap: 10 }}>
            <View
              className="rounded-lg items-center justify-center"
              style={{
                width: 36,
                height: 36,
                backgroundColor:
                  reward.type === "gems" ? "#e8f5e9" : "#fff3e0",
              }}
            >
              {reward.type === "gems" ? (
                <GemStar size={18} />
              ) : (
                <Text style={{ fontSize: 16 }}>🎫</Text>
              )}
            </View>
            <View>
              <Text
                className="font-semibold"
                style={{ fontSize: 13, color: "#1a1a1a" }}
              >
                {reward.name}
              </Text>
              <Text style={{ fontSize: 11, color: "#888" }}>
                {reward.type === "gems" ? "Auto-credited" : "Claim via email"}
              </Text>
            </View>
          </View>
        ))}
      </View>
    </View>
  );
}

export function LastWeekRow({
  rank,
  displayName,
  finalGems,
  rewards,
  isCurrentUser,
  tooltipOpen,
  onToggleTooltip,
}: LastWeekRowProps) {
  const medal = MEDAL[rank];
  const hasRewards = rewards.length > 0;

  const rowBg = isCurrentUser
    ? "#e8f5e9"
    : medal?.bg ?? "transparent";

  const borderStyle = isCurrentUser
    ? { borderWidth: 1.5, borderColor: "#81c784" }
    : {};

  return (
    <View style={{ zIndex: tooltipOpen ? 10 : 1 }}>
      <View
        className="flex-row items-center rounded-xl"
        style={{
          backgroundColor: rowBg,
          padding: 12,
          paddingLeft: medal ? 12 : 16,
          marginBottom: 2,
          ...borderStyle,
        }}
      >
        {/* Rank */}
        {medal ? (
          <Text style={{ fontSize: 20, width: 32, textAlign: "center" }}>
            {medal.emoji}
          </Text>
        ) : (
          <Text
            className="font-semibold"
            style={{
              fontSize: 14,
              color: "#888",
              width: 32,
              textAlign: "center",
            }}
          >
            {rank}
          </Text>
        )}

        {/* Display name */}
        <View className="flex-1" style={{ marginLeft: 8 }}>
          <Text
            style={{
              fontSize: 14,
              fontWeight: medal ? "600" : "500",
              color: "#1a1a1a",
            }}
          >
            {displayName}
            {isCurrentUser ? " (YOU)" : ""}
          </Text>
        </View>

        {/* Gems + gift button */}
        <View className="flex-row items-center" style={{ gap: 8 }}>
          <View className="flex-row items-center" style={{ gap: 3 }}>
            <Text
              style={{
                fontSize: 14,
                fontWeight: medal ? "700" : "600",
                color: "#1a1a1a",
              }}
            >
              {finalGems.toLocaleString()}
            </Text>
            <GemStar size={12} />
          </View>

          {hasRewards && (
            <Pressable
              onPress={onToggleTooltip}
              className="items-center justify-center rounded-lg"
              style={{
                width: 28,
                height: 28,
                backgroundColor: "#fff3e0",
                borderWidth: 1,
                borderColor: "#ffcc80",
              }}
            >
              <Text style={{ fontSize: 16 }}>🎁</Text>
            </Pressable>
          )}
        </View>
      </View>

      {/* Reward tooltip */}
      {tooltipOpen && hasRewards && (
        <RewardTooltip rank={rank} rewards={rewards} />
      )}
    </View>
  );
}
