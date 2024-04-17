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

func SetVar(port int, disableTracker bool, loglevel string, baseUrl string, username string, password string) {
	Port = port
	ShouldShowError = true
	DisableTracker = disableTracker
	LogLevel = loglevel
	BaseUrl = baseUrl
	Username = username
	Password = password
}

func GetPasswordMasked() string {
	return strings.Repeat("*", len(Password))
}
