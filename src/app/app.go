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
	Port                int
	LogLevel            string
	ExperimentalFeature ExperimentalFeatures
	Feature             Features
	URL                 string
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

	if exporterUrl != "" && !internal.IsValidURL(exporterUrl) {
		panic(fmt.Sprintf("%s is not a valid URL", exporterUrl))
	}

	QBittorrent = QBittorrentSettings{
		BaseUrl:  baseUrl,
		Username: qbitUsername,
		Password: qbitPassword,
		Timeout:  time.Duration(timeoutDuration) * time.Second,
		Cookie:   nil,
	}

	Exporter = ExporterSettings{
		Feature: Features{
			EnableHighCardinality: envSetToTrue(enableHighCardinality),
			EnableTracker:         envSetToTrue(enableTracker),
		},
		ExperimentalFeature: ExperimentalFeatures{
			EnableLabelWithHash: envSetToTrue(labelWithHash),
		},
		LogLevel: loglevel,
		Port:     exporterPort,
		URL:      exporterUrl,
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

	if Exporter.Feature.EnableHighCardinality {
		features += "High cardinality"
	}

	if Exporter.Feature.EnableTracker {
		addComma()
		features += "Trackers"
	}

	if Exporter.ExperimentalFeature.EnableLabelWithHash {
		addComma()
		features += "Label with hash (experimental)"
	}

	features = fmt.Sprintf("[%s]", features)

	return features
}
