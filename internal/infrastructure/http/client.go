package http

import (
	"fmt"
	"io"
	nethttp "net/http"
	"time"

	"github.com/bnema/archup/internal/domain/ports"
)

// HTTPClient implements the HTTPClient port using the standard library
type HTTPClient struct {
	client *nethttp.Client
}

// NewHTTPClient creates a new HTTP client with a default timeout
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &nethttp.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewHTTPClientWithTimeout creates a new HTTP client with a custom timeout
func NewHTTPClientWithTimeout(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &nethttp.Client{
			Timeout: timeout,
		},
	}
}

// Get performs an HTTP GET request
func (hc *HTTPClient) Get(url string) (ports.Response, error) {
	resp, err := hc.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET request failed: %w", err)
	}

	// Read body into memory
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if closeErr := resp.Body.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to read response body: %w (close error: %v)", err, closeErr)
		}
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &httpResponse{
		statusCode: resp.StatusCode,
		body:       body,
		rawResp:    resp,
	}, nil
}

// Post performs an HTTP POST request
func (hc *HTTPClient) Post(url string, contentType string, body []byte) (ports.Response, error) {
	resp, err := hc.client.Post(url, contentType, nil)
	if err != nil {
		return nil, fmt.Errorf("POST request failed: %w", err)
	}

	// Read body into memory
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		if closeErr := resp.Body.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to read response body: %w (close error: %v)", err, closeErr)
		}
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &httpResponse{
		statusCode: resp.StatusCode,
		body:       respBody,
		rawResp:    resp,
	}, nil
}

// httpResponse implements the Response port
type httpResponse struct {
	statusCode int
	body       []byte
	rawResp    *nethttp.Response
}

// StatusCode returns the HTTP status code
func (hr *httpResponse) StatusCode() int {
	return hr.statusCode
}

// Body returns the response body
func (hr *httpResponse) Body() []byte {
	return hr.body
}

// Close closes the response
func (hr *httpResponse) Close() error {
	if hr.rawResp != nil {
		return hr.rawResp.Body.Close()
	}
	return nil
}
