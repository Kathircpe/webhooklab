import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { listEndpoints, createEndpoint, deleteEndpoint, updateEndpoint } from "../api/endpoints";
import type { Endpoint } from "../api/types";
import CopyButton from "../components/CopyButton";
import EmptyState from "../components/EmptyState";
import Skeleton from "../components/Skeleton";
import StatusBadge from "../components/StatusBadge";

export default function EndpointsPage() {
  const qc = useQueryClient();
  const [name, setName] = useState("");
  const [ttl, setTtl] = useState("604800");
  const { data, isLoading } = useQuery({ queryKey: ["endpoints"], queryFn: () => listEndpoints() });

  const createMut = useMutation({
    mutationFn: () => createEndpoint({ name: name || undefined, ttl_seconds: Number(ttl) || undefined }),
    onSuccess: () => { setName(""); qc.invalidateQueries({ queryKey: ["endpoints"] }); },
  });
  const deleteMut = useMutation({
    mutationFn: (id: string) => deleteEndpoint(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["endpoints"] }),
  });
  const toggleMut = useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) => updateEndpoint(id, { is_enabled: enabled }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["endpoints"] }),
  });

  return (
    <div>
      <div className="mb-6 flex flex-col sm:flex-row gap-3 items-stretch sm:items-center">
        <input className="input flex-1" placeholder="Endpoint name (optional)" value={name} onChange={(e) => setName(e.target.value)} />
        <input className="input w-32" placeholder="TTL sec" value={ttl} onChange={(e) => setTtl(e.target.value)} />
        <button className="btn-primary" onClick={() => createMut.mutate()} disabled={createMut.isPending}>
          {createMut.isPending ? "Creating…" : "Create endpoint"}
        </button>
      </div>

      {isLoading && (
        <div className="grid gap-3">
          {[0, 1, 2].map((i) => <Skeleton key={i} className="h-24" />)}
        </div>
      )}

      {!isLoading && data?.items.length === 0 && (
        <EmptyState title="No endpoints yet" description="Create one above to start capturing webhooks." />
      )}

      <div className="grid gap-3">
        {data?.items.map((ep: Endpoint) => (
          <div key={ep.id} className="card p-4 flex flex-col sm:flex-row sm:items-center gap-4 justify-between">
            <div className="min-w-0">
              <div className="flex items-center gap-2 mb-1">
                <StatusBadge status={ep.status} />
                <span className="font-medium truncate">{ep.name || ep.slug}</span>
              </div>
              <div className="text-xs text-gray-500 dark:text-gray-400 mb-2">
                {ep.request_count} requests · {new Date(ep.created_at).toLocaleString()}
              </div>
              <div className="font-mono text-xs text-gray-600 dark:text-gray-300 truncate">{ep.url}</div>
            </div>
            <div className="flex flex-wrap gap-2 shrink-0">
              <CopyButton value={ep.url} label="Copy URL" />
              <button className="btn" onClick={() => toggleMut.mutate({ id: ep.id, enabled: !ep.is_enabled })}>
                {ep.is_enabled ? "Disable" : "Enable"}
              </button>
              <a href={`/endpoints/${ep.id}`} className="btn">Details</a>
              <button
                className="btn-danger"
                onClick={() => { if (confirm("Delete this endpoint and all events?")) deleteMut.mutate(ep.id); }}
              >
                Delete
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
