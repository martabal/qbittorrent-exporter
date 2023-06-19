package models

type TypeAppConfig struct {
	Port int
}

var AppConfig TypeAppConfig
var myerr bool

func SetApp(port int) {
	AppConfig = TypeAppConfig{
		Port: port,
	}
}

func GetPort() int {
	return AppConfig.Port
}

func SetPromptError(prompt bool) {
	myerr = prompt
}

func GetPromptError() bool {
	return myerr
}
