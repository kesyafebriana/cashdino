import { View, Text, ScrollView, Pressable } from "react-native";
import { GemStar } from "@/components/GemStar";

interface RewardCard {
  id: string;
  emoji: string;
  title: string;
  subtitle: string;
  buttonText: string;
  buttonVariant: "dark" | "muted";
  backgroundColor: string;
  showGemIcon?: boolean;
}

const REWARD_CARDS: RewardCard[] = [
  {
    id: "boost",
    emoji: "🚀",
    title: "x Boost",
    subtitle: "Efficient gems",
    buttonText: "Earn 3999",
    buttonVariant: "muted",
    backgroundColor: "#f5f5f5",
  },
  {
    id: "invite",
    emoji: "📨",
    title: "Invite a friend",
    subtitle: "And get 200 coins!",
    buttonText: "Get 200",
    buttonVariant: "dark",
    backgroundColor: "#fce4ec",
    showGemIcon: true,
  },
  {
    id: "more",
    emoji: "✨",
    title: "More rewards",
    subtitle: "Coming soon!",
    buttonText: "Explore",
    buttonVariant: "muted",
    backgroundColor: "#f5f5f5",
  },
];

function RewardCardItem({ card }: { card: RewardCard }) {
  const isDark = card.buttonVariant === "dark";

  return (
    <View
      className="rounded-2xl p-4 items-center"
      style={{ backgroundColor: card.backgroundColor, width: 140 }}
    >
      <Text className="text-3xl mb-2">{card.emoji}</Text>
      <Text className="text-sm font-bold text-gray-900 text-center">
        {card.title}
      </Text>
      <Text className="text-xs text-gray-500 text-center">{card.subtitle}</Text>
      <Pressable
        className="mt-2 rounded-full px-3 py-1.5 flex-row items-center gap-1"
        style={{
          backgroundColor: isDark ? "#1a1a1a" : "#e0e0e0",
        }}
      >
        <Text
          className="text-xs font-semibold"
          style={{ color: isDark ? "#fff" : "#888" }}
        >
          {card.buttonText}
        </Text>
        {card.showGemIcon && <GemStar size={12} />}
      </Pressable>
    </View>
  );
}

export function ExtraRewards() {
  return (
    <View className="mb-4">
      <Text className="text-lg font-bold text-gray-900 px-4 mb-3">
        Extra rewards
      </Text>
      <ScrollView
        horizontal
        showsHorizontalScrollIndicator={false}
        contentContainerClassName="px-4 gap-2.5"
      >
        {REWARD_CARDS.map((card) => (
          <RewardCardItem key={card.id} card={card} />
        ))}
      </ScrollView>
    </View>
  );
}
