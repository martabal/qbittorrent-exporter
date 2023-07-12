package models

type TypeInfo []struct {
	AmountLeft        int     `json:"amount_left"`
	Category          string  `json:"category"`
	Dlspeed           int     `json:"dlspeed"`
	Downloaded        int     `json:"downloaded"`
	DownloadedSession int     `json:"downloaded_session"`
	Eta               int     `json:"eta"`
	MaxRatio          float64 `json:"max_ratio"`
	Name              string  `json:"name"`
	NumLeechs         int     `json:"num_leechs"`
	NumSeeds          int     `json:"num_seeds"`
	Progress          float64 `json:"progress"`
	Ratio             float64 `json:"ratio"`
	Size              int     `json:"size"`
	State             string  `json:"state"`
	Tags              string  `json:"tags"`
	TimeActive        int     `json:"time_active"`
	Uploaded          int     `json:"uploaded"`
	UploadedSession   int     `json:"uploaded_session"`
	Upspeed           int     `json:"upspeed"`
}

type TypePreferences struct {
	AltDlLimit         int `json:"alt_dl_limit"`
	AltUpLimit         int `json:"alt_up_limit"`
	DlLimit            int `json:"dl_limit"`
	MaxActiveDownloads int `json:"max_active_downloads"`
	MaxActiveTorrents  int `json:"max_active_torrents"`
	MaxActiveUploads   int `json:"max_active_uploads"`
	UpLimit            int `json:"up_limit"`
}

type TypeMaindata struct {
	CategoryMap map[string]TypeCategory `json:"categories"`
	ServerState struct {
		AlltimeDl         int    `json:"alltime_dl"`
		AlltimeUl         int    `json:"alltime_ul"`
		DlInfoData        int    `json:"dl_info_data"`
		DlInfoSpeed       int    `json:"dl_info_speed"`
		GlobalRatio       string `json:"global_ratio"`
		UpInfoData        int    `json:"up_info_data"`
		UpInfoSpeed       int    `json:"up_info_speed"`
		UseAltSpeedLimits bool   `json:"use_alt_speed_limits"`
	} `json:"server_state"`
	Tags []string `json:"tags"`
}

type TypeCategory struct {
	Name     string `json:"name"`
	SavePath string `json:"savePath"`
}
