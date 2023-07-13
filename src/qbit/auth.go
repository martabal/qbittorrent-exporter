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

func Auth(init bool) error {
	qbit_url, username, password := models.GetQbit()
	params := url.Values{}
	params.Add("username", username)
	params.Add("password", password)
	resp, err := http.PostForm(qbit_url+"/api/v2/auth/login", params)
	if err != nil {
		if init {
			log.Panicln("Can't connect to qbittorrent with url : ", models.Getbaseurl())
		} else {
			log.Warn("Can't connect to qbittorrent with url : ", models.Getbaseurl())
		}
		return err
	} else {
		if resp.StatusCode == 200 {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatalln(err)
			}
			if string(body) == "Fails." {
				log.Panicln("Authentication Error, check your qBittorrent username / password")
				return fmt.Errorf("Authentication Error")
			} else {
				if models.GetPromptError() {
					log.Info("New cookie stored")
				}
				models.Setcookie(strings.Split(strings.Split(resp.Header["Set-Cookie"][0], ";")[0], "=")[1])
				return nil
			}
		} else {
			return fmt.Errorf("Unknown error")
		}
	}
}
