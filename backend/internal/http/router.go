package httpx

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"webhooklab/internal/config"
	"webhooklab/internal/http/handlers"
	"webhooklab/internal/http/middleware"
	"webhooklab/internal/service"
)

func NewRouter(cfg config.Config, svc *service.Service, health *handlers.HealthHandler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recover)
	r.Use(middleware.Logger)
	r.Use(middleware.CORS(cfg.CORSOrigins))
	r.Use(middleware.RequestID)

	eh := &handlers.EndpointHandler{Svc: svc}
	evh := &handlers.EventHandler{Svc: svc}
	caph := &handlers.CaptureHandler{Svc: svc}

	r.Get("/healthz", health.Healthz)
	r.Get("/readyz", health.Readyz)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/endpoints", func(r chi.Router) {
			r.Post("/", eh.Create)
			r.Get("/", eh.List)
			r.Get("/{id}", eh.Get)
			r.Patch("/{id}", eh.Update)
			r.Delete("/{id}", eh.Delete)
			r.Get("/{id}/events", evh.List)
			r.Get("/{id}/events/stream", evh.Stream)
			r.Delete("/{id}/events", evh.Clear)
			r.Get("/{id}/events/{eventId}", evh.Get)
			r.Delete("/{id}/events/{eventId}", evh.Delete)
		})
	})

	r.Handle("/w/*", http.HandlerFunc(caph.Handle))

	// serve embedded frontend? optional; for dev, frontend separate.
	return r
}
