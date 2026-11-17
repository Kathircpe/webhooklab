package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"webhooklab/internal/model"
	"webhooklab/internal/service"
)

type EndpointHandler struct {
	Svc *service.Service
}

func (h *EndpointHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateEndpointReq
	if err := decodeJSON(r, &req); err != nil && err.Error() != "EOF" {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	ep, err := h.Svc.CreateEndpoint(r.Context(), req)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, ep)
}

func (h *EndpointHandler) List(w http.ResponseWriter, r *http.Request) {
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
	search := r.URL.Query().Get("search")
	eps, total, err := h.Svc.ListEndpoints(r.Context(), limit, offset, search)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Page{Items: eps, Total: total, Limit: limit, Offset: offset})
}

func (h *EndpointHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ep, err := h.Svc.GetEndpoint(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, ep)
}

func (h *EndpointHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req model.UpdateEndpointReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	ep, err := h.Svc.UpdateEndpoint(r.Context(), id, req)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ep)
}

func (h *EndpointHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.Svc.DeleteEndpoint(r.Context(), id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
