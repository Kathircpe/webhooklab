package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"webhooklab/internal/db"
)

func TestRepositoryIntegration(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	pool, err := newPoolForTest(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	repo := New(pool)
	ctx := context.Background()

	// migrate
	if err := migrateForTest(ctx, pool); err != nil {
		t.Fatal(err)
	}

	ep, err := repo.CreateEndpoint(ctx, "testslug"+t.Name(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if ep.Slug == "" {
		t.Fatal("missing slug")
	}

	id, err := repo.InsertEvent(ctx, CapturedEvent{
		EndpointID:  ep.ID,
		Method:      "POST",
		Path:        "/x",
		QueryParams: map[string]any{"a": []string{"1"}},
		Headers:     map[string]any{"Content-Type": []string{"application/json"}},
		ContentType: strPtr("application/json"),
		Body:        []byte(`{"hi":1}`),
		BodySize:    7,
		ReceivedAt:  time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetEvent(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got.BodyText == nil || *got.BodyText != `{"hi":1}` {
		t.Fatalf("body text = %v", got.BodyText)
	}
	list, total, err := repo.ListEvents(ctx, ep.ID, 10, 0, "")
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("list = %d/%d err=%v", len(list), total, err)
	}
	if err := repo.DeleteEndpoint(ctx, ep.ID); err != nil {
		t.Fatal(err)
	}
}

func newPoolForTest(dsn string) (*pgxpool.Pool, error) {
	return pgxpool.New(context.Background(), dsn)
}

func migrateForTest(ctx context.Context, pool *pgxpool.Pool) error {
	return db.Migrate(ctx, pool)
}

func strPtr(s string) *string { return &s }
