const colors: Record<string, string> = {
  GET: "bg-emerald-100 text-emerald-700 dark:bg-emerald-900 dark:text-emerald-300",
  POST: "bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300",
  PUT: "bg-amber-100 text-amber-700 dark:bg-amber-900 dark:text-amber-300",
  PATCH: "bg-amber-100 text-amber-700 dark:bg-amber-900 dark:text-amber-300",
  DELETE: "bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300",
};

export default function MethodBadge({ method }: { method: string }) {
  return (
    <span className={`badge font-mono ${colors[method] || "bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300"}`}>
      {method}
    </span>
  );
}
