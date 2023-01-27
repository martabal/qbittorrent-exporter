package models

type User struct {
	Username string
	Password string
}

var myuser User

func mask(input string) string {
	hide := ""
	for i := 0; i < len(input); i++ {
		hide += "*"
	}
	return hide
}

func Getuser() (string, string) {
	return myuser.Username, myuser.Password
}

func Setuser(username string, password string) {
	myuser.Username = username
	myuser.Password = password
}

func GetUsername() string {
	return myuser.Username
}

func Getpassword() string {
	return myuser.Password
}
func Getpasswordmasked() string {
	return mask(myuser.Password)
}
