package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"webhooklab/internal/config"
	"webhooklab/internal/model"
	"webhooklab/internal/repository"
	"webhooklab/internal/sse"
)

type fakeRepo struct {
	endpoints map[string]model.Endpoint
	events    map[string][]repository.CapturedEvent
	nextID    int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{endpoints: map[string]model.Endpoint{}, events: map[string][]repository.CapturedEvent{}}
}

func (f *fakeRepo) CreateEndpoint(ctx context.Context, slug string, name *string, expiresAt *time.Time) (model.Endpoint, error) {
	ep := model.Endpoint{ID: slug + "-id", Slug: slug, Name: name, IsEnabled: true, ExpiresAt: expiresAt, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	f.endpoints[ep.ID] = ep
	return ep, nil
}
func (f *fakeRepo) GetEndpointByID(ctx context.Context, id string) (model.Endpoint, error) {
	ep, ok := f.endpoints[id]
	if !ok {
		return model.Endpoint{}, errors.New("not found")
	}
	return ep, nil
}
func (f *fakeRepo) GetEndpointBySlug(ctx context.Context, slug string) (model.Endpoint, error) {
	for _, ep := range f.endpoints {
		if ep.Slug == slug {
			return ep, nil
		}
	}
	return model.Endpoint{}, errors.New("not found")
}
func (f *fakeRepo) ListEndpoints(ctx context.Context, limit, offset int, search string) ([]model.Endpoint, int64, error) {
	out := []model.Endpoint{}
	for _, ep := range f.endpoints {
		out = append(out, ep)
	}
	return out, int64(len(out)), nil
}
func (f *fakeRepo) UpdateEndpoint(ctx context.Context, id string, name *string, isEnabled *bool, expiresAt *time.Time, clearExpiry bool, ttl *int, now time.Time) (model.Endpoint, error) {
	ep, ok := f.endpoints[id]
	if !ok {
		return model.Endpoint{}, errors.New("not found")
	}
	if name != nil {
		ep.Name = name
	}
	if isEnabled != nil {
		ep.IsEnabled = *isEnabled
	}
	if clearExpiry {
		ep.ExpiresAt = nil
	} else if expiresAt != nil {
		ep.ExpiresAt = expiresAt
	} else if ttl != nil {
		if *ttl == 0 {
			ep.ExpiresAt = nil
		} else {
			t := now.Add(time.Duration(*ttl) * time.Second)
			ep.ExpiresAt = &t
		}
	}
	f.endpoints[id] = ep
	return ep, nil
}
func (f *fakeRepo) DeleteEndpoint(ctx context.Context, id string) error {
	delete(f.endpoints, id)
	return nil
}
func (f *fakeRepo) TouchEndpoint(ctx context.Context, id string, at time.Time) error { return nil }
func (f *fakeRepo) InsertEvent(ctx context.Context, e repository.CapturedEvent) (string, error) {
	f.nextID++
	id := "ev" + itoa(f.nextID)
	f.events[e.EndpointID] = append(f.events[e.EndpointID], e)
	return id, nil
}
func (f *fakeRepo) TrimEvents(ctx context.Context, endpointID string, cap int) error { return nil }
func (f *fakeRepo) ListEvents(ctx context.Context, endpointID string, limit, offset int, method string) ([]model.EventSummary, int64, error) {
	return nil, 0, nil
}
func (f *fakeRepo) GetEvent(ctx context.Context, id string) (model.EventDetail, error) {
	return model.EventDetail{}, errors.New("not found")
}
func (f *fakeRepo) DeleteEvent(ctx context.Context, id string) error { return nil }
func (f *fakeRepo) DeleteEventsByEndpoint(ctx context.Context, endpointID string) error {
	delete(f.events, endpointID)
	return nil
}
func (f *fakeRepo) PurgeExpired(ctx context.Context, retention time.Duration) (int64, error) {
	return 0, nil
}
func (f *fakeRepo) RecentEvents(ctx context.Context, limit int) ([]model.RecentEvent, error) {
	return nil, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

func TestCreateEndpointDefaultTTL(t *testing.T) {
	repo := newFakeRepo()
	svc := New(config.Config{BaseURL: "http://localhost", DefaultTTL: time.Hour}, repo, sse.New())
	ep, err := svc.CreateEndpoint(context.Background(), model.CreateEndpointReq{Name: ptr("x")})
	if err != nil {
		t.Fatal(err)
	}
	if ep.ExpiresAt == nil {
		t.Fatal("expected expiry")
	}
	if ep.URL != "http://localhost/w/"+ep.Slug {
		t.Fatalf("url = %s", ep.URL)
	}
}

func TestCaptureDisabledAndExpired(t *testing.T) {
	repo := newFakeRepo()
	svc := New(config.Config{BaseURL: "http://localhost", DefaultTTL: time.Hour, EventCap: 10}, repo, sse.New())
	ep, _ := svc.CreateEndpoint(context.Background(), model.CreateEndpointReq{})
	// disable
	_, err := svc.UpdateEndpoint(context.Background(), ep.ID, model.UpdateEndpointReq{IsEnabled: boolPtr(false)})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Capture(context.Background(), ep.Slug, repository.CapturedEvent{Method: "GET", ReceivedAt: time.Now()}, 1024)
	if err == nil || err.Error() != "disabled" {
		t.Fatalf("expected disabled, got %v", err)
	}
	// expired
	past := time.Now().Add(-time.Hour)
	_, _ = svc.UpdateEndpoint(context.Background(), ep.ID, model.UpdateEndpointReq{IsEnabled: boolPtr(true), ExpiresAt: &past})
	_, err = svc.Capture(context.Background(), ep.Slug, repository.CapturedEvent{Method: "GET", ReceivedAt: time.Now()}, 1024)
	if err == nil || err.Error() != "expired" {
		t.Fatalf("expected expired, got %v", err)
	}
}

func ptr(s string) *string { return &s }
func boolPtr(b bool) *bool { return &b }
