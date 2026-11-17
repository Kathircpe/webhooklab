package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"webhooklab/internal/model"
	"webhooklab/internal/service"
)

type ReplayHandler struct {
	Svc *service.Service
}

func (h *ReplayHandler) Send(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventId")
	var req model.ReplayRequest
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.DestinationURL == "" {
		writeErr(w, http.StatusBadRequest, "destination_url is required")
		return
	}
	rep, err := h.Svc.ReplayEvent(r.Context(), eventID, req)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rep)
}

func (h *ReplayHandler) List(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventId")
	items, err := h.Svc.ListReplays(r.Context(), eventID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
