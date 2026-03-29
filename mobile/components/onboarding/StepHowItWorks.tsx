import { View, Text } from "react-native";
import { Ionicons } from "@expo/vector-icons";

export function StepHowItWorks() {
  return (
    <View
      className="items-center justify-center px-6"
      style={{ paddingTop: 32, minHeight: 380 }}
    >
      <Text style={{ fontSize: 48, marginBottom: 12 }}>💎</Text>

      {/* Gem sources */}
      <View style={{ width: 220, flexDirection: "row", gap: 8, marginBottom: 16 }}>
        <View
          className="flex-1 rounded-2xl items-center"
          style={{ backgroundColor: "#e8f5e9", paddingVertical: 14, paddingHorizontal: 12 }}
        >
          <Text style={{ fontSize: 24, marginBottom: 4 }}>🎮</Text>
          <Text className="font-semibold" style={{ fontSize: 10, color: "#2e7d32" }}>
            Games
          </Text>
        </View>
        <View
          className="flex-1 rounded-2xl items-center"
          style={{ backgroundColor: "#e3f2fd", paddingVertical: 14, paddingHorizontal: 12 }}
        >
          <Text style={{ fontSize: 24, marginBottom: 4 }}>✅</Text>
          <Text className="font-semibold" style={{ fontSize: 10, color: "#1565c0" }}>
            Check-in
          </Text>
        </View>
        <View
          className="flex-1 rounded-2xl items-center"
          style={{ backgroundColor: "#fce4ec", paddingVertical: 14, paddingHorizontal: 12 }}
        >
          <Text style={{ fontSize: 24, marginBottom: 4 }}>📋</Text>
          <Text className="font-semibold" style={{ fontSize: 10, color: "#c62828" }}>
            Surveys
          </Text>
        </View>
      </View>

      {/* Arrow down */}
      <View style={{ marginTop: 4, marginBottom: 12 }}>
        <Ionicons name="arrow-down" size={32} color="#e65100" />
      </View>

      {/* Weekly leaderboard badge */}
      <View
        className="flex-row items-center rounded-2xl"
        style={{
          backgroundColor: "#fff3e0",
          borderWidth: 1,
          borderColor: "#ffcc80",
          paddingHorizontal: 20,
          paddingVertical: 10,
          gap: 8,
          marginBottom: 24,
        }}
      >
        <Text style={{ fontSize: 18 }}>🏆</Text>
        <Text className="font-bold" style={{ fontSize: 14, color: "#e65100" }}>
          Weekly leaderboard
        </Text>
      </View>

      {/* Resets every Monday */}
      <View
        className="flex-row items-center rounded-xl"
        style={{
          backgroundColor: "#f5f5f5",
          paddingHorizontal: 14,
          paddingVertical: 6,
          gap: 6,
          marginBottom: 20,
        }}
      >
        <Ionicons name="refresh" size={14} color="#888" />
        <Text className="font-medium" style={{ fontSize: 11, color: "#888" }}>
          Resets every Monday
        </Text>
      </View>

      <Text
        className="font-bold text-center"
        style={{ fontSize: 22, color: "#1a1a1a", marginBottom: 8 }}
      >
        How it works
      </Text>
      <Text
        className="text-center"
        style={{ fontSize: 14, color: "#888", lineHeight: 22, paddingHorizontal: 8 }}
      >
        Gems from games, check-ins, and surveys all count toward your weekly
        rank. Rankings reset every Monday.
      </Text>
    </View>
  );
}
