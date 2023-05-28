package prom

import (
	"fmt"
	"qbit-exp/src/models"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

func Sendbackmessagetorrent(result *models.Response, r *prometheus.Registry) {

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
	}, []string{"name", "category", "state", "size", "progress", "seeders", "leechers", "dl_speed", "up_speed", "amount_left", "time_active", "eta", "uploaded", "uploaded_session", "downloaded", "downloaded_session", "max_ratio", "ratio"})
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
	for i := 0; i < len(*result); i++ {
		qbittorrent_eta.With(prometheus.Labels{"name": (*result)[i].Name}).Set(float64((*result)[i].Eta))
		qbittorrent_torrent_download_speed_bytes.With(prometheus.Labels{"name": (*result)[i].Name}).Set(float64((*result)[i].Dlspeed))
		qbittorrent_torrent_upload_speed_bytes.With(prometheus.Labels{"name": (*result)[i].Name}).Set(float64((*result)[i].Upspeed))
		qbittorrent_torrent_progress.With(prometheus.Labels{"name": (*result)[i].Name}).Set(float64((*result)[i].Progress))
		qbittorrent_torrent_time_active.With(prometheus.Labels{"name": (*result)[i].Name}).Set(float64((*result)[i].TimeActive))
		qbittorrent_torrent_seeders.With(prometheus.Labels{"name": (*result)[i].Name}).Set(float64((*result)[i].NumSeeds))
		qbittorrent_torrent_leechers.With(prometheus.Labels{"name": (*result)[i].Name}).Set(float64((*result)[i].NumLeechs))
		qbittorrent_torrent_ratio.With(prometheus.Labels{"name": (*result)[i].Name}).Set(float64((*result)[i].Ratio))
		qbittorrent_torrent_amount_left_bytes.With(prometheus.Labels{"name": (*result)[i].Name}).Set(float64((*result)[i].AmountLeft))
		qbittorrent_torrent_size_bytes.With(prometheus.Labels{"name": (*result)[i].Name}).Set(float64((*result)[i].Size))
		qbittorrent_torrent_session_downloaded_bytes.With(prometheus.Labels{"name": (*result)[i].Name}).Set(float64((*result)[i].DownloadedSession))
		qbittorrent_torrent_session_uploaded_bytes.With(prometheus.Labels{"name": (*result)[i].Name}).Set(float64((*result)[i].UploadedSession))
		qbittorrent_torrent_total_downloaded_bytes.With(prometheus.Labels{"name": (*result)[i].Name}).Set(float64((*result)[i].Downloaded))
		qbittorrent_torrent_total_uploaded_bytes.With(prometheus.Labels{"name": (*result)[i].Name}).Set(float64((*result)[i].Uploaded))
		if (*result)[i].State == "stalledUP" {
			count_stelledup += 1
		} else {
			count_uploading += 1
		}
		qbittorrent_torrent_info.With(prometheus.Labels{"name": (*result)[i].Name, "category": (*result)[i].Category, "state": (*result)[i].State, "size": strconv.Itoa((*result)[i].Size), "progress": strconv.Itoa(int((*result)[i].Progress)), "seeders": strconv.Itoa(int((*result)[i].NumSeeds)), "leechers": strconv.Itoa(int((*result)[i].NumLeechs)), "dl_speed": strconv.Itoa(int((*result)[i].Dlspeed)), "up_speed": strconv.Itoa(int((*result)[i].Upspeed)), "amount_left": strconv.Itoa(int((*result)[i].AmountLeft)), "time_active": strconv.Itoa(int((*result)[i].TimeActive)), "eta": strconv.Itoa(int((*result)[i].Eta)), "uploaded": strconv.Itoa(int((*result)[i].Uploaded)), "uploaded_session": strconv.Itoa(int((*result)[i].UploadedSession)), "downloaded": strconv.Itoa(int((*result)[i].Downloaded)), "downloaded_session": strconv.Itoa(int((*result)[i].DownloadedSession)), "max_ratio": strconv.Itoa(int((*result)[i].MaxRatio)), "ratio": strconv.Itoa(int((*result)[i].Ratio))}).Set(1)
		if (*result)[i].Tags != "" {
			separated_list := strings.Split((*result)[i].Tags, ", ")
			for j := 0; j < len(separated_list); j++ {
				labels := prometheus.Labels{
					"name": (*result)[i].Name,
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

func Sendbackmessagepreference(result *models.Preferences, r *prometheus.Registry) {
	qbittorrent_app_max_active_downloads := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "qbittorrent_app_max_active_downloads",
		Help: "The max number of downloads allowed",
	})
	qbittorrent_app_max_active_uploads := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "qbittorrent_app_max_active_uploads",
		Help: "The max number of active uploads allowed",
	})
	qbittorrent_app_max_active_torrents := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "qbittorrent_app_max_active_torrents",
		Help: "The max number of active torrents allowed",
	})
	qbittorrent_app_download_rate_limit_bytes := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "qbittorrent_app_download_rate_limit_bytes",
		Help: "The global download rate limit (in bytes)",
	})
	qbittorrent_app_upload_rate_limit_bytes := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "qbittorrent_app_upload_rate_limit_bytes",
		Help: "The global upload rate limit (in bytes)",
	})
	qbittorrent_app_alt_download_rate_limit_bytes := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "qbittorrent_app_alt_download_rate_limit_bytes",
		Help: "The alternate download rate limit (in bytes)",
	})
	qbittorrent_app_alt_upload_rate_limit_bytes := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "qbittorrent_app_alt_upload_rate_limit_bytes",
		Help: "The alternate upload rate limit (in bytes)",
	})
	r.MustRegister(qbittorrent_app_max_active_downloads)
	r.MustRegister(qbittorrent_app_max_active_uploads)
	r.MustRegister(qbittorrent_app_max_active_torrents)
	r.MustRegister(qbittorrent_app_download_rate_limit_bytes)
	r.MustRegister(qbittorrent_app_upload_rate_limit_bytes)
	r.MustRegister(qbittorrent_app_alt_download_rate_limit_bytes)
	r.MustRegister(qbittorrent_app_alt_upload_rate_limit_bytes)
	qbittorrent_app_max_active_downloads.Set(float64((*result).MaxActiveDownloads))
	qbittorrent_app_max_active_uploads.Set(float64((*result).MaxActiveDownloads))
	qbittorrent_app_max_active_torrents.Set(float64((*result).MaxActiveTorrents))
	qbittorrent_app_download_rate_limit_bytes.Set(float64((*result).DlLimit))
	qbittorrent_app_upload_rate_limit_bytes.Set(float64((*result).UpLimit))
	qbittorrent_app_alt_download_rate_limit_bytes.Set(float64((*result).AltDlLimit))
	qbittorrent_app_alt_upload_rate_limit_bytes.Set(float64((*result).AltUpLimit))

}

func Sendbackmessagemaindata(result *models.Maindata, r *prometheus.Registry) {
	globalratio, err := strconv.ParseFloat((*result).ServerState.GlobalRatio, 64)

	if err != nil {
		fmt.Println("error to convert ratio")
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
	qbittorrent_global_alltime_downloaded_bytes := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "qbittorrent_global_alltime_downloaded_bytes",
		Help: "The all-time total download amount of torrents (in bytes)",
	})
	qbittorrent_global_alltime_uploaded_bytes := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "qbittorrent_global_alltime_uploaded_bytes",
		Help: "The all-time total upload amount of torrents (in bytes)",
	})
	qbittorrent_global_session_downloaded_bytes := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "qbittorrent_global_session_downloaded_bytes",
		Help: "The total download amount of torrents for this session (in bytes)",
	})
	qbittorrent_global_session_uploaded_bytes := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "qbittorrent_global_session_uploaded_bytes",
		Help: "The total upload amount of torrents for this session (in bytes)",
	})
	qbittorrent_global_download_speed_bytes := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "qbittorrent_global_download_speed_bytes",
		Help: "The current download speed of all torrents (in bytes)",
	})
	qbittorrent_global_upload_speed_bytes := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "qbittorrent_global_upload_speed_bytes",
		Help: "The total current upload speed of all torrents (in bytes)",
	})
	qbittorrent_global_tags := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_global_tags",
		Help: "All tags used in qbittorrent",
	}, []string{"tag"})
	qbittorrent_global_categories := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qbittorrent_global_categories",
		Help: "All categories used in qbittorrent",
	}, []string{"category"})

	r.MustRegister(qbittorrent_app_alt_rate_limits_enabled)
	r.MustRegister(qbittorrent_global_alltime_downloaded_bytes)
	r.MustRegister(qbittorrent_global_alltime_uploaded_bytes)
	r.MustRegister(qbittorrent_global_session_downloaded_bytes)
	r.MustRegister(qbittorrent_global_session_uploaded_bytes)
	r.MustRegister(qbittorrent_global_download_speed_bytes)
	r.MustRegister(qbittorrent_global_upload_speed_bytes)
	r.MustRegister(qbittorrent_global_tags)
	r.MustRegister(qbittorrent_global_categories)

	if len((*result).Tags) > 0 {

		for j := 0; j < len((*result).Tags); j++ {
			labels := prometheus.Labels{
				"tag": (*result).Tags[j],
			}
			qbittorrent_global_tags.With(labels).Set(1)
		}

	}
	for _, category := range result.CategoryMap {
		labels := prometheus.Labels{
			"category": category.Name,
		}
		qbittorrent_global_categories.With(labels).Set(1)
	}

	qbittorrent_app_alt_rate_limits_enabled.Set(float64(UseAltSpeedLimits))
	qbittorrent_global_alltime_downloaded_bytes.Set(float64((*result).ServerState.AlltimeDl))
	qbittorrent_global_alltime_uploaded_bytes.Set(float64((*result).ServerState.AlltimeUl))
	qbittorrent_global_session_downloaded_bytes.Set(float64((*result).ServerState.DlInfoData))
	qbittorrent_global_session_uploaded_bytes.Set(float64((*result).ServerState.UpInfoData))
	qbittorrent_global_download_speed_bytes.Set(float64((*result).ServerState.DlInfoSpeed))
	qbittorrent_global_upload_speed_bytes.Set(float64((*result).ServerState.UpInfoSpeed))

}
