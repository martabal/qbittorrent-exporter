package models

type Response []struct {
	AmountLeft        float64 `json:"amount_left"`
	Dlspeed           float64 `json:"dlspeed"`
	Downloaded        float64 `json:"downloaded"`
	DownloadedSession float64 `json:"downloaded_session"`
	Eta               float64 `json:"eta"`
	MaxRatio          float64 `json:"max_ratio"`
	Name              string  `json:"name"`
	NumLeechs         float64 `json:"num_leechs"`
	NumSeeds          float64 `json:"num_seeds"`
	Progress          float64 `json:"progress"`
	Ratio             float64 `json:"ratio"`
	Size              float64 `json:"size"`
	State             string  `json:"state"`
	TimeActive        float64 `json:"time_active"`
	Uploaded          float64 `json:"uploaded"`
	UploadedSession   float64 `json:"uploaded_session"`
	Upspeed           float64 `json:"upspeed"`
}

type Preferences struct {
	AltDlLimit         float64 `json:"alt_dl_limit"`
	DlLimit            float64 `json:"dl_limit"`
	MaxActiveDownloads float64 `json:"max_active_downloads"`
	MaxActiveTorrents  float64 `json:"max_active_torrents"`
	MaxActiveUploads   float64 `json:"max_active_uploads"`
	UpLimit            float64 `json:"up_limit"`
}

type Maindata struct {
	ServerState struct {
		AlltimeDl         float64 `json:"alltime_dl"`
		AlltimeUl         float64 `json:"alltime_ul"`
		DlInfoData        float64 `json:"dl_info_data"`
		DlInfoSpeed       float64 `json:"dl_info_speed"`
		GlobalRatio       string  `json:"global_ratio"`
		UpInfoSpeed       float64 `json:"up_info_speed"`
		UseAltSpeedLimits bool    `json:"use_alt_speed_limits"`
	} `json:"server_state"`
}
