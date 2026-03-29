import { useState } from "react";
import { View, Text, TextInput, Pressable, Alert, Keyboard, TouchableWithoutFeedback } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useUser } from "@/contexts/UserContext";
import { earnGems } from "@/services/api";
import { GemStar } from "@/components/GemStar";

const SOURCES = ["gameplay", "survey", "referral", "boost"] as const;

export default function TasksScreen() {
  const { currentUser, refreshBanner } = useUser();
  const [amount, setAmount] = useState("");
  const [source, setSource] = useState<(typeof SOURCES)[number]>("gameplay");
  const [loading, setLoading] = useState(false);
  const [lastResult, setLastResult] = useState<{
    amount: number;
    weeklyGems: number;
  } | null>(null);

  const handleEarn = async () => {
    if (!currentUser) return;
    const parsed = parseInt(amount, 10);
    if (!parsed || parsed <= 0) {
      Alert.alert("Invalid amount", "Enter a number greater than 0");
      return;
    }

    setLoading(true);
    try {
      const res = await earnGems(currentUser.id, parsed, source);
      setLastResult({ amount: parsed, weeklyGems: res.weekly_gems });
      setAmount("");
      refreshBanner();
    } catch (err: any) {
      Alert.alert("Error", err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <TouchableWithoutFeedback onPress={Keyboard.dismiss}>
    <View className="flex-1 bg-white px-6 pt-6">
      {/* Header */}
      <Text className="text-xl font-bold text-gray-900 mb-1">My Tasks</Text>
      <View className="flex-row items-center gap-1.5 mb-6">
        <Ionicons name="information-circle-outline" size={14} color="#999" />
        <Text className="text-xs text-gray-400">
          Simulation only — earn gems to test the leaderboard
        </Text>
      </View>

      {/* Gem amount input */}
      <Text className="text-sm font-semibold text-gray-900 mb-2">
        Gems to earn
      </Text>
      <TextInput
        className="border border-gray-200 rounded-xl px-4 py-3 text-base text-gray-900 mb-4"
        placeholder="Enter amount (e.g. 500)"
        placeholderTextColor="#bbb"
        keyboardType="number-pad"
        value={amount}
        onChangeText={setAmount}
      />

      {/* Source selector */}
      <Text className="text-sm font-semibold text-gray-900 mb-2">Source</Text>
      <View className="flex-row flex-wrap gap-2 mb-6">
        {SOURCES.map((s) => (
          <Pressable
            key={s}
            onPress={() => setSource(s)}
            className="rounded-full px-4 py-2"
            style={{
              backgroundColor: source === s ? "#1a1a1a" : "#f5f5f5",
            }}
          >
            <Text
              className="text-xs font-semibold"
              style={{ color: source === s ? "#fff" : "#666" }}
            >
              {s}
            </Text>
          </Pressable>
        ))}
      </View>

      {/* Earn button */}
      <Pressable
        onPress={handleEarn}
        disabled={loading}
        className="rounded-xl py-3.5 items-center mb-6"
        style={{ backgroundColor: loading ? "#ccc" : "#1a1a1a" }}
      >
        <View className="flex-row items-center gap-2">
          <GemStar size={16} />
          <Text className="text-sm font-semibold text-white">
            {loading ? "Earning..." : "Earn Gems"}
          </Text>
        </View>
      </Pressable>

      {/* Result */}
      {lastResult && (
        <View className="bg-green-50 rounded-2xl p-4 border border-green-200">
          <View className="flex-row items-center gap-2 mb-1">
            <Ionicons name="checkmark-circle" size={20} color="#16a34a" />
            <Text className="text-sm font-semibold text-green-800">
              +{lastResult.amount.toLocaleString()} gems earned!
            </Text>
          </View>
          <Text className="text-xs text-green-700">
            Weekly total: {lastResult.weeklyGems.toLocaleString()} gems
          </Text>
        </View>
      )}
    </View>
    </TouchableWithoutFeedback>
  );
}
