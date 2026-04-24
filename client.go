// Package qrpro is the official Go SDK for Abundera QR Pro.
//
// See https://pro.qr.abundera.ai/docs/ for the full API reference.
package qrpro

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	defaultBaseURL    = "https://pro.qr.abundera.ai"
	defaultTimeout    = 30 * time.Second
	defaultMaxRetries = 3
	sdkVersion        = "0.1.0"
)

var retryStatus = map[int]bool{429: true, 500: true, 502: true, 503: true, 504: true}

// Client is the Abundera QR Pro API client.
type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	MaxRetries int
	UserAgent  string
}

// Option configures a Client.
type Option func(*Client)

func WithBaseURL(u string) Option          { return func(c *Client) { c.BaseURL = u } }
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.HTTPClient = h } }
func WithMaxRetries(n int) Option          { return func(c *Client) { c.MaxRetries = n } }
func WithUserAgent(ua string) Option       { return func(c *Client) { c.UserAgent = ua } }

// New returns a Client authenticated with apiKey.
func New(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("apiKey is required")
	}
	c := &Client{
		APIKey:     apiKey,
		BaseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: defaultTimeout},
		MaxRetries: defaultMaxRetries,
		UserAgent:  "abundera-qr-pro-go/" + sdkVersion,
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

// Error is returned for any non-2xx API response.
type Error struct {
	Status    int    `json:"status"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%d %s] %s", e.Status, e.Code, e.Message)
}

func (c *Client) request(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	u := c.BaseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	var bodyBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyBytes = b
	}

	var lastErr error
	for attempt := 0; attempt <= c.MaxRetries; attempt++ {
		var reqBody io.Reader
		if bodyBytes != nil {
			reqBody = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, u, reqBody)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
		req.Header.Set("User-Agent", c.UserAgent)
		req.Header.Set("Accept", "application/json")
		if bodyBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt >= c.MaxRetries {
				return err
			}
			time.Sleep(backoff(attempt))
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			defer resp.Body.Close()
			if out == nil {
				io.Copy(io.Discard, resp.Body)
				return nil
			}
			if s, ok := out.(*string); ok {
				b, err := io.ReadAll(resp.Body)
				if err != nil {
					return err
				}
				*s = string(b)
				return nil
			}
			return json.NewDecoder(resp.Body).Decode(out)
		}

		if retryStatus[resp.StatusCode] && attempt < c.MaxRetries {
			retryAfter, _ := strconv.ParseFloat(resp.Header.Get("Retry-After"), 64)
			resp.Body.Close()
			if retryAfter > 0 {
				time.Sleep(time.Duration(retryAfter * float64(time.Second)))
			} else {
				time.Sleep(backoff(attempt))
			}
			continue
		}

		apiErr := &Error{Status: resp.StatusCode, Code: "http_error", Message: "HTTP " + resp.Status, RequestID: resp.Header.Get("X-Request-Id")}
		if b, err := io.ReadAll(resp.Body); err == nil {
			var parsed struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}
			if json.Unmarshal(b, &parsed) == nil {
				if parsed.Code != "" {
					apiErr.Code = parsed.Code
				}
				if parsed.Message != "" {
					apiErr.Message = parsed.Message
				}
			}
		}
		resp.Body.Close()
		return apiErr
	}
	return lastErr
}

func backoff(attempt int) time.Duration {
	return time.Duration(250*math.Pow(2, float64(attempt))) * time.Millisecond
}

// --- Codes -------------------------------------------------------------

type ListCodesParams struct {
	Cursor  string
	Limit   int
	Tag     string
	GroupID string
}

func (p *ListCodesParams) values() url.Values {
	v := url.Values{}
	if p == nil {
		return v
	}
	if p.Cursor != "" {
		v.Set("cursor", p.Cursor)
	}
	if p.Limit > 0 {
		v.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Tag != "" {
		v.Set("tag", p.Tag)
	}
	if p.GroupID != "" {
		v.Set("group_id", p.GroupID)
	}
	return v
}

func (c *Client) ListCodes(ctx context.Context, params *ListCodesParams) (*ListResult[Code], error) {
	out := &ListResult[Code]{}
	return out, c.request(ctx, "GET", "/api/codes", params.values(), nil, out)
}

func (c *Client) GetCode(ctx context.Context, id string) (*Code, error) {
	out := &Code{}
	return out, c.request(ctx, "GET", "/api/codes/"+url.PathEscape(id), nil, nil, out)
}

func (c *Client) CreateCode(ctx context.Context, in CodeCreate) (*Code, error) {
	out := &Code{}
	return out, c.request(ctx, "POST", "/api/codes", nil, in, out)
}

func (c *Client) UpdateCode(ctx context.Context, id string, patch CodePatch) (*Code, error) {
	out := &Code{}
	return out, c.request(ctx, "PATCH", "/api/codes/"+url.PathEscape(id), nil, patch, out)
}

func (c *Client) DeleteCode(ctx context.Context, id string) error {
	return c.request(ctx, "DELETE", "/api/codes/"+url.PathEscape(id), nil, nil, nil)
}

func (c *Client) CheckSlug(ctx context.Context, slug string) (bool, error) {
	v := url.Values{"slug": {slug}}
	var out struct {
		Available bool `json:"available"`
	}
	if err := c.request(ctx, "GET", "/api/codes/check-slug", v, nil, &out); err != nil {
		return false, err
	}
	return out.Available, nil
}

// --- Analytics ---------------------------------------------------------

type AnalyticsParams struct{ From, To string }

func (c *Client) GetAnalytics(ctx context.Context, codeID string, p *AnalyticsParams) (*Analytics, error) {
	v := url.Values{}
	if p != nil {
		if p.From != "" {
			v.Set("from", p.From)
		}
		if p.To != "" {
			v.Set("to", p.To)
		}
	}
	out := &Analytics{}
	return out, c.request(ctx, "GET", "/api/codes/"+url.PathEscape(codeID)+"/analytics", v, nil, out)
}

func (c *Client) GetAnalyticsCSV(ctx context.Context, codeID string, p *AnalyticsParams) (string, error) {
	v := url.Values{}
	if p != nil {
		if p.From != "" {
			v.Set("from", p.From)
		}
		if p.To != "" {
			v.Set("to", p.To)
		}
	}
	var out string
	return out, c.request(ctx, "GET", "/api/codes/"+url.PathEscape(codeID)+"/analytics.csv", v, nil, &out)
}

// --- Groups ------------------------------------------------------------

func (c *Client) ListGroups(ctx context.Context) (*ListResult[Group], error) {
	out := &ListResult[Group]{}
	return out, c.request(ctx, "GET", "/api/groups", nil, nil, out)
}

func (c *Client) CreateGroup(ctx context.Context, name, description string) (*Group, error) {
	body := map[string]string{"name": name}
	if description != "" {
		body["description"] = description
	}
	out := &Group{}
	return out, c.request(ctx, "POST", "/api/groups", nil, body, out)
}

func (c *Client) DeleteGroup(ctx context.Context, id string) error {
	return c.request(ctx, "DELETE", "/api/groups/"+url.PathEscape(id), nil, nil, nil)
}

// --- Webhooks ----------------------------------------------------------

func (c *Client) ListWebhooks(ctx context.Context) (*ListResult[Webhook], error) {
	out := &ListResult[Webhook]{}
	return out, c.request(ctx, "GET", "/api/webhooks", nil, nil, out)
}

// CreateWebhook returns a Webhook with Secret populated — store it immediately.
func (c *Client) CreateWebhook(ctx context.Context, webhookURL string, events []string) (*Webhook, error) {
	body := map[string]any{"url": webhookURL, "events": events}
	out := &Webhook{}
	return out, c.request(ctx, "POST", "/api/webhooks", nil, body, out)
}

func (c *Client) DeleteWebhook(ctx context.Context, id string) error {
	return c.request(ctx, "DELETE", "/api/webhooks/"+url.PathEscape(id), nil, nil, nil)
}

// --- User --------------------------------------------------------------

func (c *Client) Me(ctx context.Context) (map[string]any, error) {
	out := map[string]any{}
	return out, c.request(ctx, "GET", "/api/user/me", nil, nil, &out)
}
