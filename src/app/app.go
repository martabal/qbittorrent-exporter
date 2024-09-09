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
	QBittorrentTimeout time.Duration
	Port               int
	ShouldShowError    bool
	DisableTracker     bool
	LogLevel           string
	BaseUrl            string
	Cookie             string
	Username           string
	Password           string
)

func SetVar(port int, disableTracker bool, loglevel string, baseUrl string, username string, password string, qBittorrentTimeout int) {
	Port = port
	ShouldShowError = true
	DisableTracker = disableTracker
	LogLevel = loglevel
	BaseUrl = baseUrl
	Username = username
	Password = password
	QBittorrentTimeout = time.Duration(qBittorrentTimeout)
}

func LoadEnv() {
	var envfile bool
	flag.BoolVar(&envfile, "e", false, "Use .env file")
	flag.Parse()
	_, err := os.Stat(".env")
	if !os.IsNotExist(err) && !envfile {
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
	disableTracker := getEnv(defaultDisableTracker)

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

	SetVar(exporterPort, strings.ToLower(disableTracker) == "true", loglevel, qbitURL, qbitUsername, qbitPassword, timeoutDuration)
}

func GetPasswordMasked() string {
	return strings.Repeat("*", len(Password))
}
