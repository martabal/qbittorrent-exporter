package models

type TypeAppConfig struct {
	Port           int
	Error          bool
	LogLevel       string
	DisableTracker bool
}

var AppConfig TypeAppConfig

func SetApp(setport int, seterror bool, setdisableTracker bool, setloglevel string) {
	AppConfig = TypeAppConfig{
		Port:           setport,
		Error:          seterror,
		DisableTracker: setdisableTracker,
		LogLevel:       setloglevel,
	}
}

func GetFeatureFlag() bool {
	return AppConfig.DisableTracker
}

func GetPort() int {
	return AppConfig.Port
}

func SetPromptError(prompt bool) {
	AppConfig.Error = prompt
}

func GetLogLevel() string {
	return AppConfig.LogLevel
}

func GetPromptError() bool {
	return AppConfig.Error
}
