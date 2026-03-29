import { View, Text, Pressable } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { GemStar } from "@/components/GemStar";

interface CheckInBannerProps {
  claimed: boolean;
  loading?: boolean;
  gemsEarned?: number;
  streak?: number;
  onPress?: () => void;
}

export function CheckInBanner({
  claimed,
  loading,
  gemsEarned,
  streak,
  onPress,
}: CheckInBannerProps) {
  return (
    <Pressable
      onPress={!claimed && !loading ? onPress : undefined}
      className="mx-4 mb-4 rounded-2xl px-4 py-3 flex-row items-center justify-between"
      style={{ backgroundColor: "#ffd700" }}
    >
      <View className="flex-row items-center gap-3">
        <View className="w-11 h-11 bg-white rounded-full items-center justify-center">
          <Text className="text-2xl">🎰</Text>
        </View>
        <View>
          <Text className="text-sm font-semibold text-gray-900">
            {loading
              ? "Checking in..."
              : claimed
                ? "All done! Come back tomorrow!"
                : "Tap to claim your daily reward!"}
          </Text>
          {claimed && gemsEarned != null && (
            <View className="flex-row items-center gap-1 mt-0.5">
              <Text className="text-xs text-gray-700">
                +{gemsEarned} gems
              </Text>
              <GemStar size={10} />
              {streak != null && streak > 1 && (
                <Text className="text-xs text-gray-700">
                  {" "}
                  · {streak} day streak 🔥
                </Text>
              )}
            </View>
          )}
        </View>
      </View>
      {claimed && (
        <View className="flex-row items-center gap-1">
          <Text className="text-xs font-semibold text-gray-900">Claimed</Text>
          <Ionicons name="checkmark-circle-outline" size={18} color="#1a1a1a" />
        </View>
      )}
    </Pressable>
  );
}
