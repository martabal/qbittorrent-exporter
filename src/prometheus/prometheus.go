package prom

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	API "qbit-exp/api"
	"qbit-exp/app"
	"qbit-exp/internal"
	"qbit-exp/logger"

	"github.com/prometheus/client_golang/prometheus"
)

type Gauge struct {
	Name   string
	Help   string
	Labels []string
}

const (
	BytesHelper   string = " (in " + torrentLabelBytes + ")"
	SecondsHelper string = " (in seconds)"
)

type GaugeList []Gauge

type GaugeSetting struct {
	Name  string
	Help  string
	Value float64
}

type GaugeSet []GaugeSetting

// Metric prefix
const metricPrefix string = "qbittorrent"

const separator string = "_"

// Metric categories
const (
	metricCatApp      string = metricPrefix + separator + metricNameApp + separator
	metricCatGlobal   string = metricPrefix + separator + metricNameGlobal + separator
	metricCatTorrent  string = metricPrefix + separator + metricNameTorrent + separator
	metricCatTracker  string = metricPrefix + separator + metricNameTracker + separator
	metricCatTransfer string = metricPrefix + separator + metricNameTransfer + separator
)

const (
	labelName string = "name"

	labelDownloaded string = "downloaded"
	labelSeeders    string = "seeders"

	torrentLabelAddedOn                string = "added_on"
	torrentLabelAllTimeHelper          string = "alltime"
	torrentLabelAllTimeDownloaded      string = torrentLabelAllTimeHelper + separator + torrentLabelDownload + separator + torrentLabelBytes
	torrentLabelAllTimeUploaded        string = torrentLabelAllTimeHelper + separator + torrentLabelUploaded + separator + torrentLabelBytes
	torrentLabelAltRatesLimitEnabled   string = "alt_rate_limits_enabled"
	torrentLabelAverageTimeQueue       string = "average_time_queue"
	torrentLabelAmountLeft             string = "amount_left"
	torrentLabelAmountLeftBytes        string = "amount_left" + separator + torrentLabelBytes
	torrentLabelBytes                  string = "bytes"
	torrentLabelCategories             string = "categories"
	torrentLabelCategory               string = "category"
	torrentLabelComment                string = "comment"
	torrentLabelCompletionOn           string = "completed_on"
	torrentLabelConnectionStatus       string = "connection_status"
	torrentLabelDHTNodes               string = "dht_nodes"
	torrentLabelDlSpeed                string = "dl_speed"
	torrentLabelDownload               string = "downloaded"
	torrentLabelDownloadedSession      string = "downloaded_session"
	torrentLabelDownloadSpeed          string = "download_speed" + separator + torrentLabelBytes
	torrentLabelEta                    string = "eta"
	torrentLabelFreeSpaceOnDisk        string = "free_space_on_disk" + torrentLabelBytes
	torrentLabelHash                   string = "hash"
	torrentLabelInfo                   string = "info"
	torrentLabelLeechers               string = "leechers"
	torrentLabelMaxRatio               string = "max_ratio"
	torrentLabelHelperMaxActive        string = "max_active"
	torrentLabelTag                    string = "tag"
	torrentLabelTags                   string = "tags"
	torrentLabelProgress               string = "progress"
	torrentLabelQueuedIoJobs           string = "queued_io_jobs"
	torrentLabelRatio                  string = "ratio"
	torrentLabelSavePath               string = "save_path"
	torrentLabelSeeders                string = "seeders"
	torrentLabelSessionHelper          string = "session"
	torrentLabelSessionDownloadedBytes string = torrentLabelSessionHelper + separator + torrentLabelDownload + separator + torrentLabelBytes
	torrentLabelSessionUploadedBytes   string = torrentLabelSessionHelper + separator + torrentLabelUploaded + separator + torrentLabelBytes
	torrentLabelSize                   string = "size"
	torrentLabelSizeBytes              string = "size" + separator + torrentLabelBytes
	torrentLabelState                  string = "state"
	torrentLabelStates                 string = "states"
	torrentLabelTimeActive             string = "time_active"
	torrentLabelTorrents               string = "torrents"
	torrentLabelTotalBuffersSize       string = "total_buffers_size" + separator + torrentLabelBytes
	torrentLabelTotalPeerConnections   string = "total_peer_connections"
	torrentLabelTotalQueuedSize        string = "total_queued_size" + separator + torrentLabelBytes
	torrentLabelTotalHelper            string = "total"
	torrentLabelTotalDownloaded        string = torrentLabelTotalHelper + separator + torrentLabelDownload + separator + torrentLabelBytes
	torrentLabelTotalUploaded          string = torrentLabelTotalHelper + separator + torrentLabelUploaded + separator + torrentLabelBytes
	torrentLabelTotalWastedSession     string = "total_wasted_session" + separator + torrentLabelBytes
	torrentLabelTracker                string = "tracker"
	torrentLabelUploaded               string = "uploaded"
	torrentLabelUploadedSession        string = "uploaded_session"
	torrentLabelUploadSpeed            string = "upload_speed" + separator + torrentLabelBytes
	torrentLabelUpSpeed                string = "up_speed"
	torrentLabelVersion                string = "version"

	trackerLabelURL     string = "url"
	trackerLabelMessage string = "message"
	trackerLabelLeeches string = "leeches"
	trackerLabelPeers   string = "peers"
	trackerLabelStatus  string = "status"
	trackerLabelTier    string = "tier"

	// Global-specific labels
	globalLabelDownloadRateLimit    string = "download_rate_limit" + separator + torrentLabelBytes
	globalLabelUploadRateLimit      string = "upload_rate_limit" + separator + torrentLabelBytes
	globalLabelAltDownloadRateLimit string = "alt_download_rate_limit" + separator + torrentLabelBytes
	globalLabelAltUploadRateLimit   string = "alt_upload_rate_limit" + separator + torrentLabelBytes
)

const (
	metricNameTorrent  string = "torrent"
	metricNameTracker  string = "tracker"
	metricNameGlobal   string = "global"
	metricNameApp      string = "app"
	metricNameTransfer string = "transfer"
)

var allStates = [...]string{
	"error",
	"missingFiles",
	"uploading",
	"queuedUP",
	"stalledUP",
	"checkingUP",
	"forcedUP",
	"allocating",
	"downloading",
	"metaDL",
	"queuedDL",
	"stalledDL",
	"checkingDL",
	"forcedDL",
	"checkingResumeData",
	"moving",
	"unknown",
}

// Reference the API changes
const (
	stateRenamedVersion string = "2.11.0" // https://qbittorrent-api.readthedocs.io/en/latest/apidoc/definitions.html#qbittorrentapi.definitions.TorrentState
)

const (
	// Web API >= v2.11.0
	stateStoppedUp string = "stoppedUP"
	stateStoppedDl string = "stoppedDL"

	// Web API < v2.11.0
	statePausedUp string = "pausedUP"
	statePausedDl string = "pausedDL"
)

// App
const (
	qbittorrentAppVersion                  string = metricCatApp + torrentLabelVersion
	qbittorrentAppAltRateLimitsEnabled     string = metricCatApp + torrentLabelAltRatesLimitEnabled
	helpQbittorrentAppVersion              string = "The current qBittorrent version"
	helpQbittorrentAppAltRateLimitsEnabled string = "If alternate rate limits are enabled"
)

// Global
const (
	qbittorrentGlobalTorrents                  string = metricCatGlobal + torrentLabelTorrents
	qbittorrentGlobalRatio                     string = metricCatGlobal + torrentLabelRatio
	qbittorrentGlobalTags                      string = metricCatGlobal + torrentLabelTags
	qbittorrentGlobalCategories                string = metricCatGlobal + torrentLabelCategories
	qbittorrentGlobalAlltimeDownloadedBytes    string = metricCatGlobal + torrentLabelAllTimeDownloaded
	qbittorrentGlobalAlltimeUploadedBytes      string = metricCatGlobal + torrentLabelAllTimeUploaded
	qbittorrentGlobalSessionDownloadedBytes    string = metricCatGlobal + torrentLabelSessionDownloadedBytes
	qbittorrentGlobalSessionUploadedBytes      string = metricCatGlobal + torrentLabelSessionUploadedBytes
	qbittorrentGlobalDownloadSpeedBytes        string = metricCatGlobal + torrentLabelDownloadSpeed
	qbittorrentGlobalUploadSpeedBytes          string = metricCatGlobal + torrentLabelUploadSpeed
	qbittorrentGlobalMaxActiveDownloads        string = metricCatGlobal + torrentLabelHelperMaxActive + "_downloads"
	qbittorrentGlobalMaxActiveUploads          string = metricCatGlobal + torrentLabelHelperMaxActive + "_uploads"
	qbittorrentGlobalMaxActiveTorrents         string = metricCatGlobal + torrentLabelHelperMaxActive + separator + torrentLabelTorrents
	qbittorrentGlobalDownloadRateLimitBytes    string = metricCatGlobal + globalLabelDownloadRateLimit
	qbittorrentGlobalUploadRateLimitBytes      string = metricCatGlobal + globalLabelUploadRateLimit
	qbittorrentGlobalAltDownloadRateLimitBytes string = metricCatGlobal + globalLabelAltDownloadRateLimit
	qbittorrentGlobalAltUploadRateLimitBytes   string = metricCatGlobal + globalLabelAltUploadRateLimit
	qbittorrentGlobalDHTNodes                  string = metricCatGlobal + torrentLabelDHTNodes
	qbittorrentGlobalAverageTimeQueue          string = metricCatGlobal + torrentLabelAverageTimeQueue
	qbittorrentGlobalFreeSpaceOnDiskBytes      string = metricCatGlobal + torrentLabelFreeSpaceOnDisk
	qbittorrentGlobalQueuedIoJobs              string = metricCatGlobal + torrentLabelQueuedIoJobs
	qbittorrentGlobalTotalBuffersSizeBytes     string = metricCatGlobal + torrentLabelTotalBuffersSize
	qbittorrentGlobalTotalQueuedSizeBytes      string = metricCatGlobal + torrentLabelTotalQueuedSize
	qbittorrentGlobalTotalPeerConnections      string = metricCatGlobal + torrentLabelTotalPeerConnections
	qbittorrentGlobalTotalWastedSessionBytes   string = metricCatGlobal + torrentLabelTotalWastedSession

	helpqbittorrentGlobalTorrents                  string = "The total number of torrents"
	helpqbittorrentGlobalRatio                     string = "The current global ratio of all torrents"
	helpqbittorrentGlobalTags                      string = "All tags used in qbittorrent"
	helpqbittorrentGlobalCategories                string = "All categories used in qbittorrent"
	helpqbittorrentGlobalAlltimeDownloadedBytes    string = "The all-time total download amount of torrents" + BytesHelper
	helpqbittorrentGlobalAlltimeUploadedBytes      string = "The all-time total upload amount of torrents" + BytesHelper
	helpqbittorrentGlobalSessionDownloadedBytes    string = "The total download amount of torrents for this session" + BytesHelper
	helpqbittorrentGlobalSessionUploadedBytes      string = "The total upload amount of torrents for this session" + BytesHelper
	helpqbittorrentGlobalDownloadSpeedBytes        string = "The current download speed of all torrents" + BytesHelper
	helpqbittorrentGlobalUploadSpeedBytes          string = "The total current upload speed of all torrents" + BytesHelper
	helpqbittorrentGlobalMaxActiveDownloads        string = "The max number of downloads allowed"
	helpqbittorrentGlobalMaxActiveUploads          string = "The max number of active uploads allowed"
	helpqbittorrentGlobalMaxActiveTorrents         string = "The max number of active torrents allowed"
	helpqbittorrentGlobalDownloadRateLimitBytes    string = "The global download rate limit" + BytesHelper
	helpqbittorrentGlobalUploadRateLimitBytes      string = "The global upload rate limit" + BytesHelper
	helpqbittorrentGlobalAltDownloadRateLimitBytes string = "The alternate download rate limit" + BytesHelper
	helpqbittorrentGlobalAltUploadRateLimitBytes   string = "The alternate upload rate limit" + BytesHelper
	helpqbittorrentGlobalDHTNodes                  string = "The DHT nodes connected to"
	helpqbittorrentGlobalAverageTimeQueue          string = "The DHT average time queue"
	helpqbittorrentGlobalFreeSpaceOnDisk           string = "The free space on disk" + BytesHelper
	helpqbittorrentGlobalQueuedIoJobs              string = "The DHT queued io jobs"
	helpqbittorrentGlobalTotalBuffersSize          string = "The total buffer size" + BytesHelper
	helpqbittorrentGlobalTotalQueuedSize           string = "The total queued size" + BytesHelper
	helpqbittorrentGlobalTotalPeerConnections      string = "The number of peer's connections"
	helpqbittorrentGlobalWastedSession             string = "The number of wasted session" + BytesHelper
)

// Torrent metrics
const (
	qbittorrentTorrentEta                    string = metricCatTorrent + torrentLabelEta
	qbittorrentTorrentDownloadSpeedBytes     string = metricCatTorrent + torrentLabelDownloadSpeed
	qbittorrentTorrentUploadSpeedBytes       string = metricCatTorrent + torrentLabelUploadSpeed
	qbittorrentTorrentProgress               string = metricCatTorrent + torrentLabelProgress
	qbittorrentTorrentTimeActive             string = metricCatTorrent + torrentLabelTimeActive
	qbittorrentTorrentSeeders                string = metricCatTorrent + torrentLabelSeeders
	qbittorrentTorrentLeechers               string = metricCatTorrent + torrentLabelLeechers
	qbittorrentTorrentRatio                  string = metricCatTorrent + torrentLabelRatio
	qbittorrentTorrentAmountLeftBytes        string = metricCatTorrent + torrentLabelAmountLeftBytes
	qbittorrentTorrentSizeBytes              string = metricCatTorrent + torrentLabelSizeBytes
	qbittorrentTorrentSessionDownloadedBytes string = metricCatTorrent + torrentLabelSessionDownloadedBytes
	qbittorrentTorrentSessionUploadedBytes   string = metricCatTorrent + torrentLabelSessionUploadedBytes
	qbittorrentTorrentTotalDownloadedBytes   string = metricCatTorrent + torrentLabelTotalDownloaded
	qbittorrentTorrentTotalUploadedBytes     string = metricCatTorrent + torrentLabelTotalUploaded
	qbittorrentTorrentTags                   string = metricCatTorrent + torrentLabelTags
	qbittorrentTorrentStates                 string = metricCatTorrent + torrentLabelStates
	qbittorrentTorrentInfo                   string = metricCatTorrent + torrentLabelInfo
	qbittorrentTorrentComment                string = metricCatTorrent + torrentLabelComment
	qbittorrentTorrentState                  string = metricCatTorrent + torrentLabelState
	qbittorrentTorrentSavePath               string = metricCatTorrent + torrentLabelSavePath
	qbittorrentTorrentAddedOn                string = metricCatTorrent + torrentLabelAddedOn
	qbittorrentTorrentCompletionOn           string = metricCatTorrent + torrentLabelCompletionOn

	helpQbittorrentTorrentEta                    string = "The current ETA for each torrent" + SecondsHelper
	helpQbittorrentTorrentDownloadSpeedBytes     string = "The current download speed of torrents" + BytesHelper
	helpQbittorrentTorrentUploadSpeedBytes       string = "The current upload speed of torrents" + BytesHelper
	helpQbittorrentTorrentProgress               string = "The current progress of torrents"
	helpQbittorrentTorrentTimeActive             string = "The total active time" + SecondsHelper
	helpQbittorrentTorrentSeeders                string = "The current number of seeders for each torrent"
	helpQbittorrentTorrentLeechers               string = "The current number of leechers for each torrent"
	helpQbittorrentTorrentRatio                  string = "The current ratio of each torrent"
	helpQbittorrentTorrentAmountLeftBytes        string = "The amount remaining for each torrent" + BytesHelper
	helpQbittorrentTorrentSizeBytes              string = "The size of each torrent" + BytesHelper
	helpQbittorrentTorrentSessionDownloadedBytes string = "The current session download amount of torrents" + BytesHelper
	helpQbittorrentTorrentSessionUploadedBytes   string = "The current session upload amount of torrents" + BytesHelper
	helpQbittorrentTorrentTotalDownloadedBytes   string = "The current total download amount of torrents" + BytesHelper
	helpQbittorrentTorrentTotalUploadedBytes     string = "The current total upload amount of torrents" + BytesHelper
	helpQbittorrentTorrentTags                   string = "All tags associated to this torrent"
	helpQbittorrentTorrentStates                 string = "The current state of torrents"
	helpQbittorrentTorrentInfo                   string = "All info for torrents"
	helpQbittorrentTorrentComment                string = "Comment added to this torrent"
	helpQbittorrentTorrentState                  string = "The state of this torrent"
	helpQbittorrentTorrentSavePath               string = "Save path for this torrent"
	helpQbittorrentTorrentAddedOn                string = "Timestamp when this torrent was added"
	helpQbittorrentTorrentCompletionOn           string = "Timestamp when this torrent was completed"

	qbittorrentTorrentTransferConnectionStatus     string = metricCatTransfer + torrentLabelConnectionStatus
	helpQbittorrentTorrentTransferConnectionStatus string = "Connection status (connected, firewalled or disconnected)"
)

// Trackers
const (
	qbittorrentTrackerDownloaded string = metricCatTracker + labelDownloaded
	qbittorrentTrackerLeeches    string = metricCatTracker + trackerLabelLeeches
	qbittorrentTrackerPeers      string = metricCatTracker + trackerLabelPeers
	qbittorrentTrackerSeeders    string = metricCatTracker + labelSeeders
	qbittorrentTrackerStatus     string = metricCatTracker + trackerLabelStatus
	qbittorrentTrackerTier       string = metricCatTracker + trackerLabelTier
	qbittorrentTrackerInfo       string = metricCatTracker + torrentLabelInfo

	helpqbittorrentTrackerDownloaded string = "The current number of completed downloads for each tracker"
	helpqbittorrentTrackerLeeches    string = "The current number of leechers for each tracker"
	helpqbittorrentTrackerPeers      string = "The current number of peers for each tracker"
	helpqbittorrentTrackerSeeders    string = "The current number of seeders for each tracker"
	helpqbittorrentTrackerStatus     string = "The current status of each tracker"
	helpqbittorrentTrackerTier       string = "The current tracker priority tier of each tracker"
	helpqbittorrentTrackerInfo       string = "All info for trackers"
)

func Version(result *[]byte, r *prometheus.Registry) {
	qbittorrentAppVersion := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: qbittorrentAppVersion,
		Help: helpQbittorrentAppVersion,
		ConstLabels: map[string]string{
			"version": string(*result),
		},
	})
	r.MustRegister(qbittorrentAppVersion)
	qbittorrentAppVersion.Set(1)
}

func createTorrentInfoLabels(enableHighCardinality, enableLabelWithHash bool, enableLabelWithTags bool) []string {
	baseLabels := []string{
		labelName, torrentLabelCategory, torrentLabelState, torrentLabelSize, torrentLabelProgress,
		labelSeeders, torrentLabelLeechers, torrentLabelDlSpeed, torrentLabelUpSpeed, torrentLabelAmountLeft,
		torrentLabelTimeActive, torrentLabelEta, torrentLabelUploaded, torrentLabelUploadedSession,
		labelDownloaded, torrentLabelDownloadedSession, torrentLabelMaxRatio, torrentLabelRatio,
		torrentLabelTracker, torrentLabelAddedOn, torrentLabelComment, torrentLabelCompletionOn,
		torrentLabelSavePath,
	}

	if !enableHighCardinality {
		baseLabels = []string{
			torrentLabelAddedOn, torrentLabelCategory, torrentLabelComment,
			torrentLabelCompletionOn, labelName, torrentLabelSavePath,
			torrentLabelState, torrentLabelTracker,
		}
	}

	if enableLabelWithHash {
		baseLabels = append(baseLabels, torrentLabelHash)
	}

	if enableLabelWithTags {
		baseLabels = append(baseLabels, torrentLabelTags)
	}

	return baseLabels
}

func createTorrentLabels(torrent API.Info, enableHighCardinality, enableLabelWithHash bool, enableLabelWithTags bool) prometheus.Labels {
	infoLabels := prometheus.Labels{
		labelName:                torrent.Name,
		torrentLabelCategory:     torrent.Category,
		torrentLabelState:        torrent.State,
		torrentLabelTracker:      torrent.Tracker,
		torrentLabelComment:      torrent.Comment,
		torrentLabelSavePath:     torrent.SavePath,
		torrentLabelAddedOn:      strconv.Itoa(int(torrent.AddedOn)),
		torrentLabelCompletionOn: strconv.Itoa(int(torrent.CompletionOn)),
	}

	if enableHighCardinality {
		infoLabels[torrentLabelSize] = strconv.FormatInt(torrent.Size, 10)
		infoLabels[torrentLabelProgress] = strconv.FormatFloat(torrent.Progress, 'f', 4, 64)
		infoLabels[labelSeeders] = strconv.FormatInt(torrent.NumSeeds, 10)
		infoLabels[torrentLabelLeechers] = strconv.FormatInt(torrent.NumLeechs, 10)
		infoLabels[torrentLabelDlSpeed] = strconv.FormatInt(torrent.Dlspeed, 10)
		infoLabels[torrentLabelUpSpeed] = strconv.FormatInt(torrent.Upspeed, 10)
		infoLabels[torrentLabelAmountLeft] = strconv.FormatInt(torrent.AmountLeft, 10)
		infoLabels[torrentLabelTimeActive] = strconv.FormatInt(torrent.TimeActive, 10)
		infoLabels[torrentLabelEta] = strconv.FormatInt(torrent.Eta, 10)
		infoLabels[torrentLabelUploaded] = strconv.FormatInt(torrent.Uploaded, 10)
		infoLabels[torrentLabelUploadedSession] = strconv.FormatInt(torrent.UploadedSession, 10)
		infoLabels[labelDownloaded] = strconv.FormatInt(torrent.Downloaded, 10)
		infoLabels[torrentLabelDownloadedSession] = strconv.FormatInt(torrent.DownloadedSession, 10)
		infoLabels[torrentLabelMaxRatio] = strconv.FormatFloat(torrent.MaxRatio, 'f', 3, 64)
		infoLabels[torrentLabelRatio] = strconv.FormatFloat(torrent.Ratio, 'f', 3, 64)
	}

	if enableLabelWithHash {
		infoLabels[torrentLabelHash] = torrent.Hash
	}

	if enableLabelWithTags {
		infoLabels[torrentLabelTags] = torrent.Tags
	}

	return infoLabels
}

func Torrent(result *API.SliceInfo, webUIVersion *string, r *prometheus.Registry) {
	labels := []string{labelName}
	if app.Exporter.ExperimentalFeatures.EnableLabelWithHash {
		labels = append(labels, torrentLabelHash)
	}
	if app.Exporter.ExperimentalFeatures.EnableLabelWithTracker {
		labels = append(labels, torrentLabelTracker)
	}
	labelsWithTag := append(append([]string{}, labels...), torrentLabelTag)
	labelsWithComment := append(append([]string{}, labels...), torrentLabelComment)
	labelsWithState := append(append([]string{}, labels...), torrentLabelState)
	labelsWithSavePath := append(append([]string{}, labels...), torrentLabelSavePath)

	gauges := GaugeList{
		{qbittorrentTorrentEta, helpQbittorrentTorrentEta, labels},
		{qbittorrentTorrentDownloadSpeedBytes, helpQbittorrentTorrentDownloadSpeedBytes, labels},
		{qbittorrentTorrentUploadSpeedBytes, helpQbittorrentTorrentUploadSpeedBytes, labels},
		{qbittorrentTorrentProgress, helpQbittorrentTorrentProgress, labels},
		{qbittorrentTorrentTimeActive, helpQbittorrentTorrentTimeActive, labels},
		{qbittorrentTorrentSeeders, helpQbittorrentTorrentSeeders, labels},
		{qbittorrentTorrentLeechers, helpQbittorrentTorrentLeechers, labels},
		{qbittorrentTorrentRatio, helpQbittorrentTorrentRatio, labels},
		{qbittorrentTorrentAmountLeftBytes, helpQbittorrentTorrentAmountLeftBytes, labels},
		{qbittorrentTorrentSizeBytes, helpQbittorrentTorrentSizeBytes, labels},
		{qbittorrentTorrentSessionDownloadedBytes, helpQbittorrentTorrentSessionDownloadedBytes, labels},
		{qbittorrentTorrentSessionUploadedBytes, helpQbittorrentTorrentSessionUploadedBytes, labels},
		{qbittorrentTorrentTotalDownloadedBytes, helpQbittorrentTorrentTotalDownloadedBytes, labels},
		{qbittorrentTorrentTotalUploadedBytes, helpQbittorrentTorrentTotalUploadedBytes, labels},
		{qbittorrentTorrentTags, helpQbittorrentTorrentTags, labelsWithTag},
		{qbittorrentTorrentStates, helpQbittorrentTorrentStates, []string{labelName}},
		{qbittorrentTorrentComment, helpQbittorrentTorrentComment, labelsWithComment},
		{qbittorrentTorrentState, helpQbittorrentTorrentState, labelsWithState},
		{qbittorrentTorrentSavePath, helpQbittorrentTorrentSavePath, labelsWithSavePath},
		{qbittorrentTorrentAddedOn, helpQbittorrentTorrentAddedOn, labels},
		{qbittorrentTorrentCompletionOn, helpQbittorrentTorrentCompletionOn, labels},
	}

	metrics := registerGauge(&gauges, r)

	enableLabelWithHash := app.Exporter.ExperimentalFeatures.EnableLabelWithHash
	enableLabelWithTags := app.Exporter.ExperimentalFeatures.EnableLabelWithTags

	var torrentInfoLabels []string
	if app.Exporter.Features.EnableHighCardinality {
		torrentInfoLabels = createTorrentInfoLabels(true, enableLabelWithHash, enableLabelWithTags)
	} else if app.Exporter.Features.EnableIncreasedCardinality {
		torrentInfoLabels = createTorrentInfoLabels(false, enableLabelWithHash, enableLabelWithTags)
	}

	if len(torrentInfoLabels) > 0 {
		metrics[qbittorrentTorrentInfo] = newGaugeVec(qbittorrentTorrentInfo, helpQbittorrentTorrentInfo, torrentInfoLabels)
		r.MustRegister(metrics[qbittorrentTorrentInfo])
	}

	qbittorrentGlobalTorrents := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: qbittorrentGlobalTorrents,
		Help: helpqbittorrentGlobalTorrents,
	})
	r.MustRegister(qbittorrentGlobalTorrents)

	var countStates = make(map[string]float64, len(allStates)+2)
	for _, state := range allStates {
		countStates[state] = 0.0
	}

	if result := internal.CompareSemVer(*webUIVersion, stateRenamedVersion); result == 1 || result == 0 {
		countStates[stateStoppedUp] = 0.0
		countStates[stateStoppedDl] = 0.0
	} else {
		countStates[statePausedUp] = 0.0
		countStates[statePausedDl] = 0.0
	}

	countTotal := 0.0

	baseTorrentLabels := func(t API.Info) prometheus.Labels {
		l := prometheus.Labels{
			labelName: t.Name,
		}
		if app.Exporter.ExperimentalFeatures.EnableLabelWithHash {
			l[torrentLabelHash] = t.Hash
		}
		if app.Exporter.ExperimentalFeatures.EnableLabelWithTracker {
			l[torrentLabelTracker] = t.Tracker
		}
		return l
	}

	for _, torrent := range *result {
		torrentLabels := baseTorrentLabels(torrent)

		metrics[qbittorrentTorrentEta].With(torrentLabels).Set(float64(torrent.Eta))
		metrics[qbittorrentTorrentDownloadSpeedBytes].With(torrentLabels).Set(float64(torrent.Dlspeed))
		metrics[qbittorrentTorrentUploadSpeedBytes].With(torrentLabels).Set(float64(torrent.Upspeed))
		metrics[qbittorrentTorrentProgress].With(torrentLabels).Set(math.Round(float64(torrent.Progress)*10000) / 10000)
		metrics[qbittorrentTorrentTimeActive].With(torrentLabels).Set(float64(torrent.TimeActive))
		metrics[qbittorrentTorrentSeeders].With(torrentLabels).Set(float64(torrent.NumSeeds))
		metrics[qbittorrentTorrentLeechers].With(torrentLabels).Set(float64(torrent.NumLeechs))
		metrics[qbittorrentTorrentRatio].With(torrentLabels).Set(float64(torrent.Ratio))
		metrics[qbittorrentTorrentAmountLeftBytes].With(torrentLabels).Set(float64(torrent.AmountLeft))
		metrics[qbittorrentTorrentSizeBytes].With(torrentLabels).Set(float64(torrent.Size))
		metrics[qbittorrentTorrentSessionDownloadedBytes].With(torrentLabels).Set(float64(torrent.DownloadedSession))
		metrics[qbittorrentTorrentSessionUploadedBytes].With(torrentLabels).Set(float64(torrent.UploadedSession))
		metrics[qbittorrentTorrentTotalDownloadedBytes].With(torrentLabels).Set(float64(torrent.Downloaded))
		metrics[qbittorrentTorrentTotalUploadedBytes].With(torrentLabels).Set(float64(torrent.Uploaded))
		metrics[qbittorrentTorrentCompletionOn].With(torrentLabels).Set(float64(torrent.CompletionOn))
		metrics[qbittorrentTorrentAddedOn].With(torrentLabels).Set(float64(torrent.AddedOn))

		if app.Exporter.Features.EnableIncreasedCardinality {
			tagComment := baseTorrentLabels(torrent)
			tagComment[torrentLabelComment] = torrent.Comment

			tagSavePath := baseTorrentLabels(torrent)
			tagSavePath[torrentLabelSavePath] = torrent.SavePath

			metrics[qbittorrentTorrentSavePath].With(tagSavePath).Set(1.0)
			metrics[qbittorrentTorrentComment].With(tagComment).Set(1.0)

			for _, state := range allStates {
				tagState := baseTorrentLabels(torrent)
				tagState[torrentLabelState] = state
				value := 0.0
				if torrent.State == state {
					value = 1.0
				}
				metrics[qbittorrentTorrentState].With(tagState).Set(value)
			}
		}

		if _, exists := countStates[torrent.State]; exists {
			countStates[torrent.State]++
		} else {
			logger.Error(fmt.Sprintf("Unknown state: %s", torrent.State))
		}
		countTotal++

		var infoLabels prometheus.Labels
		if app.Exporter.Features.EnableHighCardinality {
			infoLabels = createTorrentLabels(torrent, true, enableLabelWithHash, enableLabelWithTags)
		} else if app.Exporter.Features.EnableIncreasedCardinality {
			infoLabels = createTorrentLabels(torrent, false, enableLabelWithHash, enableLabelWithTags)
		}

		if len(infoLabels) > 0 {
			metrics[qbittorrentTorrentInfo].With(infoLabels).Set(1)
		}

		if torrent.Tags != "" {
			for _, tag := range strings.Split(torrent.Tags, ", ") {
				tagLabels := baseTorrentLabels(torrent)
				tagLabels[torrentLabelTag] = tag
				metrics[qbittorrentTorrentTags].With(tagLabels).Set(1)
			}
		}
	}

	for state, count := range countStates {
		metrics[qbittorrentTorrentStates].With(prometheus.Labels{labelName: state}).Set(count)
	}
	qbittorrentGlobalTorrents.Set(countTotal)
}

func Preference(result *API.Preferences, r *prometheus.Registry) {
	gauges := GaugeSet{
		{qbittorrentGlobalMaxActiveDownloads, helpqbittorrentGlobalMaxActiveDownloads, float64(result.MaxActiveDownloads)},
		{qbittorrentGlobalMaxActiveUploads, helpqbittorrentGlobalMaxActiveUploads, float64(result.MaxActiveUploads)},
		{qbittorrentGlobalMaxActiveTorrents, helpqbittorrentGlobalMaxActiveTorrents, float64(result.MaxActiveTorrents)},
		{qbittorrentGlobalDownloadRateLimitBytes, helpqbittorrentGlobalDownloadRateLimitBytes, float64(result.DlLimit)},
		{qbittorrentGlobalUploadRateLimitBytes, helpqbittorrentGlobalUploadRateLimitBytes, float64(result.UpLimit)},
		{qbittorrentGlobalAltDownloadRateLimitBytes, helpqbittorrentGlobalAltDownloadRateLimitBytes, float64(result.AltDlLimit)},
		{qbittorrentGlobalAltUploadRateLimitBytes, helpqbittorrentGlobalAltUploadRateLimitBytes, float64(result.AltUpLimit)},
	}

	registerGaugeGlobalAndSet(&gauges, r)
}

func Trackers(result []*API.Trackers, r *prometheus.Registry) {
	if len(result) == 0 {
		logger.Trace("No tracker")
		return
	}

	labels := []string{trackerLabelURL}

	gauges := GaugeList{
		{qbittorrentTrackerDownloaded, helpqbittorrentTrackerDownloaded, labels},
		{qbittorrentTrackerLeeches, helpqbittorrentTrackerLeeches, labels},
		{qbittorrentTrackerPeers, helpqbittorrentTrackerPeers, labels},
		{qbittorrentTrackerSeeders, helpqbittorrentTrackerSeeders, labels},
		{qbittorrentTrackerStatus, helpqbittorrentTrackerStatus, labels},
		{qbittorrentTrackerTier, helpqbittorrentTrackerTier, labels},
	}

	metrics := registerGauge(&gauges, r)

	if app.Exporter.Features.EnableHighCardinality {
		metrics[qbittorrentTrackerInfo] = newGaugeVec(qbittorrentTrackerInfo, helpqbittorrentTrackerInfo,
			[]string{trackerLabelMessage, labelDownloaded, trackerLabelLeeches, trackerLabelPeers, labelSeeders, trackerLabelStatus, trackerLabelTier, trackerLabelURL})
		r.MustRegister(metrics[qbittorrentTrackerInfo])
	}

	for _, listOfTracker := range result {
		for _, tracker := range *listOfTracker {
			if internal.IsValidURL(tracker.URL) {
				tier, err := strconv.Atoi(string(tracker.Tier))
				if err != nil {
					logger.Trace(fmt.Sprintf("can't convert \"%s\" to int", tracker.Tier))
					tier = 0
				}
				labels := prometheus.Labels{trackerLabelURL: tracker.URL}
				metrics[qbittorrentTrackerDownloaded].With(labels).Set(float64(tracker.NumDownloaded))
				metrics[qbittorrentTrackerLeeches].With(labels).Set(float64(tracker.NumLeeches))
				metrics[qbittorrentTrackerSeeders].With(labels).Set(float64(tracker.NumSeeds))
				metrics[qbittorrentTrackerPeers].With(labels).Set(float64(tracker.NumPeers))
				metrics[qbittorrentTrackerStatus].With(labels).Set(float64(tracker.Status))
				metrics[qbittorrentTrackerTier].With(labels).Set(float64(tier))

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
					metrics[qbittorrentTrackerInfo].With(qbittorrentTrackerInfoLabels).Set(1)
				}
			}
		}
	}
}

func MainData(result *API.MainData, r *prometheus.Registry) {
	var globalRatio float64
	var err error

	globalRatio, err = strconv.ParseFloat(result.ServerState.GlobalRatio, 64)

	if err != nil {
		logger.Trace("retrying to convert ratio...")
		newGlobalRatioState := strings.ReplaceAll(result.ServerState.GlobalRatio, ",", ".")
		globalRatio, err = strconv.ParseFloat(newGlobalRatioState, 64)
		if err != nil {
			logger.Warn(fmt.Sprintf("error to convert ratio \"%s\"", result.ServerState.GlobalRatio))
		}
	}
	if err == nil {
		qbittorrentGlobalRatio := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: qbittorrentGlobalRatio,
			Help: helpqbittorrentGlobalRatio,
		})
		r.MustRegister(qbittorrentGlobalRatio)
		qbittorrentGlobalRatio.Set(globalRatio)
	}
	useAltSpeedLimits := 0.0
	if result.ServerState.UseAltSpeedLimits {
		useAltSpeedLimits = 1.0
	}
	qbittorrentAppAltRateLimitsEnabled := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: qbittorrentAppAltRateLimitsEnabled,
		Help: helpQbittorrentAppAltRateLimitsEnabled,
	})
	r.MustRegister(qbittorrentAppAltRateLimitsEnabled)
	qbittorrentAppAltRateLimitsEnabled.Set(float64(useAltSpeedLimits))

	gauges := GaugeSet{
		{qbittorrentGlobalAlltimeDownloadedBytes, helpqbittorrentGlobalAlltimeDownloadedBytes, float64(result.ServerState.AlltimeDl)},
		{qbittorrentGlobalAlltimeUploadedBytes, helpqbittorrentGlobalAlltimeUploadedBytes, float64(result.ServerState.AlltimeUl)},
		{qbittorrentGlobalSessionDownloadedBytes, helpqbittorrentGlobalSessionDownloadedBytes, float64(result.ServerState.DlInfoData)},
		{qbittorrentGlobalSessionUploadedBytes, helpqbittorrentGlobalSessionUploadedBytes, float64(result.ServerState.UpInfoData)},
		{qbittorrentGlobalDownloadSpeedBytes, helpqbittorrentGlobalDownloadSpeedBytes, float64(result.ServerState.DlInfoSpeed)},
		{qbittorrentGlobalUploadSpeedBytes, helpqbittorrentGlobalUploadSpeedBytes, float64(result.ServerState.UpInfoSpeed)},
		{qbittorrentGlobalDHTNodes, helpqbittorrentGlobalDHTNodes, float64(result.ServerState.DHTNodes)},
		{qbittorrentGlobalAverageTimeQueue, helpqbittorrentGlobalAverageTimeQueue, float64(result.ServerState.AverageTimeQueue)},
		{qbittorrentGlobalFreeSpaceOnDiskBytes, helpqbittorrentGlobalFreeSpaceOnDisk, float64(result.ServerState.FreeSpaceOnDisk)},
		{qbittorrentGlobalQueuedIoJobs, helpqbittorrentGlobalQueuedIoJobs, float64(result.ServerState.QueuedIoJobs)},
		{qbittorrentGlobalTotalBuffersSizeBytes, helpqbittorrentGlobalTotalBuffersSize, float64(result.ServerState.TotalBuffersSize)},
		{qbittorrentGlobalTotalQueuedSizeBytes, helpqbittorrentGlobalTotalQueuedSize, float64(result.ServerState.TotalQueuedSize)},
		{qbittorrentGlobalTotalPeerConnections, helpqbittorrentGlobalTotalPeerConnections, float64(result.ServerState.TotalPeerConnections)},
		{qbittorrentGlobalTotalWastedSessionBytes, helpqbittorrentGlobalWastedSession, float64(result.ServerState.TotalWastedSession)},
	}

	qbittorrentTransferConnectionStatus := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: qbittorrentTorrentTransferConnectionStatus,
		Help: helpQbittorrentTorrentTransferConnectionStatus,
	}, []string{torrentLabelConnectionStatus})

	r.MustRegister(qbittorrentTransferConnectionStatus)
	qbittorrentTransferConnectionStatus.With(prometheus.Labels{
		torrentLabelConnectionStatus: result.ServerState.ConnectionStatus,
	}).Set(1)

	registerGaugeGlobalAndSet(&gauges, r)

	qbittorrentGlobalTags := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: qbittorrentGlobalTags,
		Help: helpqbittorrentGlobalTags,
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
		Name: qbittorrentGlobalCategories,
		Help: helpqbittorrentGlobalCategories,
	}, []string{torrentLabelCategory})
	r.MustRegister(qbittorrentGlobalCategories)
	for _, category := range result.CategoryMap {
		labels := prometheus.Labels{
			torrentLabelCategory: category.Name,
		}
		qbittorrentGlobalCategories.With(labels).Set(1)
	}
}

func registerGaugeGlobalAndSet(gauges *GaugeSet, r *prometheus.Registry) {
	for _, gauge := range *gauges {
		g := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: gauge.Name,
			Help: gauge.Help,
		})
		r.MustRegister(g)
		g.Set(gauge.Value)
	}
}

func registerGauge(gauges *GaugeList, r *prometheus.Registry) map[string]*prometheus.GaugeVec {
	metrics := make(map[string]*prometheus.GaugeVec, len(*gauges))
	for _, gauge := range *gauges {
		metrics[gauge.Name] = newGaugeVec(gauge.Name, gauge.Help, gauge.Labels)
		r.MustRegister(metrics[gauge.Name])
	}
	return metrics
}

func newGaugeVec(name, help string, labels []string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}, labels)
}
