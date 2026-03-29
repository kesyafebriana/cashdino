import { Text } from "react-native";

interface GemStarProps {
  size?: number;
}

export function GemStar({ size = 16 }: GemStarProps) {
  return <Text style={{ fontSize: size, lineHeight: size + 2 }}>⭐</Text>;
}
