package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"webhooklab/internal/capture"
	"webhooklab/internal/service"
)

type CaptureHandler struct {
	Svc *service.Service
}

func (h *CaptureHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// path: /w/{slug}/...
	p := strings.TrimPrefix(r.URL.Path, "/w/")
	parts := strings.SplitN(p, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	slug := parts[0]
	remainder := "/"
	if len(parts) == 2 && parts[1] != "" {
		remainder = "/" + parts[1]
	}

	ep, err := h.Svc.GetEndpointBySlug(r.Context(), slug)
	if err != nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	if !ep.IsEnabled {
		writeErr(w, http.StatusGone, "endpoint disabled")
		return
	}
	if ep.ExpiresAt != nil && time.Now().After(*ep.ExpiresAt) {
		writeErr(w, http.StatusGone, "endpoint expired")
		return
	}

	srcIP := capture.ClientIP(r)
	ev, err := capture.FromRequest(r, ep.ID, h.Svc.Cfg.MaxBodyBytes, srcIP, time.Now())
	if err != nil {
		if strings.Contains(err.Error(), "too large") {
			writeErr(w, http.StatusRequestEntityTooLarge, err.Error())
			return
		}
		writeErr(w, http.StatusBadRequest, "read body failed")
		return
	}
	ev.Path = remainder

	id, err := h.Svc.Capture(r.Context(), slug, ev, h.Svc.Cfg.MaxBodyBytes)
	if err != nil {
		if err.Error() == "disabled" || err.Error() == "expired" {
			writeErr(w, http.StatusGone, err.Error())
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"received": true, "event_id": id})
}
