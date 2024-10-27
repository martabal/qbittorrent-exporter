package qbit

import (
	"bytes"
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

const twoSeconds = time.Duration(2 * time.Second)

func init() {
	logger.Log = &logger.Logger{Logger: slog.New(slog.NewTextHandler(buff, &slog.HandlerOptions{}))}
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
	app.QBittorrent.Timeout = twoSeconds

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
	app.QBittorrent.Timeout = twoSeconds

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for authentication failure, got none")
		}
	}()

	Auth()
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
	app.QBittorrent.Timeout = twoSeconds

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for invalid URL")
		}
	}()

	Auth()
}

func TestAuthTimeout(t *testing.T) {
	t.Cleanup(resetState)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
	}))
	defer ts.Close()

	app.QBittorrent.BaseUrl = ts.URL
	app.QBittorrent.Username = ""
	app.QBittorrent.Password = ""
	app.QBittorrent.Timeout = twoSeconds
	Auth()

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
	app.QBittorrent.Timeout = twoSeconds
	Auth()

	if !strings.Contains(buff.String(), strconv.Itoa(http.StatusCreated)) {
		t.Errorf("expected %d, got: %s", http.StatusCreated, buff.String())
	}
}

func resetState() {
	app.QBittorrent.Cookie = ""
	app.ShouldShowError = true
	buff.Reset()
}
