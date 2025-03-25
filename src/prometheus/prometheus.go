package prom

import (
	"fmt"
	"math"
	API "qbit-exp/api"
	"qbit-exp/app"
	"qbit-exp/internal"
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
	labelName string = "name"

	labelDownloaded string = "downloaded"
	labelSeeders    string = "seeders"

	torrentLabelAddedOn                string = "added_on"
	torrentLabelAmountLeft             string = "amount_left"
	torrentLabelAmountLeftBytes        string = "amount_left_bytes"
	torrentLabelCategory               string = "category"
	torrentLabelComment                string = "comment"
	torrentLabelCompletionOn           string = "completed_on"
	torrentLabelConnectionStatus       string = "connection_status"
	torrentLabelDlSpeed                string = "dl_speed"
	torrentLabelDownloadedSession      string = "downloaded_session"
	torrentLabelDownloadSpeed          string = "download_speed_bytes"
	torrentLabelEta                    string = "eta"
	torrentLabelHash                   string = "hash"
	torrentLabelInfo                   string = "info"
	torrentLabelLeechers               string = "leechers"
	torrentLabelMaxRatio               string = "max_ratio"
	torrentLabelTag                    string = "tag"
	torrentLabelTags                   string = "tags"
	torrentLabelProgress               string = "progress"
	torrentLabelRatio                  string = "ratio"
	torrentLabelSavePath               string = "save_path"
	torrentLabelSeeders                string = "seeders"
	torrentLabelSessionDownloadedBytes string = "session_downloaded_bytes"
	torrentLabelSessionUploadedBytes   string = "session_uploaded_bytes"
	torrentLabelSize                   string = "size"
	torrentLabelSizeBytes              string = "size_bytes"
	torrentLabelState                  string = "state"
	torrentLabelStates                 string = "states"
	torrentLabelTimeActive             string = "time_active"
	torrentLabelTorrents               string = "torrents"
	torrentLabelTotalDownloaded        string = "total_downloaded_bytes"
	torrentLabelTotalUploaded          string = "total_uploaded_bytes"
	torrentLabelTracker                string = "tracker"
	torrentLabelTransfer               string = "transfer"
	torrentLabelUploaded               string = "uploaded"
	torrentLabelUploadedSession        string = "uploaded_session"
	torrentLabelUploadSpeed            string = "upload_speed_bytes"
	torrentLabelUpSpeed                string = "up_speed"

	trackerLabelURL     string = "url"
	trackerLabelMessage string = "message"
	trackerLabelLeeches string = "leeches"
	trackerLabelPeers   string = "peers"
	trackerLabelStatus  string = "status"
	trackerLabelTier    string = "tier"
)

const metricPrefix string = "qbittorrent"
const (
	metricNameTorrent string = "torrent"
	metricNameTracker string = "tracker"
	metricNameGlobal  string = "global"
	metricNameApp     string = "app"
)

const (
	stateError              string = "error"
	stateMissingFiles       string = "missingFiles"
	stateUploading          string = "uploading"
	stateQueuedUP           string = "queuedUP"
	stateStalledUP          string = "stalledUP"
	stateCheckingUP         string = "checkingUP"
	stateForcedUP           string = "forcedUP"
	stateAllocating         string = "allocating"
	stateDownloading        string = "downloading"
	stateMetaDL             string = "metaDL"
	stateQueuedDL           string = "queuedDL"
	stateStalledDL          string = "stalledDL"
	stateCheckingDL         string = "checkingDL"
	stateForcedDL           string = "forcedDL"
	stateCheckingResumeData string = "checkingResumeData"
	stateMoving             string = "moving"
	stateUnknown            string = "unknown"
)

// Reference the API changes
const (
	stateRenamed string = "2.11.0" // https://qbittorrent-api.readthedocs.io/en/latest/apidoc/definitions.html#qbittorrentapi.definitions.TorrentState
)

// Web API >= v2.11.0
const (
	stateStoppedUP string = "stoppedUP"
	stateStoppedDL string = "stoppedDL"
)

// Web API < v2.11.0
const (
	statePausedUP string = "pausedUP"
	statePausedDL string = "pausedDL"
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

func Torrent(result *API.Info, webUIVersion *string, r *prometheus.Registry) {

	var (
		torrentEta               = createMetricName(metricNameTorrent, torrentLabelEta)
		torrentDownloadSpeed     = createMetricName(metricNameTorrent, torrentLabelDownloadSpeed)
		torrentUploadSpeed       = createMetricName(metricNameTorrent, torrentLabelUploadSpeed)
		torrentProgress          = createMetricName(metricNameTorrent, torrentLabelProgress)
		torrentTimeActive        = createMetricName(metricNameTorrent, torrentLabelTimeActive)
		torrentSeeders           = createMetricName(metricNameTorrent, torrentLabelSeeders)
		torrentLeechers          = createMetricName(metricNameTorrent, torrentLabelLeechers)
		torrentRatio             = createMetricName(metricNameTorrent, torrentLabelRatio)
		torrentAmountLeft        = createMetricName(metricNameTorrent, torrentLabelAmountLeftBytes)
		torrentSize              = createMetricName(metricNameTorrent, torrentLabelSizeBytes)
		torrentSessionDownloaded = createMetricName(metricNameTorrent, torrentLabelSessionDownloadedBytes)
		torrentSessionUploaded   = createMetricName(metricNameTorrent, torrentLabelSessionUploadedBytes)
		torrentTotalDownloaded   = createMetricName(metricNameTorrent, torrentLabelTotalDownloaded)
		torrentTotalUploaded     = createMetricName(metricNameTorrent, torrentLabelTotalUploaded)
		torrentTags              = createMetricName(metricNameTorrent, torrentLabelTags)
		torrentStates            = createMetricName(metricNameTorrent, torrentLabelStates)
		torrentInfo              = createMetricName(metricNameTorrent, torrentLabelInfo)
		globalTorrents           = createMetricName(metricNameGlobal, torrentLabelTorrents)
		torrentComment           = createMetricName(metricNameTorrent, torrentLabelComment)
		torrentState             = createMetricName(metricNameTorrent, torrentLabelState)
		torrentSavePath          = createMetricName(metricNameTorrent, torrentLabelSavePath)
		torrentAddedOn           = createMetricName(metricNameTorrent, torrentLabelAddedOn)
		torrentCompletionOn      = createMetricName(metricNameTorrent, torrentLabelCompletionOn)
	)

	labels := []string{labelName}
	if app.Exporter.ExperimentalFeatures.EnableLabelWithHash {
		labels = append(labels, torrentLabelHash)
	}
	labelsWithTag := append([]string{}, labels...)
	labelsWithTag = append(labelsWithTag, torrentLabelTag)
	labelsWithComment := append([]string{}, labels...)
	labelsWithComment = append(labelsWithComment, torrentLabelComment)
	labelsWithState := append([]string{}, labels...)
	labelsWithState = append(labelsWithState, torrentLabelState)
	labelsWithSavePath := append([]string{}, labels...)
	labelsWithSavePath = append(labelsWithSavePath, torrentLabelSavePath)

	gauges := Gauge{
		{&torrentEta, &Seconds, "The current ETA for each torrent", &labels},
		{&torrentDownloadSpeed, &Bytes, "The current download speed of torrents", &labels},
		{&torrentUploadSpeed, &Bytes, "The current upload speed of torrents", &labels},
		{&torrentProgress, nil, "The current progress of torrents", &labels},
		{&torrentTimeActive, &Seconds, "The total active time", &labels},
		{&torrentSeeders, nil, "The current number of seeders for each torrent", &labels},
		{&torrentLeechers, nil, "The current number of leechers for each torrent", &labels},
		{&torrentRatio, nil, "The current ratio of each torrent", &labels},
		{&torrentAmountLeft, &Bytes, "The amount remaining for each torrent", &labels},
		{&torrentSize, &Bytes, "The size of each torrent", &labels},
		{&torrentSessionDownloaded, &Bytes, "The current session download amount of torrents", &labels},
		{&torrentSessionUploaded, &Bytes, "The current session upload amount of torrents", &labels},
		{&torrentTotalDownloaded, &Bytes, "The current total download amount of torrents", &labels},
		{&torrentTotalUploaded, &Bytes, "The current total upload amount of torrents", &labels},
		{&torrentTags, nil, "All tags associated to this torrent", &labelsWithTag},
		{&torrentStates, nil, "The current state of torrents", &[]string{labelName}},
		{&torrentComment, nil, "Comment added to this torrent", &labelsWithComment},
		{&torrentState, nil, "State this torrent", &labelsWithState},
		{&torrentSavePath, nil, "Save path for this torrent", &labelsWithSavePath},
		{&torrentAddedOn, nil, "Timestamp when this torrent was added", &labels},
		{&torrentCompletionOn, nil, "Timestamp when this torrent was completed", &labels},
	}

	metrics := registerGauge(&gauges, r)

	if app.Exporter.Features.EnableHighCardinality {
		torrentInfoLabels := []string{labelName, torrentLabelCategory, torrentLabelState, torrentLabelSize, torrentLabelProgress, labelSeeders, torrentLabelLeechers, torrentLabelDlSpeed, torrentLabelUpSpeed, torrentLabelAmountLeft, torrentLabelTimeActive, torrentLabelEta, torrentLabelUploaded, torrentLabelUploadedSession, labelDownloaded, torrentLabelDownloadedSession, torrentLabelMaxRatio, torrentLabelRatio, torrentLabelTracker, torrentLabelAddedOn, torrentLabelComment, torrentLabelCompletionOn, torrentLabelSavePath}
		if app.Exporter.ExperimentalFeatures.EnableLabelWithHash {
			torrentInfoLabels = append(torrentInfoLabels, torrentLabelHash)
		}
		metrics[torrentInfo] = newGaugeVec(torrentInfo, "All info for torrents",
			torrentInfoLabels)

		r.MustRegister(metrics[torrentInfo])
	}

	qbittorrentGlobalTorrents := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: globalTorrents,
		Help: "The total number of torrents",
	})
	r.MustRegister(qbittorrentGlobalTorrents)

	var countStates = map[string]float64{
		stateError:              0.0,
		stateMissingFiles:       0.0,
		stateUploading:          0.0,
		stateQueuedUP:           0.0,
		stateStalledUP:          0.0,
		stateCheckingUP:         0.0,
		stateForcedUP:           0.0,
		stateAllocating:         0.0,
		stateDownloading:        0.0,
		stateMetaDL:             0.0,
		stateQueuedDL:           0.0,
		stateStalledDL:          0.0,
		stateCheckingDL:         0.0,
		stateForcedDL:           0.0,
		stateCheckingResumeData: 0.0,
		stateMoving:             0.0,
		stateUnknown:            0.0,
	}

	if result := internal.CompareSemVer(*webUIVersion, stateRenamed); result == 1 || result == 0 {
		countStates[stateStoppedUP] = 0.0
		countStates[stateStoppedDL] = 0.0
	} else {
		countStates[statePausedUP] = 0.0
		countStates[statePausedDL] = 0.0
	}
	countTotal := 0.0
	for _, torrent := range *result {
		torrentLabels := prometheus.Labels{labelName: torrent.Name}
		if app.Exporter.ExperimentalFeatures.EnableLabelWithHash {
			torrentLabels[torrentLabelHash] = torrent.Hash
		}

		metrics[torrentEta].With(torrentLabels).Set(float64(torrent.Eta))
		metrics[torrentDownloadSpeed].With(torrentLabels).Set(float64(torrent.Dlspeed))
		metrics[torrentUploadSpeed].With(torrentLabels).Set(float64(torrent.Upspeed))
		metrics[torrentProgress].With(torrentLabels).Set(math.Round(float64(torrent.Progress)*10000) / 10000)
		metrics[torrentTimeActive].With(torrentLabels).Set(float64(torrent.TimeActive))
		metrics[torrentSeeders].With(torrentLabels).Set(float64(torrent.NumSeeds))
		metrics[torrentLeechers].With(torrentLabels).Set(float64(torrent.NumLeechs))
		metrics[torrentRatio].With(torrentLabels).Set(float64(torrent.Ratio))
		metrics[torrentAmountLeft].With(torrentLabels).Set(float64(torrent.AmountLeft))
		metrics[torrentSize].With(torrentLabels).Set(float64(torrent.Size))
		metrics[torrentSessionDownloaded].With(torrentLabels).Set(float64(torrent.DownloadedSession))
		metrics[torrentSessionUploaded].With(torrentLabels).Set(float64(torrent.UploadedSession))
		metrics[torrentTotalDownloaded].With(torrentLabels).Set(float64(torrent.Downloaded))
		metrics[torrentTotalUploaded].With(torrentLabels).Set(float64(torrent.Uploaded))
		metrics[torrentCompletionOn].With(torrentLabels).Set(float64(torrent.CompletionOn))
		metrics[torrentAddedOn].With(torrentLabels).Set(float64(torrent.AddedOn))

		if app.Exporter.Features.EnableIncreasedCardinality {
			tagComment := prometheus.Labels{labelName: torrent.Name, torrentLabelComment: torrent.Comment}
			tagState := prometheus.Labels{labelName: torrent.Name, torrentLabelState: torrent.State}
			tagSavePath := prometheus.Labels{labelName: torrent.Name, torrentLabelSavePath: torrent.SavePath}
			if app.Exporter.ExperimentalFeatures.EnableLabelWithHash {
				tagState[torrentLabelHash] = torrent.Hash
				tagSavePath[torrentLabelHash] = torrent.Hash
				tagComment[torrentLabelHash] = torrent.Hash
			}
			metrics[torrentState].With(tagState).Set(1.0)
			metrics[torrentSavePath].With(tagSavePath).Set(1.0)
			metrics[torrentComment].With(tagComment).Set(1.0)
		}

		_, exists := countStates[torrent.State]
		if exists {
			countStates[torrent.State]++
		} else {
			logger.Log.Error(fmt.Sprintf("Unknown state: %s", torrent.State))
		}
		countTotal++

		if app.Exporter.Features.EnableHighCardinality {
			infoLabels := prometheus.Labels{
				labelName:                     torrent.Name,
				torrentLabelCategory:          torrent.Category,
				torrentLabelState:             torrent.State,
				torrentLabelSize:              strconv.FormatInt(torrent.Size, 10),
				torrentLabelProgress:          strconv.FormatFloat(torrent.Progress, 'f', 4, 64),
				labelSeeders:                  strconv.FormatInt(torrent.NumSeeds, 10),
				torrentLabelLeechers:          strconv.FormatInt(torrent.NumLeechs, 10),
				torrentLabelDlSpeed:           strconv.FormatInt(torrent.Dlspeed, 10),
				torrentLabelUpSpeed:           strconv.FormatInt(torrent.Upspeed, 10),
				torrentLabelAmountLeft:        strconv.FormatInt(torrent.AmountLeft, 10),
				torrentLabelTimeActive:        strconv.FormatInt(torrent.TimeActive, 10),
				torrentLabelEta:               strconv.FormatInt(torrent.Eta, 10),
				torrentLabelUploaded:          strconv.FormatInt(torrent.Uploaded, 10),
				torrentLabelUploadedSession:   strconv.FormatInt(torrent.UploadedSession, 10),
				labelDownloaded:               strconv.FormatInt(torrent.Downloaded, 10),
				torrentLabelDownloadedSession: strconv.FormatInt(torrent.DownloadedSession, 10),
				torrentLabelMaxRatio:          strconv.FormatFloat(torrent.MaxRatio, 'f', 3, 64),
				torrentLabelRatio:             strconv.FormatFloat(torrent.Ratio, 'f', 3, 64),
				torrentLabelTracker:           torrent.Tracker,
				torrentLabelComment:           torrent.Comment,
				torrentLabelSavePath:          torrentSavePath,
				torrentLabelAddedOn:           torrentAddedOn,
				torrentLabelCompletionOn:      torrentCompletionOn,
			}
			if app.Exporter.ExperimentalFeatures.EnableLabelWithHash {
				infoLabels[torrentLabelHash] = torrent.Hash
			}
			metrics[torrentInfo].With(infoLabels).Set(1)
		}

		if torrent.Tags != "" {
			for _, tag := range strings.Split(torrent.Tags, ", ") {
				tagLabels := prometheus.Labels{labelName: torrent.Name, torrentLabelTag: tag}
				if app.Exporter.ExperimentalFeatures.EnableLabelWithHash {
					tagLabels[torrentLabelHash] = torrent.Hash
				}
				metrics[torrentTags].With(tagLabels).Set(1)
			}
		}
	}

	for state, count := range countStates {
		metrics[torrentStates].With(prometheus.Labels{labelName: state}).Set(count)
	}
	qbittorrentGlobalTorrents.Set(countTotal)
}

func Preference(result *API.Preferences, r *prometheus.Registry) {
	gauges := GaugeSet{
		{"max_active_downloads", nil, "The max number of downloads allowed", float64(result.MaxActiveDownloads)},
		{"max_active_uploads", nil, "The max number of active uploads allowed", float64(result.MaxActiveUploads)},
		{"max_active_torrents", nil, "The max number of active torrents allowed", float64(result.MaxActiveTorrents)},
		{"download_rate_limit", &Bytes, "The global download rate limit", float64(result.DlLimit)},
		{"upload_rate_limit", &Bytes, "The global upload rate limit", float64(result.UpLimit)},
		{"alt_download_rate_limit", &Bytes, "The alternate download rate limit", float64(result.AltDlLimit)},
		{"alt_upload_rate_limit", &Bytes, "The alternate upload rate limit", float64(result.AltUpLimit)},
	}

	registerGaugeGlobalAndSet(&gauges, r)
}

func Transfer(result *API.Transfer, r *prometheus.Registry) {
	gauges := GaugeSet{
		{"dht_nodes", nil, "The DHT nodes connected to", float64(result.DhtNodes)},
	}
	qbittorrentTransferConnectionStatus := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: createMetricName(torrentLabelTransfer, torrentLabelConnectionStatus),
		Help: "Connection status (connected, firewalled or disconnected)",
	}, []string{torrentLabelConnectionStatus})

	r.MustRegister(qbittorrentTransferConnectionStatus)
	qbittorrentTransferConnectionStatus.With(prometheus.Labels{
		torrentLabelConnectionStatus: result.ConnectionStatus,
	}).Set(1)

	registerGaugeGlobalAndSet(&gauges, r)

}

func Trackers(result []*API.Trackers, r *prometheus.Registry) {
	if len(result) == 0 {
		logger.Log.Trace("No tracker")
		return
	}

	labels := []string{trackerLabelURL}
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

	if app.Exporter.Features.EnableHighCardinality {
		metrics[QbittorrentTrackerInfo] = newGaugeVec(QbittorrentTrackerInfo, "All info for trackers",
			[]string{trackerLabelMessage, labelDownloaded, trackerLabelLeeches, trackerLabelPeers, labelSeeders, trackerLabelStatus, trackerLabelTier, trackerLabelURL})
	}

	for _, listOfTracker := range result {
		for _, tracker := range *listOfTracker {
			if internal.IsValidURL(tracker.URL) {
				tier, err := strconv.Atoi(string(tracker.Tier))
				if err != nil {
					tier = 0
				}
				labels := prometheus.Labels{trackerLabelURL: tracker.URL}
				metrics[QbittorrentTrackerDownloaded].With(labels).Set((float64(tracker.NumDownloaded)))
				metrics[QbittorrentTrackerLeeches].With(labels).Set((float64(tracker.NumLeeches)))
				metrics[QbittorrentTrackerSeeders].With(labels).Set((float64(tracker.NumSeeds)))
				metrics[QbittorrentTrackerPeers].With(labels).Set((float64(tracker.NumPeers)))
				metrics[QbittorrentTrackerStatus].With(labels).Set((float64(tracker.Status)))
				metrics[QbittorrentTrackerTier].With(labels).Set((float64(tier)))

				if app.Exporter.Features.EnableHighCardinality {
					qbittorrentTrackerInfoLabels := prometheus.Labels{
						trackerLabelMessage: tracker.Message,
						labelDownloaded:     strconv.Itoa(tracker.NumDownloaded),
						trackerLabelLeeches: strconv.Itoa(tracker.NumLeeches),
						trackerLabelPeers:   strconv.Itoa(tracker.NumPeers),
						labelSeeders:        strconv.Itoa(int(tracker.NumSeeds)),
						trackerLabelStatus:  strconv.Itoa((tracker.Status)),
						trackerLabelTier:    strconv.Itoa(tier),
						trackerLabelURL:     tracker.URL}
					metrics[QbittorrentTrackerInfo].With(qbittorrentTrackerInfoLabels).Set(1)
				}

			}
		}

	}

}

func MainData(result *API.MainData, r *prometheus.Registry) {
	globalratio, err := strconv.ParseFloat(result.ServerState.GlobalRatio, 64)

	if err != nil {
		logger.Log.Warn(fmt.Sprintf("error to convert ratio \"%s\"", result.ServerState.GlobalRatio))
	} else {
		qbittorrentGlobalRatio := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: createMetricName(metricNameGlobal, "ratio"),
			Help: "The current global ratio of all torrents",
		})
		r.MustRegister(qbittorrentGlobalRatio)
		qbittorrentGlobalRatio.Set(globalratio)

	}
	useAltSpeedLimits := 0.0
	if result.ServerState.UseAltSpeedLimits {
		useAltSpeedLimits = 1.0
	}
	qbittorrentAppAltRateLimitsEnabled := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: createMetricName(metricNameApp, "alt_rate_limits_enabled"),
		Help: "If alternate rate limits are enabled",
	})
	r.MustRegister(qbittorrentAppAltRateLimitsEnabled)
	qbittorrentAppAltRateLimitsEnabled.Set(float64(useAltSpeedLimits))

	gauges := GaugeSet{
		{"alltime_downloaded", &Bytes, "The all-time total download amount of torrents", float64(result.ServerState.AlltimeDl)},
		{"alltime_uploaded", &Bytes, "The all-time total upload amount of torrents", float64(result.ServerState.AlltimeUl)},
		{"session_downloaded", &Bytes, "The total download amount of torrents for this session", float64(result.ServerState.DlInfoData)},
		{"session_uploaded", &Bytes, "The total upload amount of torrents for this session", float64(result.ServerState.UpInfoData)},
		{"download_speed", &Bytes, "The current download speed of all torrents", float64(result.ServerState.DlInfoSpeed)},
		{"upload_speed", &Bytes, "The total current upload speed of all torrents", float64(result.ServerState.UpInfoSpeed)},
	}

	registerGaugeGlobalAndSet(&gauges, r)

	qbittorrentGlobalTags := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: createMetricName(metricNameGlobal, "tags"),
		Help: "All tags used in qbittorrent",
	}, []string{torrentLabelTag})
	r.MustRegister(qbittorrentGlobalTags)
	if len(result.Tags) > 0 {
		for _, tag := range result.Tags {
			labels := prometheus.Labels{
				torrentLabelTag: tag,
			}

			qbittorrentGlobalTags.With(labels).Set(1)
		}
	}
	qbittorrentGlobalCategories := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: createMetricName(metricNameGlobal, "categories"),
		Help: "All categories used in qbittorrent",
	}, []string{torrentLabelCategory})
	r.MustRegister(qbittorrentGlobalCategories)
	for _, category := range result.CategoryMap {
		labels := prometheus.Labels{
			torrentLabelCategory: category.Name,
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

func createMetricName(metricName string, metric string) string {
	return fmt.Sprintf("%s_%s_%s", metricPrefix, metricName, metric)
}
