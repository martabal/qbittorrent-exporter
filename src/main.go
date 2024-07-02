package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"runtime"

	"qbit-exp/qbit"

	app "qbit-exp/app"
	logger "qbit-exp/logger"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	Version     = "dev"
	Author      = "martabal"
	ProjectName = "qbittorrent-exporter"
)

func main() {
	loadenv()
	fmt.Printf("%s (version %s)\n", ProjectName, Version)
	fmt.Println("Author:", Author)
	fmt.Println("Using log level: " + app.LogLevel)

	qbit.Auth()

	logger.Log.Info("qbittorrent URL: " + app.BaseUrl)
	logger.Log.Info("username: " + app.Username)
	logger.Log.Info("password: " + app.GetPasswordMasked())
	logger.Log.Info("Started")
	http.HandleFunc("/metrics", metrics)
	addr := ":" + strconv.Itoa(app.Port)
	if app.Port != app.DEFAULT_PORT {
		logger.Log.Info("Listening on port " + strconv.Itoa(app.Port))
	}
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}

func metrics(w http.ResponseWriter, req *http.Request) {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err == nil {
		logger.Log.Debug("New request from " + ip)
	} else {
		logger.Log.Debug("New request")
	}

	registry := prometheus.NewRegistry()
	err = qbit.AllRequests(registry)
	if err != nil {
		http.Error(w, "", http.StatusServiceUnavailable)
		runtime.GC()
	} else {
		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, req)
	}

}

func loadenv() {
	var envfile bool
	flag.BoolVar(&envfile, "e", false, "Use .env file")
	flag.Parse()
	_, err := os.Stat(".env")
	if !os.IsNotExist(err) && !envfile {
		err := godotenv.Load(".env")
		if err != nil {
			errormessage := "Error loading .env file:" + err.Error()
			panic(errormessage)
		}
	}

	loglevel := setLogLevel(getEnv(app.LogLevelEnv))
	qbitUsername := getEnv(app.UsernameEnv)
	qbitPassword := getEnv(app.PasswordEnv)
	qbitURL := strings.TrimSuffix(getEnv(app.BaseUrlEnv), "/")
	exporterPortEnv := getEnv(app.PortEnv)
	timeoutDurationEnv := getEnv(app.TimeoutEnv)
	disableTracker := getEnv(app.DisableTrackerEnv)

	exporterPort, errExporterPort := strconv.Atoi(exporterPortEnv)
	if errExporterPort != nil {
		panic(fmt.Sprintf("%s must be an integer", app.PortEnv.Key))
	}
	if exporterPort < 0 || exporterPort > 65353 {
		panic(fmt.Sprintf("%s must be > 0 and < 65353", app.PortEnv.Key))
	}

	timeoutDuration, errTimeoutDuration := strconv.Atoi(timeoutDurationEnv)
	if errTimeoutDuration != nil {
		panic(fmt.Sprintf("%s must be an integer", app.PortEnv.Key))
	}
	if timeoutDuration < 0 {
		panic(fmt.Sprintf("%s must be > 0", app.PortEnv.Key))
	}

	app.SetVar(exporterPort, strings.ToLower(disableTracker) == "true", loglevel, qbitURL, qbitUsername, qbitPassword, timeoutDuration)
}

func setLogLevel(logLevel string) string {
	upperLogLevel := strings.ToUpper(logLevel)
	level, found := logger.LogLevels[upperLogLevel]
	if !found {
		upperLogLevel = "INFO"
		level = logger.LogLevels[upperLogLevel]
	}

	opts := slog.HandlerOptions{
		Level: slog.Level(level),
	}

	handler := logger.NewPrettyHandler(os.Stdout, opts)
	logger.Log = slog.New(handler)
	return upperLogLevel
}

func getEnv(env app.Env) string {
	value, ok := os.LookupEnv(env.Key)
	if !ok || value == "" {
		if env.Help != "" {
			logger.Log.Warn(env.Help)
		}
		return env.DefaultValue
	}
	return value
}
