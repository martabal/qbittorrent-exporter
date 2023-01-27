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
	FullUpdate  bool    `json:"full_update"`
	Rid         float64 `json:"rid"`
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
var myerr bool

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

func SetPromptError(prompt bool) {
	myerr = prompt
}

func GetPromptError() bool {
	return myerr
}
