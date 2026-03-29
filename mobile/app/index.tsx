import { useState, useEffect, useCallback } from "react";
import { ScrollView, View, Alert } from "react-native";
import { useRouter } from "expo-router";
import AsyncStorage from "@react-native-async-storage/async-storage";

import { CheckInBanner } from "@/components/home/CheckInBanner";
import { WeeklyChallengeBanner } from "@/components/home/WeeklyChallengeBanner";
import { WeeklyChallengeSkeleton } from "@/components/home/WeeklyChallengeBanner";
import { ExtraRewards } from "@/components/home/ExtraRewards";
import { KeepPlaying } from "@/components/home/KeepPlaying";
import { useUser } from "@/contexts/UserContext";
import { checkin } from "@/services/api";

const MOCK_GAMES = [
  {
    id: "1",
    name: "Match Triple Goods : Sort Game",
    gemsEarned: 1423,
    totalGems: 190847,
  },
];

function parseRank(rankDisplay: string): number | null {
  if (rankDisplay === "99+") return null;
  const match = rankDisplay.match(/^#(\d+)$/);
  return match ? parseInt(match[1], 10) : null;
}

function formatTimeLeft(endTime: string): string {
  const end = new Date(endTime).getTime();
  const now = Date.now();
  const diff = end - now;

  if (diff <= 0) return "Ended";

  const days = Math.floor(diff / (1000 * 60 * 60 * 24));
  const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));

  if (days > 0) return `${days}d ${hours}h left`;
  const mins = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
  return `${hours}h ${mins}m left`;
}

export default function HomeScreen() {
  const router = useRouter();
  const { currentUser, banner, bannerLoading, rankChange, refreshBanner } =
    useUser();

  const [checkedIn, setCheckedIn] = useState(false);
  const [checkinLoading, setCheckinLoading] = useState(false);
  const [checkinResult, setCheckinResult] = useState<{
    gemsEarned: number;
    streak: number;
  } | null>(null);

  const checkinPrefix =
    process.env.EXPO_PUBLIC_STORAGE_KEY_CHECKIN || "checkin";

  const todayKey = useCallback(
    (userId: string) => {
      const d = new Date().toISOString().slice(0, 10);
      return `${checkinPrefix}:${userId}:${d}`;
    },
    [checkinPrefix]
  );

  // Check if already claimed today (from local storage)
  useEffect(() => {
    if (!currentUser) return;
    setCheckedIn(false);
    setCheckinResult(null);
    AsyncStorage.getItem(todayKey(currentUser.id)).then((val) => {
      if (val) {
        setCheckedIn(true);
      }
    });
  }, [currentUser, todayKey]);

  const rank = banner ? parseRank(banner.rank_display) : null;
  const is99Plus = banner?.rank_display === "99+";
  const isNotRanked = banner?.weekly_gems === 0;

  const nextRank =
    banner?.gap_to_next != null
      ? is99Plus
        ? 99
        : rank != null
          ? rank - 1
          : undefined
      : undefined;

  const handleCheckin = async () => {
    if (!currentUser || checkedIn || checkinLoading) return;
    setCheckinLoading(true);
    try {
      const res = await checkin(currentUser.id);
      setCheckedIn(true);
      setCheckinResult({
        gemsEarned: res.gems_earned,
        streak: res.current_streak,
      });
      await AsyncStorage.setItem(todayKey(currentUser.id), "1");
      refreshBanner();
    } catch (err: any) {
      if (err.message?.includes("already checked in")) {
        setCheckedIn(true);
        await AsyncStorage.setItem(todayKey(currentUser.id), "1");
      } else {
        Alert.alert("Check-in failed", err.message);
      }
    } finally {
      setCheckinLoading(false);
    }
  };

  return (
    <ScrollView className="flex-1" showsVerticalScrollIndicator={false}>
      <CheckInBanner
        claimed={checkedIn}
        loading={checkinLoading}
        gemsEarned={checkinResult?.gemsEarned}
        streak={checkinResult?.streak}
        onPress={handleCheckin}
      />

      {bannerLoading ? (
        <WeeklyChallengeSkeleton />
      ) : banner ? (
        <WeeklyChallengeBanner
          rank={rank}
          rankChange={
            !is99Plus && !isNotRanked ? (rankChange ?? undefined) : undefined
          }
          weeklyGems={banner.weekly_gems}
          timeLeft={formatTimeLeft(banner.end_time)}
          gemsToNextRank={
            !isNotRanked ? (banner.gap_to_next ?? undefined) : undefined
          }
          nextRank={!isNotRanked ? nextRank : undefined}
          notRanked={isNotRanked}
          onViewLeaderboard={() => router.push("/leaderboard")}
        />
      ) : (
        <View className="mx-4 mb-5" />
      )}

      <ExtraRewards />

      <KeepPlaying games={MOCK_GAMES} />
    </ScrollView>
  );
}
