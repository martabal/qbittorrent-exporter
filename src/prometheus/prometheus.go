package prom

import (
	"net/url"
	API "qbit-exp/api"
	"qbit-exp/app"
	"qbit-exp/logger"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type Unit string

var (
	Bytes   Unit = "bytes"
	Seconds Unit = "seconds"
)

type Gauge []struct {
	name  string
	unit  *Unit
	help  string
	value float64
}

const TorrentLabelName = "name"
const TorrentLabelHash = "hash"

const TrackerLabelName = "url"

func isValidURL(input string) bool {
	u, err := url.Parse(input)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func Torrent(result *API.Info, r *prometheus.Registry) {
	labels := []string{TorrentLabelName}
	if app.ExperimentalFeature.EnableLabelWithHash {
		labels = []string{TorrentLabelName, TorrentLabelHash}
	}

	metrics := map[string]*prometheus.GaugeVec{
		"qbittorrent_torrent_eta":                      newGaugeVec("qbittorrent_torrent_eta", "The current ETA for each torrent (in seconds)", labels),
		"qbittorrent_torrent_download_speed_bytes":     newGaugeVec("qbittorrent_torrent_download_speed_bytes", "The current download speed of torrents (in bytes)", labels),
		"qbittorrent_torrent_upload_speed_bytes":       newGaugeVec("qbittorrent_torrent_upload_speed_bytes", "The current upload speed of torrents (in bytes)", labels),
		"qbittorrent_torrent_progress":                 newGaugeVec("qbittorrent_torrent_progress", "The current progress of torrents", labels),
		"qbittorrent_torrent_time_active":              newGaugeVec("qbittorrent_torrent_time_active", "The total active time (in seconds)", labels),
		"qbittorrent_torrent_seeders":                  newGaugeVec("qbittorrent_torrent_seeders", "The current number of seeders for each torrent", labels),
		"qbittorrent_torrent_leechers":                 newGaugeVec("qbittorrent_torrent_leechers", "The current number of leechers for each torrent", labels),
		"qbittorrent_torrent_ratio":                    newGaugeVec("qbittorrent_torrent_ratio", "The current ratio of each torrent", labels),
		"qbittorrent_torrent_amount_left_bytes":        newGaugeVec("qbittorrent_torrent_amount_left_bytes", "The amount remaining for each torrent (in bytes)", labels),
		"qbittorrent_torrent_size_bytes":               newGaugeVec("qbittorrent_torrent_size_bytes", "The size of each torrent (in bytes)", labels),
		"qbittorrent_torrent_session_downloaded_bytes": newGaugeVec("qbittorrent_torrent_session_downloaded_bytes", "The current session download amount of torrents (in bytes)", labels),
		"qbittorrent_torrent_session_uploaded_bytes":   newGaugeVec("qbittorrent_torrent_session_uploaded_bytes", "The current session upload amount of torrents (in bytes)", labels),
		"qbittorrent_torrent_total_downloaded_bytes":   newGaugeVec("qbittorrent_torrent_total_downloaded_bytes", "The current total download amount of torrents (in bytes)", labels),
		"qbittorrent_torrent_total_uploaded_bytes":     newGaugeVec("qbittorrent_torrent_total_uploaded_bytes", "The current total upload amount of torrents (in bytes)", labels),
		"qbittorrent_torrent_tags":                     newGaugeVec("qbittorrent_tags", "All tags associated to this torrent", labels),
		"qbittorrent_torrent_states":                   newGaugeVec("qbittorrent_torrent_states", "The current state of torrents", []string{TorrentLabelName}),
	}

	if app.Feature.EnableHighCardinality {
		metrics["qbittorrent_torrent_info"] = newGaugeVec("qbittorrent_torrent_info", "All info for torrents",
			[]string{TorrentLabelName, "category", "state", "size", "progress", "seeders", "leechers", "dl_speed", "up_speed", "amount_left", "time_active", "eta", "uploaded", "uploaded_session", "downloaded", "downloaded_session", "max_ratio", "ratio", "tracker", TorrentLabelHash})
	}

	qbittorrent_global_torrents := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "qbittorrent_global_torrents",
		Help: "The total number of torrents",
	})

	for _, metric := range metrics {
		r.MustRegister(metric)
	}
	r.MustRegister(qbittorrent_global_torrents)

	count_stelledup, count_uploading := 0, 0
	for _, torrent := range *result {
		labels := prometheus.Labels{TorrentLabelName: torrent.Name}
		if app.ExperimentalFeature.EnableLabelWithHash {
			labels = prometheus.Labels{TorrentLabelName: torrent.Name, TorrentLabelHash: torrent.Hash}
		}
		metrics["qbittorrent_torrent_eta"].With(labels).Set(float64(torrent.Eta))
		metrics["qbittorrent_torrent_download_speed_bytes"].With(labels).Set(float64(torrent.Dlspeed))
		metrics["qbittorrent_torrent_upload_speed_bytes"].With(labels).Set(float64(torrent.Upspeed))
		metrics["qbittorrent_torrent_progress"].With(labels).Set(float64(torrent.Progress))
		metrics["qbittorrent_torrent_time_active"].With(labels).Set(float64(torrent.TimeActive))
		metrics["qbittorrent_torrent_seeders"].With(labels).Set(float64(torrent.NumSeeds))
		metrics["qbittorrent_torrent_leechers"].With(labels).Set(float64(torrent.NumLeechs))
		metrics["qbittorrent_torrent_ratio"].With(labels).Set(float64(torrent.Ratio))
		metrics["qbittorrent_torrent_amount_left_bytes"].With(labels).Set(float64(torrent.AmountLeft))
		metrics["qbittorrent_torrent_size_bytes"].With(labels).Set(float64(torrent.Size))
		metrics["qbittorrent_torrent_session_downloaded_bytes"].With(labels).Set(float64(torrent.DownloadedSession))
		metrics["qbittorrent_torrent_session_uploaded_bytes"].With(labels).Set(float64(torrent.UploadedSession))
		metrics["qbittorrent_torrent_total_downloaded_bytes"].With(labels).Set(float64(torrent.Downloaded))
		metrics["qbittorrent_torrent_total_uploaded_bytes"].With(labels).Set(float64(torrent.Uploaded))
		if torrent.State == "stalledUP" {
			count_stelledup++
		} else {
			count_uploading++
		}
		infoLabels := prometheus.Labels{
			TorrentLabelName:     torrent.Name,
			"category":           torrent.Category,
			"state":              torrent.State,
			"size":               strconv.FormatInt(torrent.Size, 10),
			"progress":           strconv.Itoa(int(torrent.Progress)),
			"seeders":            strconv.FormatInt(torrent.NumSeeds, 10),
			"leechers":           strconv.FormatInt(torrent.NumLeechs, 10),
			"dl_speed":           strconv.FormatInt(torrent.Dlspeed, 10),
			"up_speed":           strconv.FormatInt(torrent.Upspeed, 10),
			"amount_left":        strconv.FormatInt(torrent.AmountLeft, 10),
			"time_active":        strconv.FormatInt(torrent.TimeActive, 10),
			"eta":                strconv.FormatInt(torrent.Eta, 10),
			"uploaded":           strconv.FormatInt(torrent.Uploaded, 10),
			"uploaded_session":   strconv.FormatInt(torrent.UploadedSession, 10),
			"downloaded":         strconv.FormatInt(torrent.Downloaded, 10),
			"downloaded_session": strconv.FormatInt(torrent.DownloadedSession, 10),
			"max_ratio":          strconv.FormatFloat(torrent.MaxRatio, 'f', 3, 64),
			"ratio":              strconv.FormatFloat(torrent.Ratio, 'f', 3, 64),
			"tracker":            torrent.Tracker,
			TorrentLabelHash:     torrent.Hash,
		}
		if app.Feature.EnableHighCardinality {
			metrics["qbittorrent_torrent_info"].With(infoLabels).Set(1)
		}

		if torrent.Tags != "" {
			tags := strings.Split(torrent.Tags, ", ")
			for _, tag := range tags {
				tagLabels := prometheus.Labels{
					TorrentLabelName: torrent.Name,
					TorrentLabelHash: torrent.Hash,
					"tag":            tag,
				}
				metrics["qbittorrent_torrent_tags"].With(tagLabels).Set(1)
			}
		}
	}

	metrics["qbittorrent_torrent_states"].With(prometheus.Labels{TorrentLabelName: "stalledUP"}).Set(float64(count_stelledup))
	metrics["qbittorrent_torrent_states"].With(prometheus.Labels{TorrentLabelName: "uploading"}).Set(float64(count_uploading))
	qbittorrent_global_torrents.Set(float64(count_stelledup + count_uploading))
}

func Preference(result *API.Preferences, r *prometheus.Registry) {
	gauges := Gauge{
		{"max active downloads", nil, "The max number of downloads allowed", float64((*result).MaxActiveDownloads)},
		{"max active uploads", nil, "The max number of active uploads allowed", float64((*result).MaxActiveDownloads)},
		{"max active torrents", nil, "The max number of active torrents allowed", float64((*result).MaxActiveTorrents)},
		{"download rate limit", &Bytes, "The global download rate limit", float64((*result).DlLimit)},
		{"upload rate limit", &Bytes, "The global upload rate limit", float64((*result).UpLimit)},
		{"alt download rate limit", &Bytes, "The alternate download rate limit", float64((*result).AltDlLimit)},
		{"alt upload rate limit", &Bytes, "The alternate upload rate limit", float64((*result).AltUpLimit)},
	}

	register_and_set(&gauges, r)

}

func Transfer(result *API.Transfer, r *prometheus.Registry) {
	gauges := Gauge{
		{"dht nodes", nil, "The DHT nodes connected to", float64(result.DhtNodes)},
	}
	qbittorrent_transfer_connection_status := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_transfer_connection_status",
		Help: "Connection status (connected, firewalled or disconnected)",
	}, []string{"connection_status"})

	r.MustRegister(qbittorrent_transfer_connection_status)
	qbittorrent_transfer_connection_status.With(prometheus.Labels{
		"connection_status": result.ConnectionStatus,
	}).Set(1)

	register_and_set(&gauges, r)

}

func Trackers(result []*API.Trackers, r *prometheus.Registry) {
	if len(result) == 0 {
		logger.Log.Trace("No tracker")
		return
	}

	label := []string{TrackerLabelName}

	metrics := map[string]*prometheus.GaugeVec{
		"qbittorrent_tracker_downloaded": newGaugeVec("qbittorrent_tracker_downloaded", "The current number of completed downloads for each trackers", label),
		"qbittorrent_tracker_leeches":    newGaugeVec("qbittorrent_tracker_leeches", "The current number of leechers for each trackers", label),
		"qbittorrent_tracker_peers":      newGaugeVec("qbittorrent_tracker_peers", "The current number of peers for each trackers", label),
		"qbittorrent_tracker_seeders":    newGaugeVec("qbittorrent_tracker_seeders", "The current number of seeders for each trackers", label),
		"qbittorrent_tracker_status":     newGaugeVec("qbittorrent_tracker_status", "The current status of each trackers", label),
		"qbittorrent_tracker_tier":       newGaugeVec("qbittorrent_tracker_tier", "The current tracker priority tier of each trackers", label),
	}

	if app.Feature.EnableHighCardinality {
		metrics["qbittorrent_tracker_info"] = newGaugeVec("qbittorrent_tracker_info", "All info for trackers",
			[]string{"message", "downloaded", "leeches", "peers", "seeders", "status", "tier", TrackerLabelName})
	}

	for _, metric := range metrics {
		r.MustRegister(metric)
	}
	for _, listOfTracker := range result {
		for _, tracker := range *listOfTracker {
			if isValidURL(tracker.URL) {
				tier, err := strconv.Atoi(string(tracker.Tier))
				if err != nil {
					tier = 0
				}
				label := prometheus.Labels{TrackerLabelName: tracker.URL}
				metrics["qbittorrent_tracker_downloaded"].With(label).Set((float64(tracker.NumDownloaded)))
				metrics["qbittorrent_tracker_leeches"].With(label).Set((float64(tracker.NumLeeches)))
				metrics["qbittorrent_tracker_seeders"].With(label).Set((float64(tracker.NumSeeds)))
				metrics["qbittorrent_tracker_peers"].With(label).Set((float64(tracker.NumPeers)))
				metrics["qbittorrent_tracker_status"].With(label).Set((float64(tracker.Status)))

				if app.Feature.EnableHighCardinality {
					qbittorrent_tracker_info_labels := prometheus.Labels{
						"message":        tracker.Message,
						"downloaded":     strconv.Itoa(tracker.NumDownloaded),
						"leeches":        strconv.Itoa(tracker.NumLeeches),
						"peers":          strconv.Itoa(tracker.NumPeers),
						"seeders":        strconv.Itoa(int(tracker.NumSeeds)),
						"status":         strconv.Itoa((tracker.Status)),
						"tier":           strconv.Itoa(tier),
						TrackerLabelName: tracker.URL}
					metrics["qbittorrent_tracker_info"].With(qbittorrent_tracker_info_labels).Set(1)
				}

			}
		}

	}

}

func MainData(result *API.MainData, r *prometheus.Registry) {
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
		{"alltime downloaded", &Bytes, "The all-time total download amount of torrents", float64((*result).ServerState.AlltimeDl)},
		{"alltime uploaded", &Bytes, "The all-time total upload amount of torrents", float64((*result).ServerState.AlltimeUl)},
		{"session downloaded", &Bytes, "The total download amount of torrents for this session", float64((*result).ServerState.DlInfoData)},
		{"session uploaded", &Bytes, "The total upload amount of torrents for this session", float64((*result).ServerState.UpInfoData)},
		{"download speed", &Bytes, "The current download speed of all torrents", float64((*result).ServerState.DlInfoSpeed)},
		{"upload speed", &Bytes, "The total current upload speed of all torrents", float64((*result).ServerState.UpInfoSpeed)},
	}

	register_and_set(&gauges, r)

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

func register_and_set(gauges *Gauge, r *prometheus.Registry) {
	for _, gauge := range *gauges {
		name := "qbittorrent_global_" + strings.Replace(gauge.name, " ", "_", -1)
		help := gauge.help
		if gauge.unit != nil {
			if gauge.unit == &Bytes {
				name += "_" + string(*gauge.unit)
			}
			help += " (in " + string(*gauge.unit) + ")"
		}
		g := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: name,
			Help: help,
		})
		r.MustRegister(g)
		g.Set(gauge.value)
	}
}

func newGaugeVec(name, help string, labels []string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}, labels)
}
