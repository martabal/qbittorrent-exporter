package qbit

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	API "qbit-exp/api"
	app "qbit-exp/app"
	"qbit-exp/logger"
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
		panic(API.ErrorWithUrl + err.Error())
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)

	if ctx.Err() == context.DeadlineExceeded {
		if app.ShouldShowError {
			logger.Log.Error(API.QbittorrentTimeOut)
			app.ShouldShowError = false
		}
		return
	}
	if err != nil {
		if app.ShouldShowError {
			logger.Log.Error(fmt.Sprintf("%s: %v", API.ErrorConnect, err))
			app.ShouldShowError = false
		}
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Log.Error(fmt.Sprintf("Authentication failed, status code: %d", resp.StatusCode))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic("Error reading the body" + err.Error())
	}
	if string(body) == "Fails." {
		panic("Authentication Error, check your qBittorrent username / password")
	}

	logger.Log.Info("New cookie for auth stored")

	cookie := resp.Header.Get("Set-Cookie")
	cookieValue := strings.Split(strings.Split(cookie, ";")[0], "=")[1]
	app.QBittorrent.Cookie = cookieValue
}
