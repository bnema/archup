package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPClient_NewHTTPClient(t *testing.T) {
	client := NewHTTPClient()
	if client == nil {
		t.Fatal("expected non-nil client")
	}

	if client.client == nil {
		t.Fatal("expected non-nil internal client")
	}
}

func TestHTTPClient_NewHTTPClientWithTimeout(t *testing.T) {
	timeout := 5 * time.Second
	client := NewHTTPClientWithTimeout(timeout)

	if client == nil {
		t.Fatal("expected non-nil client")
	}

	if client.client.Timeout != timeout {
		t.Errorf("expected timeout %v, got %v", timeout, client.client.Timeout)
	}
}

func TestHTTPClient_Get_Success(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		if _, err := w.Write([]byte("test response")); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := NewHTTPClient()
	resp, err := client.Get(server.URL)
	defer func() {
		if err := resp.Close(); err != nil {
			t.Fatalf("failed to close response: %v", err)
		}
	}()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.StatusCode() != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode())
	}

	if string(resp.Body()) != "test response" {
		t.Errorf("expected 'test response', got %s", string(resp.Body()))
	}
}

func TestHTTPClient_Get_StatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer server.Close()

	client := NewHTTPClient()
	resp, err := client.Get(server.URL)
	defer func() {
		if err := resp.Close(); err != nil {
			t.Fatalf("failed to close response: %v", err)
		}
	}()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.StatusCode() != 404 {
		t.Errorf("expected status 404, got %d", resp.StatusCode())
	}
}

func TestHTTPClient_Get_EmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer server.Close()

	client := NewHTTPClient()
	resp, err := client.Get(server.URL)
	defer func() {
		if err := resp.Close(); err != nil {
			t.Fatalf("failed to close response: %v", err)
		}
	}()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(resp.Body()) != 0 {
		t.Errorf("expected empty body, got %s", string(resp.Body()))
	}
}

func TestHTTPClient_Get_LargeBody(t *testing.T) {
	largeBody := strings.Repeat("a", 10000)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		if _, err := w.Write([]byte(largeBody)); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := NewHTTPClient()
	resp, err := client.Get(server.URL)
	defer func() {
		if err := resp.Close(); err != nil {
			t.Fatalf("failed to close response: %v", err)
		}
	}()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if string(resp.Body()) != largeBody {
		t.Error("expected body to match large response")
	}
}

func TestHTTPClient_Get_InvalidURL(t *testing.T) {
	client := NewHTTPClient()
	_, err := client.Get("http://nonexistent.invalid.domain.example/")

	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestHTTPClient_Post_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		if _, err := w.Write([]byte("created")); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := NewHTTPClient()
	resp, err := client.Post(server.URL, "application/json", []byte(`{"key":"value"}`))
	defer func() {
		if err := resp.Close(); err != nil {
			t.Fatalf("failed to close response: %v", err)
		}
	}()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.StatusCode() != 201 {
		t.Errorf("expected status 201, got %d", resp.StatusCode())
	}

	if string(resp.Body()) != "created" {
		t.Errorf("expected 'created', got %s", string(resp.Body()))
	}
}

func TestHTTPClient_Post_InvalidURL(t *testing.T) {
	client := NewHTTPClient()
	_, err := client.Post("http://nonexistent.invalid.domain.example/", "application/json", []byte("test"))

	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestHTTPClient_Response_Close(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		if _, err := w.Write([]byte("test")); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := NewHTTPClient()
	resp, _ := client.Get(server.URL)

	err := resp.Close()
	if err != nil {
		t.Errorf("expected no error closing response, got %v", err)
	}
}

func TestHTTPClient_Get_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(200)
	}))
	defer server.Close()

	client := NewHTTPClientWithTimeout(100 * time.Millisecond)
	_, err := client.Get(server.URL)

	if err == nil {
		t.Fatal("expected error due to timeout")
	}
}
