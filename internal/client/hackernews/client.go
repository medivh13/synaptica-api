package hackernews

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client interface {
	FetchStoryIDs(ctx context.Context, feed string) ([]int64, error)
	FetchItem(ctx context.Context, id int64) (*Item, error)
}

type client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient(baseURL string, timeout time.Duration) Client {
	return &client{
		httpClient: &http.Client{Timeout: timeout},
		baseURL:    strings.TrimRight(baseURL, "/"),
	}
}

func (c *client) FetchStoryIDs(ctx context.Context, feed string) ([]int64, error) {
	endpoint, err := c.buildURL("v0", feed+".json")
	if err != nil {
		return nil, err
	}

	var ids []int64
	if err := c.getJSON(ctx, endpoint, &ids); err != nil {
		return nil, fmt.Errorf("fetch hacker news story ids: %w", err)
	}
	if ids == nil {
		return nil, fmt.Errorf("hacker news story ids response was empty")
	}

	return ids, nil
}

func (c *client) FetchItem(ctx context.Context, id int64) (*Item, error) {
	endpoint, err := c.buildURL("v0", "item", fmt.Sprintf("%d.json", id))
	if err != nil {
		return nil, err
	}

	var item *Item
	if err := c.getJSON(ctx, endpoint, &item); err != nil {
		return nil, fmt.Errorf("fetch hacker news item %d: %w", id, err)
	}

	return item, nil
}

func (c *client) buildURL(path ...string) (string, error) {
	endpoint, err := url.JoinPath(c.baseURL, path...)
	if err != nil {
		return "", fmt.Errorf("build hacker news request URL: %w", err)
	}
	return endpoint, nil
}

func (c *client) getJSON(ctx context.Context, endpoint string, target any) error {
	backoffs := []time.Duration{0, 500 * time.Millisecond, time.Second}
	var lastErr error

	for attempt, backoff := range backoffs {
		if backoff > 0 {
			timer := time.NewTimer(backoff)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return fmt.Errorf("create hacker news request: %w", err)
		}
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt < len(backoffs)-1 && isRetryableError(err) {
				continue
			}
			return fmt.Errorf("send hacker news request: %w", err)
		}

		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			lastErr = fmt.Errorf("hacker news returned non-2xx status: %d", resp.StatusCode)
			resp.Body.Close()
			if attempt < len(backoffs)-1 && isRetryableStatus(resp.StatusCode) {
				continue
			}
			return lastErr
		}

		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return fmt.Errorf("decode hacker news response: %w", err)
		}
		return nil
	}

	return lastErr
}

func isRetryableStatus(status int) bool {
	switch status {
	case http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func isRetryableError(err error) bool {
	var netErr net.Error
	return errors.Is(err, context.DeadlineExceeded) || errors.As(err, &netErr) && netErr.Timeout()
}
