package jwtrevokeapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ClientOption func(*Client)

type Client struct {
	apiKey            string
	baseURL           string
	client            *http.Client
	maxRetries        int
	rateLimitDelay    time.Duration
	requestTimeout    time.Duration
}

type ClientError struct {
	StatusCode int
	Message    string
	Data       interface{}
}

func (e *ClientError) Error() string {
	return fmt.Sprintf("jwt-revoke error: %s (status: %d)", e.Message, e.StatusCode)
}

func WithMaxRetries(retries int) ClientOption {
	return func(c *Client) {
		c.maxRetries = retries
	}
}

func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.requestTimeout = timeout
	}
}

func WithRateLimitDelay(delay time.Duration) ClientOption {
	return func(c *Client) {
		c.rateLimitDelay = delay
	}
}

func NewClient(apiKey string, options ...ClientOption) *Client {
	c := &Client{
		apiKey:         apiKey,
		baseURL:        "https://api.jwtrevoke.com",
		maxRetries:     3,
		rateLimitDelay: time.Second,
		requestTimeout: 10 * time.Second,
		client:        &http.Client{},
	}

	for _, option := range options {
		option(c)
	}

	c.client.Timeout = c.requestTimeout
	return c
}

func (c *Client) doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		resp, err = c.client.Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			time.Sleep(c.rateLimitDelay)
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		if resp.StatusCode >= 500 {
			continue
		}

		// Client error, don't retry
		var errorResponse struct {
			Message string      `json:"message"`
			Data    interface{} `json:"data"`
		}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		return nil, &ClientError{
			StatusCode: resp.StatusCode,
			Message:    errorResponse.Message,
			Data:      errorResponse.Data,
		}
	}

	return resp, err
}

type RevokedToken struct {
	ID            string    `json:"id"`
	JwtID         string    `json:"jwt_id"`
	Reason        string    `json:"reason"`
	ExpiryDate    time.Time `json:"expiry_date"`
	RevokedByEmail string   `json:"revoked_by_email,omitempty"`
}

type RevokeRequest struct {
	JwtID      string    `json:"jwtId"`
	Reason     string    `json:"reason"`
	ExpiryDate time.Time `json:"expiryDate"`
}

func (c *Client) ListRevokedTokens() ([]RevokedToken, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/revocations/list", c.baseURL), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", c.apiKey)
	
	resp, err := c.doRequest(context.Background(), req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []RevokedToken `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (c *Client) RevokeToken(jwtID string, reason string, expiryDate time.Time) (*RevokedToken, error) {
	payload := RevokeRequest{
		JwtID:      jwtID,
		Reason:     reason,
		ExpiryDate: expiryDate,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/revocations/revoke", c.baseURL), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(context.Background(), req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Token RevokedToken `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Token, nil
}

func (c *Client) DeleteRevokedToken(jwtID string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/revocations/%s", c.baseURL, jwtID), nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.doRequest(context.Background(), req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
} 