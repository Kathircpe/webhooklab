import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { listEndpoints } from "../api/endpoints";
import { recentEvents } from "../api/endpoints";
import type { RecentEvent } from "../api/types";
import MethodBadge from "../components/MethodBadge";

import Skeleton from "../components/Skeleton";
import EmptyState from "../components/EmptyState";

export default function DashboardPage() {
  const { data: eps } = useQuery({ queryKey: ["endpoints"], queryFn: () => listEndpoints() });
  const { data: recent, isLoading } = useQuery({ queryKey: ["recent"], queryFn: () => recentEvents(25) });

  const endpoints = eps?.items || [];
  const totalRequests = endpoints.reduce((a, e) => a + e.request_count, 0);
  const active = endpoints.filter((e) => e.status === "active").length;
  const expired = endpoints.filter((e) => e.status === "expired").length;

  const stats = [
    { label: "Endpoints", value: endpoints.length },
    { label: "Active", value: active },
    { label: "Expired", value: expired },
    { label: "Total requests", value: totalRequests },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-semibold">Dashboard</h2>
        <p className="text-sm text-gray-500">Overview of your webhook endpoints and recent activity.</p>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        {stats.map((s) => (
          <div key={s.label} className="card p-4">
            <div className="text-xs text-gray-500 uppercase tracking-wide">{s.label}</div>
            <div className="text-2xl font-semibold mt-1">{s.value}</div>
          </div>
        ))}
      </div>

      <div className="card overflow-hidden">
        <div className="px-4 py-3 border-b border-gray-200 dark:border-gray-800 flex items-center justify-between">
          <h3 className="font-medium">Recent activity</h3>
          <Link to="/endpoints" className="text-sm text-blue-600">Manage endpoints →</Link>
        </div>
        <div className="overflow-auto">
          <table className="w-full text-sm">
            <thead className="text-left text-xs text-gray-500 uppercase">
              <tr>
                <th className="px-4 py-2">Time</th>
                <th className="px-4 py-2">Endpoint</th>
                <th className="px-4 py-2">Method</th>
                <th className="px-4 py-2">Path</th>
                <th className="px-4 py-2">Size</th>
                <th className="px-4 py-2">Duration</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-800">
              {isLoading && [0, 1, 2, 3].map((i) => (
                <tr key={i}><td colSpan={6} className="px-4 py-3"><Skeleton className="h-4" /></td></tr>
              ))}
              {!isLoading && (recent?.items || []).length === 0 && (
                <tr><td colSpan={6}><EmptyState title="No activity yet" description="Create an endpoint and send a request." /></td></tr>
              )}
              {(recent?.items || []).map((e: RecentEvent) => (
                <tr key={e.id} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                  <td className="px-4 py-2 text-xs text-gray-500">{new Date(e.received_at).toLocaleString()}</td>
                  <td className="px-4 py-2">
                    <Link to={`/endpoints/${e.endpoint_id}`} className="text-blue-600 hover:underline">
                      {e.endpoint_name || e.endpoint_slug}
                    </Link>
                  </td>
                  <td className="px-4 py-2"><MethodBadge method={e.method} /></td>
                  <td className="px-4 py-2 font-mono text-xs">{e.path}</td>
                  <td className="px-4 py-2 text-xs">{e.body_size}B</td>
                  <td className="px-4 py-2 text-xs">{e.processing_duration_ms}ms</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
