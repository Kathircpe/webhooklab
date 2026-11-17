import { useEffect, useMemo, useState } from "react";
import { useParams, Link } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { getEndpoint, listEvents, getEvent, deleteEvent, clearEvents, updateEndpoint } from "../api/endpoints";
import type { EventSummary } from "../api/types";
import MethodBadge from "../components/MethodBadge";
import StatusBadge from "../components/StatusBadge";
import CopyButton from "../components/CopyButton";
import EmptyState from "../components/EmptyState";
import Skeleton from "../components/Skeleton";

export default function EndpointDetailPage() {
  const { id } = useParams();
  const qc = useQueryClient();
  const { data: ep } = useQuery({ queryKey: ["endpoint", id], queryFn: () => getEndpoint(id!) });
  const { data: evs, isLoading } = useQuery({ queryKey: ["events", id], queryFn: () => listEvents(id!, 100, 0) });

  const [selected, setSelected] = useState<string | null>(null);
  const [tab, setTab] = useState<"body" | "headers" | "query">("body");
  const [search, setSearch] = useState("");
  const [methodFilter, setMethodFilter] = useState("");
  const [pretty, setPretty] = useState(true);

  const { data: detail } = useQuery({
    queryKey: ["event", id, selected],
    queryFn: () => getEvent(id!, selected!),
    enabled: !!selected,
  });

  const prettyBody = useMemo(() => {
    if (!detail?.body_text) return null;
    try {
      return JSON.stringify(JSON.parse(detail.body_text), null, 2);
    } catch {
      return null;
    }
  }, [detail]);

  const rawBody = detail?.body_text || detail?.body_base64 || "";
  const bodyToShow = pretty && prettyBody ? prettyBody : rawBody;
  const copyText = async (text: string) => { await navigator.clipboard.writeText(text); };

  useEffect(() => {
    if (!id) return;
    const es = new EventSource(`/api/v1/endpoints/${id}/events/stream`);
    es.addEventListener("created", () => qc.invalidateQueries({ queryKey: ["events", id] }));
    return () => es.close();
  }, [id, qc]);

  const clearMut = useMutation({
    mutationFn: () => clearEvents(id!),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["events", id] }),
  });
  const delMut = useMutation({
    mutationFn: (eid: string) => deleteEvent(id!, eid),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["events", id] }); setSelected(null); },
  });
  const toggleMut = useMutation({
    mutationFn: (enabled: boolean) => updateEndpoint(id!, { is_enabled: enabled }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["endpoint", id] }),
  });

  if (!ep) return <Skeleton className="h-32" />;

  const filtered = (evs?.items || []).filter((e: EventSummary) => {
    if (methodFilter && e.method !== methodFilter) return false;
    if (search && !(`${e.method} ${e.path} ${e.content_type || ""}`.toLowerCase().includes(search.toLowerCase()))) return false;
    return true;
  });

  const downloadBody = () => {
    if (!detail?.body_base64) return;
    const bytes = Uint8Array.from(atob(detail.body_base64), (c) => c.charCodeAt(0));
    const blob = new Blob([bytes], { type: detail.content_type || "application/octet-stream" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `event-${detail.id}.bin`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className="space-y-4">
      <Link to="/" className="text-sm text-gray-500 hover:text-gray-900 dark:hover:text-gray-100">← Back to endpoints</Link>

      <div className="card p-4 flex flex-col md:flex-row md:items-center gap-4 justify-between">
        <div className="min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <StatusBadge status={ep.status} />
            <span className="font-medium">{ep.name || ep.slug}</span>
          </div>
          <div className="font-mono text-xs text-gray-600 dark:text-gray-300 break-all">{ep.url}</div>
          <div className="text-xs text-gray-500 mt-1">{ep.request_count} requests · expires {ep.expires_at ? new Date(ep.expires_at).toLocaleString() : "never"}</div>
        </div>
        <div className="flex flex-wrap gap-2 shrink-0">
          <CopyButton value={ep.url} label="Copy URL" />
          <button className="btn" onClick={() => toggleMut.mutate(!ep.is_enabled)}>{ep.is_enabled ? "Disable" : "Enable"}</button>
          <button className="btn-danger" onClick={() => { if (confirm("Clear all events?")) clearMut.mutate(); }}>Clear events</button>
        </div>
      </div>

      <div className="grid md:grid-cols-5 gap-4">
        <div className="md:col-span-2 card overflow-hidden">
          <div className="p-3 border-b border-gray-200 dark:border-gray-800 flex gap-2">
            <input className="input flex-1" placeholder="Search…" value={search} onChange={(e) => setSearch(e.target.value)} />
            <select className="input" value={methodFilter} onChange={(e) => setMethodFilter(e.target.value)}>
              <option value="">All</option>
              {["GET", "POST", "PUT", "PATCH", "DELETE"].map((m) => <option key={m} value={m}>{m}</option>)}
            </select>
          </div>
          <div className="max-h-[70vh] overflow-auto divide-y divide-gray-100 dark:divide-gray-800">
            {isLoading && <div className="p-4"><Skeleton className="h-6 mb-2" /><Skeleton className="h-6" /></div>}
            {!isLoading && filtered.length === 0 && <EmptyState title="No events" description="Send a request to the URL above." />}
            {filtered.map((e: EventSummary) => (
              <button
                key={e.id}
                onClick={() => setSelected(e.id)}
                className={`w-full text-left px-3 py-2 hover:bg-gray-50 dark:hover:bg-gray-800 transition ${selected === e.id ? "bg-blue-50 dark:bg-blue-950" : ""}`}
              >
                <div className="flex items-center justify-between gap-2">
                  <MethodBadge method={e.method} />
                  <span className="text-xs text-gray-400">{new Date(e.received_at).toLocaleTimeString()}</span>
                </div>
                <div className="text-xs font-mono text-gray-600 dark:text-gray-300 truncate mt-1">{e.path}</div>
                <div className="text-xs text-gray-400 mt-0.5">{e.content_type || "-"} · {e.body_size}B · {e.processing_duration_ms}ms</div>
              </button>
            ))}
          </div>
        </div>

        <div className="md:col-span-3 card p-4 min-h-[40vh]">
          {!detail && <EmptyState title="Select an event" description="Choose a request on the left to inspect it." />}
          {detail && (
            <div className="space-y-3">
              <div className="flex items-center justify-between gap-2 flex-wrap">
                <div className="flex items-center gap-2">
                  <MethodBadge method={detail.method} />
                  <span className="font-mono text-sm">{detail.path}</span>
                </div>
                <div className="flex gap-2">
                  <button className="btn" onClick={downloadBody}>Download</button>
                  <button className="btn-danger" onClick={() => { if (confirm("Delete event?")) delMut.mutate(detail.id); }}>Delete</button>
                </div>
              </div>
              <div className="text-xs text-gray-500">
                {detail.source_ip || "-"} · {new Date(detail.received_at).toLocaleString()} · {detail.body_size} bytes · {detail.processing_duration_ms}ms{detail.body_truncated ? " (truncated)" : ""}
              </div>

              <div className="flex gap-2 border-b border-gray-200 dark:border-gray-800">
                {(["body", "headers", "query"] as const).map((t) => (
                  <button
                    key={t}
                    onClick={() => setTab(t)}
                    className={`px-3 py-1.5 text-sm border-b-2 -mb-px ${tab === t ? "border-blue-600 text-blue-600" : "border-transparent text-gray-500"}`}
                  >
                    {t.charAt(0).toUpperCase() + t.slice(1)}
                  </button>
                ))}
              </div>

              {tab === "body" && (
                <div className="space-y-2">
                  <div className="flex items-center justify-between gap-2">
                    <div className="flex gap-1">
                      {prettyBody && (
                        <>
                          <button className={`btn ${pretty ? "border-blue-600 text-blue-600" : ""}`} onClick={() => setPretty(true)}>Pretty</button>
                          <button className={`btn ${!pretty ? "border-blue-600 text-blue-600" : ""}`} onClick={() => setPretty(false)}>Raw</button>
                        </>
                      )}
                    </div>
                    <button className="btn" onClick={() => copyText(rawBody)}>Copy payload</button>
                  </div>
                  <pre className="bg-gray-900 text-gray-100 rounded-md p-3 overflow-auto text-xs max-h-[50vh]">{bodyToShow}</pre>
                </div>
              )}
              {tab === "headers" && (
                <div className="space-y-2">
                  <div className="flex justify-end">
                    <button className="btn" onClick={() => copyText(JSON.stringify(detail.headers, null, 2))}>Copy headers</button>
                  </div>
                  <pre className="bg-gray-900 text-gray-100 rounded-md p-3 overflow-auto text-xs max-h-[50vh]">{JSON.stringify(detail.headers, null, 2)}</pre>
                </div>
              )}
              {tab === "query" && (
                <div className="space-y-2">
                  <div className="flex justify-end">
                    <button className="btn" onClick={() => copyText(JSON.stringify(detail.query_params, null, 2))}>Copy query</button>
                  </div>
                  <pre className="bg-gray-900 text-gray-100 rounded-md p-3 overflow-auto text-xs max-h-[50vh]">{JSON.stringify(detail.query_params, null, 2)}</pre>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
