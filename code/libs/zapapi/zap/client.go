package zap

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	DefaultBase           = "http://zap/JSON/"
	DefaultBaseOther      = "http://zap/OTHER/"
	DefaultHTTPSBase      = "https://zap/JSON/"
	DefaultHTTPSBaseOther = "https://zap/OTHER/"
	DefaultProxy          = "tcp://127.0.0.1:8080"
	ZapAPIKeyParam        = "apikey"
	ZapAPIKeyHeader       = "X-ZAP-API-Key"
)

type Config struct {
	Base      string
	BaseOther string
	Proxy     string
	APIKey    string
	TLSConfig tls.Config
}

type Client struct {
	*Config
	httpClient *http.Client
}

// NewClient returns a new ZAP client based on the passed in config
// func NewClient(cfg *Config) (Interface, error) {
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if cfg.Base == "" {
		cfg.Base = DefaultBase
	}

	if cfg.BaseOther == "" {
		cfg.BaseOther = DefaultBaseOther
	}

	if cfg.Proxy == "" {
		cfg.Proxy = DefaultProxy
	}

	proxyURL, err := url.Parse(cfg.Proxy)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyURL(proxyURL),
			TLSClientConfig: &cfg.TLSConfig,
		},
	}

	return &Client{
		Config:     cfg,
		httpClient: httpClient,
	}, nil
}

// Requesting to the API
func (c *Client) Request(ctx context.Context, path string, queryParams map[string]string) ([]byte, error) {
	return c.request(ctx, c.Base+path, queryParams)
}

// RequestOther sends HTTP request to zap other("http://zap/OTHER/") API group
func (c *Client) RequestOther(ctx context.Context, path string, queryParams map[string]string) ([]byte, error) {
	return c.request(ctx, c.BaseOther+path, queryParams)
}

func (c *Client) request(ctx context.Context, path string, queryParams map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if len(queryParams) == 0 {
		queryParams = map[string]string{}
	}

	// Send the API key even if there are no parameters,
	// older ZAP versions might need API key as (query) parameter.
	queryParams[ZapAPIKeyParam] = c.APIKey

	query := req.URL.Query()
	for k, v := range queryParams {
		if v == "" {
			continue
		}
		query.Add(k, v)
	}
	req.URL.RawQuery = query.Encode()

	req.Header.Add("Accept", "application/json")
	req.Header.Add(ZapAPIKeyHeader, c.APIKey)

	// Close the connection
	req.Close = true
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("errored when sending request to the server: %v", err)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
