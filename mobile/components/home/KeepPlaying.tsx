import { View, Text, Pressable } from "react-native";
import { GemStar } from "@/components/GemStar";

interface Game {
  id: string;
  name: string;
  gemsEarned: number;
  totalGems: number;
}

interface KeepPlayingProps {
  games: Game[];
  onSeeAll?: () => void;
}

function GameCard({ game }: { game: Game }) {
  return (
    <View className="rounded-2xl overflow-hidden border border-gray-200">
      {/* Game title */}
      <View className="bg-gray-50 py-2.5 px-3 items-center">
        <Text className="text-sm font-semibold text-gray-900">{game.name}</Text>
      </View>

      {/* Game preview placeholder */}
      <View className="h-28 bg-gray-100 items-center justify-center">
        <Text className="text-sm text-gray-400">Game preview</Text>
      </View>

      {/* Gem stats */}
      <View className="flex-row justify-between px-3 py-2.5 bg-white">
        <View className="flex-row items-center gap-1">
          <Text className="text-sm font-semibold text-gray-900">
            {game.gemsEarned.toLocaleString()}
          </Text>
          <GemStar size={12} />
        </View>
        <View className="flex-row items-center gap-1">
          <Text className="text-sm font-semibold text-gray-900">
            {game.totalGems.toLocaleString()}
          </Text>
          <GemStar size={12} />
        </View>
      </View>
    </View>
  );
}

export function KeepPlaying({ games, onSeeAll }: KeepPlayingProps) {
  return (
    <View className="px-4 pb-4">
      {/* Section header */}
      <View className="flex-row justify-between items-center mb-3">
        <Text className="text-lg font-bold text-gray-900">Keep playing</Text>
        <Pressable onPress={onSeeAll}>
          <Text className="text-sm font-medium text-gray-900">See All →</Text>
        </Pressable>
      </View>

      {/* Game cards */}
      <View className="gap-3">
        {games.map((game) => (
          <GameCard key={game.id} game={game} />
        ))}
      </View>
    </View>
  );
}
