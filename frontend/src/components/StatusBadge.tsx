export default function StatusBadge({ status }: { status: string }) {
  const map: Record<string, string> = {
    active: "bg-emerald-100 text-emerald-700 dark:bg-emerald-900 dark:text-emerald-300",
    disabled: "bg-amber-100 text-amber-700 dark:bg-amber-900 dark:text-amber-300",
    expired: "bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300",
  };
  return <span className={`badge ${map[status] || "bg-gray-100 text-gray-700"}`}>{status}</span>;
}
