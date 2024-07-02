package app

import "strconv"

const DEFAULT_PORT = 8090
const DEFAULT_TIMEOUT = 30

type Env struct {
	Key          string
	DefaultValue string
	Help         string
}

var LogLevelEnv = Env{
	Key:          "LOG_LEVEL",
	DefaultValue: "INFO",
	Help:         "",
}

var PortEnv = Env{
	Key:          "EXPORTER_PORT",
	DefaultValue: strconv.Itoa(DEFAULT_PORT),
	Help:         "",
}

var TimeoutEnv = Env{
	Key:          "EXPORTER_PORT",
	DefaultValue: strconv.Itoa(DEFAULT_TIMEOUT),
	Help:         "",
}

var UsernameEnv = Env{
	Key:          "QBITTORRENT_USERNAME",
	DefaultValue: "admin",
	Help:         "Qbittorrent username is not set. Using default username",
}

var PasswordEnv = Env{
	Key:          "QBITTORRENT_PASSWORD",
	DefaultValue: "adminadmin",
	Help:         "Qbittorrent password is not set. Using default password",
}

var BaseUrlEnv = Env{
	Key:          "QBITTORRENT_BASE_URL",
	DefaultValue: "http://localhost:8080",
	Help:         "Qbittorrent base_url is not set. Using default base_url",
}

var DisableTrackerEnv = Env{
	Key:          "QBITTORRENT_BASE_URL",
	DefaultValue: "http://localhost:8080",
	Help:         "Qbittorrent base_url is not set. Using default base_url",
}
