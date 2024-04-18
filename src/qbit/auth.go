package qbit

import (
	"io"
	"net/http"
	"net/url"
	app "qbit-exp/app"
	"qbit-exp/logger"
	"strconv"
	"strings"
)

func Auth() {
	params := url.Values{
		"username": {app.Username},
		"password": {app.Password},
	}
	resp, err := http.PostForm(app.BaseUrl+"/api/v2/auth/login", params)
	if err != nil {
		if app.ShouldShowError {
			app.ShouldShowError = false
			logger.Log.Warn("Can't connect to qbittorrent with url : " + app.BaseUrl)
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

	if !app.ShouldShowError {
		logger.Log.Info("New cookie stored")
	}

	cookie := resp.Header.Get("Set-Cookie")
	cookieValue := strings.Split(strings.Split(cookie, ";")[0], "=")[1]
	app.Cookie = cookieValue
}
