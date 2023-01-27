package models

type Response []struct {
	AddedOn           int     `json:"added_on"`
	AmountLeft        int     `json:"amount_left"`
	AutoTmm           bool    `json:"auto_tmm"`
	Availability      float64 `json:"availability"`
	Category          string  `json:"category"`
	Completed         int     `json:"completed"`
	CompletionOn      int     `json:"completion_on"`
	ContentPath       string  `json:"content_path"`
	DlLimit           float64 `json:"dl_limit"`
	Dlspeed           float64 `json:"dlspeed"`
	DownloadPath      string  `json:"download_path"`
	Downloaded        int     `json:"downloaded"`
	DownloadedSession int     `json:"downloaded_session"`
	Eta               int     `json:"eta"`
	FLPiecePrio       bool    `json:"f_l_piece_prio"`
	ForceStart        bool    `json:"force_start"`
	Hash              string  `json:"hash"`
	InfohashV1        string  `json:"infohash_v1"`
	InfohashV2        string  `json:"infohash_v2"`
	LastActivity      int     `json:"last_activity"`
	MagnetURI         string  `json:"magnet_uri"`
	MaxRatio          float64 `json:"max_ratio"`
	MaxSeedingTime    float64 `json:"max_seeding_time"`
	Name              string  `json:"name"`
	NumComplete       int     `json:"num_complete"`
	NumIncomplete     int     `json:"num_incomplete"`
	NumLeechs         int     `json:"num_leechs"`
	NumSeeds          int     `json:"num_seeds"`
	Priority          int     `json:"priority"`
	Progress          float64 `json:"progress"`
	Ratio             float64 `json:"ratio"`
	RatioLimit        int     `json:"ratio_limit"`
	SavePath          string  `json:"save_path"`
	SeedingTime       int     `json:"seeding_time"`
	SeedingTimeLimit  int     `json:"seeding_time_limit"`
	SeenComplete      int     `json:"seen_complete"`
	SeqDl             bool    `json:"seq_dl"`
	Size              int     `json:"size"`
	State             string  `json:"state"`
	SuperSeeding      bool    `json:"super_seeding"`
	Tags              string  `json:"tags"`
	TimeActive        int     `json:"time_active"`
	TotalSize         int     `json:"total_size"`
	Tracker           string  `json:"tracker"`
	TrackersCount     int     `json:"trackers_count"`
	UpLimit           int     `json:"up_limit"`
	Uploaded          int     `json:"uploaded"`
	UploadedSession   int     `json:"uploaded_session"`
	Upspeed           int     `json:"upspeed"`
}

type Info struct {
	Bitness    int64  `json:"bitness"`
	Boost      string `json:"boost"`
	Libtorrent string `json:"libtorrent"`
	Openssl    string `json:"openssl"`
	Qt         string `json:"qt"`
	Zlib       string `json:"zlib"`
}

type User struct {
	Username string
	Password string
}

type Request struct {
	Base_url string
	Cookie   string
}

func mask(input string) string {
	hide := ""
	for i := 0; i < len(input); i++ {
		hide += "*"
	}
	return hide
}

var myuser User
var myrequest Request

func Getrequest() (string, string) {
	return myrequest.Base_url, myrequest.Cookie
}

func Setrequest(base_url string, cookie string) {
	myrequest.Base_url = base_url
	myrequest.Cookie = cookie
}
func Setbaseurl(base_url string) {
	myrequest.Base_url = base_url

}
func Getbaseurl() string {
	return myrequest.Base_url

}
func Getcookie() string {
	return myrequest.Cookie

}
func Setcookie(cookie string) {
	myrequest.Cookie = cookie
}

func Getuser() (string, string) {
	return myuser.Username, myuser.Password
}

func Setuser(username string, password string) {
	myuser.Username = username
	myuser.Password = password
}

func GetUsername() string {
	return myuser.Username
}

func Getpassword() string {
	return myuser.Password
}
func Getpasswordmasked() string {
	return mask(myuser.Password)
}
