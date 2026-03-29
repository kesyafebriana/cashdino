import { View, Text } from "react-native";

export function StepIntro() {
  return (
    <View
      className="items-center justify-center px-6"
      style={{ paddingTop: 32, minHeight: 380 }}
    >
      <Text style={{ fontSize: 48, marginBottom: 8 }}>🏆</Text>

      {/* Leaderboard preview illustration */}
      <View
        className="rounded-2xl mb-6"
        style={{
          width: 220,
          backgroundColor: "#fff3e0",
          padding: 12,
          gap: 5,
        }}
      >
        {/* Gold */}
        <View
          className="flex-row items-center rounded-lg"
          style={{ backgroundColor: "#fffde7", padding: 7, paddingHorizontal: 8, gap: 6 }}
        >
          <Text style={{ fontSize: 14 }}>🥇</Text>
          <View
            className="flex-1 rounded-full"
            style={{ height: 8, backgroundColor: "#e0e0e0" }}
          />
          <View
            className="rounded-full"
            style={{ width: 44, height: 8, backgroundColor: "#ffcc80" }}
          />
        </View>

        {/* Silver */}
        <View
          className="flex-row items-center rounded-lg"
          style={{ backgroundColor: "#f5f5f5", padding: 7, paddingHorizontal: 8, gap: 6 }}
        >
          <Text style={{ fontSize: 14 }}>🥈</Text>
          <View
            className="flex-1 rounded-full"
            style={{ height: 8, backgroundColor: "#e0e0e0" }}
          />
          <View
            className="rounded-full"
            style={{ width: 36, height: 8, backgroundColor: "#e0e0e0" }}
          />
        </View>

        {/* Bronze */}
        <View
          className="flex-row items-center rounded-lg"
          style={{ backgroundColor: "#fff3e0", padding: 7, paddingHorizontal: 8, gap: 6 }}
        >
          <Text style={{ fontSize: 14 }}>🥉</Text>
          <View
            className="flex-1 rounded-full"
            style={{ height: 8, backgroundColor: "#e0e0e0" }}
          />
          <View
            className="rounded-full"
            style={{ width: 30, height: 8, backgroundColor: "#e0e0e0" }}
          />
        </View>

        {/* You */}
        <View
          className="flex-row items-center"
          style={{ padding: 7, paddingHorizontal: 8, gap: 6 }}
        >
          <View
            className="rounded-full"
            style={{ backgroundColor: "#4cd964", paddingHorizontal: 7, paddingVertical: 2 }}
          >
            <Text
              className="font-bold text-white"
              style={{ fontSize: 9 }}
            >
              YOU
            </Text>
          </View>
          <View
            className="flex-1 rounded-full"
            style={{ height: 8, backgroundColor: "#c8e6c9" }}
          />
          <View
            className="rounded-full"
            style={{ width: 22, height: 8, backgroundColor: "#a5d6a7" }}
          />
        </View>
      </View>

      <Text
        className="font-bold text-center"
        style={{ fontSize: 22, color: "#1a1a1a", marginBottom: 8 }}
      >
        Introducing weekly challenges
      </Text>
      <Text
        className="text-center"
        style={{ fontSize: 14, color: "#888", lineHeight: 22, paddingHorizontal: 8 }}
      >
        Every gem you earn now counts toward a weekly ranking. Compete with all
        players!
      </Text>
    </View>
  );
}
