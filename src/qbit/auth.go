package qbit

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"qbit-exp/src/models"
	"strings"

	log "github.com/sirupsen/logrus"
)

func Auth(init bool) (string, error) {
	username, password := models.Getuser()
	qbit_url := models.Getbaseurl()
	params := url.Values{}
	params.Add("username", username)
	params.Add("password", password)
	resp, err := http.PostForm(qbit_url+"/api/v2/auth/login", params)
	if err != nil {
		if init {
			log.Panicln("Can't connect to qbittorrent with url", models.Getbaseurl())
		} else {
			log.Warn("Can't connect to qbittorrent with url", models.Getbaseurl())
		}
		return "", err
	} else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		if resp.StatusCode == 200 {
			if string(body) == "Fails." {
				autherror := fmt.Errorf("Authentication Error")
				log.Panicln("Authentication Error, check your qBittorrent username / password")
				return "", autherror
			} else {
				if models.GetPromptError() {
					log.Info("New cookie stored")
				}
				return strings.Split(strings.Split(resp.Header["Set-Cookie"][0], ";")[0], "=")[1], nil
			}
		} else {
			return "", fmt.Errorf("Unknown error")
		}
	}
}
