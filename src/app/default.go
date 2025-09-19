package app

import (
	"fmt"
	"os"
	"qbit-exp/logger"
	"strconv"
)

type Env struct {
	Key          string
	DefaultValue string
	Help         string
}

const defaultExporterPort int = 8090
const DefaultTimeout int = 30
const defaultExporterPath string = "/metrics"

const TLS12 string = "TLS_1_2"
const TLS13 string = "TLS_1_3"

// Exporter

var defaultEnableTracker = Env{
	Key:          "ENABLE_TRACKER",
	DefaultValue: "true",
	Help:         "",
}

var defaultLabelWithTracker = Env{
	Key:          "ENABLE_LABEL_WITH_TRACKER",
	DefaultValue: "false",
	Help:         "",
}

var defaultExporterURL = "EXPORTER_URL"

var defaultExporterPathEnv = Env{
	Key:          "EXPORTER_PATH",
	DefaultValue: defaultExporterPath,
	Help:         "",
}

var defaultExporterShowPassword = Env{
	Key:          "DANGEROUS_SHOW_PASSWORD",
	DefaultValue: "false",
	Help:         "",
}

var defaultHighCardinality = Env{
	Key:          "ENABLE_HIGH_CARDINALITY",
	DefaultValue: "false",
	Help:         "",
}

var defaultIncreasedCardinality = Env{
	Key:          "ENABLE_INCREASED_CARDINALITY",
	DefaultValue: "false",
	Help:         "",
}

var defaultLabelWithHash = Env{
	Key:          "ENABLE_LABEL_WITH_HASH",
	DefaultValue: "false",
	Help:         "",
}

var defaultLogLevel = Env{
	Key:          "LOG_LEVEL",
	DefaultValue: "INFO",
	Help:         "",
}
var defaultPort = Env{
	Key:          "EXPORTER_PORT",
	DefaultValue: strconv.Itoa(defaultExporterPort),
	Help:         "",
}

// QBittorrent

var defaultBaseUrl = Env{
	Key:          "QBITTORRENT_BASE_URL",
	DefaultValue: "http://localhost:8080",
	Help:         "qBittorrent base_url is not set. Using default base_url",
}

var defaultBasicAuthUsername = "EXPORTER_BASIC_AUTH_USERNAME"

var defaultBasicAuthPassword = "EXPORTER_BASIC_AUTH_PASSWORD"

var defaultCertificateAuthorityPath = "CERTIFICATE_AUTHORITY_PATH"

var defaultInsecureSkipVerify = Env{
	Key:          "INSECURE_SKIP_VERIFY",
	DefaultValue: "false",
	Help:         "",
}

var defaultMinTlsVersion = Env{
	Key:          "MIN_TLS_VERSION",
	DefaultValue: TLS13,
	Help:         "",
}

var defaultPassword = Env{
	Key:          "QBITTORRENT_PASSWORD",
	DefaultValue: "adminadmin",
	Help:         "qBittorrent password is not set. Using default password",
}

var defaultPasswordFile = "QBITTORRENT_PASSWORD_FILE"

var defaultQbitBasicAuthUsername = "QBITTORRENT_BASIC_AUTH_USERNAME"

var defaultQbitBasicAuthPassword = "QBITTORRENT_BASIC_AUTH_PASSWORD"

var defaultUsername = Env{
	Key:          "QBITTORRENT_USERNAME",
	DefaultValue: "admin",
	Help:         "qBittorrent username is not set. Using default username",
}

var defaultTimeout = Env{
	Key:          "QBITTORRENT_TIMEOUT",
	DefaultValue: strconv.Itoa(DefaultTimeout),
	Help:         "",
}

func getEnv(env Env) (string, bool) {
	if value, ok := os.LookupEnv(env.Key); ok && value != "" {
		return value, false
	}
	if env.Help != "" {
		logger.Log.Warn(fmt.Sprintf("%s (%s)", env.Help, env.DefaultValue))
	}
	return env.DefaultValue, true
}

func getOptionalEnv(env string) *string {
	if value, ok := os.LookupEnv(env); ok {
		return &value
	}
	return nil
}
