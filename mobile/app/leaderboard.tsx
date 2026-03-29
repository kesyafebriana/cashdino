import { useState, useEffect, useCallback } from "react";
import { View, Text, FlatList, Pressable, ActivityIndicator } from "react-native";
import { useRouter } from "expo-router";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { Ionicons } from "@expo/vector-icons";
import AsyncStorage from "@react-native-async-storage/async-storage";

import { LeaderboardRow } from "@/components/leaderboard/LeaderboardRow";
import { LastWeekRow } from "@/components/leaderboard/LastWeekRow";
import { PrizeBanner } from "@/components/leaderboard/PrizeBanner";
import { CurrentUserFooter } from "@/components/leaderboard/CurrentUserFooter";
import { RewardsModal } from "@/components/leaderboard/RewardsModal";
import { useUser } from "@/contexts/UserContext";
import {
  fetchCurrentLeaderboard,
  fetchLastWeekLeaderboard,
  type CurrentLeaderboardResponse,
  type CurrentLeaderboardRow,
  type LastWeekResponse,
} from "@/services/api";

function formatTimeLeft(endTime: string): string {
  const diff = new Date(endTime).getTime() - Date.now();
  if (diff <= 0) return "Ended";
  const days = Math.floor(diff / (1000 * 60 * 60 * 24));
  const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
  if (days > 0) return `${days}d ${hours}h left`;
  const mins = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
  return `${hours}h ${mins}m left`;
}

function formatDateRange(startTime: string, endTime: string): string {
  const start = new Date(startTime);
  const end = new Date(endTime);
  const monthNames = [
    "Jan", "Feb", "Mar", "Apr", "May", "Jun",
    "Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
  ];
  const sMonth = monthNames[start.getUTCMonth()];
  const eMonth = monthNames[end.getUTCMonth()];
  const sDay = start.getUTCDate();
  const eDay = end.getUTCDate();
  if (sMonth === eMonth) {
    return `Results — ${sMonth} ${sDay}–${eDay}`;
  }
  return `Results — ${sMonth} ${sDay} – ${eMonth} ${eDay}`;
}

function computeRankChanges(
  current: CurrentLeaderboardRow[],
  previous: Record<string, number> | null
): Record<string, number> {
  const changes: Record<string, number> = {};
  if (!previous) return changes;
  for (const entry of current) {
    const oldRank = previous[entry.display_name];
    if (oldRank != null) {
      changes[entry.display_name] = oldRank - entry.rank;
    }
  }
  return changes;
}

export default function LeaderboardScreen() {
  const router = useRouter();
  const insets = useSafeAreaInsets();
  const { currentUser } = useUser();

  const [activeTab, setActiveTab] = useState<"this" | "last">("this");

  // This week state
  const [thisWeekData, setThisWeekData] =
    useState<CurrentLeaderboardResponse | null>(null);
  const [thisWeekLoading, setThisWeekLoading] = useState(true);
  const [thisWeekError, setThisWeekError] = useState<string | null>(null);
  const [rewardsVisible, setRewardsVisible] = useState(false);
  const [rankChanges, setRankChanges] = useState<Record<string, number>>({});

  // Last week state
  const [lastWeekData, setLastWeekData] = useState<LastWeekResponse | null>(
    null
  );
  const [lastWeekLoading, setLastWeekLoading] = useState(true);
  const [lastWeekError, setLastWeekError] = useState<string | null>(null);
  const [openTooltipRank, setOpenTooltipRank] = useState<number | null>(null);

  const loadThisWeek = useCallback(async () => {
    if (!currentUser) return;
    setThisWeekLoading(true);
    setThisWeekError(null);
    try {
      const resp = await fetchCurrentLeaderboard(currentUser.id);
      const storageKey = `leaderboard_ranks:${currentUser.id}:${resp.challenge.id}`;
      const stored = await AsyncStorage.getItem(storageKey);
      const previous: Record<string, number> | null = stored
        ? JSON.parse(stored)
        : null;
      setRankChanges(computeRankChanges(resp.leaderboard, previous));
      const currentRanks: Record<string, number> = {};
      for (const entry of resp.leaderboard) {
        currentRanks[entry.display_name] = entry.rank;
      }
      await AsyncStorage.setItem(storageKey, JSON.stringify(currentRanks));
      setThisWeekData(resp);
    } catch (err: any) {
      setThisWeekError(err.message);
    } finally {
      setThisWeekLoading(false);
    }
  }, [currentUser]);

  const loadLastWeek = useCallback(async () => {
    if (!currentUser) return;
    setLastWeekLoading(true);
    setLastWeekError(null);
    try {
      const resp = await fetchLastWeekLeaderboard(currentUser.id);
      setLastWeekData(resp);
    } catch (err: any) {
      setLastWeekError(err.message);
    } finally {
      setLastWeekLoading(false);
    }
  }, [currentUser]);

  useEffect(() => {
    loadThisWeek();
    loadLastWeek();
  }, [loadThisWeek, loadLastWeek]);

  const thisWeekUserRank = thisWeekData?.current_user?.rank;
  const showThisWeekFooter = thisWeekData?.current_user != null;

  const lastWeekUserRank = lastWeekData?.current_user?.rank;
  const showLastWeekFooter = lastWeekData?.current_user != null;

  const loading = activeTab === "this" ? thisWeekLoading : lastWeekLoading;
  const error = activeTab === "this" ? thisWeekError : lastWeekError;
  const retry = activeTab === "this" ? loadThisWeek : loadLastWeek;

  return (
    <View className="flex-1 bg-white" style={{ paddingTop: insets.top }}>
      {/* Header */}
      <View
        className="flex-row items-center"
        style={{
          paddingHorizontal: 16,
          paddingTop: 12,
          paddingBottom: 8,
          gap: 8,
        }}
      >
        <Pressable onPress={() => router.back()}>
          <Ionicons name="chevron-back" size={24} color="#1a1a1a" />
        </Pressable>
        <Text
          className="flex-1 font-bold"
          style={{ fontSize: 20, color: "#1a1a1a" }}
        >
          Weekly challenge
        </Text>
        {activeTab === "this" && thisWeekData && (
          <View
            className="rounded-full"
            style={{
              backgroundColor: "#e65100",
              paddingHorizontal: 10,
              paddingVertical: 3,
            }}
          >
            <Text className="font-semibold text-white" style={{ fontSize: 11 }}>
              {formatTimeLeft(thisWeekData.challenge.end_time)}
            </Text>
          </View>
        )}
      </View>

      {/* Tab toggle */}
      <View
        className="flex-row mx-4 rounded-xl"
        style={{ backgroundColor: "#f5f5f5", padding: 3, marginTop: 8 }}
      >
        {(["this", "last"] as const).map((tab) => (
          <Pressable
            key={tab}
            onPress={() => {
              setActiveTab(tab);
              setOpenTooltipRank(null);
            }}
            className="flex-1 items-center rounded-lg"
            style={{
              padding: 8,
              backgroundColor: activeTab === tab ? "#fff" : "transparent",
              ...(activeTab === tab
                ? {
                    shadowColor: "#000",
                    shadowOffset: { width: 0, height: 1 },
                    shadowOpacity: 0.08,
                    shadowRadius: 3,
                    elevation: 2,
                  }
                : {}),
            }}
          >
            <Text
              className={activeTab === tab ? "font-semibold" : "font-medium"}
              style={{
                fontSize: 13,
                color: activeTab === tab ? "#1a1a1a" : "#999",
              }}
            >
              {tab === "this" ? "This week" : "Last week"}
            </Text>
          </Pressable>
        ))}
      </View>

      {/* Content */}
      {loading ? (
        <View className="flex-1 items-center justify-center">
          <ActivityIndicator size="large" color="#e65100" />
        </View>
      ) : error ? (
        <View className="flex-1 items-center justify-center px-8">
          <Text
            style={{
              fontSize: 14,
              color: "#888",
              textAlign: "center",
              marginBottom: 12,
            }}
          >
            {error}
          </Text>
          <Pressable
            onPress={retry}
            className="rounded-xl items-center"
            style={{
              backgroundColor: "#e65100",
              paddingHorizontal: 24,
              paddingVertical: 10,
            }}
          >
            <Text className="font-semibold text-white" style={{ fontSize: 14 }}>
              Retry
            </Text>
          </Pressable>
        </View>
      ) : activeTab === "this" && thisWeekData ? (
        <>
          <FlatList
            data={thisWeekData.leaderboard}
            keyExtractor={(item) => String(item.rank)}
            contentContainerStyle={{
              paddingHorizontal: 16,
              paddingTop: 0,
              paddingBottom: 16,
            }}
            showsVerticalScrollIndicator={false}
            ListHeaderComponent={
              thisWeekData.campaign ? (
                <PrizeBanner
                  rewardsSummary={thisWeekData.campaign.rewards_summary}
                  onSeeAll={() => setRewardsVisible(true)}
                />
              ) : null
            }
            ListHeaderComponentStyle={{
              marginHorizontal: -16,
              marginBottom: 12,
            }}
            renderItem={({ item }) => (
              <LeaderboardRow
                rank={item.rank}
                displayName={item.display_name}
                weeklyGems={item.weekly_gems}
                rankChange={rankChanges[item.display_name] ?? null}
                isCurrentUser={thisWeekUserRank === item.rank}
              />
            )}
            ListEmptyComponent={
              <View className="items-center py-12">
                <Text style={{ fontSize: 14, color: "#888" }}>
                  No entries yet. Start earning gems!
                </Text>
              </View>
            }
          />

          {showThisWeekFooter && (
            <CurrentUserFooter
              rankDisplay={thisWeekData.current_user!.rank_display}
              weeklyGems={thisWeekData.current_user!.weekly_gems}
            />
          )}

          {thisWeekData.campaign && (
            <RewardsModal
              visible={rewardsVisible}
              onClose={() => setRewardsVisible(false)}
              rewardsSummary={thisWeekData.campaign.rewards_summary}
            />
          )}
        </>
      ) : activeTab === "last" && lastWeekData?.challenge ? (
        <>
          <FlatList
            data={lastWeekData.leaderboard}
            keyExtractor={(item) => String(item.rank)}
            contentContainerStyle={{
              paddingHorizontal: 16,
              paddingTop: 0,
              paddingBottom: 16,
            }}
            showsVerticalScrollIndicator={false}
            ListHeaderComponent={
              <Text
                className="font-medium"
                style={{
                  fontSize: 12,
                  color: "#888",
                  paddingTop: 12,
                  paddingBottom: 4,
                }}
              >
                {formatDateRange(
                  lastWeekData.challenge!.start_time,
                  lastWeekData.challenge!.end_time
                )}
              </Text>
            }
            renderItem={({ item }) => (
              <LastWeekRow
                rank={item.rank}
                displayName={item.display_name}
                finalGems={item.final_gems}
                rewards={item.rewards}
                isCurrentUser={lastWeekUserRank === item.rank}
                tooltipOpen={openTooltipRank === item.rank}
                onToggleTooltip={() =>
                  setOpenTooltipRank(
                    openTooltipRank === item.rank ? null : item.rank
                  )
                }
              />
            )}
            ListEmptyComponent={
              <View className="items-center py-12">
                <Text style={{ fontSize: 14, color: "#888" }}>
                  No results for last week.
                </Text>
              </View>
            }
          />

          {showLastWeekFooter ? (
            <CurrentUserFooter
              rankDisplay={lastWeekData.current_user!.rank_display}
              weeklyGems={lastWeekData.current_user!.final_gems}
            />
          ) : (
            <CurrentUserFooter
              rankDisplay="—"
              weeklyGems={0}
              notRanked
            />
          )}
        </>
      ) : activeTab === "last" ? (
        <View className="flex-1 items-center justify-center px-8">
          <Text style={{ fontSize: 48, marginBottom: 12 }}>📅</Text>
          <Text
            className="font-bold"
            style={{ fontSize: 18, color: "#1a1a1a", marginBottom: 4 }}
          >
            No previous results
          </Text>
          <Text style={{ fontSize: 14, color: "#888", textAlign: "center" }}>
            Last week's results will appear here after the first weekly reset.
          </Text>
        </View>
      ) : null}
    </View>
  );
}
