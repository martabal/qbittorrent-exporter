package prom

import (
	"fmt"
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
	name   *string
	unit   *Unit
	help   string
	labels *[]string
}

type GaugeSet []struct {
	name  string
	unit  *Unit
	help  string
	value float64
}

const (
	LabelName = "name"

	LabelDownloaded = "downloaded"
	LabelSeeders    = "seeders"

	TorrentLabelAmountLeft             = "amount_left"
	TorrentLabelAmountLeftBytes        = "amount_left_bytes"
	TorrentLabelCategory               = "category"
	TorrentLabelConnectionStatus       = "connection_status"
	TorrentLabelDlSpeed                = "dl_speed"
	TorrentLabelDownloadedSession      = "downloaded_session"
	TorrentLabelDownloadSpeed          = "download_speed_bytes"
	TorrentLabelEta                    = "eta"
	TorrentLabelHash                   = "hash"
	TorrentLabelInfo                   = "info"
	TorrentLabelLeechers               = "leechers"
	TorrentLabelMaxRatio               = "max_ratio"
	TorrentLabelTag                    = "tag"
	TorrentLabelTags                   = "tags"
	TorrentLabelProgress               = "progress"
	TorrentLabelRatio                  = "ratio"
	TorrentLabelSeeders                = "seeders"
	TorrentLabelSessionDownloadedBytes = "session_downloaded_bytes"
	TorrentLabelSessionUploadedBytes   = "session_uploaded_bytes"
	TorrentLabelSize                   = "size"
	TorrentLabelSizeBytes              = "size_bytes"
	TorrentLabelState                  = "state"
	TorrentLabelStates                 = "states"
	TorrentLabelTimeActive             = "time_active"
	TorrentLabelTorrents               = "torrents"
	TorrentLabelTotalDownloaded        = "total_downloaded_bytes"
	TorrentLabelTotalUploaded          = "total_uploaded_bytes"
	TorrentLabelTracker                = "tracker"
	TorrentLabelTransfer               = "transfer"
	TorrentLabelUploaded               = "uploaded"
	TorrentLabelUploadedSession        = "uploaded_session"
	TorrentLabelUploadSpeed            = "upload_speed_bytes"
	TorrentLabelUpSpeed                = "up_speed"

	TrackerLabelURL     = "url"
	TrackerLabelMessage = "message"
	TrackerLabelLeeches = "leeches"
	TrackerLabelPeers   = "peers"
	TrackerLabelStatus  = "status"
	TrackerLabelTier    = "tier"
)

const metricPrefix = "qbittorrent"
const (
	metricNameTorrent = "torrent"
	metricNameTracker = "tracker"
	metricNameGlobal  = "global"
	metricNameApp     = "app"
)

func Version(result *[]byte, r *prometheus.Registry) {
	qbittorrent_app_version := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: createMetricName(metricNameApp, "version"),
		Help: "The current qBittorrent version",
		ConstLabels: map[string]string{
			"version": string(*result),
		},
	})
	r.MustRegister(qbittorrent_app_version)
	qbittorrent_app_version.Set(1)
}

func Torrent(result *API.Info, r *prometheus.Registry) {

	var (
		TorrentEta               = createMetricName(metricNameTorrent, TorrentLabelEta)
		TorrentDownloadSpeed     = createMetricName(metricNameTorrent, TorrentLabelDownloadSpeed)
		TorrentUploadSpeed       = createMetricName(metricNameTorrent, TorrentLabelUploadSpeed)
		TorrentProgress          = createMetricName(metricNameTorrent, TorrentLabelProgress)
		TorrentTimeActive        = createMetricName(metricNameTorrent, TorrentLabelTimeActive)
		TorrentSeeders           = createMetricName(metricNameTorrent, TorrentLabelSeeders)
		TorrentLeechers          = createMetricName(metricNameTorrent, TorrentLabelLeechers)
		TorrentRatio             = createMetricName(metricNameTorrent, TorrentLabelRatio)
		TorrentAmountLeft        = createMetricName(metricNameTorrent, TorrentLabelAmountLeftBytes)
		TorrentSize              = createMetricName(metricNameTorrent, TorrentLabelSizeBytes)
		TorrentSessionDownloaded = createMetricName(metricNameTorrent, TorrentLabelSessionDownloadedBytes)
		TorrentSessionUploaded   = createMetricName(metricNameTorrent, TorrentLabelSessionUploadedBytes)
		TorrentTotalDownloaded   = createMetricName(metricNameTorrent, TorrentLabelTotalDownloaded)
		TorrentTotalUploaded     = createMetricName(metricNameTorrent, TorrentLabelTotalUploaded)
		TorrentTags              = createMetricName(metricNameTorrent, TorrentLabelTags)
		TorrentStates            = createMetricName(metricNameTorrent, TorrentLabelStates)
		TorrentInfo              = createMetricName(metricNameTorrent, TorrentLabelInfo)
		GlobalTorrents           = createMetricName(metricNameGlobal, TorrentLabelTorrents)
	)

	labels := []string{LabelName}
	if app.Exporter.ExperimentalFeature.EnableLabelWithHash {
		labels = append(labels, TorrentLabelHash)
	}
	labelsWithTag := append(labels, TorrentLabelTag)

	gauges := Gauge{
		{&TorrentEta, &Seconds, "The current ETA for each torrent", &labels},
		{&TorrentDownloadSpeed, &Bytes, "The current download speed of torrents", &labels},
		{&TorrentUploadSpeed, &Bytes, "The current upload speed of torrents", &labels},
		{&TorrentProgress, nil, "The current progress of torrents", &labels},
		{&TorrentTimeActive, &Seconds, "The total active time", &labels},
		{&TorrentSeeders, nil, "The current number of seeders for each torrent", &labels},
		{&TorrentLeechers, nil, "The current number of leechers for each torrent", &labels},
		{&TorrentRatio, nil, "The current ratio of each torrent", &labels},
		{&TorrentAmountLeft, &Bytes, "The amount remaining for each torrent", &labels},
		{&TorrentSize, &Bytes, "The size of each torrent", &labels},
		{&TorrentSessionDownloaded, &Bytes, "The current session download amount of torrents", &labels},
		{&TorrentSessionUploaded, &Bytes, "The current session upload amount of torrents", &labels},
		{&TorrentTotalDownloaded, &Bytes, "The current total download amount of torrents", &labels},
		{&TorrentTotalUploaded, &Bytes, "The current total upload amount of torrents", &labels},
		{&TorrentTags, nil, "All tags associated to this torrent", &labelsWithTag},
		{&TorrentStates, nil, "The current state of torrents", &[]string{LabelName}},
	}

	metrics := registerGauge(&gauges, r)

	if app.Exporter.Feature.EnableHighCardinality {
		torrentInfoLabels := []string{LabelName, TorrentLabelCategory, TorrentLabelState, TorrentLabelSize, TorrentLabelProgress, LabelSeeders, TorrentLabelLeechers, TorrentLabelDlSpeed, TorrentLabelUpSpeed, TorrentLabelAmountLeft, TorrentLabelTimeActive, TorrentLabelEta, TorrentLabelUploaded, TorrentLabelUploadedSession, LabelDownloaded, TorrentLabelDownloadedSession, TorrentLabelMaxRatio, TorrentLabelRatio, TorrentLabelTracker}
		if app.Exporter.ExperimentalFeature.EnableLabelWithHash {
			torrentInfoLabels = append(torrentInfoLabels, TorrentLabelHash)
		}
		metrics[TorrentInfo] = newGaugeVec(TorrentInfo, "All info for torrents",
			torrentInfoLabels)

		r.MustRegister(metrics[TorrentInfo])
	}

	qbittorrentGlobalTorrents := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: GlobalTorrents,
		Help: "The total number of torrents",
	})
	r.MustRegister(qbittorrentGlobalTorrents)

	countStalledUP, countUploading := 0, 0
	for _, torrent := range *result {
		torrentLabels := prometheus.Labels{LabelName: torrent.Name}
		if app.Exporter.ExperimentalFeature.EnableLabelWithHash {
			torrentLabels[TorrentLabelHash] = torrent.Hash
		}

		metrics[TorrentEta].With(torrentLabels).Set(float64(torrent.Eta))
		metrics[TorrentDownloadSpeed].With(torrentLabels).Set(float64(torrent.Dlspeed))
		metrics[TorrentUploadSpeed].With(torrentLabels).Set(float64(torrent.Upspeed))
		metrics[TorrentProgress].With(torrentLabels).Set(float64(torrent.Progress))
		metrics[TorrentTimeActive].With(torrentLabels).Set(float64(torrent.TimeActive))
		metrics[TorrentSeeders].With(torrentLabels).Set(float64(torrent.NumSeeds))
		metrics[TorrentLeechers].With(torrentLabels).Set(float64(torrent.NumLeechs))
		metrics[TorrentRatio].With(torrentLabels).Set(float64(torrent.Ratio))
		metrics[TorrentAmountLeft].With(torrentLabels).Set(float64(torrent.AmountLeft))
		metrics[TorrentSize].With(torrentLabels).Set(float64(torrent.Size))
		metrics[TorrentSessionDownloaded].With(torrentLabels).Set(float64(torrent.DownloadedSession))
		metrics[TorrentSessionUploaded].With(torrentLabels).Set(float64(torrent.UploadedSession))
		metrics[TorrentTotalDownloaded].With(torrentLabels).Set(float64(torrent.Downloaded))
		metrics[TorrentTotalUploaded].With(torrentLabels).Set(float64(torrent.Uploaded))

		if torrent.State == "stalledUP" {
			countStalledUP++
		} else {
			countUploading++
		}

		if app.Exporter.Feature.EnableHighCardinality {
			infoLabels := prometheus.Labels{
				LabelName:                     torrent.Name,
				TorrentLabelCategory:          torrent.Category,
				TorrentLabelState:             torrent.State,
				TorrentLabelSize:              strconv.FormatInt(torrent.Size, 10),
				TorrentLabelProgress:          strconv.Itoa(int(torrent.Progress)),
				LabelSeeders:                  strconv.FormatInt(torrent.NumSeeds, 10),
				TorrentLabelLeechers:          strconv.FormatInt(torrent.NumLeechs, 10),
				TorrentLabelDlSpeed:           strconv.FormatInt(torrent.Dlspeed, 10),
				TorrentLabelUpSpeed:           strconv.FormatInt(torrent.Upspeed, 10),
				TorrentLabelAmountLeft:        strconv.FormatInt(torrent.AmountLeft, 10),
				TorrentLabelTimeActive:        strconv.FormatInt(torrent.TimeActive, 10),
				TorrentLabelEta:               strconv.FormatInt(torrent.Eta, 10),
				TorrentLabelUploaded:          strconv.FormatInt(torrent.Uploaded, 10),
				TorrentLabelUploadedSession:   strconv.FormatInt(torrent.UploadedSession, 10),
				LabelDownloaded:               strconv.FormatInt(torrent.Downloaded, 10),
				TorrentLabelDownloadedSession: strconv.FormatInt(torrent.DownloadedSession, 10),
				TorrentLabelMaxRatio:          strconv.FormatFloat(torrent.MaxRatio, 'f', 3, 64),
				TorrentLabelRatio:             strconv.FormatFloat(torrent.Ratio, 'f', 3, 64),
				TorrentLabelTracker:           torrent.Tracker,
			}
			if app.Exporter.ExperimentalFeature.EnableLabelWithHash {
				infoLabels[TorrentLabelHash] = torrent.Hash
			}
			metrics[TorrentInfo].With(infoLabels).Set(1)
		}

		if torrent.Tags != "" {
			for _, tag := range strings.Split(torrent.Tags, ", ") {
				tagLabels := prometheus.Labels{LabelName: torrent.Name, TorrentLabelTag: tag}
				if app.Exporter.ExperimentalFeature.EnableLabelWithHash {
					tagLabels[TorrentLabelHash] = torrent.Hash
				}
				metrics[TorrentTags].With(tagLabels).Set(1)
			}
		}
	}

	metrics[TorrentStates].With(prometheus.Labels{LabelName: "stalledUP"}).Set(float64(countStalledUP))
	metrics[TorrentStates].With(prometheus.Labels{LabelName: "uploading"}).Set(float64(countUploading))
	qbittorrentGlobalTorrents.Set(float64(countStalledUP + countUploading))
}

func Preference(result *API.Preferences, r *prometheus.Registry) {
	gauges := GaugeSet{
		{"max_active_downloads", nil, "The max number of downloads allowed", float64((*result).MaxActiveDownloads)},
		{"max_active_uploads", nil, "The max number of active uploads allowed", float64((*result).MaxActiveDownloads)},
		{"max_active_torrents", nil, "The max number of active torrents allowed", float64((*result).MaxActiveTorrents)},
		{"download_rate_limit", &Bytes, "The global download rate limit", float64((*result).DlLimit)},
		{"upload_rate_limit", &Bytes, "The global upload rate limit", float64((*result).UpLimit)},
		{"alt_download_rate_limit", &Bytes, "The alternate download rate limit", float64((*result).AltDlLimit)},
		{"alt_upload_rate_limit", &Bytes, "The alternate upload rate limit", float64((*result).AltUpLimit)},
	}

	registerGaugeGlobalAndSet(&gauges, r)
}

func Transfer(result *API.Transfer, r *prometheus.Registry) {
	gauges := GaugeSet{
		{"dht_nodes", nil, "The DHT nodes connected to", float64(result.DhtNodes)},
	}
	qbittorrentTransferConnectionStatus := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: createMetricName(TorrentLabelTransfer, TorrentLabelConnectionStatus),
		Help: "Connection status (connected, firewalled or disconnected)",
	}, []string{TorrentLabelConnectionStatus})

	r.MustRegister(qbittorrentTransferConnectionStatus)
	qbittorrentTransferConnectionStatus.With(prometheus.Labels{
		TorrentLabelConnectionStatus: result.ConnectionStatus,
	}).Set(1)

	registerGaugeGlobalAndSet(&gauges, r)

}

func Trackers(result []*API.Trackers, r *prometheus.Registry) {
	if len(result) == 0 {
		logger.Log.Trace("No tracker")
		return
	}

	labels := []string{TrackerLabelURL}
	var (
		QbittorrentTrackerDownloaded = createMetricName(metricNameTracker, "downloaded")
		QbittorrentTrackerLeeches    = createMetricName(metricNameTracker, "leeches")
		QbittorrentTrackerPeers      = createMetricName(metricNameTracker, "peers")
		QbittorrentTrackerSeeders    = createMetricName(metricNameTracker, "seeders")
		QbittorrentTrackerStatus     = createMetricName(metricNameTracker, "status")
		QbittorrentTrackerTier       = createMetricName(metricNameTracker, "tier")
		QbittorrentTrackerInfo       = createMetricName(metricNameTracker, "info")
	)
	gauges := Gauge{
		{&QbittorrentTrackerDownloaded, nil, "The current number of completed downloads for each trackers", &labels},
		{&QbittorrentTrackerLeeches, nil, "The current number of leechers for each trackers", &labels},
		{&QbittorrentTrackerPeers, nil, "The current number of peers for each trackers", &labels},
		{&QbittorrentTrackerSeeders, nil, "The current number of seeders for each trackers", &labels},
		{&QbittorrentTrackerStatus, nil, "The current status of each trackers", &labels},
		{&QbittorrentTrackerTier, nil, "The current tracker priority tier of each trackers", &labels},
	}

	metrics := registerGauge(&gauges, r)

	if app.Exporter.Feature.EnableHighCardinality {
		metrics[QbittorrentTrackerInfo] = newGaugeVec(QbittorrentTrackerInfo, "All info for trackers",
			[]string{TrackerLabelMessage, LabelDownloaded, TrackerLabelLeeches, TrackerLabelPeers, LabelSeeders, TrackerLabelStatus, TrackerLabelTier, TrackerLabelURL})
	}

	for _, listOfTracker := range result {
		for _, tracker := range *listOfTracker {
			if isValidURL(tracker.URL) {
				tier, err := strconv.Atoi(string(tracker.Tier))
				if err != nil {
					tier = 0
				}
				labels := prometheus.Labels{TrackerLabelURL: tracker.URL}
				metrics[QbittorrentTrackerDownloaded].With(labels).Set((float64(tracker.NumDownloaded)))
				metrics[QbittorrentTrackerLeeches].With(labels).Set((float64(tracker.NumLeeches)))
				metrics[QbittorrentTrackerSeeders].With(labels).Set((float64(tracker.NumSeeds)))
				metrics[QbittorrentTrackerPeers].With(labels).Set((float64(tracker.NumPeers)))
				metrics[QbittorrentTrackerStatus].With(labels).Set((float64(tracker.Status)))

				if app.Exporter.Feature.EnableHighCardinality {
					qbittorrentTrackerInfoLabels := prometheus.Labels{
						TrackerLabelMessage: tracker.Message,
						LabelDownloaded:     strconv.Itoa(tracker.NumDownloaded),
						TrackerLabelLeeches: strconv.Itoa(tracker.NumLeeches),
						TrackerLabelPeers:   strconv.Itoa(tracker.NumPeers),
						LabelSeeders:        strconv.Itoa(int(tracker.NumSeeds)),
						TrackerLabelStatus:  strconv.Itoa((tracker.Status)),
						TrackerLabelTier:    strconv.Itoa(tier),
						TrackerLabelURL:     tracker.URL}
					metrics[QbittorrentTrackerInfo].With(qbittorrentTrackerInfoLabels).Set(1)
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
		qbittorrentGlobalRatio := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: createMetricName(metricNameGlobal, "ratio"),
			Help: "The current global ratio of all torrents",
		})
		r.MustRegister(qbittorrentGlobalRatio)
		qbittorrentGlobalRatio.Set(globalratio)

	}
	useAltSpeedLimits := 0.0
	if (*result).ServerState.UseAltSpeedLimits {
		useAltSpeedLimits = 1.0
	}
	qbittorrentAppAltRateLimitsEnabled := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: createMetricName(metricNameApp, "alt_rate_limits_enabled"),
		Help: "If alternate rate limits are enabled",
	})
	r.MustRegister(qbittorrentAppAltRateLimitsEnabled)
	qbittorrentAppAltRateLimitsEnabled.Set(float64(useAltSpeedLimits))

	gauges := GaugeSet{
		{"alltime_downloaded", &Bytes, "The all-time total download amount of torrents", float64((*result).ServerState.AlltimeDl)},
		{"alltime_uploaded", &Bytes, "The all-time total upload amount of torrents", float64((*result).ServerState.AlltimeUl)},
		{"session_downloaded", &Bytes, "The total download amount of torrents for this session", float64((*result).ServerState.DlInfoData)},
		{"session_uploaded", &Bytes, "The total upload amount of torrents for this session", float64((*result).ServerState.UpInfoData)},
		{"download_speed", &Bytes, "The current download speed of all torrents", float64((*result).ServerState.DlInfoSpeed)},
		{"upload_speed", &Bytes, "The total current upload speed of all torrents", float64((*result).ServerState.UpInfoSpeed)},
	}

	registerGaugeGlobalAndSet(&gauges, r)

	qbittorrentGlobalTags := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: createMetricName(metricNameGlobal, "tags"),
		Help: "All tags used in qbittorrent",
	}, []string{TorrentLabelTag})
	r.MustRegister(qbittorrentGlobalTags)
	if len((*result).Tags) > 0 {
		for _, tag := range result.Tags {
			labels := prometheus.Labels{
				TorrentLabelTag: tag,
			}

			qbittorrentGlobalTags.With(labels).Set(1)
		}
	}
	qbittorrentGlobalCategories := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: createMetricName(metricNameGlobal, "categories"),
		Help: "All categories used in qbittorrent",
	}, []string{TorrentLabelCategory})
	r.MustRegister(qbittorrentGlobalCategories)
	for _, category := range result.CategoryMap {
		labels := prometheus.Labels{
			TorrentLabelCategory: category.Name,
		}
		qbittorrentGlobalCategories.With(labels).Set(1)
	}
}

func createNameAndHelp(name *string, help *string, unit *Unit, changeName bool) {
	if unit != nil {
		if unit == &Bytes && changeName {
			*name += "_" + string(*unit)
		}
		*help += " (in " + string(*unit) + ")"
	}
}

func registerGaugeGlobalAndSet(gauges *GaugeSet, r *prometheus.Registry) {
	for _, gauge := range *gauges {
		gauge.name = createMetricName(metricNameGlobal, gauge.name)
		createNameAndHelp(&gauge.name, &gauge.help, gauge.unit, true)
		g := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: gauge.name,
			Help: gauge.help,
		})
		r.MustRegister(g)
		g.Set(gauge.value)
	}
}

func registerGauge(gauges *Gauge, r *prometheus.Registry) map[string]*prometheus.GaugeVec {
	metrics := make(map[string]*prometheus.GaugeVec)
	for _, gauge := range *gauges {
		createNameAndHelp(gauge.name, &gauge.help, gauge.unit, false)

		metrics[*gauge.name] = newGaugeVec(*gauge.name, gauge.help, *gauge.labels)
		r.MustRegister(metrics[*gauge.name])

	}
	return metrics
}

func newGaugeVec(name, help string, labels []string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}, labels)
}

func isValidURL(input string) bool {
	u, err := url.Parse(input)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func createMetricName(metricName string, metric string) string {
	return fmt.Sprintf("%s_%s_%s", metricPrefix, metricName, metric)
}
