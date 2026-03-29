import { View, Text } from "react-native";
import { Ionicons } from "@expo/vector-icons";

export function StepPrivacy() {
  return (
    <View
      className="items-center justify-center px-6"
      style={{ paddingTop: 32, minHeight: 380 }}
    >
      <Text style={{ fontSize: 48, marginBottom: 12 }}>🔒</Text>

      {/* Privacy illustration */}
      <View style={{ width: 240, marginBottom: 20 }}>
        {/* Masked identity card */}
        <View
          className="flex-row items-center rounded-2xl"
          style={{
            backgroundColor: "#f5f5f5",
            padding: 14,
            gap: 12,
            marginBottom: 10,
          }}
        >
          <View
            className="items-center justify-center rounded-full"
            style={{ width: 40, height: 40, backgroundColor: "#e0e0e0" }}
          >
            <Ionicons name="person" size={20} color="#999" />
          </View>
          <View>
            <Text
              className="font-bold"
              style={{ fontSize: 15, color: "#1a1a1a", letterSpacing: 0.5 }}
            >
              ja****s
            </Text>
            <Text style={{ fontSize: 11, color: "#888" }}>
              Your identity is masked
            </Text>
          </View>
        </View>

        {/* What others see vs real */}
        <View style={{ flexDirection: "row", gap: 8 }}>
          <View
            className="flex-1 rounded-xl items-center"
            style={{ backgroundColor: "#e8f5e9", padding: 10 }}
          >
            <Text
              className="font-semibold"
              style={{ fontSize: 10, color: "#2e7d32", marginBottom: 4 }}
            >
              Others see
            </Text>
            <Text
              className="font-bold"
              style={{ fontSize: 14, color: "#2e7d32" }}
            >
              ja****s
            </Text>
          </View>
          <View
            className="flex-1 rounded-xl items-center"
            style={{ backgroundColor: "#ffebee", padding: 10 }}
          >
            <Text
              className="font-semibold"
              style={{ fontSize: 10, color: "#c62828", marginBottom: 4 }}
            >
              Hidden
            </Text>
            <Text
              className="font-bold"
              style={{
                fontSize: 14,
                color: "#c62828",
                textDecorationLine: "line-through",
              }}
            >
              james92
            </Text>
          </View>
        </View>
      </View>

      <Text
        className="font-bold text-center"
        style={{ fontSize: 22, color: "#1a1a1a", marginBottom: 8 }}
      >
        See where you stand
      </Text>
      <Text
        className="text-center"
        style={{ fontSize: 14, color: "#888", lineHeight: 22, paddingHorizontal: 8 }}
      >
        Your name is masked for privacy. No one can see your real username.
        Check the leaderboard now!
      </Text>
    </View>
  );
}
