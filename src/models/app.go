package models

type TypeAppConfig struct {
	Port  int
	Error bool
}

var AppConfig TypeAppConfig

func SetApp(setport int, seterror bool) {
	AppConfig = TypeAppConfig{
		Port:  setport,
		Error: seterror,
	}
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
