export type Endpoint = {
  id: string;
  slug: string;
  name: string | null;
  is_enabled: boolean;
  expires_at: string | null;
  created_at: string;
  updated_at: string;
  request_count: number;
  last_request_at: string | null;
  status: string;
  url: string;
};

export type EventSummary = {
  id: string;
  method: string;
  path: string;
  content_type: string | null;
  body_size: number;
  source_ip: string | null;
  received_at: string;
};

export type EventDetail = {
  id: string;
  endpoint_id: string;
  method: string;
  path: string;
  raw_query: string;
  query_params: Record<string, any>;
  headers: Record<string, any>;
  content_type: string | null;
  body_base64: string;
  body_text: string | null;
  body_truncated: boolean;
  body_size: number;
  source_ip: string | null;
  received_at: string;
};

export type Page<T> = { items: T[]; total: number; limit: number; offset: number };

export type RecentEvent = {
  id: string;
  endpoint_id: string;
  endpoint_slug: string;
  endpoint_name: string | null;
  method: string;
  path: string;
  content_type: string | null;
  body_size: number;
  source_ip: string | null;
  received_at: string;
};
export type CreateEndpointReq = { name?: string; ttl_seconds?: number };
export type UpdateEndpointReq = { name?: string; is_enabled?: boolean; expires_at?: string | null; ttl_seconds?: number };
