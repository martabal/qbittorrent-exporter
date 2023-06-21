package qbit

import (
	"io"
	"net/http"
	"net/url"
	"qbit-exp/src/models"
	"strings"

	log "github.com/sirupsen/logrus"
)

func Auth() {
	username, password := models.Getuser()
	qbit_url := models.Getbaseurl()
	params := url.Values{}
	params.Add("username", username)
	params.Add("password", password)
	resp, err := http.PostForm(qbit_url+"/api/v2/auth/login", params)
	if err != nil {
		log.Warn("Can't connect to ", models.Getbaseurl())
	} else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		if resp.StatusCode == 200 {
			if string(body) == "Fails." {
				log.Error("Authentication Error")
			} else {
				if models.GetPromptError() {
					log.Info("New cookie stored")
				}
				models.Setcookie(strings.Split(strings.Split(resp.Header["Set-Cookie"][0], ";")[0], "=")[1])
			}
		}
	}
}
