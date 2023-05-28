package models

type QbitConfig struct {
	Base_url string
	Cookie   string
	Username string
	Password string
}

var Config QbitConfig

func Init(baseurl string, username string, password string) {
	Config = QbitConfig{
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
	hide := ""
	for i := 0; i < len(input); i++ {
		hide += "*"
	}
	return hide
}

func Getuser() (string, string) {
	return Config.Username, Config.Password
}

func GetUsername() string {
	return Config.Username
}

func Getpasswordmasked() string {
	return mask(Config.Password)
}
