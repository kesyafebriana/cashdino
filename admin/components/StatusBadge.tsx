const STYLES: Record<string, string> = {
  active: "bg-green-50 text-green-700",
  completed: "bg-gray-100 text-gray-500",
  draft: "bg-orange-50 text-[#e65100]",
  scheduled: "bg-blue-50 text-blue-700",
  delivered: "bg-green-50 text-green-700",
  failed: "bg-red-50 text-red-600",
  pending: "bg-yellow-50 text-yellow-700",
};

export function StatusBadge({ status }: { status: string }) {
  const style = STYLES[status] ?? "bg-gray-100 text-gray-500";
  return (
    <span
      className={`inline-block rounded-full px-2.5 py-0.5 text-xs font-semibold capitalize ${style}`}
    >
      {status}
    </span>
  );
}
