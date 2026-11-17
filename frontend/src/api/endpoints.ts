import { api } from "./client";
import type { Endpoint, Page, CreateEndpointReq, UpdateEndpointReq, EventSummary, EventDetail } from "./types";

export const listEndpoints = (limit = 50, offset = 0) =>
  api<Page<Endpoint>>(`/api/v1/endpoints?limit=${limit}&offset=${offset}`);

export const getEndpoint = (id: string) => api<Endpoint>(`/api/v1/endpoints/${id}`);

export const createEndpoint = (req: CreateEndpointReq) =>
  api<Endpoint>(`/api/v1/endpoints`, { method: "POST", body: JSON.stringify(req) });

export const updateEndpoint = (id: string, req: UpdateEndpointReq) =>
  api<Endpoint>(`/api/v1/endpoints/${id}`, { method: "PATCH", body: JSON.stringify(req) });

export const deleteEndpoint = (id: string) =>
  api<void>(`/api/v1/endpoints/${id}`, { method: "DELETE" });

export const listEvents = (id: string, limit = 50, offset = 0) =>
  api<Page<EventSummary>>(`/api/v1/endpoints/${id}/events?limit=${limit}&offset=${offset}`);

export const getEvent = (id: string, eventId: string) =>
  api<EventDetail>(`/api/v1/endpoints/${id}/events/${eventId}`);

export const deleteEvent = (id: string, eventId: string) =>
  api<void>(`/api/v1/endpoints/${id}/events/${eventId}`, { method: "DELETE" });

export const clearEvents = (id: string) =>
  api<void>(`/api/v1/endpoints/${id}/events`, { method: "DELETE" });
