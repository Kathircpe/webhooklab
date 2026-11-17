package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"webhooklab/internal/model"
)

type Repository interface {
	CreateEndpoint(ctx context.Context, slug string, name *string, expiresAt *time.Time) (model.Endpoint, error)
	GetEndpointByID(ctx context.Context, id string) (model.Endpoint, error)
	GetEndpointBySlug(ctx context.Context, slug string) (model.Endpoint, error)
	ListEndpoints(ctx context.Context, limit, offset int, search string) ([]model.Endpoint, int64, error)
	UpdateEndpoint(ctx context.Context, id string, name *string, isEnabled *bool, expiresAt *time.Time, clearExpiry bool, ttl *int, now time.Time) (model.Endpoint, error)
	DeleteEndpoint(ctx context.Context, id string) error
	TouchEndpoint(ctx context.Context, id string, at time.Time) error
	InsertEvent(ctx context.Context, e CapturedEvent) (string, error)
	TrimEvents(ctx context.Context, endpointID string, cap int) error
	ListEvents(ctx context.Context, endpointID string, limit, offset int, method string) ([]model.EventSummary, int64, error)
	GetEvent(ctx context.Context, id string) (model.EventDetail, error)
	UpdateEventDuration(ctx context.Context, id string, ms int64) error
	DeleteEvent(ctx context.Context, id string) error
	DeleteEventsByEndpoint(ctx context.Context, endpointID string) error
	PurgeExpired(ctx context.Context, retention time.Duration) (int64, error)
	RecentEvents(ctx context.Context, limit int) ([]model.RecentEvent, error)
	InsertReplay(ctx context.Context, eventID, destinationURL, method string, reqHeaders map[string]any, reqBody []byte, status *int, respHeaders map[string]any, respBody []byte, durationMS int64, errMsg *string) (model.Replay, error)
	ListReplaysByEvent(ctx context.Context, eventID string) ([]model.Replay, error)
	GetReplay(ctx context.Context, id string) (model.Replay, error)
}

type Repo struct {
	Pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo { return &Repo{Pool: pool} }

func statusOf(isEnabled bool, expiresAt *time.Time) string {
	if expiresAt != nil && time.Now().After(*expiresAt) {
		return "expired"
	}
	if !isEnabled {
		return "disabled"
	}
	return "active"
}

func (r *Repo) CreateEndpoint(ctx context.Context, slug string, name *string, expiresAt *time.Time) (model.Endpoint, error) {
	row := r.Pool.QueryRow(ctx, `
		INSERT INTO endpoints (slug, name, expires_at)
		VALUES ($1,$2,$3)
		RETURNING id, slug, name, is_enabled, expires_at, created_at, updated_at, request_count, last_request_at`,
		slug, name, expiresAt)
	var e model.Endpoint
	var id string
	var nm *string
	var exp, last *time.Time
	var rc int64
	var enabled bool
	var created, updated time.Time
	if err := row.Scan(&id, &e.Slug, &nm, &enabled, &exp, &created, &updated, &rc, &last); err != nil {
		return e, err
	}
	e.ID = id
	e.Name = nm
	e.IsEnabled = enabled
	e.ExpiresAt = exp
	e.CreatedAt = created
	e.UpdatedAt = updated
	e.RequestCount = rc
	e.LastRequestAt = last
	e.Status = statusOf(e.IsEnabled, e.ExpiresAt)
	return e, nil
}

func (r *Repo) GetEndpointByID(ctx context.Context, id string) (model.Endpoint, error) {
	return r.scanEndpoint(r.Pool.QueryRow(ctx, `
		SELECT id, slug, name, is_enabled, expires_at, created_at, updated_at, request_count, last_request_at
		FROM endpoints WHERE id=$1`, id))
}

func (r *Repo) GetEndpointBySlug(ctx context.Context, slug string) (model.Endpoint, error) {
	return r.scanEndpoint(r.Pool.QueryRow(ctx, `
		SELECT id, slug, name, is_enabled, expires_at, created_at, updated_at, request_count, last_request_at
		FROM endpoints WHERE slug=$1`, slug))
}

func (r *Repo) scanEndpoint(row pgx.Row) (model.Endpoint, error) {
	var e model.Endpoint
	var id string
	var nm *string
	var enabled bool
	var exp, last *time.Time
	var rc int64
	var created, updated time.Time
	var slug string
	if err := row.Scan(&id, &slug, &nm, &enabled, &exp, &created, &updated, &rc, &last); err != nil {
		return e, err
	}
	e.ID = id
	e.Slug = slug
	e.Name = nm
	e.IsEnabled = enabled
	e.ExpiresAt = exp
	e.CreatedAt = created
	e.UpdatedAt = updated
	e.RequestCount = rc
	e.LastRequestAt = last
	e.Status = statusOf(e.IsEnabled, e.ExpiresAt)
	return e, nil
}

func (r *Repo) ListEndpoints(ctx context.Context, limit, offset int, search string) ([]model.Endpoint, int64, error) {
	where := ""
	args := []any{}
	if search != "" {
		where = "WHERE name ILIKE $1 OR slug ILIKE $1"
		args = append(args, "%"+search+"%")
	}
	// total
	var total int64
	q := "SELECT COUNT(*) FROM endpoints " + where
	if err := r.Pool.QueryRow(ctx, q, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	// rows
	limitIdx := len(args) + 1
	offsetIdx := len(args) + 2
	args = append(args, limit, offset)
	rows, err := r.Pool.Query(ctx, `
		SELECT id, slug, name, is_enabled, expires_at, created_at, updated_at, request_count, last_request_at
		FROM endpoints `+where+` ORDER BY created_at DESC LIMIT $`+itoa(limitIdx)+` OFFSET $`+itoa(offsetIdx), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []model.Endpoint{}
	for rows.Next() {
		e, err := r.scanEndpoint(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

func (r *Repo) UpdateEndpoint(ctx context.Context, id string, name *string, isEnabled *bool, expiresAt *time.Time, clearExpiry bool, ttl *int, now time.Time) (model.Endpoint, error) {
	// build expiry
	var exp *time.Time
	if clearExpiry {
		exp = nil
		_ = expiresAt
	} else if expiresAt != nil {
		exp = expiresAt
	} else if ttl != nil {
		if *ttl == 0 {
			exp = nil
		} else {
			t := now.Add(time.Duration(*ttl) * time.Second)
			exp = &t
		}
	} else {
		// no change: keep existing by not updating expires_at? We'll set to current via subselect
	}
	// We'll do a COALESCE approach: if exp pointer nil and not clear, keep existing.
	if clearExpiry || expiresAt != nil || ttl != nil {
		_, err := r.Pool.Exec(ctx, `
			UPDATE endpoints SET
				name = COALESCE($2, name),
				is_enabled = COALESCE($3, is_enabled),
				expires_at = $4,
				updated_at = now()
			WHERE id=$1`, id, name, isEnabled, exp)
		if err != nil {
			return model.Endpoint{}, err
		}
	} else {
		_, err := r.Pool.Exec(ctx, `
			UPDATE endpoints SET
				name = COALESCE($2, name),
				is_enabled = COALESCE($3, is_enabled),
				updated_at = now()
			WHERE id=$1`, id, name, isEnabled)
		if err != nil {
			return model.Endpoint{}, err
		}
	}
	return r.GetEndpointByID(ctx, id)
}

func (r *Repo) DeleteEndpoint(ctx context.Context, id string) error {
	_, err := r.Pool.Exec(ctx, `DELETE FROM endpoints WHERE id=$1`, id)
	return err
}

func (r *Repo) TouchEndpoint(ctx context.Context, id string, at time.Time) error {
	_, err := r.Pool.Exec(ctx, `UPDATE endpoints SET request_count=request_count+1, last_request_at=$2, updated_at=now() WHERE id=$1`, id, at)
	return err
}

// Events

type CapturedEvent struct {
	EndpointID    string
	Method        string
	Path          string
	RawQuery      string
	QueryParams   map[string]any
	Headers       map[string]any
	ContentType   *string
	Body          []byte
	BodySize      int64
	BodyTruncated bool
	SourceIP      *string
	ReceivedAt    time.Time
}

func (r *Repo) InsertEvent(ctx context.Context, e CapturedEvent) (string, error) {
	qp, _ := json.Marshal(e.QueryParams)
	hd, _ := json.Marshal(e.Headers)
	var id string
	err := r.Pool.QueryRow(ctx, `
		INSERT INTO events (endpoint_id, method, path, raw_query, query_params, headers, content_type, body, body_size, body_truncated, source_ip, received_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id`,
		e.EndpointID, e.Method, e.Path, e.RawQuery, qp, hd, e.ContentType, e.Body, e.BodySize, e.BodyTruncated, e.SourceIP, e.ReceivedAt).Scan(&id)
	return id, err
}

func (r *Repo) TrimEvents(ctx context.Context, endpointID string, cap int) error {
	if cap <= 0 {
		return nil
	}
	_, err := r.Pool.Exec(ctx, `
		DELETE FROM events
		WHERE endpoint_id=$1
		  AND id NOT IN (
		    SELECT id FROM events WHERE endpoint_id=$1 ORDER BY received_at DESC LIMIT $2
		  )`, endpointID, cap)
	return err
}

func (r *Repo) ListEvents(ctx context.Context, endpointID string, limit, offset int, method string) ([]model.EventSummary, int64, error) {
	where := "WHERE endpoint_id=$1"
	args := []any{endpointID}
	if method != "" {
		where += " AND method=$2"
		args = append(args, method)
	}
	var total int64
	if err := r.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM events "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	idx := len(args) + 1
	args = append(args, limit, offset)
	rows, err := r.Pool.Query(ctx, `
		SELECT id, method, path, content_type, body_size, host(source_ip), received_at, processing_duration_ms
		FROM events `+where+` ORDER BY received_at DESC LIMIT $`+itoa(idx)+` OFFSET $`+itoa(idx+1), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []model.EventSummary{}
	for rows.Next() {
		var es model.EventSummary
		var ct *string
		var sip *string
		if err := rows.Scan(&es.ID, &es.Method, &es.Path, &ct, &es.BodySize, &sip, &es.ReceivedAt, &es.DurationMS); err != nil {
			return nil, 0, err
		}
		es.ContentType = ct
		es.SourceIP = sip
		out = append(out, es)
	}
	return out, total, rows.Err()
}

func (r *Repo) GetEvent(ctx context.Context, id string) (model.EventDetail, error) {
	var ed model.EventDetail
	var ct *string
	var sip *string
	var qp, hd []byte
	var body []byte
	var truncated bool
	row := r.Pool.QueryRow(ctx, `
		SELECT id, endpoint_id, method, path, raw_query, query_params, headers, content_type, body, body_size, body_truncated, host(source_ip), received_at, processing_duration_ms
		FROM events WHERE id=$1`, id)
	var rawQuery string
	if err := row.Scan(&ed.ID, &ed.EndpointID, &ed.Method, &ed.Path, &rawQuery, &qp, &hd, &ct, &body, &ed.BodySize, &truncated, &sip, &ed.ReceivedAt, &ed.DurationMS); err != nil {
		return ed, err
	}
	ed.RawQuery = rawQuery
	ed.ContentType = ct
	ed.SourceIP = sip
	ed.BodyTruncated = truncated
	// parse jsonb
	var qm map[string]any
	if len(qp) > 0 {
		json.Unmarshal(qp, &qm)
	}
	if qm == nil {
		qm = map[string]any{}
	}
	ed.QueryParams = qm
	var hm map[string]any
	if len(hd) > 0 {
		json.Unmarshal(hd, &hm)
	}
	if hm == nil {
		hm = map[string]any{}
	}
	ed.Headers = hm
	// base64 body
	ed.BodyBase64 = base64Std(body)
	// body text if valid utf8
	if len(body) > 0 {
		s := string(body)
		// basic validity: check no invalid utf8
		if isValidUTF8(body) {
			ed.BodyText = &s
		}
	}
	return ed, nil
}

func (r *Repo) UpdateEventDuration(ctx context.Context, id string, ms int64) error {
	_, err := r.Pool.Exec(ctx, `UPDATE events SET processing_duration_ms=$2 WHERE id=$1`, id, ms)
	return err
}

func (r *Repo) DeleteEvent(ctx context.Context, id string) error {
	_, err := r.Pool.Exec(ctx, `DELETE FROM events WHERE id=$1`, id)
	return err
}

func (r *Repo) RecentEvents(ctx context.Context, limit int) ([]model.RecentEvent, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT e.id, e.endpoint_id, ep.slug, ep.name, e.method, e.path, e.content_type, e.body_size, host(e.source_ip), e.received_at, e.processing_duration_ms
		FROM events e JOIN endpoints ep ON ep.id = e.endpoint_id
		ORDER BY e.received_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.RecentEvent{}
	for rows.Next() {
		var re model.RecentEvent
		var ct *string
		var sip *string
		var name *string
		if err := rows.Scan(&re.ID, &re.EndpointID, &re.EndpointSlug, &name, &re.Method, &re.Path, &ct, &re.BodySize, &sip, &re.ReceivedAt, &re.DurationMS); err != nil {
			return nil, err
		}
		re.EndpointName = name
		re.ContentType = ct
		re.SourceIP = sip
		out = append(out, re)
	}
	return out, rows.Err()
}

func (r *Repo) InsertReplay(ctx context.Context, eventID, destinationURL, method string, reqHeaders map[string]any, reqBody []byte, status *int, respHeaders map[string]any, respBody []byte, durationMS int64, errMsg *string) (model.Replay, error) {
	rh, _ := json.Marshal(reqHeaders)
	resph, _ := json.Marshal(respHeaders)
	var id string
	err := r.Pool.QueryRow(ctx, `
		INSERT INTO replays (event_id, destination_url, method, request_headers, request_body, response_status, response_headers, response_body, response_duration_ms, error)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`,
		eventID, destinationURL, method, rh, reqBody, status, resph, respBody, durationMS, errMsg).Scan(&id)
	if err != nil {
		return model.Replay{}, err
	}
	return r.GetReplay(ctx, id)
}

func (r *Repo) scanReplay(row pgx.Row) (model.Replay, error) {
	var rep model.Replay
	var rh, resph, reqBody, respBody []byte
	var status *int
	var errMsg *string
	if err := row.Scan(&rep.ID, &rep.EventID, &rep.DestinationURL, &rep.Method, &rh, &reqBody, &status, &resph, &respBody, &rep.ResponseDurationMS, &errMsg, &rep.CreatedAt); err != nil {
		return rep, err
	}
	rep.ResponseStatus = status
	rep.Error = errMsg
	var rm, resm map[string]any
	if len(rh) > 0 {
		json.Unmarshal(rh, &rm)
	}
	if len(resph) > 0 {
		json.Unmarshal(resph, &resm)
	}
	if rm == nil {
		rm = map[string]any{}
	}
	if resm == nil {
		resm = map[string]any{}
	}
	rep.RequestHeaders = rm
	rep.ResponseHeaders = resm
	rep.RequestBodyBase64 = base64Std(reqBody)
	rep.ResponseBodyBase64 = base64Std(respBody)
	if len(respBody) > 0 && isValidUTF8(respBody) {
		s := string(respBody)
		rep.ResponseBodyText = &s
	}
	return rep, nil
}

func (r *Repo) ListReplaysByEvent(ctx context.Context, eventID string) ([]model.Replay, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT id, event_id, destination_url, method, request_headers, request_body, response_status, response_headers, response_body, response_duration_ms, error, created_at
		FROM replays WHERE event_id=$1 ORDER BY created_at DESC`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Replay{}
	for rows.Next() {
		rep, err := r.scanReplay(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rep)
	}
	return out, rows.Err()
}

func (r *Repo) GetReplay(ctx context.Context, id string) (model.Replay, error) {
	return r.scanReplay(r.Pool.QueryRow(ctx, `
		SELECT id, event_id, destination_url, method, request_headers, request_body, response_status, response_headers, response_body, response_duration_ms, error, created_at
		FROM replays WHERE id=$1`, id))
}

func (r *Repo) DeleteEventsByEndpoint(ctx context.Context, endpointID string) error {
	_, err := r.Pool.Exec(ctx, `DELETE FROM events WHERE endpoint_id=$1`, endpointID)
	return err
}

func (r *Repo) PurgeExpired(ctx context.Context, retention time.Duration) (int64, error) {
	cutoff := time.Now().Add(-retention)
	tag, err := r.Pool.Exec(ctx, `
		DELETE FROM endpoints
		WHERE expires_at IS NOT NULL
		  AND expires_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func itoa(n int) string {
	// simple
	if n == 0 {
		return "0"
	}
	b := []byte{}
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}
