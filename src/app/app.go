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
	"sync"
	"time"

	"github.com/joho/godotenv"
)

var loadEnvOnce sync.Once

var (
	QBittorrent  QBittorrentSettings
	Exporter     ExporterSettings
	HttpClient   http.Client
	UsingEnvFile bool
)

type ExporterSettings struct {
	Port                 int
	LogLevel             string
	ExperimentalFeatures ExperimentalFeatures
	Features             Features
	URL                  string
	Path                 string
	BasicAuth            BasicAuth
}

type BasicAuth struct {
	Username *string
	Password *string
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
	EnableHighCardinality bool
	EnableTracker         bool
	ShowPassword          bool
	// SendBasicAuthRequestHeader sets the Authorization basic request header to qBittorrent endpoint.
	EnableBasicAuthRequestHeader bool
}

func LoadEnv() {
	loadEnvOnce.Do(func() {
		var envfile bool
		flag.BoolVar(&envfile, "e", false, "Use .env file")
		flag.Parse()
		_, err := os.Stat(".env")
		UsingEnvFile = false
		if !os.IsNotExist(err) && !envfile {
			UsingEnvFile = true
			err := godotenv.Load(".env")
			if err != nil {
				errormessage := "Error loading .env file:" + err.Error()
				panic(errormessage)
			}
		}
	})

	loglevel := logger.SetLogLevel(getEnv(defaultLogLevel))
	qbitUsername := getEnv(defaultUsername)
	qbitPassword := getEnv(defaultPassword)
	baseUrl := strings.TrimSuffix(getEnv(defaultBaseUrl), "/")
	enableQbittorrentBasicAuth := getEnv(defaultEnableQbittorrentBasicAuth)
	qbitBasicAuthUsername := getEnv(defaultQbitBasicAuthUsername)
	qbitBasicAuthPassword := getEnv(defaultQbitBasicAuthPassword)
	exporterPortEnv := getEnv(defaultPort)
	timeoutDurationEnv := getEnv(defaultTimeout)
	enableTracker := getEnv(defaultDisableTracker)
	enableHighCardinality := getEnv(defaultHighCardinality)
	labelWithHash := getEnv(defaultLabelWithHash)
	exporterUrl := getEnv(defaultExporterURL)
	exporterPath := getEnv(defaultExporterPath)
	showPassword := getEnv(defaultExporterShowPassword)
	basicAuthUsername := getEnv(defaultBasicAuthUsername)
	basicAuthPassword := getEnv(defaultBasicAuthPassword)
	certificateAuthorityPath := getEnv(defaultCertificateAuthorityPath)
	insecureSkipVerify := getEnv(defaultInsecureSkipVerify)
	minTlsVersionStr := getEnv(defaultMinTlsVersion)

	exporterPort, errExporterPort := strconv.Atoi(exporterPortEnv)
	if errExporterPort != nil {
		panic(fmt.Sprintf("%s must be an integer (check %s)", exporterPortEnv, defaultPort.Key))
	}
	if exporterPort < 0 || exporterPort > 65353 {
		panic(fmt.Sprintf("%d must be > 0 and < 65353 (check %s)", exporterPort, defaultPort.Key))
	}

	timeoutDuration, errTimeoutDuration := strconv.Atoi(timeoutDurationEnv)
	if errTimeoutDuration != nil {
		panic(fmt.Sprintf("%s must be an integer (check %s)", timeoutDurationEnv, defaultTimeout.Key))
	}
	if timeoutDuration < 0 {
		panic(fmt.Sprintf("%d must be > 0 (check %s)", timeoutDuration, defaultTimeout.Key))
	}

	if exporterUrl != "" {
		exporterUrl = strings.TrimSuffix(exporterUrl, "/")
		if !internal.IsValidURL(exporterUrl) {
			panic(fmt.Sprintf("%s is not a valid URL (check %s)", exporterUrl, defaultExporterURL.Key))
		}
	}

	basicAuth := BasicAuth{Username: nil, Password: nil}
	if basicAuthUsername != "" && basicAuthPassword == "" {
		logger.Log.Warn(fmt.Sprintf("You set a basic auth username but not password (check %s and %s)",
			defaultBasicAuthUsername.Key, defaultBasicAuthPassword.Key))
	} else if basicAuthUsername == "" && basicAuthPassword != "" {
		logger.Log.Warn(fmt.Sprintf("You set a basic auth password but not username (check %s and %s)",
			defaultBasicAuthUsername.Key, defaultBasicAuthPassword.Key))
	} else if basicAuthUsername != "" && basicAuthPassword != "" {
		basicAuth = BasicAuth{
			Username: &basicAuthUsername,
			Password: &basicAuthPassword,
		}
	}

	if !internal.IsValidURL(baseUrl) {
		panic(fmt.Sprintf("%s is not a valid URL (check %s)", baseUrl, defaultBaseUrl.Key))
	}

	// If a custom CA is provided and INSECURE_SKIP_VERIFY is set, that's kinda sus
	if certificateAuthorityPath != "" && envSetToTrue(insecureSkipVerify) {
		logger.Log.Warn(fmt.Sprintf("You provided a custom CA and disabled certificate validation (check %s and %s)",
			defaultCertificateAuthorityPath.Key, defaultInsecureSkipVerify.Key))
	}

	// If a custom CA is provided or INSECURE_SKIP_VERIFY is set and the exporter URL is not HTTPS, that's kinda sus
	if (certificateAuthorityPath != "" || envSetToTrue(insecureSkipVerify)) && !internal.IsValidHttpsURL(baseUrl) {
		logger.Log.Warn(fmt.Sprintf("You provided a custom CA or disabled certificate validation but the qBittorrent URL is not HTTPS. (check %s, %s and %s)",
			defaultCertificateAuthorityPath.Key, defaultInsecureSkipVerify.Key, defaultBaseUrl.Key))
	}

	// If a custom CA is provided, load the root CAs from the system and append the custom CA
	var caCertPool *x509.CertPool
	if certificateAuthorityPath != "" {
		caCert, errCaCert := os.ReadFile(certificateAuthorityPath)
		if errCaCert != nil {
			panic(fmt.Sprintf("Error reading certificate authority file: %s (check %s)",
				errCaCert, defaultCertificateAuthorityPath.Key))
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
	case "TLS_1_2":
		minTlsVersion = tls.VersionTLS12
	case "TLS_1_3":
		minTlsVersion = tls.VersionTLS13
	default:
		panic(fmt.Sprintf("Invalid minimum TLS version: %s (valid options are TLS_1_2 and TLS_1_3) (check %s)",
			minTlsVersionStr, defaultMinTlsVersion.Key))
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
		BaseUrl:  baseUrl,
		Username: qbitUsername,
		Password: qbitPassword,
		Timeout:  time.Duration(timeoutDuration) * time.Second,
		Cookie:   nil,
		BasicAuth: BasicAuth{
			Username: &qbitBasicAuthUsername,
			Password: &qbitBasicAuthPassword,
		},
	}

	Exporter = ExporterSettings{
		Features: Features{
			EnableHighCardinality:        envSetToTrue(enableHighCardinality),
			EnableTracker:                envSetToTrue(enableTracker),
			ShowPassword:                 envSetToTrue(showPassword),
			EnableBasicAuthRequestHeader: envSetToTrue(enableQbittorrentBasicAuth),
		},
		ExperimentalFeatures: ExperimentalFeatures{
			EnableLabelWithHash: envSetToTrue(labelWithHash),
		},
		LogLevel:  loglevel,
		Port:      exporterPort,
		URL:       exporterUrl,
		Path:      exporterPath,
		BasicAuth: basicAuth,
	}

}

func envSetToTrue(env string) bool {
	return strings.ToLower(env) == "true"
}

func GetPasswordMasked() string {
	return strings.Repeat("*", len(QBittorrent.Password))
}

func GetFeaturesEnabled() string {
	features := ""

	addComma := func() {
		if features != "" {
			features += ", "
		}
	}

	if Exporter.Features.EnableHighCardinality {
		features += "High cardinality"
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

	if Exporter.Features.EnableBasicAuthRequestHeader {
		addComma()
		features += "Send HTTP Basic Authorization request header to qBittorrent"
	}

	features = fmt.Sprintf("[%s]", features)

	return features
}
