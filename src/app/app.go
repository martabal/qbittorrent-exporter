package app

import (
	"flag"
	"fmt"
	"os"
	"qbit-exp/logger"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var (
	QBittorrent     QBittorrentSettings
	Exporter        ExporterSettings
	ShouldShowError bool
	UsingEnvFile    bool
)

type ExporterSettings struct {
	Port                int
	LogLevel            string
	ExperimentalFeature ExperimentalFeatures
	Feature             Features
}

type QBittorrentSettings struct {
	Timeout  time.Duration
	BaseUrl  string
	Cookie   string
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

func SetVar(port int, enableTracker bool, loglevel string, baseUrl string, username string, password string, timeout int, enableHighCardinality bool, enableLabelWithHash bool) {
	ShouldShowError = true
	QBittorrent = QBittorrentSettings{
		BaseUrl:  baseUrl,
		Username: username,
		Password: password,
		Timeout:  time.Duration(timeout) * time.Second,
	}
	Exporter = ExporterSettings{
		Feature: Features{
			EnableHighCardinality: enableHighCardinality,
			EnableTracker:         enableTracker,
		},
		ExperimentalFeature: ExperimentalFeatures{
			EnableLabelWithHash: enableLabelWithHash,
		},
		LogLevel: loglevel,
		Port:     port,
	}

}

func LoadEnv() {
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

	loglevel := logger.SetLogLevel(getEnv(defaultLogLevel))
	qbitUsername := getEnv(defaultUsername)
	qbitPassword := getEnv(defaultPassword)
	qbitURL := strings.TrimSuffix(getEnv(defaultBaseUrl), "/")
	exporterPortEnv := getEnv(defaultPort)
	timeoutDurationEnv := getEnv(defaultTimeout)
	enableTracker := getEnv(defaultDisableTracker)
	enableHighCardinality := getEnv(defaultHighCardinality)
	labelWithHash := getEnv(defaultLabelWithHash)

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

	SetVar(exporterPort, envSetToTrue(enableTracker), loglevel, qbitURL, qbitUsername, qbitPassword, timeoutDuration, envSetToTrue(enableHighCardinality), envSetToTrue(labelWithHash))
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

	features = "[" + features + "]"

	return features
}
