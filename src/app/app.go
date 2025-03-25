package app

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"net/http"
	"os"
	"qbit-exp/internal"
	"qbit-exp/logger"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const devVersion string = "dev"

var (
	version     = devVersion
	author      = "martabal"
	projectName = "qbittorrent-exporter"
)

var (
	QBittorrent QBittorrentSettings
	Exporter    ExporterSettings
	HttpClient  http.Client
)

type ExporterSettings struct {
	Port                 int
	LogLevel             string
	ExperimentalFeatures ExperimentalFeatures
	Features             Features
	Path                 string
	BasicAuth            *BasicAuth
}

type BasicAuth struct {
	Username string
	Password string
}

type QBittorrentSettings struct {
	Timeout  time.Duration
	BaseUrl  string
	Cookie   *string
	Username string
	Password string

	// BasicAuth sets the Authorization header for requests to BaseUrl.
	BasicAuth *BasicAuth
}

type ExperimentalFeatures struct {
	EnableLabelWithHash bool
}

type Features struct {
	EnableIncreasedCardinality bool
	EnableHighCardinality      bool
	EnableTracker              bool
	ShowPassword               bool
}

func LoadEnv() {
	envfile := flag.Bool("e", false, "Use .env file")
	flag.Parse()

	envFileMessage := "Using environment variables"

	if _, err := os.Stat(".env"); err == nil && !*envfile {
		if err := godotenv.Load(".env"); err != nil {
			errormessage := "Error loading .env file:" + err.Error()
			panic(errormessage)
		}
		envFileMessage = "Using .env file"
	}
	defaultLogLevelEnv, _ := getEnv(defaultLogLevel)
	loglevel := logger.SetLogLevel(defaultLogLevelEnv)

	fmt.Printf("%s (version %s)\n", projectName, version)
	fmt.Printf("Author: %s\n", author)
	fmt.Printf("Using log level: %s%s%s\n", logger.ColorLogLevel[logger.LogLevels[loglevel]], loglevel, logger.Reset)

	qbitUsername, usingDefaultValue := getEnv(defaultUsername)
	if !usingDefaultValue {
		logger.Log.Info(fmt.Sprintf("username: %s", qbitUsername))
	}
	showPasswordString, _ := getEnv(defaultExporterShowPassword)
	showPassword := envSetToTrue(showPasswordString)
	qbitPassword, usingDefaultValue := getEnv(defaultPassword)
	if !usingDefaultValue {
		password := GetPasswordMasked()
		if showPassword {
			password = qbitPassword
		}
		logger.Log.Info(fmt.Sprintf("password: %s", password))
	}
	baseUrlEnv, usingDefaultValue := getEnv(defaultBaseUrl)
	baseUrl := strings.TrimSuffix(baseUrlEnv, "/")
	if !internal.IsValidURL(baseUrl) {
		panic(fmt.Sprintf("%s is not a valid URL (check %s)", baseUrl, defaultBaseUrl.Key))
	}
	if !usingDefaultValue {
		logger.Log.Info(fmt.Sprintf("qBittorrent URL: %s", baseUrl))
	}

	qbitBasicAuthUsername := getOptionalEnv(defaultQbitBasicAuthUsername)
	qbitBasicAuthPassword := getOptionalEnv(defaultQbitBasicAuthPassword)
	exporterPortEnv, _ := getEnv(defaultPort)
	timeoutDurationEnv, _ := getEnv(defaultTimeout)
	enableTracker, _ := getEnv(defaultDisableTracker)
	enableHighCardinality, _ := getEnv(defaultHighCardinality)
	enableIncreasedCardinality, _ := getEnv(defaultIncreasedCardinality)
	labelWithHash, _ := getEnv(defaultLabelWithHash)
	exporterUrlEnv := getOptionalEnv(defaultExporterURL)
	exporterPath, _ := getEnv(defaultExporterPathEnv)

	basicAuthUsername := getOptionalEnv(defaultBasicAuthUsername)
	basicAuthPassword := getOptionalEnv(defaultBasicAuthPassword)
	certificateAuthorityPath := getOptionalEnv(defaultCertificateAuthorityPath)
	insecureSkipVerify, _ := getEnv(defaultInsecureSkipVerify)
	minTlsVersionStr, _ := getEnv(defaultMinTlsVersion)

	logger.Log.Debug(envFileMessage)

	exporterPort, errExporterPort := strconv.Atoi(exporterPortEnv)
	if errExporterPort != nil {
		panic(fmt.Sprintf("%s must be an integer (check %s)", exporterPortEnv, defaultPort.Key))
	}
	if exporterPort < 0 || exporterPort > 65353 {
		panic(fmt.Sprintf("%d must be > 0 and < 65353 (check %s)", exporterPort, defaultPort.Key))
	}
	if exporterPort != defaultExporterPort {
		logger.Log.Info(fmt.Sprintf("Listening on port %d", exporterPort))
	}

	timeoutDuration, errTimeoutDuration := strconv.Atoi(timeoutDurationEnv)
	if errTimeoutDuration != nil {
		panic(fmt.Sprintf("%s must be an integer (check %s)", timeoutDurationEnv, defaultTimeout.Key))
	}
	if timeoutDuration < 0 {
		panic(fmt.Sprintf("%d must be > 0 (check %s)", timeoutDuration, defaultTimeout.Key))
	}

	exporterUrl := ""
	if version == devVersion {
		exporterUrl = fmt.Sprintf("http://localhost:%d%s", exporterPort, exporterPath)
	}
	if exporterUrlEnv != nil {
		exporterUrl = strings.TrimSuffix(*exporterUrlEnv, "/")
		if !internal.IsValidURL(exporterUrl) {
			panic(fmt.Sprintf("%s is not a valid URL (check %s)", exporterUrl, defaultExporterURL))
		}
	}
	if exporterUrl != "" {
		logger.Log.Info(fmt.Sprintf("qbittorrent-exporter URL: %s", exporterUrl))
	}

	// If a custom CA is provided and INSECURE_SKIP_VERIFY is set, that's kinda sus
	if certificateAuthorityPath != nil && envSetToTrue(insecureSkipVerify) {
		logger.Log.Warn(fmt.Sprintf("You provided a custom CA and disabled certificate validation (check %s and %s)",
			defaultCertificateAuthorityPath, defaultInsecureSkipVerify.Key))
	}

	// If a custom CA is provided or INSECURE_SKIP_VERIFY is set and the exporter URL is not HTTPS, that's kinda sus
	if (certificateAuthorityPath != nil || envSetToTrue(insecureSkipVerify)) && !internal.IsValidHttpsURL(baseUrl) {
		logger.Log.Warn(fmt.Sprintf("You provided a custom CA or disabled certificate validation but the qBittorrent URL is not HTTPS. (check %s, %s and %s)",
			defaultCertificateAuthorityPath, defaultInsecureSkipVerify.Key, defaultBaseUrl.Key))
	}

	// If a custom CA is provided, load the root CAs from the system and append the custom CA
	var caCertPool *x509.CertPool
	if certificateAuthorityPath != nil {
		caCert, errCaCert := os.ReadFile(*certificateAuthorityPath)
		if errCaCert != nil {
			panic(fmt.Sprintf("Error reading certificate authority file: %s (check %s)",
				errCaCert, defaultCertificateAuthorityPath))
		}

		var errCaCertPool error
		caCertPool, errCaCertPool = x509.SystemCertPool()
		if errCaCertPool != nil {
			panic(fmt.Sprintf("Error getting system certificate pool: %s", errCaCertPool))
		}

		if !caCertPool.AppendCertsFromPEM(caCert) {
			panic(fmt.Sprintf("Error adding custom certificate authority to pool: %s", errCaCert))
		}
	}

	var minTlsVersion uint16
	switch minTlsVersionStr {
	case TLS12:
		minTlsVersion = tls.VersionTLS12
	case TLS13:
		minTlsVersion = tls.VersionTLS13
	default:
		panic(fmt.Sprintf("Invalid minimum TLS version: %s (valid options are %s and %s) (check %s)",
			minTlsVersionStr, TLS12, TLS13, defaultMinTlsVersion.Key))
	}

	qbittorrentBasicAuth := getBasicAuth(qbitBasicAuthUsername, qbitBasicAuthPassword, defaultBasicAuthUsername, defaultBasicAuthPassword)
	exporterBasicAuth := getBasicAuth(basicAuthUsername, basicAuthPassword, defaultBasicAuthUsername, defaultBasicAuthPassword)

	if qbittorrentBasicAuth != nil {
		logger.Log.Info("Enabling qBittorrent Basic Auth request header.")
	}

	if exporterBasicAuth != nil {
		logger.Log.Info("Using basic auth to protect the exporter instance")
	} else {
		logger.Log.Trace("Not using basic auth to protect the exporter instance")
	}

	HttpClient = http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            caCertPool,
				InsecureSkipVerify: envSetToTrue(insecureSkipVerify),
				MinVersion:         minTlsVersion,
			},
		},
	}

	internal.EnsureLeadingSlash(&exporterPath)

	QBittorrent = QBittorrentSettings{
		BaseUrl:   baseUrl,
		Username:  qbitUsername,
		Password:  qbitPassword,
		Timeout:   time.Duration(timeoutDuration) * time.Second,
		Cookie:    nil,
		BasicAuth: qbittorrentBasicAuth,
	}

	Exporter = ExporterSettings{
		Features: Features{
			EnableIncreasedCardinality: envSetToTrue(enableIncreasedCardinality),
			EnableHighCardinality:      envSetToTrue(enableHighCardinality),
			EnableTracker:              envSetToTrue(enableTracker),
			ShowPassword:               showPassword,
		},
		ExperimentalFeatures: ExperimentalFeatures{
			EnableLabelWithHash: envSetToTrue(labelWithHash),
		},
		LogLevel:  loglevel,
		Port:      exporterPort,
		Path:      exporterPath,
		BasicAuth: exporterBasicAuth,
	}

	logger.Log.Info(fmt.Sprintf("Features enabled: %s", getFeaturesEnabled()))
}

func getBasicAuth(basicAuthUsername *string, basicAuthPassword *string, defaultBasicAuth string, defaultBasicPassword string) *BasicAuth {
	var basicAuth *BasicAuth
	if basicAuthUsername != nil || basicAuthPassword != nil {
		var username, password string

		if basicAuthUsername != nil {
			username = *basicAuthUsername
		} else {
			logger.Log.Info(fmt.Sprintf("You set a basic auth password but not username (check %s and %s)",
				defaultBasicAuth, defaultBasicAuth))
		}

		if basicAuthPassword != nil {
			password = *basicAuthPassword
		} else {
			logger.Log.Info(fmt.Sprintf("You set a basic auth username but not password (check %s and %s)",
				defaultBasicPassword, defaultBasicPassword))
		}

		basicAuth = &BasicAuth{Username: username, Password: password}
	}
	return basicAuth
}

func envSetToTrue(env string) bool {
	return strings.ToLower(env) == "true"
}

func GetPasswordMasked() string {
	return strings.Repeat("*", len(QBittorrent.Password))
}

func getFeaturesEnabled() string {
	features := ""

	addComma := func() {
		if features != "" {
			features += ", "
		}
	}

	if Exporter.Features.EnableHighCardinality {
		features += "High cardinality"
	}

	if Exporter.Features.EnableIncreasedCardinality {
		addComma()
		features += "Increased cardinality"
	}

	if Exporter.Features.EnableTracker {
		addComma()
		features += "Trackers"
	}

	if Exporter.Features.ShowPassword {
		addComma()
		features += "Show password"
	}

	if Exporter.ExperimentalFeatures.EnableLabelWithHash {
		addComma()
		features += "Label with hash (experimental)"
	}

	features = fmt.Sprintf("[%s]", features)

	return features
}
