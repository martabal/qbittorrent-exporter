package models

var myerr bool

func SetPromptError(prompt bool) {
	myerr = prompt
}

func GetPromptError() bool {
	return myerr
}
