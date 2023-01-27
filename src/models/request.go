package models

type Request struct {
	Base_url string
	Cookie   string
}

var myrequest Request

func Getrequest() (string, string) {
	return myrequest.Base_url, myrequest.Cookie
}

func Setrequest(base_url string, cookie string) {
	myrequest.Base_url = base_url
	myrequest.Cookie = cookie
}

func Setbaseurl(base_url string) {
	myrequest.Base_url = base_url
}

func Setcookie(cookie string) {
	myrequest.Cookie = cookie
}

func Getbaseurl() string {
	return myrequest.Base_url
}

func Getcookie() string {
	return myrequest.Cookie
}
