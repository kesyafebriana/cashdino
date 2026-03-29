import { View, Text, Pressable } from "react-native";
import { Ionicons } from "@expo/vector-icons";

export type TabKey = "discover" | "tasks" | "rewards" | "wallet" | "profile";

interface Tab {
  key: TabKey;
  label: string;
  icon: keyof typeof Ionicons.glyphMap;
  iconFilled: keyof typeof Ionicons.glyphMap;
}

const TABS: Tab[] = [
  { key: "discover", label: "Discover", icon: "home-outline", iconFilled: "home" },
  { key: "tasks", label: "My Tasks", icon: "grid-outline", iconFilled: "grid" },
  { key: "rewards", label: "Rewards", icon: "cash-outline", iconFilled: "cash" },
  { key: "wallet", label: "Wallet", icon: "card-outline", iconFilled: "card" },
  { key: "profile", label: "Profile", icon: "person-outline", iconFilled: "person" },
];

interface BottomNavbarProps {
  activeTab?: TabKey;
  onTabPress?: (tab: TabKey) => void;
}

export function BottomNavbar({
  activeTab = "discover",
  onTabPress,
}: BottomNavbarProps) {
  return (
    <View
      className="flex-row justify-around bg-white pt-2 pb-7 border-t border-gray-200"
    >
      {TABS.map((tab) => {
        const isActive = activeTab === tab.key;
        return (
          <Pressable
            key={tab.key}
            onPress={() => onTabPress?.(tab.key)}
            className="items-center"
          >
            <Ionicons
              name={isActive ? tab.iconFilled : tab.icon}
              size={22}
              color={isActive ? "#1a1a1a" : "#999"}
            />
            <Text
              className="text-[10px] mt-0.5"
              style={{
                color: isActive ? "#1a1a1a" : "#999",
                fontWeight: isActive ? "600" : "400",
              }}
            >
              {tab.label}
            </Text>
          </Pressable>
        );
      })}
    </View>
  );
}
