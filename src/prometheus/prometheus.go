package prom

import (
	"net/url"
	API "qbit-exp/api"
	"qbit-exp/logger"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type Unit string

const (
	Bytes   Unit = "bytes"
	Seconds Unit = "seconds"
)

type Gauge []struct {
	name  string
	unit  Unit
	help  string
	value float64
}

func IsValidURL(input string) bool {
	u, err := url.Parse(input)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func Sendbackmessagetorrent(result *API.Info, r *prometheus.Registry) {

	qbittorrent_eta := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_torrent_eta",
		Help: "The current ETA for each torrent (in seconds)",
	}, []string{"name"})
	qbittorrent_torrent_download_speed_bytes := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_torrent_download_speed_bytes",
		Help: "The current download speed of torrents (in bytes)",
	}, []string{"name"})
	qbittorrent_torrent_upload_speed_bytes := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_torrent_upload_speed_bytes",
		Help: "The current upload speed of torrents (in bytes)",
	}, []string{"name"})
	qbittorrent_torrent_progress := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_torrent_progress",
		Help: "The current progress of torrents",
	}, []string{"name"})
	qbittorrent_torrent_time_active := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_torrent_time_active",
		Help: "The total active time (in seconds)",
	}, []string{"name"})
	qbittorrent_torrent_states := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_torrent_states",
		Help: "The current state of torrents",
	}, []string{"name"})
	qbittorrent_torrent_seeders := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_torrent_seeders",
		Help: "The current number of seeders for each torrent",
	}, []string{"name"})
	qbittorrent_torrent_leechers := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_torrent_leechers",
		Help: "The current number of leechers for each torrent",
	}, []string{"name"})
	qbittorrent_torrent_ratio := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_torrent_ratio",
		Help: "The current ratio each torrent",
	}, []string{"name"})
	qbittorrent_torrent_amount_left_bytes := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_torrent_amount_left_bytes",
		Help: "The amount remaining for each torrent (in bytes)",
	}, []string{"name"})
	qbittorrent_torrent_size_bytes := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_torrent_size_bytes",
		Help: "The size for each torrent (in bytes)",
	}, []string{"name"})
	qbittorrent_torrent_session_downloaded_bytes := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_torrent_session_downloaded_bytes",
		Help: "The current session download amount of torrents (in bytes)",
	}, []string{"name"})
	qbittorrent_torrent_session_uploaded_bytes := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_torrent_session_uploaded_bytes",
		Help: "The current session upload amount of torrents (in bytes)",
	}, []string{"name"})
	qbittorrent_torrent_total_downloaded_bytes := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_torrent_total_downloaded_bytes",
		Help: "The current total download amount of torrents (in bytes)",
	}, []string{"name"})
	qbittorrent_torrent_total_uploaded_bytes := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_torrent_total_uploaded_bytes",
		Help: "The current total upload amount of torrents (in bytes)",
	}, []string{"name"})
	qbittorrent_global_torrents := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "qbittorrent_global_torrents",
		Help: "The total number of torrents",
	})
	qbittorrent_torrent_info := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_torrent_info",
		Help: "All info for torrents",
	}, []string{"name", "category", "state", "size", "progress", "seeders", "leechers", "dl_speed", "up_speed", "amount_left", "time_active", "eta", "uploaded", "uploaded_session", "downloaded", "downloaded_session", "max_ratio", "ratio", "tracker"})
	qbittorrent_torrent_tags := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_tags",
		Help: "All tags associated to this torrent",
	}, []string{"name", "tag"})
	r.MustRegister(qbittorrent_eta)
	r.MustRegister(qbittorrent_torrent_download_speed_bytes)
	r.MustRegister(qbittorrent_torrent_upload_speed_bytes)
	r.MustRegister(qbittorrent_torrent_progress)
	r.MustRegister(qbittorrent_torrent_time_active)
	r.MustRegister(qbittorrent_torrent_states)
	r.MustRegister(qbittorrent_torrent_seeders)
	r.MustRegister(qbittorrent_torrent_leechers)
	r.MustRegister(qbittorrent_torrent_ratio)
	r.MustRegister(qbittorrent_torrent_amount_left_bytes)
	r.MustRegister(qbittorrent_torrent_size_bytes)
	r.MustRegister(qbittorrent_torrent_session_downloaded_bytes)
	r.MustRegister(qbittorrent_torrent_session_uploaded_bytes)
	r.MustRegister(qbittorrent_torrent_total_downloaded_bytes)
	r.MustRegister(qbittorrent_torrent_total_uploaded_bytes)
	r.MustRegister(qbittorrent_global_torrents)
	r.MustRegister(qbittorrent_torrent_info)
	r.MustRegister(qbittorrent_torrent_tags)

	count_stelledup := 0
	count_uploading := 0
	for _, torrent := range *result {
		qbittorrent_eta.With(prometheus.Labels{"name": torrent.Name}).Set(float64(torrent.Eta))
		qbittorrent_torrent_download_speed_bytes.With(prometheus.Labels{"name": torrent.Name}).Set(float64(torrent.Dlspeed))
		qbittorrent_torrent_upload_speed_bytes.With(prometheus.Labels{"name": torrent.Name}).Set(float64(torrent.Upspeed))
		qbittorrent_torrent_progress.With(prometheus.Labels{"name": torrent.Name}).Set(float64(torrent.Progress))
		qbittorrent_torrent_time_active.With(prometheus.Labels{"name": torrent.Name}).Set(float64(torrent.TimeActive))
		qbittorrent_torrent_seeders.With(prometheus.Labels{"name": torrent.Name}).Set(float64(torrent.NumSeeds))
		qbittorrent_torrent_leechers.With(prometheus.Labels{"name": torrent.Name}).Set(float64(torrent.NumLeechs))
		qbittorrent_torrent_ratio.With(prometheus.Labels{"name": torrent.Name}).Set(float64(torrent.Ratio))
		qbittorrent_torrent_amount_left_bytes.With(prometheus.Labels{"name": torrent.Name}).Set(float64(torrent.AmountLeft))
		qbittorrent_torrent_size_bytes.With(prometheus.Labels{"name": torrent.Name}).Set(float64(torrent.Size))
		qbittorrent_torrent_session_downloaded_bytes.With(prometheus.Labels{"name": torrent.Name}).Set(float64(torrent.DownloadedSession))
		qbittorrent_torrent_session_uploaded_bytes.With(prometheus.Labels{"name": torrent.Name}).Set(float64(torrent.UploadedSession))
		qbittorrent_torrent_total_downloaded_bytes.With(prometheus.Labels{"name": torrent.Name}).Set(float64(torrent.Downloaded))
		qbittorrent_torrent_total_uploaded_bytes.With(prometheus.Labels{"name": torrent.Name}).Set(float64(torrent.Uploaded))
		if torrent.State == "stalledUP" {
			count_stelledup += 1
		} else {
			count_uploading += 1
		}
		qbittorrent_torrent_info.With(prometheus.Labels{
			"name":               torrent.Name,
			"category":           torrent.Category,
			"state":              torrent.State,
			"size":               strconv.Itoa(torrent.Size),
			"progress":           strconv.Itoa(int(torrent.Progress)),
			"seeders":            strconv.Itoa((torrent.NumSeeds)),
			"leechers":           strconv.Itoa((torrent.NumLeechs)),
			"dl_speed":           strconv.Itoa((torrent.Dlspeed)),
			"up_speed":           strconv.Itoa((torrent.Upspeed)),
			"amount_left":        strconv.Itoa((torrent.AmountLeft)),
			"time_active":        strconv.Itoa((torrent.TimeActive)),
			"eta":                strconv.Itoa((torrent.Eta)),
			"uploaded":           strconv.Itoa((torrent.Uploaded)),
			"uploaded_session":   strconv.Itoa((torrent.UploadedSession)),
			"downloaded":         strconv.Itoa((torrent.Downloaded)),
			"downloaded_session": strconv.Itoa((torrent.DownloadedSession)),
			"max_ratio":          strconv.FormatFloat((torrent.MaxRatio), 'f', 3, 64),
			"ratio":              strconv.FormatFloat((torrent.Ratio), 'f', 3, 64),
			"tracker":            torrent.Tracker}).Set(1)
		if torrent.Tags != "" {
			separated_list := strings.Split(torrent.Tags, ", ")
			for j := 0; j < len(separated_list); j++ {
				labels := prometheus.Labels{
					"name": torrent.Name,
					"tag":  separated_list[j],
				}
				qbittorrent_torrent_tags.With(labels).Set(1)
			}

		}
	}

	qbittorrent_torrent_states.With(prometheus.Labels{"name": "stalledUP"}).Set(float64(count_stelledup))
	qbittorrent_torrent_states.With(prometheus.Labels{"name": "uploading"}).Set(float64(count_uploading))
	qbittorrent_global_torrents.Set(float64(count_stelledup + count_uploading))

}

func Sendbackmessagepreference(result *API.Preferences, r *prometheus.Registry) {
	gauges := Gauge{
		{"max active downloads", "", "The max number of downloads allowed", float64((*result).MaxActiveDownloads)},
		{"max active uploads", "", "The max number of active uploads allowed", float64((*result).MaxActiveDownloads)},
		{"max active torrents", "", "The max number of active torrents allowed", float64((*result).MaxActiveTorrents)},
		{"download rate limit", Bytes, "The global download rate limit", float64((*result).DlLimit)},
		{"upload rate limite", Bytes, "The global upload rate limit", float64((*result).UpLimit)},
		{"alt download rate limit", Bytes, "The alternate download rate limit", float64((*result).AltDlLimit)},
		{"alt upload rate limit", Bytes, "The alternate upload rate limit", float64((*result).AltUpLimit)},
	}

	register(gauges, r)

}

func Sendbackmessagetransfer(result *API.Transfer, r *prometheus.Registry) {
	gauges := Gauge{
		{"dht nodes", "", "The DHT nodes connected to", float64(result.DhtNodes)},
	}
	qbittorrent_tracker_info := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_transfer_connection_status",
		Help: "Connection status (connected, firewalled or disconnected)",
	}, []string{"connection_status"})

	r.MustRegister(qbittorrent_tracker_info)
	qbittorrent_tracker_info.With(prometheus.Labels{
		"connection_status": result.ConnectionStatus,
	}).Set(1)

	register(gauges, r)

}

func Sendbackmessagetrackers(result []*API.Trackers, r *prometheus.Registry) {
	if len(result) == 0 {
		logger.Log.Debug("No tracker")
		return
	}
	qbittorrent_tracker_info := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_tracker_info",
		Help: "All info for trackers",
	}, []string{"message", "downloaded", "leeches", "peers", "seeders", "status", "tier", "url"})

	r.MustRegister(qbittorrent_tracker_info)
	for _, listOfTracker := range result {
		for _, tracker := range *listOfTracker {
			if IsValidURL(tracker.URL) {
				qbittorrent_tracker_info.With(prometheus.Labels{
					"message":    tracker.Message,
					"downloaded": strconv.Itoa(tracker.NumDownloaded),
					"leeches":    strconv.Itoa(tracker.NumLeeches),
					"peers":      strconv.Itoa(tracker.NumPeers),
					"seeders":    strconv.Itoa(int(tracker.NumSeeds)),
					"status":     strconv.Itoa((tracker.Status)),
					"tier":       strconv.Itoa((tracker.Tier)),
					"url":        tracker.URL}).Set(1)
			}
		}

	}

}

func Sendbackmessagemaindata(result *API.Maindata, r *prometheus.Registry) {
	globalratio, err := strconv.ParseFloat((*result).ServerState.GlobalRatio, 64)

	if err != nil {
		logger.Log.Warn("error to convert ratio")
	} else {
		qbittorrent_global_ratio := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "qbittorrent_global_ratio",
			Help: "The current global ratio of all torrents",
		})
		r.MustRegister(qbittorrent_global_ratio)
		qbittorrent_global_ratio.Set(globalratio)

	}
	UseAltSpeedLimits := 0.0
	if (*result).ServerState.UseAltSpeedLimits {
		UseAltSpeedLimits = 1.0
	}
	qbittorrent_app_alt_rate_limits_enabled := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "qbittorrent_app_alt_rate_limits_enabled",
		Help: "If alternate rate limits are enabled",
	})
	r.MustRegister(qbittorrent_app_alt_rate_limits_enabled)
	qbittorrent_app_alt_rate_limits_enabled.Set(float64(UseAltSpeedLimits))

	gauges := Gauge{
		{"alltime downloaded", Bytes, "The all-time total download amount of torrents", float64((*result).ServerState.AlltimeDl)},
		{"alltime uploaded", Bytes, "The all-time total upload amount of torrents", float64((*result).ServerState.AlltimeUl)},
		{"session downloaded", Bytes, "The total download amount of torrents for this session", float64((*result).ServerState.DlInfoData)},
		{"session uploaded", Bytes, "The total upload amount of torrents for this session", float64((*result).ServerState.UpInfoData)},
		{"download speed", Bytes, "The current download speed of all torrents", float64((*result).ServerState.DlInfoSpeed)},
		{"upload speed", Bytes, "The total current upload speed of all torrents", float64((*result).ServerState.UpInfoSpeed)},
	}

	register(gauges, r)

	qbittorrent_global_tags := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_global_tags",
		Help: "All tags used in qbittorrent",
	}, []string{"tag"})
	r.MustRegister(qbittorrent_global_tags)
	if len((*result).Tags) > 0 {
		for _, tag := range result.Tags {
			labels := prometheus.Labels{
				"tag": tag,
			}

			qbittorrent_global_tags.With(labels).Set(1)
		}
	}
	qbittorrent_global_categories := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_global_categories",
		Help: "All categories used in qbittorrent",
	}, []string{"category"})
	r.MustRegister(qbittorrent_global_categories)
	for _, category := range result.CategoryMap {
		labels := prometheus.Labels{
			"category": category.Name,
		}
		qbittorrent_global_categories.With(labels).Set(1)
	}
}

func register(gauges Gauge, r *prometheus.Registry) {
	for _, gauge := range gauges {
		name := "qbittorrent_global_" + strings.Replace(gauge.name, " ", "_", -1)
		help := gauge.help
		if gauge.unit != "" {
			if gauge.unit == Bytes {
				name += "_" + string(gauge.unit)
			}
			help += " (in " + string(gauge.unit) + ")"
		}
		g := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: name,
			Help: help,
		})
		r.MustRegister(g)
		g.Set(gauge.value)
	}
}
