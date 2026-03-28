import React from "react";
import { View, Text, TouchableOpacity, StyleSheet } from "react-native";

interface UserSwitcherProps {
  userId: number;
  onChangeUser: (id: number) => void;
}

const TEST_USERS = [
  { id: 1, name: "User 1" },
  { id: 2, name: "User 2" },
  { id: 3, name: "User 3" },
];

export default function UserSwitcher({ userId, onChangeUser }: UserSwitcherProps) {
  return (
    <View style={styles.container}>
      <Text style={styles.label}>Current User:</Text>
      <View style={styles.buttons}>
        {TEST_USERS.map((u) => (
          <TouchableOpacity
            key={u.id}
            style={[styles.button, userId === u.id && styles.activeButton]}
            onPress={() => onChangeUser(u.id)}
          >
            <Text style={[styles.buttonText, userId === u.id && styles.activeText]}>
              {u.name}
            </Text>
          </TouchableOpacity>
        ))}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: "row",
    alignItems: "center",
    padding: 12,
    backgroundColor: "#f5f5f5",
    borderRadius: 8,
  },
  label: {
    fontSize: 14,
    fontWeight: "600",
    marginRight: 8,
  },
  buttons: {
    flexDirection: "row",
    gap: 6,
  },
  button: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 6,
    backgroundColor: "#e0e0e0",
  },
  activeButton: {
    backgroundColor: "#4f46e5",
  },
  buttonText: {
    fontSize: 13,
    color: "#333",
  },
  activeText: {
    color: "#fff",
  },
});
