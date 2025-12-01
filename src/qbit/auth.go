package qbit

import (
	"context"
	"errors"
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, app.QBittorrent.BaseUrl+"/api/v2/auth/login", strings.NewReader(params.Encode()))
	if err != nil {
		panic(fmt.Sprintf("%s %s", API.ErrorWithUrl, err.Error()))
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if app.QBittorrent.BasicAuth != nil {
		req.SetBasicAuth(app.QBittorrent.BasicAuth.Username, app.QBittorrent.BasicAuth.Password)
	}

	resp, err := app.HttpClient.Do(req)

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		logger.Error(API.QbittorrentTimeOut)

		return context.DeadlineExceeded
	}

	if err != nil {
		err := fmt.Errorf("%s: %w", API.ErrorConnect, err)
		logger.Error(err.Error())

		return err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logger.Error(fmt.Sprintf("Error closing body %v", err))
		}
	}()

	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic("Error reading the body " + err.Error())
		}

		if string(body) == "Fails." {
			panic("Authentication Error, check your qBittorrent username / password")
		}
	} else if resp.StatusCode != http.StatusNoContent {
		err := fmt.Errorf("authentication failed, status code: %d", resp.StatusCode)
		if resp.StatusCode == http.StatusForbidden && app.QBittorrent.Cookie == nil {
			panic(err.Error() + ". qBittorrent has probably banned your IP")
		}

		logger.Error(err.Error())

		return err
	}

	logger.Info("New cookie for auth stored")

	cookie := resp.Header.Get("Set-Cookie")
	cookieValue := strings.Split(strings.Split(cookie, ";")[0], "=")[1]
	app.QBittorrent.Cookie = &cookieValue

	return nil
}
