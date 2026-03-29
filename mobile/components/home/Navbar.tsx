import { View, Text } from "react-native";
import { GemStar } from "@/components/GemStar";
import { Skeleton } from "@/components/Skeleton";

interface NavbarProps {
  totalGems: number | null;
}

export function Navbar({ totalGems }: NavbarProps) {
  return (
    <View className="flex-row justify-between items-center px-6 pt-3 pb-4">
      <Text className="text-3xl font-bold text-gray-900">Discover</Text>
      <View className="flex-row items-center gap-1">
        {totalGems === null ? (
          <Skeleton width={60} height={24} borderRadius={6} />
        ) : (
          <>
            <Text className="text-2xl font-bold text-gray-900">
              {totalGems.toLocaleString()}
            </Text>
            <GemStar size={20} />
          </>
        )}
      </View>
    </View>
  );
}
