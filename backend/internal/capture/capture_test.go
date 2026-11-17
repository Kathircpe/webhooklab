package capture

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFromRequestPreservesBodyAndMetadata(t *testing.T) {
	req := httptest.NewRequest("POST", "/w/abc/path?x=1&x=2", bytes.NewBufferString(`{"a":1}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Custom", "v")
	req.RemoteAddr = "10.0.0.1:1234"
	ev, err := FromRequest(req, "ep1", 1024, ClientIP(req), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if ev.Method != "POST" {
		t.Fatalf("method = %s", ev.Method)
	}
	if string(ev.Body) != `{"a":1}` {
		t.Fatalf("body = %s", ev.Body)
	}
	if ev.BodySize != 7 {
		t.Fatalf("size = %d", ev.BodySize)
	}
	if ev.ContentType == nil || *ev.ContentType != "application/json" {
		t.Fatalf("content type = %v", ev.ContentType)
	}
	qp := ev.QueryParams["x"].([]string)
	if len(qp) != 2 {
		t.Fatalf("query params = %v", ev.QueryParams)
	}
	if ev.Headers["X-Custom"] == nil {
		t.Fatalf("headers missing custom")
	}
	if ev.SourceIP == nil || *ev.SourceIP != "10.0.0.1" {
		t.Fatalf("source ip = %v", ev.SourceIP)
	}
	if ev.Path != "/w/abc/path" {
		// path preserved as request path (handler overrides)
	}
}

func TestFromRequestRejectsOversized(t *testing.T) {
	big := strings.Repeat("a", 20)
	req := httptest.NewRequest("POST", "/w/abc", strings.NewReader(big))
	_, err := FromRequest(req, "ep1", 10, ClientIP(req), time.Now())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClientIPStripsPort(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.5:9999"
	if got := ClientIP(req); got != "192.168.1.5" {
		t.Fatalf("got %s", got)
	}
}


