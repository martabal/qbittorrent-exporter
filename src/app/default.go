package app

import (
	"os"
	"qbit-exp/logger"
	"strconv"
)

const DefaultExporterPort = 8090
const DefaultTimeout = 30
const DefaultExporterPath = "/metrics"

type Env struct {
	Key          string
	DefaultValue string
	Help         string
}

var defaultLogLevel = Env{
	Key:          "LOG_LEVEL",
	DefaultValue: "INFO",
	Help:         "",
}

var defaultPort = Env{
	Key:          "EXPORTER_PORT",
	DefaultValue: strconv.Itoa(DefaultExporterPort),
	Help:         "",
}

var defaultTimeout = Env{
	Key:          "QBITTORRENT_TIMEOUT",
	DefaultValue: strconv.Itoa(DefaultTimeout),
	Help:         "",
}

var defaultUsername = Env{
	Key:          "QBITTORRENT_USERNAME",
	DefaultValue: "admin",
	Help:         "Qbittorrent username is not set. Using default username",
}

var defaultPassword = Env{
	Key:          "QBITTORRENT_PASSWORD",
	DefaultValue: "adminadmin",
	Help:         "Qbittorrent password is not set. Using default password",
}

var defaultBaseUrl = Env{
	Key:          "QBITTORRENT_BASE_URL",
	DefaultValue: "http://localhost:8080",
	Help:         "Qbittorrent base_url is not set. Using default base_url",
}

var defaultDisableTracker = Env{
	Key:          "ENABLE_TRACKER",
	DefaultValue: "true",
	Help:         "",
}

var defaultHighCardinality = Env{
	Key:          "ENABLE_HIGH_CARDINALITY",
	DefaultValue: "false",
	Help:         "",
}

var defaultLabelWithHash = Env{
	Key:          "ENABLE_LABEL_WITH_HASH",
	DefaultValue: "false",
	Help:         "",
}

var defaultExporterURL = Env{
	Key:          "EXPORTER_URL",
	DefaultValue: "",
	Help:         "",
}

var defaultExporterPath = Env{
	Key:          "EXPORTER_PATH",
	DefaultValue: DefaultExporterPath,
	Help:         "",
}

func getEnv(env Env) string {
	value, ok := os.LookupEnv(env.Key)
	if !ok || value == "" {
		if env.Help != "" {
			logger.Log.Warn(env.Help)
		}
		return env.DefaultValue
	}
	return value
}
