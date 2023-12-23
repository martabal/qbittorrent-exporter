package qbit

import (
	"io"
	"net/http"
	"net/url"
	"qbit-exp/src/models"
	"strings"

	log "github.com/sirupsen/logrus"
)

func Auth(init bool) {
	params := url.Values{
		"username": {models.GetUsername()},
		"password": {models.Getpassword()},
	}
	resp, err := http.PostForm(models.Getbaseurl()+"/api/v2/auth/login", params)
	if err != nil {
		if !models.GetPromptError() {
			models.SetPromptError(true)
			log.Warn("Can't connect to qbittorrent with url : " + models.Getbaseurl())
		}
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("Unknown error")
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
	}

	if string(body) == "Fails." {
		log.Panicln("Authentication Error, check your qBittorrent username / password")
		return
	}

	if models.GetPromptError() {
		log.Info("New cookie stored")
	}

	cookie := resp.Header.Get("Set-Cookie")
	cookieValue := strings.Split(strings.Split(cookie, ";")[0], "=")[1]
	models.Setcookie(cookieValue)

	return
}
