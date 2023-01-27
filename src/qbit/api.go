package qbit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"qbit-prom/src/models"
	"strconv"
	"strings"
)

func Gettorrent() string {
	resp, err := Apirequest("/api/v2/torrents/info", "GET")
	if err != nil {
		if err.Error() == "403" {
			log.Println("Cookie changed, try to reconnect ...")
			Auth()
		} else {
			if models.GetPromptError() == false {
				log.Println("Error : ", err)
			}
		}
	} else {
		if models.GetPromptError() == true {
			models.SetPromptError(false)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		} else {

			var result models.Response
			if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
				log.Println("Can not unmarshal JSON")
			}
			message := Sendbackmessagetorrent(result)
			return message

		}
	}

	return ""
}

func getPreferences() string {
	resp, err := Apirequest("/api/v2/app/preferences", "GET")
	if err != nil {
		if err.Error() == "403" {
			log.Println("Cookie changed, try to reconnect ...")
			Auth()
		} else {
			if models.GetPromptError() == false {
				log.Println("Error : ", err)
			}
		}

	} else {
		if models.GetPromptError() == true {
			models.SetPromptError(false)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		} else {

			var result models.Preferences
			if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
				log.Println("Can not unmarshal JSON")
			}
			message := Sendbackmessagepreference(result)
			return message

		}
	}

	return ""
}
func getMainData() string {
	resp, err := Apirequest("/api/v2/sync/maindata", "GET")
	if err != nil {
		if err.Error() == "403" {
			log.Println("Cookie changed, try to reconnect ...")
			Auth()
		} else {
			if models.GetPromptError() == false {
				log.Println("Error : ", err)
			}
		}
	} else {
		if models.GetPromptError() == true {
			models.SetPromptError(false)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		} else {

			var result models.Maindata
			if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
				log.Println("Can not unmarshal JSON")
			}
			message := Sendbackmessagemaindata(result)
			return message

		}
	}

	return ""
}

func Handlerequest(uri string, method string) string {

	resp, err := Apirequest(uri, method)
	if err != nil {

		if err.Error() == "403" {
			log.Println("Cookie changed, try to reconnect ...")
			Auth()
		} else {
			if models.GetPromptError() == false {
				log.Println("Error : ", err)
			}

		}

	} else {
		if models.GetPromptError() == true {
			models.SetPromptError(false)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		} else {
			sb := string(body)

			return sb
		}
	}

	return ""
}

func qbitversion() string {
	version := Handlerequest("/api/v2/app/version", "GET")
	if version == "" {
		return ""
	} else {
		message := "# HELP qbittorrent_app_version The current qBittorrent version\n# TYPE qbittorrent_app_version gauge\n"
		message = message + `qbittorrent_app_version{version="` + version + `",} 1.0` + "\n"
		return message
	}

}

func Allrequests() string {
	message := qbitversion() + Gettorrent() + getPreferences() + getMainData()
	return message
}

func Apirequest(uri string, method string) (*http.Response, error) {

	req, err := http.NewRequest(method, models.Getbaseurl()+uri, nil)
	if err != nil {
		log.Println("Error with url")
	}

	req.AddCookie(&http.Cookie{Name: "SID", Value: models.Getcookie()})
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		err := fmt.Errorf("Can't connect to server")
		log.Println(err.Error())
		return resp, err

	} else {
		if resp.StatusCode == 200 {

			return resp, nil

		} else {
			err := fmt.Errorf("%d", resp.StatusCode)
			if models.GetPromptError() == false {
				models.SetPromptError(true)

				log.Println("Error code ", err.Error())

			}
			return resp, err

		}

	}

}

func Sendbackmessagetorrent(result models.Response) string {
	qbittorrent_torrent_info := "# HELP qbittorrent_torrent_info All info for torrents\n# TYPE qbittorrent_torrent_info gauge\n"
	qbittorrent_torrent_download_speed_bytes := "# HELP qbittorrent_torrent_download_speed_bytes The current download speed of torrents (in bytes)\n# TYPE qbittorrent_torrent_download_speed_bytes gauge\n"
	qbittorrent_torrent_upload_speed_bytes := "# HELP qbittorrent_torrent_upload_speed_bytes The current upload speed of torrents (in bytes)\n# TYPE qbittorrent_torrent_upload_speed_bytes gauge\n"
	qbittorrent_torrent_eta := "# HELP qbittorrent_torrent_eta The current ETA for each torrent (in seconds)\n# TYPE qbittorrent_torrent_eta gauge\n"
	qbittorrent_torrent_progress := "# HELP qbittorrent_torrent_progress The current progress of torrents\n# TYPE qbittorrent_torrent_progress gauge\n"
	qbittorrent_torrent_time_active := "# HELP qbittorrent_torrent_time_active The total active time (in seconds)\n# TYPE qbittorrent_torrent_time_active gauge\n"
	qbittorrent_torrent_states := "# HELP qbittorrent_torrent_states The current state of torrents\n# TYPE qbittorrent_torrent_states gauge\n"
	qbittorrent_torrent_seeders := "# HELP qbittorrent_torrent_seeders The current number of seeders for each torrent\n# TYPE qbittorrent_torrent_seeders gauge\n"
	qbittorrent_torrent_leechers := "# HELP qbittorrent_torrent_leechers The current number of leechers for each torrent\n# TYPE qbittorrent_torrent_leechers gauge\n"
	qbittorrent_torrent_ratio := "# HELP qbittorrent_torrent_ratio The current ratio each torrent\n# TYPE qbittorrent_torrent_ratio gauge\n"
	qbittorrent_torrent_amount_left_bytes := "# HELP qbittorrent_torrent_amount_left_bytes The amount remaining for each torrent (in bytes)\n# TYPE qbittorrent_torrent_amount_left_bytes gauge\n"
	qbittorrent_torrent_size_bytes := "# HELP qbittorrent_torrent_size_bytes The size for each torrent (in bytes)\n# TYPE qbittorrent_torrent_size_bytes gauge\n"
	qbittorrent_torrent_session_downloaded_bytes := "# HELP qbittorrent_torrent_session_downloaded_bytes The current session download amount of torrents (in bytes)\n# TYPE qbittorrent_torrent_session_downloaded_bytes gauge\n"
	qbittorrent_torrent_session_uploaded_bytes := "# HELP qbittorrent_torrent_session_uploaded_bytes The current session upload amount of torrents (in bytes)\n# TYPE qbittorrent_torrent_session_uploaded_bytes gauge\n"
	qbittorrent_torrent_total_downloaded_bytes := "# HELP qbittorrent_torrent_total_downloaded_bytes The current total download amount of torrents (in bytes)\n# TYPE qbittorrent_torrent_total_downloaded_bytes gauge\n"
	qbittorrent_torrent_total_uploaded_bytes := "# HELP qbittorrent_torrent_total_uploaded_bytes The current total upload amount of torrents (in bytes)\n# TYPE qbittorrent_torrent_total_uploaded_bytes gauge\n"
	qbittorrent_global_torrents := "# HELP qbittorrent_global_torrents The total number of torrents\n# TYPE qbittorrent_global_torrents gauge\n"
	count_stelledup := 0
	count_uploading := 0
	for i := 0; i < len(result); i++ {

		if result[i].State == "stalledUP" {
			count_stelledup += 1
		} else {
			count_uploading += 1
		}
		qbittorrent_torrent_info = qbittorrent_torrent_info + `qbittorrent_torrent_info{name="` + result[i].Name + `",state="` + result[i].State + `",size="` + fmt.Sprintf("%g", result[i].Size) + `",progress="` + fmt.Sprintf("%g", result[i].Progress) + `",seeders="` + fmt.Sprintf("%g", result[i].NumSeeds) + `",leechers="` + fmt.Sprintf("%g", result[i].NumLeechs) + `",dl_speed="` + fmt.Sprintf("%g", result[i].Dlspeed) + `",up_speed="` + fmt.Sprintf("%g", result[i].Upspeed) + `",amount_left="` + fmt.Sprintf("%g", result[i].AmountLeft) + `",time_active="` + fmt.Sprintf("%g", result[i].TimeActive) + `",eta="` + fmt.Sprintf("%g", result[i].Eta) + `",uploaded="` + fmt.Sprintf("%g", result[i].Uploaded) + `",uploaded_session="` + fmt.Sprintf("%g", result[i].UploadedSession) + `",downloaded="` + fmt.Sprintf("%g", result[i].Downloaded) + `",downloaded_session="` + fmt.Sprintf("%g", result[i].DownloadedSession) + `",max_ratio="` + fmt.Sprintf("%g", result[i].MaxRatio) + `",ratio="` + fmt.Sprintf("%g", result[i].Ratio) + `",} 1.0` + "\n"
		qbittorrent_torrent_download_speed_bytes = qbittorrent_torrent_download_speed_bytes + `qbittorrent_torrent_download_speed_bytes{name="` + result[i].Name + `",} ` + fmt.Sprintf("%.1f", result[i].Dlspeed) + "\n"
		qbittorrent_torrent_upload_speed_bytes = qbittorrent_torrent_upload_speed_bytes + `qbittorrent_torrent_upload_speed_bytes{name="` + result[i].Name + `",} ` + fmt.Sprintf("%.1f", result[i].Upspeed) + "\n"
		qbittorrent_torrent_eta = qbittorrent_torrent_eta + `qbittorrent_torrent_eta{name="` + result[i].Name + `",} ` + fmt.Sprintf("%.1f", result[i].Eta) + "\n"
		qbittorrent_torrent_progress = qbittorrent_torrent_progress + `qbittorrent_torrent_progress{name="` + result[i].Name + `",} ` + fmt.Sprintf("%.1f", result[i].Progress) + "\n"
		qbittorrent_torrent_time_active = qbittorrent_torrent_time_active + `qbittorrent_torrent_time_active{name="` + result[i].Name + `",} ` + fmt.Sprintf("%.1f", result[i].TimeActive) + "\n"
		qbittorrent_torrent_seeders = qbittorrent_torrent_seeders + `qbittorrent_torrent_seeders{name="` + result[i].Name + `",} ` + fmt.Sprintf("%.1f", result[i].NumSeeds) + "\n"
		qbittorrent_torrent_leechers = qbittorrent_torrent_leechers + `qbittorrent_torrent_leechers{name="` + result[i].Name + `",} ` + fmt.Sprintf("%.1f", result[i].NumLeechs) + "\n"
		qbittorrent_torrent_ratio = qbittorrent_torrent_ratio + `qbittorrent_torrent_ratio{name="` + result[i].Name + `",} ` + fmt.Sprintf("%v", result[i].Ratio) + "\n"
		qbittorrent_torrent_amount_left_bytes = qbittorrent_torrent_amount_left_bytes + `qbittorrent_torrent_amount_left_bytes{name="` + result[i].Name + `",} ` + fmt.Sprintf("%.1f", result[i].AmountLeft) + "\n"
		qbittorrent_torrent_size_bytes = qbittorrent_torrent_size_bytes + `qbittorrent_torrent_size_bytes{name="` + result[i].Name + `",} ` + fmt.Sprintf("%.1f", result[i].Size) + "\n"
		qbittorrent_torrent_session_downloaded_bytes = qbittorrent_torrent_session_downloaded_bytes + `qbittorrent_torrent_session_downloaded_bytes{name="` + result[i].Name + `",} ` + fmt.Sprintf("%.1f", result[i].DownloadedSession) + "\n"
		qbittorrent_torrent_session_uploaded_bytes = qbittorrent_torrent_session_uploaded_bytes + `qbittorrent_torrent_session_uploaded_bytes{name="` + result[i].Name + `",} ` + fmt.Sprintf("%.1f", result[i].UploadedSession) + "\n"
		qbittorrent_torrent_total_downloaded_bytes = qbittorrent_torrent_total_downloaded_bytes + `qbittorrent_torrent_total_downloaded_bytes{name="` + result[i].Name + `",} ` + fmt.Sprintf("%.1f", result[i].DownloadedSession) + "\n"
		qbittorrent_torrent_total_uploaded_bytes = qbittorrent_torrent_total_uploaded_bytes + `qbittorrent_torrent_total_uploaded_bytes{name="` + result[i].Name + `",} ` + convert(result[i].UploadedSession) + "\n"

	}
	qbittorrent_torrent_states = qbittorrent_torrent_states + `qbittorrent_torrent_states{name="stalledUP",} ` + strconv.Itoa(count_stelledup) + "\n"
	qbittorrent_torrent_states = qbittorrent_torrent_states + `qbittorrent_torrent_states{name="uploading",} ` + strconv.Itoa(count_uploading) + "\n"
	qbittorrent_global_torrents = qbittorrent_global_torrents + "qbittorrent_global_torrents " + strconv.Itoa(count_stelledup+count_uploading) + "\n"
	total := qbittorrent_torrent_download_speed_bytes + qbittorrent_torrent_upload_speed_bytes + qbittorrent_torrent_eta + qbittorrent_torrent_progress + qbittorrent_torrent_time_active + qbittorrent_torrent_states + qbittorrent_torrent_seeders + qbittorrent_torrent_leechers + qbittorrent_torrent_ratio + qbittorrent_torrent_amount_left_bytes + qbittorrent_torrent_size_bytes + qbittorrent_torrent_info + qbittorrent_torrent_session_downloaded_bytes + qbittorrent_torrent_session_uploaded_bytes + qbittorrent_torrent_total_downloaded_bytes + qbittorrent_torrent_total_uploaded_bytes
	return total
}

func convert(input float64) string {

	if input < 10000000 {
		return fmt.Sprintf("%.1f", input)
	} else {
		temp := fmt.Sprintf(strconv.FormatFloat(input, 'E', -1, 64))
		temp = strings.ReplaceAll(temp, "+0", "")
		temp = strings.ReplaceAll(temp, "+", "")
		return temp
	}

}

func Sendbackmessagepreference(result models.Preferences) string {
	total := ""
	total = total + "# HELP qbittorrent_app_max_active_downloads The max number of downloads allowed\n# TYPE qbittorrent_app_max_active_downloads gauge\nqbittorrent_app_max_active_downloads " + fmt.Sprintf("%.1f", result.MaxActiveDownloads) + "\n"
	total = total + "# HELP qbittorrent_app_max_active_uploads The max number of active uploads allowed\n# TYPE qbittorrent_app_max_active_uploads gauge\nqbittorrent_app_max_active_uploads " + fmt.Sprintf("%.1f", result.MaxActiveUploads) + "\n"
	total = total + "# HELP qbittorrent_app_max_active_torrents The max number of active torrents allowed\n# TYPE qbittorrent_app_max_active_torrents gauge\nqbittorrent_app_max_active_torrents " + fmt.Sprintf("%.1f", result.MaxActiveTorrents) + "\n"
	total = total + "# HELP qbittorrent_app_download_rate_limit_bytes The global download rate limit (in bytes)\n# TYPE qbittorrent_app_download_rate_limit_bytes gauge\nqbittorrent_app_download_rate_limit_bytes " + fmt.Sprintf("%.1f", result.DlLimit) + "\n"
	total = total + "# HELP qbittorrent_app_upload_rate_limit_bytes The global upload rate limit (in bytes)\n# TYPE qbittorrent_app_upload_rate_limit_bytes gauge\nqbittorrent_app_upload_rate_limit_bytes " + fmt.Sprintf("%.1f", result.UpLimit) + "\n"
	total = total + "# HELP qbittorrent_app_alt_download_rate_limit_bytes The alternate download rate limit (in bytes)\n# TYPE qbittorrent_app_alt_download_rate_limit_bytes gauge\nqbittorrent_app_alt_download_rate_limit_bytes " + fmt.Sprintf("%.1f", result.AltDlLimit) + "\n"

	return total
}

func Sendbackmessagemaindata(result models.Maindata) string {

	UseAltSpeedLimits := "0.0"
	if result.ServerState.UseAltSpeedLimits == true {
		UseAltSpeedLimits = "1.0"
	}
	total := ""
	total = total + "# HELP qbittorrent_app_alt_rate_limits_enabled If alternate rate limits are enabled\n# TYPE qbittorrent_app_alt_rate_limits_enabled gauge\nqbittorrent_app_alt_rate_limits_enabled " + UseAltSpeedLimits + "\n"
	total = total + "# HELP qbittorrent_global_alltime_downloaded_bytes The all-time total download amount of torrents (in bytes)\n# TYPE qbittorrent_global_alltime_downloaded_bytes gauge\nqbittorrent_global_alltime_downloaded_bytes " + fmt.Sprintf("%.1f", result.ServerState.AlltimeDl) + "\n"
	total = total + "# HELP qbittorrent_global_alltime_uploaded_bytes The all-time total upload amount of torrents (in bytes)\n# TYPE qbittorrent_global_alltime_uploaded_bytes gauge\nqbittorrent_global_alltime_uploaded_bytes " + fmt.Sprintf("%.1f", result.ServerState.AlltimeUl) + "\n"
	total = total + "# HELP qbittorrent_global_session_downloaded_bytes The total download amount of torrents for this session (in bytes)\n# TYPE qbittorrent_global_session_downloaded_bytes gauge\nqbittorrent_global_session_downloaded_bytes " + fmt.Sprintf("%.1f", result.ServerState.DlInfoData) + "\n"
	total = total + "# HELP qbittorrent_global_download_speed_bytes The current download speed of all torrents (in bytes)\n# TYPE qbittorrent_global_download_speed_bytes gauge\nqbittorrent_global_download_speed_bytes " + fmt.Sprintf("%.1f", result.ServerState.DlInfoSpeed) + "\n"
	total = total + "# HELP qbittorrent_global_upload_speed_bytes The total current upload speed of all torrents (in bytes)\n# TYPE qbittorrent_global_upload_speed_bytes gauge\nqbittorrent_global_upload_speed_bytes " + fmt.Sprintf("%.1f", result.ServerState.UpInfoSpeed) + "\n"
	total = total + "# HELP qbittorrent_global_ratio The current global ratio of all torrents\n# TYPE qbittorrent_global_ratio gauge\nqbittorrent_global_ratio " + result.ServerState.GlobalRatio + "\n"
	return total
}
