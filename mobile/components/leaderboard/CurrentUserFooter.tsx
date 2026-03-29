import { View, Text } from "react-native";
import { GemStar } from "@/components/GemStar";

interface CurrentUserFooterProps {
  rankDisplay: string;
  weeklyGems: number;
  notRanked?: boolean;
}

export function CurrentUserFooter({
  rankDisplay,
  weeklyGems,
  notRanked,
}: CurrentUserFooterProps) {
  return (
    <View
      style={{
        backgroundColor: "#fff",
        borderTopWidth: 1,
        borderTopColor: "#e0e0e0",
        padding: 12,
        paddingHorizontal: 16,
      }}
    >
      <View
        className="flex-row items-center rounded-xl"
        style={{
          backgroundColor: notRanked ? "#f5f5f5" : "#e8f5e9",
          borderWidth: 1.5,
          borderColor: notRanked ? "#e0e0e0" : "#81c784",
          padding: 12,
        }}
      >
        {/* Rank badge */}
        <View
          className="rounded-full items-center justify-center"
          style={{
            backgroundColor: notRanked ? "#888" : "#e65100",
            paddingHorizontal: 10,
            paddingVertical: 2,
            minWidth: 36,
          }}
        >
          <Text className="font-bold text-white" style={{ fontSize: 12 }}>
            {notRanked ? "—" : rankDisplay}
          </Text>
        </View>

        {/* Label */}
        <View className="flex-1" style={{ marginLeft: 10 }}>
          <Text className="font-bold" style={{ fontSize: 14, color: "#1a1a1a" }}>
            YOU ★
          </Text>
          {notRanked && (
            <Text style={{ fontSize: 11, color: "#888", marginTop: 1 }}>
              You didn't participate last week
            </Text>
          )}
        </View>

        {/* Gems */}
        {!notRanked && (
          <View className="flex-row items-center" style={{ gap: 3 }}>
            <Text className="font-bold" style={{ fontSize: 16, color: "#1a1a1a" }}>
              {weeklyGems.toLocaleString()}
            </Text>
            <GemStar size={14} />
          </View>
        )}
      </View>
    </View>
  );
}
