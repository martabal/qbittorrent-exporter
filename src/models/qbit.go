package models

import "strings"

type TypeQbitConfig struct {
	Base_url string
	Cookie   string
	Username string
	Password string
}

var Config TypeQbitConfig

func SetQbit(baseurl string, username string, password string) {
	Config = TypeQbitConfig{
		Base_url: baseurl,
		Username: username,
		Password: password,
	}
}

func Setcookie(cookie string) {
	Config.Cookie = cookie
}

func Getbaseurl() string {
	return Config.Base_url
}

func Getcookie() string {
	return Config.Cookie
}

func mask(input string) string {
	return strings.Repeat("*", len(input))
}

func Getuser() (string, string) {
	return Config.Username, Config.Password
}

func GetUsername() string {
	return Config.Username
}

func GetQbit() (string, string, string) {
	return Config.Base_url, Config.Username, Config.Password
}

func Getpasswordmasked() string {
	return mask(Config.Password)
}
