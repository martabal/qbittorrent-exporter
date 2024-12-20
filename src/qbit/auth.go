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

	client := &http.Client{}
	resp, err := client.Do(req)

	if ctx.Err() == context.DeadlineExceeded {
		logger.Log.Error(API.QbittorrentTimeOut)

		return context.DeadlineExceeded
	}
	if err != nil {
		err := fmt.Errorf("%s: %v", API.ErrorConnect, err)
		logger.Log.Error(err.Error())
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("Authentication failed, status code: %d", resp.StatusCode)
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
