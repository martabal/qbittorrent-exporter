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
	QBittorrentTimeout  time.Duration
	Port                int
	ShouldShowError     bool
	LogLevel            string
	BaseUrl             string
	Cookie              string
	Username            string
	Password            string
	ExperimentalFeature ExperimentalFeatures
	Feature             Features
	UsingEnvFile        bool
)

type ExperimentalFeatures struct {
	EnableLabelWithHash bool
}

type Features struct {
	EnableHighCardinality bool
	EnableTracker         bool
}

func SetVar(port int, enableTracker bool, loglevel string, baseUrl string, username string, password string, qBittorrentTimeout int, enableHighCardinality bool, enableLabelWithHash bool) {
	Port = port
	ShouldShowError = true
	Feature.EnableTracker = enableTracker
	LogLevel = loglevel
	BaseUrl = baseUrl
	Username = username
	Password = password
	QBittorrentTimeout = time.Duration(qBittorrentTimeout)
	Feature.EnableHighCardinality = enableHighCardinality
	ExperimentalFeature.EnableLabelWithHash = enableLabelWithHash
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
		panic(fmt.Sprintf("%s must be an integer", defaultPort.Key))
	}
	if exporterPort < 0 || exporterPort > 65353 {
		panic(fmt.Sprintf("%s must be > 0 and < 65353", defaultPort.Key))
	}

	timeoutDuration, errTimeoutDuration := strconv.Atoi(timeoutDurationEnv)
	if errTimeoutDuration != nil {
		panic(fmt.Sprintf("%s must be an integer", defaultPort.Key))
	}
	if timeoutDuration < 0 {
		panic(fmt.Sprintf("%s must be > 0", defaultPort.Key))
	}

	SetVar(exporterPort, envSetToTrue(enableTracker), loglevel, qbitURL, qbitUsername, qbitPassword, timeoutDuration, envSetToTrue(enableHighCardinality), envSetToTrue(labelWithHash))
}

func envSetToTrue(env string) bool {
	return strings.ToLower(env) == "true"
}

func GetPasswordMasked() string {
	return strings.Repeat("*", len(Password))
}

func GetFeaturesEnabled() string {
	features := ""

	addComma := func() {
		if features != "" {
			features += ", "
		}
	}

	if Feature.EnableHighCardinality {
		features += "High cardinality"
	}

	if Feature.EnableTracker {
		addComma()
		features += "Trackers"
	}

	if ExperimentalFeature.EnableLabelWithHash {
		addComma()
		features += "Label with hash (experimental)"
	}

	features = "[" + features + "]"

	return features
}
