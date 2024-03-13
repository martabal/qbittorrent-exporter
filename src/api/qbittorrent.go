package API

import "encoding/json"

type Info []struct {
	AmountLeft        int     `json:"amount_left"`
	Category          string  `json:"category"`
	Dlspeed           int     `json:"dlspeed"`
	Downloaded        int     `json:"downloaded"`
	DownloadedSession int     `json:"downloaded_session"`
	Eta               int     `json:"eta"`
	Hash              string  `json:"hash"`
	MaxRatio          float64 `json:"max_ratio"`
	Name              string  `json:"name"`
	NumLeechs         int     `json:"num_leechs"`
	NumSeeds          int     `json:"num_seeds"`
	Progress          float64 `json:"progress"`
	Ratio             float64 `json:"ratio"`
	Size              int     `json:"size"`
	State             string  `json:"state"`
	Tags              string  `json:"tags"`
	Tracker           string  `json:"tracker"`
	TimeActive        int     `json:"time_active"`
	Uploaded          int     `json:"uploaded"`
	UploadedSession   int     `json:"uploaded_session"`
	Upspeed           int     `json:"upspeed"`
}

type Preferences struct {
	AltDlLimit         int `json:"alt_dl_limit"`
	AltUpLimit         int `json:"alt_up_limit"`
	DlLimit            int `json:"dl_limit"`
	MaxActiveDownloads int `json:"max_active_downloads"`
	MaxActiveTorrents  int `json:"max_active_torrents"`
	MaxActiveUploads   int `json:"max_active_uploads"`
	UpLimit            int `json:"up_limit"`
}

type Maindata struct {
	CategoryMap map[string]Category `json:"categories"`
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

type Category struct {
	Name     string `json:"name"`
	SavePath string `json:"savePath"`
}

type Trackers []struct {
	Message       string          `json:"msg"`
	NumDownloaded int             `json:"num_downloaded"`
	NumLeeches    int             `json:"num_leeches"`
	NumPeers      int             `json:"num_peers"`
	NumSeeds      int             `json:"num_seeds"`
	Status        int             `json:"status"`
	Tier          json.RawMessage `json:"tier"`
	URL           string          `json:"url"`
}

type Transfer struct {
	ConnectionStatus string `json:"connection_status"`
	DhtNodes         int    `json:"dht_nodes"`
}
