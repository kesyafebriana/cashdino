import "../global.css";
import { View } from "react-native";
import { Slot, useRouter, usePathname } from "expo-router";
import { StatusBar } from "expo-status-bar";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import { Navbar } from "@/components/home/Navbar";
import { BottomNavbar, type TabKey } from "@/components/home/BottomNavbar";
import { OnboardingModal } from "@/components/onboarding/OnboardingModal";
import { UserProvider, useUser } from "@/contexts/UserContext";

function AppContent() {
  const insets = useSafeAreaInsets();
  const router = useRouter();
  const pathname = usePathname();
  const { banner, bannerLoading } = useUser();

  const totalGems = bannerLoading ? null : (banner?.total_gems ?? 0);

  const hideChrome = pathname.startsWith("/leaderboard");

  const activeTab: TabKey =
    pathname === "/profile"
      ? "profile"
      : pathname === "/tasks"
        ? "tasks"
        : "discover";

  const handleTabPress = (tab: TabKey) => {
    if (tab === "profile") {
      router.push("/profile");
    } else if (tab === "tasks") {
      router.push("/tasks");
    } else if (tab === "discover") {
      router.push("/");
    }
  };

  return (
    <View className="flex-1 bg-white">
      <StatusBar style="dark" />

      {!hideChrome && (
        <View style={{ paddingTop: insets.top }}>
          <Navbar totalGems={totalGems} />
        </View>
      )}

      <View className="flex-1">
        <Slot />
      </View>

      <BottomNavbar activeTab={activeTab} onTabPress={handleTabPress} />

      <OnboardingModal />
    </View>
  );
}

export default function RootLayout() {
  return (
    <UserProvider>
      <AppContent />
    </UserProvider>
  );
}
