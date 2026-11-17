package model

import (
	"encoding/json"
	"strconv"
	"time"
)

type Endpoint struct {
	ID            string     `json:"id"`
	Slug          string     `json:"slug"`
	Name          *string    `json:"name"`
	IsEnabled     bool       `json:"is_enabled"`
	ExpiresAt     *time.Time `json:"expires_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	RequestCount  int64      `json:"request_count"`
	LastRequestAt *time.Time `json:"last_request_at"`
	Status        string     `json:"status"`
	URL           string     `json:"url"`
}

type EventSummary struct {
	ID          string    `json:"id"`
	Method      string    `json:"method"`
	Path        string    `json:"path"`
	ContentType *string   `json:"content_type"`
	BodySize    int64     `json:"body_size"`
	SourceIP    *string   `json:"source_ip"`
	ReceivedAt  time.Time `json:"received_at"`
}

type RecentEvent struct {
	ID           string     `json:"id"`
	EndpointID   string     `json:"endpoint_id"`
	EndpointSlug string     `json:"endpoint_slug"`
	EndpointName *string    `json:"endpoint_name"`
	Method       string     `json:"method"`
	Path         string     `json:"path"`
	ContentType  *string    `json:"content_type"`
	BodySize     int64      `json:"body_size"`
	SourceIP     *string    `json:"source_ip"`
	ReceivedAt   time.Time  `json:"received_at"`
}

type EventDetail struct {
	ID            string         `json:"id"`
	EndpointID    string         `json:"endpoint_id"`
	Method        string         `json:"method"`
	Path          string         `json:"path"`
	RawQuery      string         `json:"raw_query"`
	QueryParams   map[string]any `json:"query_params"`
	Headers       map[string]any `json:"headers"`
	ContentType   *string        `json:"content_type"`
	BodyBase64    string         `json:"body_base64"`
	BodyText      *string        `json:"body_text"`
	BodyTruncated bool           `json:"body_truncated"`
	BodySize      int64          `json:"body_size"`
	SourceIP      *string        `json:"source_ip"`
	ReceivedAt    time.Time      `json:"received_at"`
}

type CreateEndpointReq struct {
	Name       *string `json:"name"`
	TTLSeconds *int    `json:"ttl_seconds"`
}

type UpdateEndpointReq struct {
	Name        *string    `json:"name"`
	IsEnabled   *bool      `json:"is_enabled"`
	ExpiresAt   *time.Time `json:"-"`
	ClearExpiry bool       `json:"-"`
	TTLSeconds  *int       `json:"ttl_seconds"`
	hasExpires  bool
}

func (u *UpdateEndpointReq) UnmarshalJSON(b []byte) error {
	type alias UpdateEndpointReq
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	u.Name = a.Name
	u.IsEnabled = a.IsEnabled
	u.TTLSeconds = a.TTLSeconds
	// detect expires_at key
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if v, ok := raw["expires_at"]; ok {
		u.hasExpires = true
		if string(v) == "null" {
			u.ClearExpiry = true
			u.ExpiresAt = nil
		} else {
			var t time.Time
			if err := json.Unmarshal(v, &t); err != nil {
				// try RFC3339
				s, _ := strconv.Unquote(string(v))
				parsed, err := time.Parse(time.RFC3339, s)
				if err != nil {
					return err
				}
				t = parsed
			}
			u.ExpiresAt = &t
		}
	}
	return nil
}

type Page struct {
	Items  any   `json:"items"`
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}
