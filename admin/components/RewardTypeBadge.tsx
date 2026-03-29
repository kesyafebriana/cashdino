const STYLES: Record<string, string> = {
  gems: "bg-green-50 text-green-700 border-green-200",
  gift_card: "bg-orange-50 text-orange-700 border-orange-200",
  cash: "bg-blue-50 text-blue-700 border-blue-200",
  other: "bg-gray-50 text-gray-600 border-gray-200",
};

export function RewardTypeBadge({ type }: { type: string }) {
  const style = STYLES[type] ?? STYLES.other;
  return (
    <span
      className={`inline-block rounded-md border px-2 py-0.5 text-xs font-medium ${style}`}
    >
      {type}
    </span>
  );
}
