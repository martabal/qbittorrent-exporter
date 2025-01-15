package app

import (
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
