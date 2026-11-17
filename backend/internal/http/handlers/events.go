package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"webhooklab/internal/model"
	"webhooklab/internal/service"
)

type EventHandler struct {
	Svc *service.Service
}

func (h *EventHandler) List(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	limit := 50
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	method := r.URL.Query().Get("method")
	eps, total, err := h.Svc.ListEvents(r.Context(), id, limit, offset, method)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Page{Items: eps, Total: total, Limit: limit, Offset: offset})
}

func (h *EventHandler) Get(w http.ResponseWriter, r *http.Request) {
	eid := chi.URLParam(r, "eventId")
	ed, err := h.Svc.GetEvent(r.Context(), eid)
	if err != nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, ed)
}

func (h *EventHandler) Delete(w http.ResponseWriter, r *http.Request) {
	eid := chi.URLParam(r, "eventId")
	if err := h.Svc.DeleteEvent(r.Context(), eid); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *EventHandler) Clear(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.Svc.ClearEvents(r.Context(), id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *EventHandler) Stream(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fl, ok := w.(http.Flusher)
	if !ok {
		writeErr(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch := h.Svc.Broker.Subscribe(id)
	defer h.Svc.Broker.Unsubscribe(id, ch)

	// initial ping
	w.Write([]byte(": connected\n\n"))
	fl.Flush()

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case data := <-ch:
			w.Write([]byte("event: created\ndata: "))
			w.Write(data)
			w.Write([]byte("\n\n"))
			fl.Flush()
		case <-ticker.C:
			w.Write([]byte(": ping\n\n"))
			fl.Flush()
		}
	}
}
