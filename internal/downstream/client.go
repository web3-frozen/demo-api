package downstream

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/web3-frozen/demo-api/internal/model"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) (*Client, error) {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return nil, fmt.Errorf("parse service url: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("service url must include scheme and host")
	}
	return &Client{
		baseURL: strings.TrimRight(parsed.String(), "/"),
		client:  &http.Client{Timeout: 5 * time.Second},
	}, nil
}

func (c *Client) Reserve(ctx context.Context, req model.ReservationRequest) (*model.Reservation, error) {
	var response model.Reservation
	if err := c.doJSON(ctx, http.MethodPost, "/reserve", req, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) Quote(ctx context.Context, priority string) (*model.Quote, error) {
	var response model.Quote
	path := "/quote?priority=" + url.QueryEscape(priority)
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, body, target any) error {
	var payload io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		payload = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, payload)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		message, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("request %s returned %d: %s", path, resp.StatusCode, strings.TrimSpace(string(message)))
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
