import { useEffect, useState } from "react";
import { useParams, Link } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { getEndpoint, listEvents, getEvent, deleteEvent, clearEvents, updateEndpoint } from "../api/endpoints";
import type { EventSummary } from "../api/types";

export default function EndpointDetailPage() {
  const { id } = useParams();
  const qc = useQueryClient();
  const { data: ep } = useQuery({ queryKey: ["endpoint", id], queryFn: () => getEndpoint(id!) });
  const { data: evs } = useQuery({ queryKey: ["events", id], queryFn: () => listEvents(id!) });

  const [selected, setSelected] = useState<string | null>(null);
  const { data: detail } = useQuery({
    queryKey: ["event", id, selected],
    queryFn: () => getEvent(id!, selected!),
    enabled: !!selected,
  });

  useEffect(() => {
    if (!id) return;
    const es = new EventSource(`/api/v1/endpoints/${id}/events/stream`);
    es.addEventListener("created", () => {
      qc.invalidateQueries({ queryKey: ["events", id] });
    });
    return () => es.close();
  }, [id, qc]);

  const clearMut = useMutation({
    mutationFn: () => clearEvents(id!),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["events", id] }); },
  });
  const delMut = useMutation({
    mutationFn: (eid: string) => deleteEvent(id!, eid),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["events", id] }); setSelected(null); },
  });
  const toggleMut = useMutation({
    mutationFn: (enabled: boolean) => updateEndpoint(id!, { is_enabled: enabled }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["endpoint", id] }),
  });

  if (!ep) return <p>Loading…</p>;

  return (
    <div>
      <Link to="/">← Back</Link>
      <div className="card">
        <div className="row" style={{ justifyContent: "space-between" }}>
          <div>
            <span className={`badge ${ep.status}`}>{ep.status}</span>
            <strong style={{ marginLeft: 8 }}>{ep.name || ep.slug}</strong>
            <div className="mono">{ep.url}</div>
            <div className="muted">{ep.request_count} requests</div>
          </div>
          <div className="row">
            <button onClick={() => navigator.clipboard.writeText(ep.url)}>Copy URL</button>
            <button onClick={() => toggleMut.mutate(!ep.is_enabled)}>{ep.is_enabled ? "Disable" : "Enable"}</button>
            <button onClick={() => { if (confirm("Clear all events?")) clearMut.mutate(); }}>Clear events</button>
          </div>
        </div>
      </div>

      <h3>Events</h3>
      <table>
        <thead><tr><th>Time</th><th>Method</th><th>Path</th><th>Type</th><th>Size</th><th></th></tr></thead>
        <tbody>
          {evs?.items.map((e: EventSummary) => (
            <tr key={e.id} style={{ cursor: "pointer", background: selected === e.id ? "#eff6ff" : undefined }} onClick={() => setSelected(e.id)}>
              <td>{new Date(e.received_at).toLocaleString()}</td>
              <td>{e.method}</td>
              <td>{e.path}</td>
              <td>{e.content_type || "-"}</td>
              <td>{e.body_size}</td>
              <td><button onClick={(ev) => { ev.stopPropagation(); if (confirm("Delete?")) delMut.mutate(e.id); }}>x</button></td>
            </tr>
          ))}
        </tbody>
      </table>
      {evs && evs.items.length === 0 && <p className="muted">No events yet. Send a request to {ep.url}</p>}

      {detail && (
        <div className="card" style={{ marginTop: 16 }}>
          <h4>Event detail</h4>
          <div className="muted">{detail.method} {detail.path} · {detail.source_ip} · {new Date(detail.received_at).toLocaleString()}</div>
          <h5>Headers</h5>
          <pre>{JSON.stringify(detail.headers, null, 2)}</pre>
          <h5>Query</h5>
          <pre>{JSON.stringify(detail.query_params, null, 2)}</pre>
          <h5>Body</h5>
          {detail.body_text ? <pre>{detail.body_text}</pre> : <pre>{detail.body_base64}</pre>}
          {detail.body_truncated && <p className="muted">Truncated</p>}
        </div>
      )}
    </div>
  );
}
