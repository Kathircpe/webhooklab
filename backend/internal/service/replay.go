package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"time"

	"webhooklab/internal/model"
)

var blockedCIDRs = func() []*net.IPNet {
	cidrs := []string{
		"0.0.0.0/8",
		"100.64.0.0/10",
		"169.254.0.0/16",
		"192.0.0.0/24",
		"192.0.2.0/24",
		"198.18.0.0/15",
		"198.51.100.0/24",
		"203.0.113.0/24",
		"240.0.0.0/4",
		"255.255.255.255/32",
		"fc00::/7",
		"fe80::/10",
		"ff00::/8",
	}
	out := make([]*net.IPNet, 0, len(cidrs))
	for _, c := range cidrs {
		_, n, _ := net.ParseCIDR(c)
		out = append(out, n)
	}
	return out
}()

func blockedHostname(host string, allowPrivate bool) bool {
	h := strings.ToLower(strings.TrimSuffix(host, "."))
	switch h {
	case "metadata", "metadata.google.internal", "metadata.azure.com", "instance-data":
		return true
	}
	if allowPrivate {
		return false
	}
	return h == "localhost" || strings.HasSuffix(h, ".localhost") || strings.HasSuffix(h, ".internal") || strings.HasSuffix(h, ".local")
}

func ipBlocked(ip net.IP, allowPrivate bool) bool {
	if ip == nil {
		return true
	}
	if ip.IsUnspecified() || ip.IsMulticast() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	for _, n := range blockedCIDRs {
		if n.Contains(ip) {
			return true
		}
	}
	if allowPrivate {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate()
}

func validateDestination(ctx context.Context, raw string, allowPrivate bool) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("destination must use http or https")
	}
	host := u.Hostname()
	if host == "" {
		return nil, errors.New("destination host is required")
	}
	if blockedHostname(host, allowPrivate) {
		return nil, errors.New("destination host is blocked")
	}
	if ip := net.ParseIP(host); ip != nil {
		if ipBlocked(ip, allowPrivate) {
			return nil, errors.New("destination address is blocked")
		}
		return u, nil
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("could not resolve destination: %w", err)
	}
	if len(ips) == 0 {
		return nil, errors.New("destination did not resolve")
	}
	for _, a := range ips {
		if ipBlocked(a.IP, allowPrivate) {
			return nil, errors.New("destination resolves to a blocked address")
		}
	}
	return u, nil
}

func safeClient(timeout time.Duration, allowPrivate bool) *http.Client {
	dialer := &net.Dialer{
		Timeout: timeout,
		Control: func(network, address string, c syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return err
			}
			ip := net.ParseIP(host)
			if ipBlocked(ip, allowPrivate) {
				return errors.New("destination address is blocked")
			}
			return nil
		},
	}
	transport := &http.Transport{
		DialContext:         dialer.DialContext,
		TLSHandshakeTimeout: timeout,
		DisableKeepAlives:   true,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func headerMap(h http.Header) map[string]any {
	out := map[string]any{}
	for k, v := range h {
		out[k] = v
	}
	return out
}

func addHeaders(dst http.Header, src map[string]any) {
	for k, v := range src {
		switch vals := v.(type) {
		case []any:
			for _, item := range vals {
				if s, ok := item.(string); ok {
					dst.Add(k, s)
				}
			}
		case []string:
			for _, s := range vals {
				dst.Add(k, s)
			}
		case string:
			dst.Add(k, vals)
		}
	}
}

var hopHeaders = map[string]bool{
	"Connection":          true,
	"Proxy-Connection":    true,
	"Keep-Alive":          true,
	"Proxy-Authenticate":  true,
	"Proxy-Authorization": true,
	"Te":                  true,
	"Trailer":             true,
	"Transfer-Encoding":   true,
	"Upgrade":             true,
	"Host":                true,
	"Content-Length":      true,
}

func (s *Service) ReplayEvent(ctx context.Context, eventID string, req model.ReplayRequest) (model.Replay, error) {
	ev, err := s.Repo.GetEvent(ctx, eventID)
	if err != nil {
		return model.Replay{}, err
	}
	dest, err := validateDestination(ctx, req.DestinationURL, s.Cfg.ReplayAllowPrivate)
	if err != nil {
		return model.Replay{}, err
	}

	method := strings.ToUpper(strings.TrimSpace(ev.Method))
	if req.Method != nil && strings.TrimSpace(*req.Method) != "" {
		method = strings.ToUpper(strings.TrimSpace(*req.Method))
	}
	switch method {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
	default:
		return model.Replay{}, errors.New("unsupported replay method")
	}

	var body []byte
	if req.BodyBase64 != nil {
		body, err = base64.StdEncoding.DecodeString(*req.BodyBase64)
		if err != nil {
			return model.Replay{}, errors.New("invalid body encoding")
		}
	} else {
		body, err = base64.StdEncoding.DecodeString(ev.BodyBase64)
		if err != nil {
			return model.Replay{}, errors.New("invalid stored body")
		}
	}

	outReq, err := http.NewRequestWithContext(ctx, method, dest.String(), bytes.NewReader(body))
	if err != nil {
		return model.Replay{}, err
	}

	headers := req.Headers
	if headers == nil {
		headers = ev.Headers
	}
	addHeaders(outReq.Header, headers)
	for h := range hopHeaders {
		outReq.Header.Del(h)
	}
	if outReq.Header.Get("User-Agent") == "" {
		outReq.Header.Set("User-Agent", "WebhookLab-Replay/1.0")
	}

	client := safeClient(s.Cfg.ReplayTimeout, s.Cfg.ReplayAllowPrivate)
	started := time.Now()
	resp, sendErr := client.Do(outReq)
	duration := time.Since(started).Milliseconds()

	var status *int
	var respHeaders map[string]any
	var respBody []byte
	var errMsg *string
	if sendErr != nil {
		m := sendErr.Error()
		errMsg = &m
	} else {
		defer resp.Body.Close()
		code := resp.StatusCode
		status = &code
		respHeaders = headerMap(resp.Header)
		limited := &io.LimitedReader{R: resp.Body, N: s.Cfg.ReplayMaxResponse}
		respBody, _ = io.ReadAll(limited)
	}

	rep, recErr := s.Repo.InsertReplay(ctx, eventID, dest.String(), method, headers, body, status, respHeaders, respBody, duration, errMsg)
	if recErr != nil {
		return rep, recErr
	}
	if sendErr != nil {
		return rep, nil
	}
	return rep, nil
}

func (s *Service) ListReplays(ctx context.Context, eventID string) ([]model.Replay, error) {
	return s.Repo.ListReplaysByEvent(ctx, eventID)
}
