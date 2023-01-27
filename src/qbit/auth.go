package qbit

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"qbit-prom/src/models"
	"strings"
)

func Auth() {
	username, password := models.Getuser()
	qbit_url := models.Getbaseurl()
	params := url.Values{}
	params.Add("username", username)
	params.Add("password", password)
	resp, err := http.PostForm(qbit_url+"/api/v2/auth/login", params)
	if err != nil {
		log.Println("Can't connect to ", models.Getbaseurl())
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		if string(body) == "Fails." {
			log.Println("Authentication Error")
		} else {
			cookies := resp.Header["Set-Cookie"][0]
			onlycookie := strings.Split(cookies, ";")[0]
			cookie := strings.Split(onlycookie, "=")[1]
			log.Println("New cookie :", cookie)
			models.Setcookie(cookie)
		}

	}

}
