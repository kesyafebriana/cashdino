import { View, Text, Pressable } from "react-native";
import { GemStar } from "@/components/GemStar";
import { Skeleton } from "@/components/Skeleton";

interface WeeklyChallengeBannerProps {
  rank: number | null;
  rankChange?: number;
  weeklyGems: number;
  timeLeft: string;
  gemsToNextRank?: number;
  nextRank?: number;
  notRanked?: boolean;
  onViewLeaderboard?: () => void;
}

export function WeeklyChallengeBanner({
  rank,
  rankChange,
  weeklyGems,
  timeLeft,
  gemsToNextRank,
  nextRank,
  notRanked,
  onViewLeaderboard,
}: WeeklyChallengeBannerProps) {
  const isRank1 = rank === 1;

  return (
    <View
      className="mx-4 mb-5 rounded-2xl p-4"
      style={{
        backgroundColor: "#fff3e0",
        borderWidth: 1.5,
        borderColor: "#ffb74d",
      }}
    >
      {/* Header */}
      <View className="flex-row justify-between items-center mb-3">
        <View className="flex-row items-center gap-2">
          <Text className="text-lg">🏆</Text>
          <Text className="text-base font-bold" style={{ color: "#e65100" }}>
            Weekly challenge
          </Text>
        </View>
        <View
          className="rounded-full px-3 py-1"
          style={{ backgroundColor: "#e65100" }}
        >
          <Text className="text-xs font-semibold text-white">{timeLeft}</Text>
        </View>
      </View>

      {/* Stats */}
      <View className="flex-row justify-between items-center mb-2">
        <View>
          <Text className="text-xs font-medium" style={{ color: "#bf360c" }}>
            Your rank
          </Text>
          <View className="flex-row items-baseline gap-1.5">
            <Text className="text-3xl font-bold" style={{ color: "#e65100" }}>
              {notRanked ? "Not Ranked" : rank != null ? `#${rank}` : "99+"}
            </Text>
            {rankChange != null && rankChange > 0 && (
              <Text className="text-xs font-semibold" style={{ color: "#2e7d32" }}>
                ▲ {rankChange}
              </Text>
            )}
            {rankChange != null && rankChange < 0 && (
              <Text className="text-xs font-semibold" style={{ color: "#d32f2f" }}>
                ▼ {Math.abs(rankChange)}
              </Text>
            )}
          </View>
        </View>
        <View className="items-end">
          <Text className="text-xs font-medium" style={{ color: "#bf360c" }}>
            Weekly gems
          </Text>
          <View className="flex-row items-center gap-1">
            <Text className="text-3xl font-bold" style={{ color: "#e65100" }}>
              {weeklyGems.toLocaleString()}
            </Text>
            <GemStar size={16} />
          </View>
        </View>
      </View>

      {/* Gap info */}
      {notRanked ? (
        <Text className="text-xs mb-3" style={{ color: "#bf360c" }}>
          Play games to earn gems and join the leaderboard!
        </Text>
      ) : isRank1 ? (
        <Text className="text-xs mb-3" style={{ color: "#2e7d32" }}>
          You're in the lead!
        </Text>
      ) : gemsToNextRank != null && nextRank != null ? (
        <Text className="text-xs mb-3" style={{ color: "#bf360c" }}>
          {gemsToNextRank.toLocaleString()} gems to reach{" "}
          <Text className="font-bold">#{nextRank}</Text>
        </Text>
      ) : null}

      {/* CTA Button */}
      <Pressable
        onPress={onViewLeaderboard}
        className="rounded-xl py-2.5 items-center"
        style={{ backgroundColor: "#e65100" }}
      >
        <Text className="text-sm font-semibold text-white">
          View leaderboard →
        </Text>
      </Pressable>
    </View>
  );
}

export function WeeklyChallengeSkeleton() {
  return (
    <View
      className="mx-4 mb-5 rounded-2xl p-4"
      style={{
        backgroundColor: "#fff3e0",
        borderWidth: 1.5,
        borderColor: "#ffb74d",
      }}
    >
      <View className="flex-row justify-between items-center mb-3">
        <Skeleton width={160} height={20} borderRadius={6} />
        <Skeleton width={80} height={24} borderRadius={12} />
      </View>
      <View className="flex-row justify-between items-center mb-2">
        <View>
          <Skeleton width={60} height={12} borderRadius={4} style={{ marginBottom: 4 }} />
          <Skeleton width={70} height={32} borderRadius={6} />
        </View>
        <View className="items-end">
          <Skeleton width={70} height={12} borderRadius={4} style={{ marginBottom: 4 }} />
          <Skeleton width={90} height={32} borderRadius={6} />
        </View>
      </View>
      <Skeleton width={160} height={12} borderRadius={4} style={{ marginBottom: 12 }} />
      <Skeleton width={"100%" as unknown as number} height={40} borderRadius={12} />
    </View>
  );
}
