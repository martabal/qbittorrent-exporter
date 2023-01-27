package qbit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"qbit-prom/src/models"
	"strconv"
)

func Gettorrent() string {
	resp, err := Apirequest("/api/v2/torrents/info?filter=all", "GET")
	if err != nil {
		log.Println("Cookie changed")
		log.Println(err)
		Auth()
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	} else {

		var result models.Response
		if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
			log.Println("Can not unmarshal JSON")
		}
		message := Sendbackmessage(result)
		return message

	}

	return ""
}

func Handlerequest(uri string, method string) string {

	resp, err := Apirequest(uri, method)
	if err != nil {

		if err.Error() == "403" {
			log.Println("Cookie changed, try to reconnect ...")
			Auth()
		}

	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	} else {
		sb := string(body)

		return sb
	}

	return ""
}

func qbitversion() string {
	message := "# HELP qbittorrent_app_version The current qBittorrent version\n# TYPE qbittorrent_app_version gauge\n"
	message = message + `qbittorrent_app_version{version="` + Handlerequest("/api/v2/app/version", "GET") + `",} 1.0` + "\n"
	return message
}

func Allrequests() string {
	message := qbitversion() + Gettorrent()
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
			log.Println("Error code ", err.Error())
			return resp, err
		}

	}

}

func Sendbackmessage(result models.Response) string {
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
	count_stelledup := 0
	count_uploading := 0
	for i := 0; i < len(result); i++ {

		if result[i].State == "stalledUP" {
			count_stelledup += 1
		} else {
			count_uploading += 1
		}
		qbittorrent_torrent_info = qbittorrent_torrent_info + `qbittorrent_torrent_info{name="` + result[i].Name + `",state="` + result[i].State + `",size="` + strconv.Itoa(result[i].Size) + `",progress="` + fmt.Sprintf("%v", result[i].Progress) + `",seeders="` + fmt.Sprintf("%v", result[i].NumSeeds) + `",leechers="` + fmt.Sprintf("%v", result[i].NumLeechs) + `",dl_speed="` + fmt.Sprintf("%v", result[i].Dlspeed) + `",up_speed="` + fmt.Sprintf("%v", result[i].Upspeed) + `",amount_left="` + fmt.Sprintf("%v", result[i].AmountLeft) + `",time_active="` + fmt.Sprintf("%v", result[i].TimeActive) + `",eta="` + fmt.Sprintf("%v", result[i].Eta) + `",uploaded="` + fmt.Sprintf("%v", result[i].Uploaded) + `",uploaded_session="` + fmt.Sprintf("%v", result[i].UploadedSession) + `",downloaded="` + fmt.Sprintf("%v", result[i].Downloaded) + `",downloaded_session="` + fmt.Sprintf("%v", result[i].DownloadedSession) + `",max_ratio="` + fmt.Sprintf("%v", result[i].MaxRatio) + `",ratio="` + fmt.Sprintf("%v", result[i].Ratio) + `",} 1.0` + "\n"
		qbittorrent_torrent_download_speed_bytes = qbittorrent_torrent_download_speed_bytes + `qbittorrent_torrent_download_speed_bytes{name="` + result[i].Name + `",} ` + fmt.Sprintf("%v", result[i].Dlspeed) + "\n"
		qbittorrent_torrent_upload_speed_bytes = qbittorrent_torrent_upload_speed_bytes + `qbittorrent_torrent_upload_speed_bytes{name="` + result[i].Name + `",} ` + fmt.Sprintf("%v", result[i].Upspeed) + "\n"
		qbittorrent_torrent_eta = qbittorrent_torrent_eta + `qbittorrent_torrent_eta{name="` + result[i].Name + `",} ` + fmt.Sprintf("%v", result[i].Eta) + "\n"
		qbittorrent_torrent_progress = qbittorrent_torrent_progress + `qbittorrent_torrent_progress{name="` + result[i].Name + `",} ` + fmt.Sprintf("%v", result[i].Progress) + "\n"
		qbittorrent_torrent_time_active = qbittorrent_torrent_time_active + `qbittorrent_torrent_time_active{name="` + result[i].Name + `",} ` + fmt.Sprintf("%v", result[i].TimeActive) + "\n"
		qbittorrent_torrent_seeders = qbittorrent_torrent_seeders + `qbittorrent_torrent_seeders{name="` + result[i].Name + `",} ` + fmt.Sprintf("%v", result[i].NumSeeds) + "\n"
		qbittorrent_torrent_leechers = qbittorrent_torrent_leechers + `qbittorrent_torrent_leechers{name="` + result[i].Name + `",} ` + fmt.Sprintf("%v", result[i].NumLeechs) + "\n"
		qbittorrent_torrent_ratio = qbittorrent_torrent_ratio + `qbittorrent_torrent_ratio{name="` + result[i].Name + `",} ` + fmt.Sprintf("%v", result[i].Ratio) + "\n"
		qbittorrent_torrent_amount_left_bytes = qbittorrent_torrent_amount_left_bytes + `qbittorrent_torrent_amount_left_bytes{name="` + result[i].Name + `",} ` + fmt.Sprintf("%v", result[i].AmountLeft) + "\n"
		qbittorrent_torrent_size_bytes = qbittorrent_torrent_size_bytes + `qbittorrent_torrent_size_bytes{name="` + result[i].Name + `",} ` + fmt.Sprintf("%v", result[i].Size) + "\n"
	}
	qbittorrent_torrent_states = qbittorrent_torrent_states + `qbittorrent_torrent_states{name="stalledUP",} ` + strconv.Itoa(count_stelledup) + "\n"
	qbittorrent_torrent_states = qbittorrent_torrent_states + `qbittorrent_torrent_states{name="uploading",} ` + strconv.Itoa(count_uploading) + "\n"
	total := qbittorrent_torrent_download_speed_bytes + qbittorrent_torrent_upload_speed_bytes + qbittorrent_torrent_eta + qbittorrent_torrent_progress + qbittorrent_torrent_time_active + qbittorrent_torrent_states + qbittorrent_torrent_seeders + qbittorrent_torrent_leechers + qbittorrent_torrent_ratio + qbittorrent_torrent_amount_left_bytes + qbittorrent_torrent_size_bytes + qbittorrent_torrent_info
	return total
}
