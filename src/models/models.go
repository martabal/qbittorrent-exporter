package models

type Response []struct {
	AddedOn           int     `json:"added_on"`
	AmountLeft        float64 `json:"amount_left"`
	AutoTmm           bool    `json:"auto_tmm"`
	Availability      float64 `json:"availability"`
	Category          string  `json:"category"`
	Completed         int     `json:"completed"`
	CompletionOn      int     `json:"completion_on"`
	ContentPath       string  `json:"content_path"`
	DlLimit           float64 `json:"dl_limit"`
	Dlspeed           float64 `json:"dlspeed"`
	DownloadPath      string  `json:"download_path"`
	Downloaded        float64 `json:"downloaded"`
	DownloadedSession float64 `json:"downloaded_session"`
	Eta               float64 `json:"eta"`
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
	NumLeechs         float64 `json:"num_leechs"`
	NumSeeds          float64 `json:"num_seeds"`
	Priority          int     `json:"priority"`
	Progress          float64 `json:"progress"`
	Ratio             float64 `json:"ratio"`
	RatioLimit        int     `json:"ratio_limit"`
	SavePath          string  `json:"save_path"`
	SeedingTime       int     `json:"seeding_time"`
	SeedingTimeLimit  int     `json:"seeding_time_limit"`
	SeenComplete      int     `json:"seen_complete"`
	SeqDl             bool    `json:"seq_dl"`
	Size              float64 `json:"size"`
	State             string  `json:"state"`
	SuperSeeding      bool    `json:"super_seeding"`
	Tags              string  `json:"tags"`
	TimeActive        float64 `json:"time_active"`
	TotalSize         int     `json:"total_size"`
	Tracker           string  `json:"tracker"`
	TrackersCount     int     `json:"trackers_count"`
	UpLimit           int     `json:"up_limit"`
	Uploaded          float64 `json:"uploaded"`
	UploadedSession   float64 `json:"uploaded_session"`
	Upspeed           float64 `json:"upspeed"`
}

type Preferences struct {
	AddTrackers                      string  `json:"add_trackers"`
	AddTrackersEnabled               bool    `json:"add_trackers_enabled"`
	AltDlLimit                       float64 `json:"alt_dl_limit"`
	AltUpLimit                       float64 `json:"alt_up_limit"`
	AlternativeWebuiEnabled          bool    `json:"alternative_webui_enabled"`
	AlternativeWebuiPath             string  `json:"alternative_webui_path"`
	AnnounceIP                       string  `json:"announce_ip"`
	AnnounceToAllTiers               bool    `json:"announce_to_all_tiers"`
	AnnounceToAllTrackers            bool    `json:"announce_to_all_trackers"`
	AnonymousMode                    bool    `json:"anonymous_mode"`
	AsyncIoThreads                   float64 `json:"async_io_threads"`
	AutoDeleteMode                   float64 `json:"auto_delete_mode"`
	AutoTmmEnabled                   bool    `json:"auto_tmm_enabled"`
	AutorunEnabled                   bool    `json:"autorun_enabled"`
	AutorunOnTorrentAddedEnabled     bool    `json:"autorun_on_torrent_added_enabled"`
	AutorunOnTorrentAddedProgram     string  `json:"autorun_on_torrent_added_program"`
	AutorunProgram                   string  `json:"autorun_program"`
	BannedIPs                        string  `json:"banned_IPs"`
	BittorrentProtocol               int     `json:"bittorrent_protocol"`
	BlockPeersOnPrivilegedPorts      bool    `json:"block_peers_on_privileged_ports"`
	BypassAuthSubnetWhitelist        string  `json:"bypass_auth_subnet_whitelist"`
	BypassAuthSubnetWhitelistEnabled bool    `json:"bypass_auth_subnet_whitelist_enabled"`
	BypassLocalAuth                  bool    `json:"bypass_local_auth"`
	CategoryChangedTmmEnabled        bool    `json:"category_changed_tmm_enabled"`
	CheckingMemoryUse                float64 `json:"checking_memory_use"`
	ConnectionSpeed                  float64 `json:"connection_speed"`
	CurrentInterfaceAddress          string  `json:"current_interface_address"`
	CurrentNetworkInterface          string  `json:"current_network_interface"`
	Dht                              bool    `json:"dht"`
	DiskCache                        float64 `json:"disk_cache"`
	DiskCacheTTL                     float64 `json:"disk_cache_ttl"`
	DiskIoReadMode                   float64 `json:"disk_io_read_mode"`
	DiskIoType                       float64 `json:"disk_io_type"`
	DiskIoWriteMode                  float64 `json:"disk_io_write_mode"`
	DiskQueueSize                    float64 `json:"disk_queue_size"`
	DlLimit                          float64 `json:"dl_limit"`
	DontCountSlowTorrents            bool    `json:"dont_count_slow_torrents"`
	DyndnsDomain                     string  `json:"dyndns_domain"`
	DyndnsEnabled                    bool    `json:"dyndns_enabled"`
	DyndnsPassword                   string  `json:"dyndns_password"`
	DyndnsService                    float64 `json:"dyndns_service"`
	DyndnsUsername                   string  `json:"dyndns_username"`
	EmbeddedTrackerPort              float64 `json:"embedded_tracker_port"`
	EmbeddedTrackerPortForwarding    bool    `json:"embedded_tracker_port_forwarding"`
	EnableCoalesceReadWrite          bool    `json:"enable_coalesce_read_write"`
	EnableEmbeddedTracker            bool    `json:"enable_embedded_tracker"`
	EnableMultiConnectionsFromSameIP bool    `json:"enable_multi_connections_from_same_ip"`
	EnablePieceExtentAffinity        bool    `json:"enable_piece_extent_affinity"`
	EnableUploadSuggestions          bool    `json:"enable_upload_suggestions"`
	Encryption                       float64 `json:"encryption"`
	ExcludedFileNames                string  `json:"excluded_file_names"`
	ExcludedFileNamesEnabled         bool    `json:"excluded_file_names_enabled"`
	ExportDir                        string  `json:"export_dir"`
	ExportDirFin                     string  `json:"export_dir_fin"`
	FilePoolSize                     float64 `json:"file_pool_size"`
	HashingThreads                   float64 `json:"hashing_threads"`
	IdnSupportEnabled                bool    `json:"idn_support_enabled"`
	IncompleteFilesExt               bool    `json:"incomplete_files_ext"`
	IPFilterEnabled                  bool    `json:"ip_filter_enabled"`
	IPFilterPath                     string  `json:"ip_filter_path"`
	IPFilterTrackers                 bool    `json:"ip_filter_trackers"`
	LimitLanPeers                    bool    `json:"limit_lan_peers"`
	LimitTCPOverhead                 bool    `json:"limit_tcp_overhead"`
	LimitUtpRate                     bool    `json:"limit_utp_rate"`
	ListenPort                       float64 `json:"listen_port"`
	Locale                           string  `json:"locale"`
	Lsd                              bool    `json:"lsd"`
	MailNotificationAuthEnabled      bool    `json:"mail_notification_auth_enabled"`
	MailNotificationEmail            string  `json:"mail_notification_email"`
	MailNotificationEnabled          bool    `json:"mail_notification_enabled"`
	MailNotificationPassword         string  `json:"mail_notification_password"`
	MailNotificationSender           string  `json:"mail_notification_sender"`
	MailNotificationSMTP             string  `json:"mail_notification_smtp"`
	MailNotificationSslEnabled       bool    `json:"mail_notification_ssl_enabled"`
	MailNotificationUsername         string  `json:"mail_notification_username"`
	MaxActiveCheckingTorrents        float64 `json:"max_active_checking_torrents"`
	MaxActiveDownloads               float64 `json:"max_active_downloads"`
	MaxActiveTorrents                float64 `json:"max_active_torrents"`
	MaxActiveUploads                 float64 `json:"max_active_uploads"`
	MaxConcurrentHTTPAnnounces       float64 `json:"max_concurrent_http_announces"`
	MaxConnec                        float64 `json:"max_connec"`
	MaxConnecPerTorrent              float64 `json:"max_connec_per_torrent"`
	MaxRatio                         float64 `json:"max_ratio"`
	MaxRatioAct                      float64 `json:"max_ratio_act"`
	MaxRatioEnabled                  bool    `json:"max_ratio_enabled"`
	MaxSeedingTime                   float64 `json:"max_seeding_time"`
	MaxSeedingTimeEnabled            bool    `json:"max_seeding_time_enabled"`
	MaxUploads                       float64 `json:"max_uploads"`
	MaxUploadsPerTorrent             float64 `json:"max_uploads_per_torrent"`
	MemoryWorkingSetLimit            float64 `json:"memory_working_set_limit"`
	OutgoingPortsMax                 float64 `json:"outgoing_ports_max"`
	OutgoingPortsMin                 float64 `json:"outgoing_ports_min"`
	PeerTos                          float64 `json:"peer_tos"`
	PeerTurnover                     float64 `json:"peer_turnover"`
	PeerTurnoverCutoff               float64 `json:"peer_turnover_cutoff"`
	PeerTurnoverInterval             float64 `json:"peer_turnover_interval"`
	PerformanceWarning               bool    `json:"performance_warning"`
	Pex                              bool    `json:"pex"`
	PreallocateAll                   bool    `json:"preallocate_all"`
	ProxyAuthEnabled                 bool    `json:"proxy_auth_enabled"`
	ProxyHostnameLookup              bool    `json:"proxy_hostname_lookup"`
	ProxyIP                          string  `json:"proxy_ip"`
	ProxyPassword                    string  `json:"proxy_password"`
	ProxyPeerConnections             bool    `json:"proxy_peer_connections"`
	ProxyPort                        float64 `json:"proxy_port"`
	ProxyTorrentsOnly                bool    `json:"proxy_torrents_only"`
	ProxyType                        float64 `json:"proxy_type"`
	ProxyUsername                    string  `json:"proxy_username"`
	QueueingEnabled                  bool    `json:"queueing_enabled"`
	RandomPort                       bool    `json:"random_port"`
	ReannounceWhenAddressChanged     bool    `json:"reannounce_when_address_changed"`
	RecheckCompletedTorrents         bool    `json:"recheck_completed_torrents"`
	RefreshInterval                  float64 `json:"refresh_interval"`
	RequestQueueSize                 float64 `json:"request_queue_size"`
	ResolvePeerCountries             bool    `json:"resolve_peer_countries"`
	RssAutoDownloadingEnabled        bool    `json:"rss_auto_downloading_enabled"`
	RssDownloadRepackProperEpisodes  bool    `json:"rss_download_repack_proper_episodes"`
	RssMaxArticlesPerFeed            float64 `json:"rss_max_articles_per_feed"`
	RssProcessingEnabled             bool    `json:"rss_processing_enabled"`
	RssRefreshInterval               float64 `json:"rss_refresh_interval"`
	RssSmartEpisodeFilters           string  `json:"rss_smart_episode_filters"`
	SavePath                         string  `json:"save_path"`
	SavePathChangedTmmEnabled        bool    `json:"save_path_changed_tmm_enabled"`
	SaveResumeDataInterval           float64 `json:"save_resume_data_interval"`
	ScanDirs                         struct {
	} `json:"scan_dirs"`
	ScheduleFromHour                   float64 `json:"schedule_from_hour"`
	ScheduleFromMin                    float64 `json:"schedule_from_min"`
	ScheduleToHour                     float64 `json:"schedule_to_hour"`
	ScheduleToMin                      float64 `json:"schedule_to_min"`
	SchedulerDays                      float64 `json:"scheduler_days"`
	SchedulerEnabled                   bool    `json:"scheduler_enabled"`
	SendBufferLowWatermark             float64 `json:"send_buffer_low_watermark"`
	SendBufferWatermark                float64 `json:"send_buffer_watermark"`
	SendBufferWatermarkFactor          float64 `json:"send_buffer_watermark_factor"`
	SlowTorrentDlRateThreshold         float64 `json:"slow_torrent_dl_rate_threshold"`
	SlowTorrentInactiveTimer           float64 `json:"slow_torrent_inactive_timer"`
	SlowTorrentUlRateThreshold         float64 `json:"slow_torrent_ul_rate_threshold"`
	SocketBacklogSize                  float64 `json:"socket_backlog_size"`
	SsrfMitigation                     bool    `json:"ssrf_mitigation"`
	StartPausedEnabled                 bool    `json:"start_paused_enabled"`
	StopTrackerTimeout                 float64 `json:"stop_tracker_timeout"`
	TempPath                           string  `json:"temp_path"`
	TempPathEnabled                    bool    `json:"temp_path_enabled"`
	TorrentChangedTmmEnabled           bool    `json:"torrent_changed_tmm_enabled"`
	TorrentContentLayout               string  `json:"torrent_content_layout"`
	TorrentStopCondition               string  `json:"torrent_stop_condition"`
	UpLimit                            float64 `json:"up_limit"`
	UploadChokingAlgorithm             float64 `json:"upload_choking_algorithm"`
	UploadSlotsBehavior                float64 `json:"upload_slots_behavior"`
	Upnp                               bool    `json:"upnp"`
	UpnpLeaseDuration                  float64 `json:"upnp_lease_duration"`
	UseCategoryPathsInManualMode       bool    `json:"use_category_paths_in_manual_mode"`
	UseHTTPS                           bool    `json:"use_https"`
	UtpTCPMixedMode                    float64 `json:"utp_tcp_mixed_mode"`
	ValidateHTTPSTrackerCertificate    bool    `json:"validate_https_tracker_certificate"`
	WebUIAddress                       string  `json:"web_ui_address"`
	WebUIBanDuration                   float64 `json:"web_ui_ban_duration"`
	WebUIClickjackingProtectionEnabled bool    `json:"web_ui_clickjacking_protection_enabled"`
	WebUICsrfProtectionEnabled         bool    `json:"web_ui_csrf_protection_enabled"`
	WebUICustomHTTPHeaders             string  `json:"web_ui_custom_http_headers"`
	WebUIDomainList                    string  `json:"web_ui_domain_list"`
	WebUIHostHeaderValidationEnabled   bool    `json:"web_ui_host_header_validation_enabled"`
	WebUIHTTPSCertPath                 string  `json:"web_ui_https_cert_path"`
	WebUIHTTPSKeyPath                  string  `json:"web_ui_https_key_path"`
	WebUIMaxAuthFailCount              float64 `json:"web_ui_max_auth_fail_count"`
	WebUIPort                          float64 `json:"web_ui_port"`
	WebUIReverseProxiesList            string  `json:"web_ui_reverse_proxies_list"`
	WebUIReverseProxyEnabled           bool    `json:"web_ui_reverse_proxy_enabled"`
	WebUISecureCookieEnabled           bool    `json:"web_ui_secure_cookie_enabled"`
	WebUISessionTimeout                float64 `json:"web_ui_session_timeout"`
	WebUIUpnp                          bool    `json:"web_ui_upnp"`
	WebUIUseCustomHTTPHeadersEnabled   bool    `json:"web_ui_use_custom_http_headers_enabled"`
	WebUIUsername                      string  `json:"web_ui_username"`
}

type Maindata struct {
	FullUpdate  bool    `json:"full_update"`
	Rid         float64 `json:"rid"`
	ServerState struct {
		AlltimeDl            float64 `json:"alltime_dl"`
		AlltimeUl            float64 `json:"alltime_ul"`
		AverageTimeQueue     float64 `json:"average_time_queue"`
		ConnectionStatus     string  `json:"connection_status"`
		DhtNodes             float64 `json:"dht_nodes"`
		DlInfoData           float64 `json:"dl_info_data"`
		DlInfoSpeed          float64 `json:"dl_info_speed"`
		DlRateLimit          float64 `json:"dl_rate_limit"`
		FreeSpaceOnDisk      float64 `json:"free_space_on_disk"`
		GlobalRatio          string  `json:"global_ratio"`
		QueuedIoJobs         float64 `json:"queued_io_jobs"`
		Queueing             bool    `json:"queueing"`
		ReadCacheHits        string  `json:"read_cache_hits"`
		ReadCacheOverload    string  `json:"read_cache_overload"`
		RefreshInterval      float64 `json:"refresh_interval"`
		TotalBuffersSize     float64 `json:"total_buffers_size"`
		TotalPeerConnections float64 `json:"total_peer_connections"`
		TotalQueuedSize      float64 `json:"total_queued_size"`
		TotalWastedSession   float64 `json:"total_wasted_session"`
		UpInfoData           float64 `json:"up_info_data"`
		UpInfoSpeed          float64 `json:"up_info_speed"`
		UpRateLimit          float64 `json:"up_rate_limit"`
		UseAltSpeedLimits    bool    `json:"use_alt_speed_limits"`
		WriteCacheOverload   string  `json:"write_cache_overload"`
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
