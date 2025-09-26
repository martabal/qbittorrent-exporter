package qbit

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	API "qbit-exp/api"
	app "qbit-exp/app"
	"qbit-exp/logger"
)

func Auth() error {
	ctx, cancel := context.WithTimeout(context.Background(), app.QBittorrent.Timeout)
	defer cancel()
	params := url.Values{
		"username": {app.QBittorrent.Username},
		"password": {app.QBittorrent.Password},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/api/v2/auth/login", app.QBittorrent.BaseUrl), strings.NewReader(params.Encode()))
	if err != nil {
		panic(fmt.Sprintf("%s %s", API.ErrorWithUrl, err.Error()))
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if app.QBittorrent.BasicAuth != nil {
		req.SetBasicAuth(app.QBittorrent.BasicAuth.Username, app.QBittorrent.BasicAuth.Password)
	}

	resp, err := app.HttpClient.Do(req)

	if ctx.Err() == context.DeadlineExceeded {
		logger.Log.Error(API.QbittorrentTimeOut)

		return context.DeadlineExceeded
	}
	if err != nil {
		err := fmt.Errorf("%s: %v", API.ErrorConnect, err)
		logger.Log.Error(err.Error())
		return err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Log.Error(fmt.Sprintf("Error closing body %v", err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("authentication failed, status code: %d", resp.StatusCode)
		if resp.StatusCode == http.StatusForbidden && app.QBittorrent.Cookie == nil {
			panic(fmt.Sprintf("%s. qBittorrent has probably banned your IP", err.Error()))
		}
		logger.Log.Error(err.Error())
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("Error reading the body %s", err.Error()))
	}
	if string(body) == "Fails." {
		panic("Authentication Error, check your qBittorrent username / password")
	}

	logger.Log.Info("New cookie for auth stored")

	cookie := resp.Header.Get("Set-Cookie")
	cookieValue := strings.Split(strings.Split(cookie, ";")[0], "=")[1]
	app.QBittorrent.Cookie = &cookieValue
	return nil
}
