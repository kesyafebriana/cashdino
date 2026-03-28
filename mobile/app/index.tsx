import { useEffect, useState } from "react";
import { View, Text, StyleSheet } from "react-native";
import { StatusBar } from "expo-status-bar";
import { useCurrentUser } from "@/hooks/useCurrentUser";
import { fetchHealth } from "@/services/api";

export default function HomeScreen() {
  const { userId } = useCurrentUser();
  const [health, setHealth] = useState<string>("loading...");

  useEffect(() => {
    fetchHealth()
      .then((data) => setHealth(data.status))
      .catch(() => setHealth("error"));
  }, []);

  return (
    <View style={styles.container}>
      <Text style={styles.title}>CashDino</Text>
      <Text style={styles.subtitle}>Weekly Challenge Leaderboard</Text>
      <Text style={styles.info}>Current User ID: {userId}</Text>
      <Text style={styles.info}>API Status: {health}</Text>
      <StatusBar style="auto" />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    padding: 24,
  },
  title: {
    fontSize: 28,
    fontWeight: "bold",
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 16,
    color: "#666",
    marginBottom: 24,
  },
  info: {
    fontSize: 14,
    color: "#333",
    marginBottom: 8,
  },
});
