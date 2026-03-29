import { View, Text } from "react-native";
import { GemStar } from "@/components/GemStar";

interface LeaderboardRowProps {
  rank: number;
  displayName: string;
  weeklyGems: number;
  rankChange?: number | null;
  isCurrentUser?: boolean;
}

const MEDAL: Record<number, { emoji: string; bg: string }> = {
  1: { emoji: "\u{1F947}", bg: "#fffde7" },
  2: { emoji: "\u{1F948}", bg: "#f5f5f5" },
  3: { emoji: "\u{1F949}", bg: "#fff3e0" },
};

function RankChangeIndicator({ change }: { change?: number | null }) {
  if (change == null) return null;

  if (change > 0) {
    return (
      <Text className="font-medium" style={{ fontSize: 11, color: "#2e7d32" }}>
        ▲ {change}
      </Text>
    );
  }
  if (change < 0) {
    return (
      <Text className="font-medium" style={{ fontSize: 11, color: "#d32f2f" }}>
        ▼ {Math.abs(change)}
      </Text>
    );
  }
  return (
    <Text className="font-medium" style={{ fontSize: 11, color: "#888" }}>
      —
    </Text>
  );
}

export function LeaderboardRow({
  rank,
  displayName,
  weeklyGems,
  rankChange,
  isCurrentUser,
}: LeaderboardRowProps) {
  const medal = MEDAL[rank];

  const rowBg = isCurrentUser
    ? "#e8f5e9"
    : medal?.bg ?? "transparent";

  const borderStyle = isCurrentUser
    ? { borderWidth: 1.5, borderColor: "#81c784" }
    : {};

  return (
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
          style={{ fontSize: 14, color: "#888", width: 32, textAlign: "center" }}
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

      {/* Gems + rank change */}
      <View className="items-end">
        <View className="flex-row items-center" style={{ gap: 3 }}>
          <Text
            style={{
              fontSize: 14,
              fontWeight: medal ? "700" : "600",
              color: "#1a1a1a",
            }}
          >
            {weeklyGems.toLocaleString()}
          </Text>
          <GemStar size={12} />
        </View>
        <RankChangeIndicator change={rankChange} />
      </View>
    </View>
  );
}
