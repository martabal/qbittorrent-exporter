package qbit

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	app "qbit-exp/app"
	"testing"
	"time"
)

func setupMockApp() {
	app.QBittorrent.Timeout = 10 * time.Millisecond
	app.QBittorrent.Cookie = "SID"
	app.ShouldShowError = true
}

func TestApiRequest_Success(t *testing.T) {
	setupMockApp()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL

	body, reAuth, err := apiRequest("/test", "GET", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if reAuth {
		t.Fatalf("Expected reAuth to be false, got %v", reAuth)
	}
	if string(body) != "success" {
		t.Fatalf("Expected body to be 'success', got %s", body)
	}
}

func TestApiRequest_Forbidden(t *testing.T) {
	setupMockApp()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL

	_, reAuth, err := apiRequest("/test", "GET", nil)
	if err == nil || err.Error() != "403" {
		t.Fatalf("Expected error '403', got %v", err)
	}
	if !reAuth {
		t.Fatalf("Expected reAuth to be true, got %v", reAuth)
	}
}

func TestApiRequest_Timeout(t *testing.T) {
	setupMockApp()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL

	body, reAuth, err := apiRequest("/test", http.MethodGet, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Expected DeadlineExceeded error, got %v", err)
	}
	if body != nil {
		t.Fatalf("Expected no body, got %v", body)
	}
	if reAuth {
		t.Fatalf("Expected reAuth to be false, got %v", reAuth)
	}
}

func TestApiRequest_WithQueryParams(t *testing.T) {
	setupMockApp()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "param1=value1&param2=value2" {
			t.Fatalf("Expected query params 'param1=value1&param2=value2', got %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("query success"))
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL

	queryParams := []QueryParams{
		{"param1", "value1"},
		{"param2", "value2"},
	}

	body, retry, err := apiRequest("/test", http.MethodGet, &queryParams)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if retry {
		t.Fatalf("Expected no retry, got %v", retry)
	}
	if string(body) != "query success" {
		t.Fatalf("Expected body to be 'query success', got %s", body)
	}
}

func TestApiRequest_Non200Status(t *testing.T) {
	setupMockApp()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL

	body, retry, err := apiRequest("/test", http.MethodGet, nil)
	if err == nil || err.Error() != "500" {
		t.Fatalf("Expected error '500', got %v", err)
	}
	if body != nil {
		t.Fatalf("Expected no body, got %v", body)
	}
	if retry {
		t.Fatalf("Expected reAuth to be false, got %v", retry)
	}
}
