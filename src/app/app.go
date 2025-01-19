package app

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
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
	TlsConfig    tls.Config
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
}

type ExperimentalFeatures struct {
	EnableLabelWithHash bool
}

type Features struct {
	EnableHighCardinality bool
	EnableTracker         bool
	ShowPassword          bool
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
		panic(fmt.Sprintf("%s must be an integer", exporterPortEnv))
	}
	if exporterPort < 0 || exporterPort > 65353 {
		panic(fmt.Sprintf("%d must be > 0 and < 65353", exporterPort))
	}

	timeoutDuration, errTimeoutDuration := strconv.Atoi(timeoutDurationEnv)
	if errTimeoutDuration != nil {
		panic(fmt.Sprintf("%s must be an integer", timeoutDurationEnv))
	}
	if timeoutDuration < 0 {
		panic(fmt.Sprintf("%d must be > 0", timeoutDuration))
	}

	if exporterUrl != "" {
		exporterUrl = strings.TrimSuffix(exporterUrl, "/")
		if !internal.IsValidURL(exporterUrl) {
			panic(fmt.Sprintf("%s is not a valid URL", exporterUrl))
		}
	}

	basicAuth := BasicAuth{Username: nil, Password: nil}
	if basicAuthUsername != "" && basicAuthPassword == "" {
		logger.Log.Warn("You set a basic auth username but not password")
	} else if basicAuthUsername == "" && basicAuthPassword != "" {
		logger.Log.Warn("You set a basic auth password but not username")
	} else if basicAuthUsername != "" && basicAuthPassword != "" {
		basicAuth = BasicAuth{
			Username: &basicAuthUsername,
			Password: &basicAuthPassword,
		}
	}

	// If a custom CA is provided, load the root CAs from the system and append the custom CA
	var caCertPool *x509.CertPool
	if certificateAuthorityPath != "" {
		caCert, errCaCert := os.ReadFile(certificateAuthorityPath)
		if errCaCert != nil {
			panic(fmt.Sprintf("Error reading certificate authority file: %s", errCaCert))
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
	case "TLS_1_0":
		minTlsVersion = tls.VersionTLS10
	case "TLS_1_1":
		minTlsVersion = tls.VersionTLS11
	case "TLS_1_2":
		minTlsVersion = tls.VersionTLS12
	case "TLS_1_3":
		minTlsVersion = tls.VersionTLS13
	default:
		panic(fmt.Sprintf("Invalid minimum TLS version: %s (valid options are TLS_1_0, TLS_1_1, TLS_1_2, TLS_1_3)", minTlsVersionStr))
	}

	TlsConfig = tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: envSetToTrue(insecureSkipVerify),
		MinVersion:         minTlsVersion,
	}

	internal.EnsureLeadingSlash(&exporterPath)

	QBittorrent = QBittorrentSettings{
		BaseUrl:  baseUrl,
		Username: qbitUsername,
		Password: qbitPassword,
		Timeout:  time.Duration(timeoutDuration) * time.Second,
		Cookie:   nil,
	}

	Exporter = ExporterSettings{
		Features: Features{
			EnableHighCardinality: envSetToTrue(enableHighCardinality),
			EnableTracker:         envSetToTrue(enableTracker),
			ShowPassword:          envSetToTrue(showPassword),
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

	features = fmt.Sprintf("[%s]", features)

	return features
}
