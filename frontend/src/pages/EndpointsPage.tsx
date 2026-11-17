import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { listEndpoints, createEndpoint, deleteEndpoint, updateEndpoint } from "../api/endpoints";
import type { Endpoint } from "../api/types";

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

  const copy = (url: string) => navigator.clipboard.writeText(url);

  return (
    <div>
      <div className="card row">
        <input placeholder="name (optional)" value={name} onChange={(e) => setName(e.target.value)} />
        <input placeholder="ttl seconds" value={ttl} onChange={(e) => setTtl(e.target.value)} style={{ width: 120 }} />
        <button className="primary" onClick={() => createMut.mutate()} disabled={createMut.isPending}>
          Create endpoint
        </button>
      </div>

      {isLoading && <p>Loading…</p>}
      {data?.items.map((ep: Endpoint) => (
        <div className="card" key={ep.id}>
          <div className="row" style={{ justifyContent: "space-between" }}>
            <div>
              <span className={`badge ${ep.status}`}>{ep.status}</span>
              <strong style={{ marginLeft: 8 }}>{ep.name || ep.slug}</strong>
              <div className="muted">{ep.request_count} requests · created {new Date(ep.created_at).toLocaleString()}</div>
              <div className="mono">{ep.url}</div>
            </div>
            <div className="row">
              <button onClick={() => copy(ep.url)}>Copy URL</button>
              <button onClick={() => toggleMut.mutate({ id: ep.id, enabled: !ep.is_enabled })}>
                {ep.is_enabled ? "Disable" : "Enable"}
              </button>
              <a href={`/endpoints/${ep.id}`}><button>Details</button></a>
              <button onClick={() => { if (confirm("Delete?")) deleteMut.mutate(ep.id); }}>Delete</button>
            </div>
          </div>
        </div>
      ))}
      {data && data.items.length === 0 && <p className="muted">No endpoints yet.</p>}
    </div>
  );
}
