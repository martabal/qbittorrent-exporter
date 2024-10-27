package qbit

import (
	"context"
	"io"
	"net/http"
	"net/url"
	API "qbit-exp/api"
	app "qbit-exp/app"
	"qbit-exp/logger"
	"strconv"
	"strings"
)

func Auth() {
	ctx, cancel := context.WithTimeout(context.Background(), app.QBittorrent.Timeout)
	defer cancel()
	params := url.Values{
		"username": {app.QBittorrent.Username},
		"password": {app.QBittorrent.Password},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, app.QBittorrent.BaseUrl+"/api/v2/auth/login", strings.NewReader(params.Encode()))
	if err != nil {
		panic("Error with url " + err.Error())
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)

	if ctx.Err() == context.DeadlineExceeded {
		if app.ShouldShowError {
			app.ShouldShowError = false
			logger.Log.Error(API.QbittorrentTimeOut)
		}
	}

	if err != nil {
		logger.Log.Debug(err.Error())
		if app.ShouldShowError {
			app.ShouldShowError = false
			logger.Log.Warn("Can't connect to qbittorrent with url : " + app.QBittorrent.BaseUrl)
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
		panic("Error reading the body" + err.Error())
	}

	if string(body) == "Fails." {
		panic("Authentication Error, check your qBittorrent username / password")
	}
	logFunc := logger.Log.Info
	if app.ShouldShowError {
		logFunc = logger.Log.Debug
	}
	logFunc("New cookie for auth stored")

	cookie := resp.Header.Get("Set-Cookie")
	cookieValue := strings.Split(strings.Split(cookie, ";")[0], "=")[1]
	app.QBittorrent.Cookie = cookieValue
}
