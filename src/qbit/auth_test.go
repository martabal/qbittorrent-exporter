package qbit

import (
	"bytes"
	"encoding/base64"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	API "qbit-exp/api"
	"qbit-exp/app"
	"qbit-exp/logger"
)

var buff = &bytes.Buffer{}

const defaultTimeout time.Duration = 10 * time.Millisecond

const defaultUsername = "testuser"
const defaultPassword = "testpass"

func init() {
	logger.Log = &logger.Logger{Logger: slog.New(slog.NewTextHandler(buff, &slog.HandlerOptions{}))}
}

func TestAuthSuccess(t *testing.T) {
	password := "abc123"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		w.Header().Set("Set-Cookie", "SID=abc123; Path=/")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte("Success"))
		if err != nil {
			panic("Error with the response " + err.Error())
		}
	}))
	defer ts.Close()

	app.QBittorrent.BaseUrl = ts.URL
	app.QBittorrent.Username = defaultUsername
	app.QBittorrent.Password = defaultPassword
	app.QBittorrent.Timeout = defaultTimeout

	err := Auth()
	if err != nil {
		t.Errorf("There was an error: %s", err.Error())
	}

	if *app.QBittorrent.Cookie != password {
		t.Errorf("expected cookie value to be 'abc123', got '%s'", *app.QBittorrent.Cookie)
	}
}

func TestAuthFail(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte("Fails."))
		if err != nil {
			panic("Error with the response " + err.Error())
		}
	}))
	defer ts.Close()

	app.QBittorrent.BaseUrl = ts.URL
	app.QBittorrent.Username = "wronguser"
	app.QBittorrent.Password = "wrongpass"
	app.QBittorrent.Timeout = defaultTimeout

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for authentication failure, got none")
		}
	}()

	err := Auth()
	if err == nil {
		t.Errorf("There wasn't an error")
	}
}

func TestAuthInvalidUrl(t *testing.T) {
	t.Cleanup(resetState)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	app.QBittorrent.BaseUrl = ts.URL + "//"
	app.QBittorrent.Username = ""
	app.QBittorrent.Password = ""
	app.QBittorrent.Timeout = defaultTimeout

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for invalid URL")
		}
	}()

	_ = Auth()
}

func TestAuthTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(defaultTimeout * 5)
	}))
	defer ts.Close()

	app.QBittorrent.BaseUrl = ts.URL
	app.QBittorrent.Username = ""
	app.QBittorrent.Password = ""
	app.QBittorrent.Timeout = defaultTimeout
	_ = Auth()

	if !strings.Contains(buff.String(), API.QbittorrentTimeOut) {
		t.Errorf("expected timeout log, got: %s", buff.String())
	}
}

func TestUnknownStatusCode(t *testing.T) {
	t.Cleanup(resetState)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	app.QBittorrent.BaseUrl = ts.URL
	app.QBittorrent.Username = ""
	app.QBittorrent.Password = ""
	app.QBittorrent.Timeout = defaultTimeout
	_ = Auth()

	if !strings.Contains(buff.String(), strconv.Itoa(http.StatusCreated)) {
		t.Errorf("expected %d, got: %s", http.StatusCreated, buff.String())
	}
}

func TestAuth_BasicAuthSuccess(t *testing.T) {
	t.Cleanup(resetState)

	httpBasicAuthUsername := "your-username"
	httpBasicAuthPassword := "your-password"
	password := "abc123"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(httpBasicAuthUsername+":"+httpBasicAuthPassword))
		if r.Header.Get("Authorization") != expectedAuth {
			t.Fatalf("Expected Authorization header %q, got %q", expectedAuth, r.Header.Get("Authorization"))
		}

		w.Header().Set("Set-Cookie", "SID=abc123; Path=/")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte("Success"))
		if err != nil {
			panic("Error with the response " + err.Error())
		}
	}))
	defer ts.Close()

	app.QBittorrent.BaseUrl = ts.URL
	app.QBittorrent.Username = defaultUsername
	app.QBittorrent.Password = defaultPassword
	app.QBittorrent.Timeout = defaultTimeout
	app.QBittorrent.BasicAuth = &app.BasicAuth{
		Username: httpBasicAuthUsername,
		Password: httpBasicAuthPassword,
	}

	err := Auth()
	if err != nil {
		t.Errorf("There was an error: %s", err.Error())
	}

	if *app.QBittorrent.Cookie != password {
		t.Errorf("expected cookie value to be 'abc123', got '%s'", *app.QBittorrent.Cookie)
	}
}

func TestAuth_BasicAuthInvalidAuthentication(t *testing.T) {
	t.Cleanup(resetState)

	httpBasicAuthUsername := "wrong-username"
	httpBasicAuthPassword := "wrong-password"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("your-username:your-password"))
		if r.Header.Get("Authorization") != expectedAuth {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("invalid auth"))

			return
		}

		w.Header().Set("Set-Cookie", "SID=abc123; Path=/")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte("Success"))
		if err != nil {
			panic("Error with the response " + err.Error())
		}
	}))
	defer ts.Close()

	app.QBittorrent.BaseUrl = ts.URL
	app.QBittorrent.Username = defaultUsername
	app.QBittorrent.Password = defaultPassword
	app.QBittorrent.Timeout = defaultTimeout
	app.QBittorrent.BasicAuth = &app.BasicAuth{
		Username: httpBasicAuthUsername,
		Password: httpBasicAuthPassword,
	}

	err := Auth()
	if err == nil {
		t.Fatalf("Expected error due to invalid authentication, but got nil")
	}
	// Use string matching to check the expected 401, until we update Auth() to return status-code or body.
	if err.Error() != "authentication failed, status code: 401" {
		t.Fatalf("Expected error to be 'authentication failed, status code: 401', but got %s", err)
	}
}

func TestAuthStatusNoContent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		w.Header().Set("Set-Cookie", "SID=xyz789; Path=/")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	app.QBittorrent.BaseUrl = ts.URL
	app.QBittorrent.Username = "user"
	app.QBittorrent.Password = "pass"
	app.QBittorrent.Timeout = defaultTimeout

	err := Auth()
	if err != nil {
		t.Fatalf("unexpected error for 204 status: %v", err)
	}

	if app.QBittorrent.Cookie == nil {
		t.Fatalf("expected cookie to be set for 204, got nil")
	}

	if *app.QBittorrent.Cookie != "xyz789" {
		t.Fatalf("expected cookie 'xyz789', got '%s'", *app.QBittorrent.Cookie)
	}

	if !strings.Contains(buff.String(), "New cookie for auth stored") {
		t.Fatalf("expected log entry for stored cookie, got: %s", buff.String())
	}
}

func resetState() {
	app.QBittorrent.Cookie = nil

	buff.Reset()
}
