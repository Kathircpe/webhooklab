package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"webhooklab/internal/config"
	"webhooklab/internal/model"
	"webhooklab/internal/repository"
	"webhooklab/internal/sse"
)

type Service struct {
	Cfg    config.Config
	Repo   repository.Repository
	Broker *sse.Broker
}

func New(cfg config.Config, repo repository.Repository, broker *sse.Broker) *Service {
	return &Service{Cfg: cfg, Repo: repo, Broker: broker}
}

const slugChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func genSlug() (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	out := make([]byte, 12)
	for i := 0; i < 12; i++ {
		out[i] = slugChars[int(b[i])%len(slugChars)]
	}
	return string(out), nil
}

func (s *Service) newSlug(ctx context.Context) (string, error) {
	for i := 0; i < 5; i++ {
		sl, err := genSlug()
		if err != nil {
			return "", err
		}
		// check uniqueness
		_, err = s.Repo.GetEndpointBySlug(ctx, sl)
		if err != nil {
			// not found expected
			return sl, nil
		}
	}
	return "", errors.New("could not generate unique slug")
}

func (s *Service) CreateEndpoint(ctx context.Context, req model.CreateEndpointReq) (model.Endpoint, error) {
	sl, err := s.newSlug(ctx)
	if err != nil {
		return model.Endpoint{}, err
	}
	var name *string
	if req.Name != nil && *req.Name != "" {
		n := *req.Name
		name = &n
	}
	var expiresAt *time.Time
	ttl := s.Cfg.DefaultTTL
	if req.TTLSeconds != nil {
		if *req.TTLSeconds == 0 {
			ttl = 0
		} else {
			ttl = time.Duration(*req.TTLSeconds) * time.Second
		}
	}
	if ttl > 0 {
		t := time.Now().Add(ttl)
		expiresAt = &t
	}
	ep, err := s.Repo.CreateEndpoint(ctx, sl, name, expiresAt)
	if err != nil {
		return model.Endpoint{}, err
	}
	ep.URL = s.Cfg.BaseURL + "/w/" + ep.Slug
	return ep, nil
}

func (s *Service) GetEndpoint(ctx context.Context, id string) (model.Endpoint, error) {
	ep, err := s.Repo.GetEndpointByID(ctx, id)
	if err != nil {
		return ep, err
	}
	ep.URL = s.Cfg.BaseURL + "/w/" + ep.Slug
	return ep, nil
}

func (s *Service) GetEndpointBySlug(ctx context.Context, slug string) (model.Endpoint, error) {
	ep, err := s.Repo.GetEndpointBySlug(ctx, slug)
	if err != nil {
		return ep, err
	}
	ep.URL = s.Cfg.BaseURL + "/w/" + ep.Slug
	return ep, nil
}

func (s *Service) ListEndpoints(ctx context.Context, limit, offset int, search string) ([]model.Endpoint, int64, error) {
	eps, total, err := s.Repo.ListEndpoints(ctx, limit, offset, search)
	if err != nil {
		return nil, 0, err
	}
	for i := range eps {
		eps[i].URL = s.Cfg.BaseURL + "/w/" + eps[i].Slug
	}
	return eps, total, nil
}

func (s *Service) UpdateEndpoint(ctx context.Context, id string, req model.UpdateEndpointReq) (model.Endpoint, error) {
	now := time.Now()
	var expiresAt *time.Time
	clearExpiry := req.ClearExpiry
	if req.ExpiresAt != nil {
		expiresAt = req.ExpiresAt
	}
	ep, err := s.Repo.UpdateEndpoint(ctx, id, req.Name, req.IsEnabled, expiresAt, clearExpiry, req.TTLSeconds, now)
	if err != nil {
		return ep, err
	}
	ep.URL = s.Cfg.BaseURL + "/w/" + ep.Slug
	return ep, nil
}

func (s *Service) DeleteEndpoint(ctx context.Context, id string) error {
	return s.Repo.DeleteEndpoint(ctx, id)
}

func (s *Service) Capture(ctx context.Context, slug string, ev repository.CapturedEvent, maxBody int64) (string, error) {
	ep, err := s.Repo.GetEndpointBySlug(ctx, slug)
	if err != nil {
		return "", err
	}
	if !ep.IsEnabled {
		return "", errors.New("disabled")
	}
	if ep.ExpiresAt != nil && time.Now().After(*ep.ExpiresAt) {
		return "", errors.New("expired")
	}
	id, err := s.Repo.InsertEvent(ctx, ev)
	if err != nil {
		return "", err
	}
	if err := s.Repo.TouchEndpoint(ctx, ep.ID, ev.ReceivedAt); err != nil {
		// non-fatal
	}
	if err := s.Repo.TrimEvents(ctx, ep.ID, s.Cfg.EventCap); err != nil {
		// non-fatal
	}
	// publish summary
	s.Broker.Publish(ep.ID, map[string]any{
		"type": "created",
		"event": map[string]any{
			"id":           id,
			"method":       ev.Method,
			"path":         ev.Path,
			"content_type": ev.ContentType,
			"body_size":    ev.BodySize,
			"source_ip":    ev.SourceIP,
			"received_at":  ev.ReceivedAt,
		},
	})
	return id, nil
}

func (s *Service) ListEvents(ctx context.Context, endpointID string, limit, offset int, method string) ([]model.EventSummary, int64, error) {
	return s.Repo.ListEvents(ctx, endpointID, limit, offset, method)
}

func (s *Service) GetEvent(ctx context.Context, id string) (model.EventDetail, error) {
	ed, err := s.Repo.GetEvent(ctx, id)
	if err != nil {
		return ed, err
	}
	// ensure body base64 not empty handling already
	_ = base64.StdEncoding
	return ed, nil
}

func (s *Service) SetEventDuration(ctx context.Context, id string, ms int64) error {
	return s.Repo.UpdateEventDuration(ctx, id, ms)
}

func (s *Service) DeleteEvent(ctx context.Context, id string) error {
	return s.Repo.DeleteEvent(ctx, id)
}

func (s *Service) ClearEvents(ctx context.Context, endpointID string) error {
	return s.Repo.DeleteEventsByEndpoint(ctx, endpointID)
}

func (s *Service) RecentEvents(ctx context.Context, limit int) ([]model.RecentEvent, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.Repo.RecentEvents(ctx, limit)
}

// Worker
func (s *Service) CleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(s.Cfg.CleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.Repo.PurgeExpired(ctx, s.Cfg.ExpiredRetention)
		}
	}
}
