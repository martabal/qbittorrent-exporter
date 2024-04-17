package app

import "strings"

var (
	Port            int
	ShouldShowError bool
	DisableTracker  bool
	LogLevel        string
	BaseUrl         string
	Cookie          string
	Username        string
	Password        string
)

func SetApp(port int, disableTracker bool, loglevel string) {
	Port = port
	ShouldShowError = false
	DisableTracker = disableTracker
	LogLevel = loglevel
}

func SetQbit(baseUrl string, username string, password string) {
	BaseUrl = baseUrl
	Username = username
	Password = password
}

func GetPasswordMasked() string {
	return strings.Repeat("*", len(Password))
}
