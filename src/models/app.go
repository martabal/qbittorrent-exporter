package models

type TypeAppConfig struct {
	Port           int
	Error          bool
	DisableTracker bool
}

var AppConfig TypeAppConfig

func SetApp(setport int, seterror bool, setdisableTracker bool) {
	AppConfig = TypeAppConfig{
		Port:           setport,
		Error:          seterror,
		DisableTracker: setdisableTracker,
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

func GetPromptError() bool {
	return AppConfig.Error
}
