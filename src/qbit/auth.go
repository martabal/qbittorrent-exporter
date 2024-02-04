package qbit

import (
	"io"
	"net/http"
	"net/url"
	"qbit-exp/logger"
	"qbit-exp/models"
	"strconv"
	"strings"
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
			logger.Log.Warn("Can't connect to qbittorrent with url : " + models.Getbaseurl())
		}
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Log.Error("Unknown error, status code " + strconv.Itoa(resp.StatusCode))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errormessage := "Error reading the body" + err.Error()
		panic(errormessage)
	}

	if string(body) == "Fails." {
		panic("Authentication Error, check your qBittorrent username / password")
	}

	if models.GetPromptError() {
		logger.Log.Info("New cookie stored")
	}

	cookie := resp.Header.Get("Set-Cookie")
	cookieValue := strings.Split(strings.Split(cookie, ";")[0], "=")[1]
	models.Setcookie(cookieValue)

}
