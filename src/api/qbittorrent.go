package API

import "encoding/json"

type Info []struct {
	AmountLeft        int64   `json:"amount_left"`
	Category          string  `json:"category"`
	Dlspeed           int64   `json:"dlspeed"`
	Downloaded        int64   `json:"downloaded"`
	DownloadedSession int64   `json:"downloaded_session"`
	Eta               int64   `json:"eta"`
	Hash              string  `json:"hash"`
	MaxRatio          float64 `json:"max_ratio"`
	Name              string  `json:"name"`
	NumLeechs         int64   `json:"num_leechs"`
	NumSeeds          int64   `json:"num_seeds"`
	Progress          float64 `json:"progress"`
	Ratio             float64 `json:"ratio"`
	Size              int64   `json:"size"`
	State             string  `json:"state"`
	Tags              string  `json:"tags"`
	Tracker           string  `json:"tracker"`
	TimeActive        int64   `json:"time_active"`
	Uploaded          int64   `json:"uploaded"`
	UploadedSession   int64   `json:"uploaded_session"`
	Upspeed           int64   `json:"upspeed"`
}

type Preferences struct {
	AltDlLimit         int64 `json:"alt_dl_limit"`
	AltUpLimit         int64 `json:"alt_up_limit"`
	DlLimit            int64 `json:"dl_limit"`
	MaxActiveDownloads int64 `json:"max_active_downloads"`
	MaxActiveTorrents  int64 `json:"max_active_torrents"`
	MaxActiveUploads   int64 `json:"max_active_uploads"`
	UpLimit            int64 `json:"up_limit"`
}

type MainData struct {
	CategoryMap map[string]Category `json:"categories"`
	ServerState struct {
		AlltimeDl         int64  `json:"alltime_dl"`
		AlltimeUl         int64  `json:"alltime_ul"`
		DlInfoData        int64  `json:"dl_info_data"`
		DlInfoSpeed       int64  `json:"dl_info_speed"`
		GlobalRatio       string `json:"global_ratio"`
		UpInfoData        int64  `json:"up_info_data"`
		UpInfoSpeed       int64  `json:"up_info_speed"`
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
