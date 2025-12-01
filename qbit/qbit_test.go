package qbit

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	app "qbit-exp/app"
)

var cookie = "SID"

func setupMockApp() {
	app.QBittorrent.Timeout = 10 * time.Millisecond
	app.QBittorrent.Cookie = &cookie
}

func createTlsServer(t *testing.T, discardServerLogs bool, maxTlsVersion uint16, handler http.Handler) (*httptest.Server, *x509.Certificate) {
	t.Helper()

	// Generate ECC private key for CA
	caPrivKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate CA private key: %v", err)
	}

	// Create CA certificate
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		t.Fatalf("Failed to create CA certificate: %v", err)
	}

	caCert, err := x509.ParseCertificate(caCertDER)
	if err != nil {
		t.Fatalf("Failed to parse CA certificate: %v", err)
	}

	// Generate ECC private key for server
	serverPrivKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate server private key: %v", err)
	}

	// Create server certificate
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}

	serverCertDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverPrivKey.PublicKey, caPrivKey)
	if err != nil {
		t.Fatalf("Failed to create server certificate: %v", err)
	}

	// Create TLS config for server
	serverCert := tls.Certificate{
		Certificate: [][]byte{serverCertDER, caCertDER},
		PrivateKey:  serverPrivKey,
	}

	// Create test server with custom TLS config
	server := httptest.NewUnstartedServer(handler)

	server.TLS = &tls.Config{ //nolint:gosec
		Certificates: []tls.Certificate{serverCert},
		MaxVersion:   maxTlsVersion,
	}
	if discardServerLogs {
		server.Config.ErrorLog = log.New(io.Discard, "", 0)
	}

	server.StartTLS()

	return server, caCert
}

func TestApiRequest_Success(t *testing.T) {
	setupMockApp()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL
	url := createUrl("/test")

	body, reAuth, err := apiRequest(url, "GET", nil)
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
	app.QBittorrent.Cookie = &cookie
	url := createUrl("/test")

	_, reAuth, err := apiRequest(url, "GET", nil)
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
	url := createUrl("/test")

	body, reAuth, err := apiRequest(url, http.MethodGet, nil)
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
	url := createUrl("/test")

	queryParams := []QueryParams{
		{"param1", "value1"},
		{"param2", "value2"},
	}

	body, retry, err := apiRequest(url, http.MethodGet, &queryParams)
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
	url := createUrl("/test")

	body, retry, err := apiRequest(url, http.MethodGet, nil)
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

func TestApiRequest_WithRequestAuthorization_Success(t *testing.T) {
	setupMockApp()

	httpBasicAuthUsername := "your-username"
	httpBasicAuthPassword := "your-password"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(httpBasicAuthUsername+":"+httpBasicAuthPassword))
		if r.Header.Get("Authorization") != expectedAuth {
			t.Fatalf("Expected Authorization header %q, got %q", expectedAuth, r.Header.Get("Authorization"))
		}

		// Respond with success
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("basic auth success"))
	}))
	defer server.Close()

	// Set base URL with mock server
	app.QBittorrent.BaseUrl = server.URL
	app.QBittorrent.BasicAuth = &app.BasicAuth{
		Username: httpBasicAuthUsername,
		Password: httpBasicAuthPassword,
	}

	url := createUrl("/test")

	body, retry, err := apiRequest(url, http.MethodGet, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retry {
		t.Fatalf("Expected no retry, got %v", retry)
	}

	if string(body) != "basic auth success" {
		t.Fatalf("Expected body to be 'basic auth success', got %s", body)
	}
}

func TestApiRequest_ServerWithoutAuthRequirement(t *testing.T) {
	setupMockApp()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Auth header should be ignored, server doesn't require authentication
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("no auth needed"))
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL
	httpBasicAuthUsername := "user"
	httpBasicAuthPassword := "pass"
	app.QBittorrent.BasicAuth = &app.BasicAuth{
		Username: httpBasicAuthUsername,
		Password: httpBasicAuthPassword,
	}

	url := createUrl("/test")

	body, retry, err := apiRequest(url, http.MethodGet, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retry {
		t.Fatalf("Expected no retry, got %v", retry)
	}

	if string(body) != "no auth needed" {
		t.Fatalf("Expected body to be 'no auth needed', got %s", body)
	}
}

func TestApiRequest_EmptyCredentials(t *testing.T) {
	setupMockApp()

	httpBasicAuthUsername := ""
	httpBasicAuthPassword := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL
	app.QBittorrent.BasicAuth = &app.BasicAuth{
		Username: httpBasicAuthUsername,
		Password: httpBasicAuthPassword,
	}

	url := createUrl("/test")

	_, retry, err := apiRequest(url, http.MethodGet, nil)
	if err == nil {
		t.Fatalf("Expected error due to empty credentials, but got nil")
	}

	if retry {
		t.Fatalf("Expected no retry when authentication failure, but got retry=%v", retry)
	}
}

func TestApiRequest_InvalidAuthorization(t *testing.T) {
	setupMockApp()

	httpBasicAuthUsername := "wrong-user"
	httpBasicAuthPassword := "wrong-pass"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("your-username:your-password"))
		if r.Header.Get("Authorization") != expectedAuth {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("invalid auth"))

			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("basic auth success"))
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL
	app.QBittorrent.BasicAuth = &app.BasicAuth{
		Username: httpBasicAuthUsername,
		Password: httpBasicAuthPassword,
	}

	url := createUrl("/test")

	_, retry, err := apiRequest(url, http.MethodGet, nil)
	if err == nil {
		t.Fatalf("Expected error due to invalid authentication, but got nil")
	}

	if retry {
		t.Fatalf("Expected no retry when authentication failure, but got retry=%v", retry)
	}
}

func TestCustomCA(t *testing.T) {
	setupMockApp()

	app.QBittorrent.Timeout = 2 * time.Second

	server, caCert := createTlsServer(t, false, 0,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
	defer server.Close()

	caPool, err := x509.SystemCertPool()
	if err != nil {
		t.Fatalf("Failed to get system cert pool: %v", err)
	}

	caPool.AddCert(caCert)

	app.HttpClient = http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{ //nolint:gosec
				RootCAs: caPool,
			},
		},
	}

	app.QBittorrent.BaseUrl = server.URL
	url := createUrl("/test")

	body, retry, err := apiRequest(url, http.MethodGet, nil)
	if err != nil || retry || string(body) != "" {
		t.Fatalf("Request failed! {body: %v, retry: %v}: %v", body, retry, err)
	}
}

func TestSkipCertValidation(t *testing.T) {
	setupMockApp()

	server, _ := createTlsServer(t, false, 0,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
	defer server.Close()

	app.HttpClient = http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec
			},
		},
	}

	app.QBittorrent.BaseUrl = server.URL
	url := createUrl("/test")

	body, retry, err := apiRequest(url, http.MethodGet, nil)
	if err != nil || retry || string(body) != "" {
		t.Fatalf("Request failed! {body: %v, retry: %v}: %v", body, retry, err)
	}
}

func TestMinTlsVersion(t *testing.T) {
	setupMockApp()

	server, _ := createTlsServer(t, true, tls.VersionTLS12,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
	defer server.Close()

	app.HttpClient = http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
			},
		},
	}

	app.QBittorrent.BaseUrl = server.URL
	url := createUrl("/test")

	body, retry, err := apiRequest(url, http.MethodGet, nil)
	if body != nil || retry {
		t.Fatalf("Expected no body and no retry, got {body: %v, retry: %v}", body, retry)
	}

	if !strings.HasSuffix(err.Error(), "tls: protocol version not supported") {
		t.Fatalf("Expected the error to end with `tls: protocol version not supported`, got: %v", err)
	}
}
