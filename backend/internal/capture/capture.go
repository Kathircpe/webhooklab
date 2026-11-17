package capture

import (
	"io"
	"net/http"
	"strings"
	"time"

	"webhooklab/internal/repository"
)

func ClientIP(r *http.Request) string {
	addr := r.RemoteAddr
	if i := strings.LastIndex(addr, ":"); i >= 0 && !strings.Contains(addr[i+1:], ":") {
		addr = addr[:i]
	}
	return addr
}

func FromRequest(r *http.Request, endpointID string, maxBody int64, sourceIP string, now time.Time) (repository.CapturedEvent, error) {
	lr := &io.LimitedReader{R: r.Body, N: maxBody + 1}
	body, err := io.ReadAll(lr)
	if err != nil {
		return repository.CapturedEvent{}, err
	}
	originalSize := int64(len(body))
	truncated := originalSize > maxBody
	if truncated {
		body = body[:maxBody]
	}

	qp := map[string]any{}
	for k, v := range r.URL.Query() {
		qp[k] = v
	}
	hd := map[string]any{}
	for k, v := range r.Header {
		hd[k] = v
	}

	var ct *string
	if c := r.Header.Get("Content-Type"); c != "" {
		ct = &c
	}
	var sip *string
	if sourceIP != "" {
		s := sourceIP
		sip = &s
	}
	path := r.URL.Path
	if path == "" {
		path = "/"
	}

	return repository.CapturedEvent{
		EndpointID:    endpointID,
		Method:        r.Method,
		Path:          path,
		RawQuery:      r.URL.RawQuery,
		QueryParams:   qp,
		Headers:       hd,
		ContentType:   ct,
		Body:          body,
		BodySize:      originalSize,
		BodyTruncated: truncated,
		SourceIP:      sip,
		ReceivedAt:    now,
	}, nil
}
