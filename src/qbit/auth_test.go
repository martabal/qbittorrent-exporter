package qbit

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"qbit-exp/app"
	"qbit-exp/logger"
)

var mockLogger *logger.Logger

func init() {
	handler := logger.NewPrettyHandler(os.Stdout, slog.HandlerOptions{})
	mockLogger = &logger.Logger{Logger: slog.New(handler)}

	logger.Log = mockLogger
}

func TestAuthSuccess(t *testing.T) {
	t.Cleanup(resetState)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		w.Header().Set("Set-Cookie", "SID=abc123; Path=/")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Success"))
		if err != nil {
			panic("Error with the response" + err.Error())
		}
	}))
	defer ts.Close()

	app.QBittorrent.BaseUrl = ts.URL
	app.QBittorrent.Username = "testuser"
	app.QBittorrent.Password = "testpass"
	app.QBittorrent.Timeout = 2

	Auth()

	if app.QBittorrent.Cookie != "abc123" {
		t.Errorf("expected cookie value to be 'abc123', got '%s'", app.QBittorrent.Cookie)
	}
}

func TestAuthFail(t *testing.T) {

	t.Cleanup(resetState)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Fails."))
		if err != nil {
			panic("Error with the response" + err.Error())
		}
	}))
	defer ts.Close()

	app.QBittorrent.BaseUrl = ts.URL
	app.QBittorrent.Username = "wronguser"
	app.QBittorrent.Password = "wrongpass"
	app.QBittorrent.Timeout = 2

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for authentication failure, got none")
		}
	}()

	Auth()
}

func resetState() {
	app.QBittorrent.Cookie = ""
}
