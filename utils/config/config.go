package config

import (
	"bufio"
	"encoding/base32"
	"flag"
	"fmt"
	"log/slog"
	"net/netip"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/authclients"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/geoip"
	"nighttroop.t1d.nz/utils/security"
	"nighttroop.t1d.nz/utils/usage"
)

// DefaultSupportEmail is the support email address when SUPPORT_EMAIL is not set
const DefaultSupportEmail = "nighttroop@t1d.nz"

// ServerStatus records which optional services are working. It is filled in
// at startup and read by anything that depends on them.
type ServerStatus struct {
	S3Working     bool
	SentryWorking bool
	// HighResources are the server's resources at or over their alert levels at its last check
	HighResources []ResourceLevel
	// Sites is how many sites the server has, and MaxSites how many MAX_SITES lets the public create
	Sites    int
	MaxSites int
}

// SitesNearLimit is whether the server has 90% or more of the sites MAX_SITES allows
func (s ServerStatus) SitesNearLimit() bool {
	return s.MaxSites > 0 && s.Sites*10 >= s.MaxSites*9
}

// SitesAtLimit is whether the server has as many sites as MAX_SITES allows, or more
func (s ServerStatus) SitesAtLimit() bool {
	return s.MaxSites > 0 && s.Sites >= s.MaxSites
}

// ResourceLevel is one of the server's resources at a check: memory or disk, its percent used of its total bytes, or
// cpu, the seconds it was at 100%, and its alert level
type ResourceLevel struct {
	Resource string   `json:"resource"`
	Value    float64  `json:"value"`
	Limit    float64  `json:"limit"`
	Total    *float64 `json:"total,omitempty"`
}

type ServerConfig struct {
	Mode consts.ServerMode

	Status ServerStatus

	HostName     string
	HostPort     int
	HostProtocol string // http or https, used for generating links.
	// MaxRequestBodyBytes is the largest request body read, so a huge upload cannot exhaust memory while it is decoded
	MaxRequestBodyBytes int64

	// Some common Host URLs cached here so we dont fmt.Sprintf them all the time
	HostNameWithPort string
	HostURL          string
	HostURLWithPort  string

	// File system
	DataDir string

	// s3 config
	S3Bucket          string
	S3URL             string
	S3AccessKeyID     string
	S3AccessSecretKey string

	// sentry
	SentryDSN string

	// Deployment names this server among those running Nighttroop, e.g. one of several staging servers, as Sentry's
	// environment its errors are reported under
	Deployment string

	// SupportEmail is where users contact the server's operator, linked from every page's footer and in the terms
	// of service and privacy policy
	SupportEmail string

	// FeedbackURL, when set, is where users give feedback, e.g. a Google Form or GitHub issues, linked from every
	// page's footer and the home page
	FeedbackURL string

	// TestSite shows every page a large banner saying it is a test site whose data is wiped, so it is not taken for
	// a real one: a staging server always is
	TestSite bool

	// StagingAdminPassword and StagingAdminTOTPSecret, in staging, are the control admin made when it starts without
	// one, so wiping its data never needs it set up again. The TOTP secret is base32, as authenticator apps take it.
	StagingAdminPassword   string
	StagingAdminTOTPSecret string

	// GeoIPDBPath is a DB-IP IP to City Lite database, for the city sign ins are from; without one only the country is
	GeoIPDBPath string

	// AllowedCountries are the only countries, by ISO 3166-1 alpha-2 code, that can create sites, sign in and
	// recover; empty allows every country, and an address whose country is not known is allowed. To limit a server,
	// copy every country DB-IP knows, below, into ALLOWED_COUNTRIES and remove the ones it is not for:
	//   AD,AE,AF,AG,AI,AL,AM,AO,AQ,AR,AS,AT,AU,AW,AX,AZ,BA,BB,BD,BE,BF,BG,BH,BI,BJ,BL,BM,BN,BO,BQ,BR,BS,BT,BV,BW,BY,BZ,
	//   CA,CC,CD,CF,CG,CH,CI,CK,CL,CM,CN,CO,CR,CU,CV,CW,CX,CY,CZ,DE,DJ,DK,DM,DO,DZ,EC,EE,EG,EH,ER,ES,ET,FI,FJ,FK,FM,FO,
	//   FR,GA,GB,GD,GE,GF,GG,GH,GI,GL,GM,GN,GP,GQ,GR,GS,GT,GU,GW,GY,HK,HM,HN,HR,HT,HU,ID,IE,IL,IM,IN,IO,IQ,IR,IS,IT,JE,
	//   JM,JO,JP,KE,KG,KH,KI,KM,KN,KP,KR,KW,KY,KZ,LA,LB,LC,LI,LK,LR,LS,LT,LU,LV,LY,MA,MC,MD,ME,MF,MG,MH,MK,ML,MM,MN,MO,
	//   MP,MQ,MR,MS,MT,MU,MV,MW,MX,MY,MZ,NA,NC,NE,NF,NG,NI,NL,NO,NP,NR,NU,NZ,OM,PA,PE,PF,PG,PH,PK,PL,PM,PN,PR,PS,PT,PW,
	//   PY,QA,RE,RO,RS,RU,RW,SA,SB,SC,SD,SE,SG,SH,SI,SJ,SK,SL,SM,SN,SO,SR,SS,ST,SV,SX,SY,SZ,TC,TD,TF,TG,TH,TJ,TK,TL,TM,
	//   TN,TO,TR,TT,TV,TW,TZ,UA,UG,UM,US,UY,UZ,VA,VC,VE,VG,VI,VN,VU,WF,WS,XK,YE,YT,ZA,ZM,ZW
	AllowedCountries []string

	// Unauthenticated Rate Limits
	GlobalRateLimit int
	GlobalRateBurst int

	GlobalIPPerMinuteLimit int
	GlobalIPPerMinuteBurst int

	SiteLoginPerMinuteLimit int
	SiteLoginPerMinuteBurst int

	SiteUnauthPerMinuteLimit int
	SiteUnauthPerMinuteBurst int

	// Authenticated Rate Limits
	SiteGETPerMinuteLimit        int
	SiteGETPerMinuteBurst        int
	SitePOSTPerMinuteLimit       int
	SitePOSTPerMinuteBurst       int
	SiteUpdateSyncPerMinuteLimit int
	SiteUpdateSyncPerMinuteBurst int

	// Each identity create hashes a password with bcrypt
	SiteIdentityCreatePerMinuteLimit int
	SiteIdentityCreatePerMinuteBurst int

	// Each download of a site's data copies its whole DB
	SiteDownloadPerMinuteLimit int
	SiteDownloadPerMinuteBurst int
	// SitePushPerMinuteLimit and Burst cap the pushes a site sends, each a request to Apple or Firebase
	SitePushPerMinuteLimit int
	SitePushPerMinuteBurst int

	//
	// Site Create Rate Limits
	SiteCreateGlobalPerMinuteLimit int
	SiteCreateGlobalPerMinuteBurst int
	SiteCreateIPPerMinuteLimit     float64
	SiteCreateIPPerMinuteBurst     int

	//
	// Site Limits
	MaxSites             int
	MaxIdentitiesPerSite int

	DeleteAfterInactivityMins int
	// DisableAfterInactivityDays is how long a site goes unwritten or unseen before the control plane can disable it
	DisableAfterInactivityDays int
	// TruncateRecordsAfterDays is how old a site's records are before truncating it clears their raw JSON
	TruncateRecordsAfterDays int

	// MemoryAlertPercent and DiskAlertPercent are the memory and DB disk used, and CPUPeggedAlertSeconds how long the
	// CPU is at 100%, at or over which a check of the server's resources warns the control site; disk only warns while
	// under DiskAlertFreeGB is free
	MemoryAlertPercent    int
	DiskAlertPercent      int
	DiskAlertFreeGB       int
	CPUPeggedAlertSeconds int

	// Backup Config
	ControlBackupIntervalInSeconds int
	BatchBackupIntervalInSeconds   int
	// SiteBackupRetentionDays is how long S3 keeps a site batch backup or archive after a newer upload replaces it
	SiteBackupRetentionDays int
	// ControlBackupRetentionDays is how long S3 keeps a control backup after a newer upload replaces it
	ControlBackupRetentionDays int

	// Sync Config
	BackgroundSyncWorkers int
	SyncIntervalInSeconds int
	// BackfillDays is how far back before a site's oldest glucose reading a backfill fetches records
	BackfillDays int

	// Truncate Config
	TruncateIntervalInSeconds int
	// LogsDBMaxRowsPerTable is how many of the newest logs, events and errors are kept in each of their tables
	LogsDBMaxRowsPerTable int

	// SlowQueryMS is how many milliseconds a DB query runs before it is logged as slow, with its name; 0 logs none
	SlowQueryMS int
	// QuerySamplePercent is how many in a hundred DB queries are kept in the usage page's slowest queries
	QuerySamplePercent int

	// Socket.IO Config, what one site's live update sessions may hold and send
	SocketPingIntervalInSeconds int
	SocketPingTimeoutInSeconds  int
	SocketMaxPayloadBytes       int
	SocketMaxSessionsPerSite    int
	SocketMaxBufferedPackets    int
	SocketMaxBufferedBytes      int
	SocketMaxPacketsPerSecond   float64

	RateLimiters *RateLimiters
	// RecentClients is where each identity recently authenticated from, for its site's admin to see
	RecentClients *authclients.Recent
	// GeoIP estimates where a client's IP address is, as recorded on its sign ins
	GeoIP *geoip.Locator

	// Events is every event on the server, by site ID: each change to a site's models, for live updates and alarms,
	// and each one done to a site or the server, which the logs DB also keeps
	Events *events.SiteBus

	// Usage counts how the server is used since it started, for control's usage page
	Usage *usage.Tracker

	// These keys are used to encrypt secrets in the DB and S3 at rest
	// This has the option to add additional keys
	Security   *security.SecurityConfig
	RawKeySets string

	// TrustedProxies are the reverse proxies, such as Caddy in front of the server, whose X-Forwarded-For is believed
	TrustedProxies []netip.Prefix
}

const SITE_GET_LIMIT = 180
const SITE_POST_LIMIT = 60

// Priority is
// 1. Args
// 2. ENVAR
// 4. .env
// 5. Default
func NewServerConfig() (*ServerConfig, error) {
	mode, dotenv, err := readFlagsAndDotEnv()
	if err != nil {
		return nil, err
	}

	setupLogging(mode)

	serverConfig, err := readServerConfig(mode, dotenv)
	if err != nil {
		return nil, err
	}

	if err := rejectUnknownVariables(dotenv); err != nil {
		return nil, err
	}
	if err := serverConfig.ValidateClean(); err != nil {
		return nil, err
	}

	serverConfig.setHostURLs()
	serverConfig.startServices()

	printServerConfig(serverConfig)
	return &serverConfig, nil
}

// readFlagsAndDotEnv are the server's mode and the variables from its flags over its mode's .env file
func readFlagsAndDotEnv() (consts.ServerMode, map[string]string, error) {
	flagsnev := compileFlags()

	mode := findMode(flagsnev)
	dotenv, err := compileDotEnvFile(mode)
	if err != nil {
		return mode, nil, err
	}

	dotenv = MergeMaps(dotenv, flagsnev) // flags override .env
	// MODE has already been read by findMode, so it is not left over as an unknown variable
	delete(dotenv, "MODE")
	return mode, dotenv, nil
}

// readServerConfig reads each setting from the environment, then dotenv, then its default, removing what it reads
// from dotenv, so the variables left in it are unknown
func readServerConfig(mode consts.ServerMode, dotenv map[string]string) (ServerConfig, error) {
	// DEFAULTS
	serverConfig := ServerConfig{
		Mode: mode,

		// Host config
		HostName:     configStringSelect(dotenv, "HOST_NAME", "nt.local.t1d.nz"),
		HostPort:     configIntSelect(dotenv, "HOST_PORT", 3000),
		HostProtocol: configStringSelect(dotenv, "HOST_PROTOCOL", "https"),
		// Far above an uploader's batch of 100 records
		MaxRequestBodyBytes: int64(configIntSelect(dotenv, "MAX_REQUEST_BODY_BYTES", 4000000)),

		// Filesystem
		DataDir: configStringSelect(dotenv, "DATA_DIR", "./.dbs"),

		// s3 config
		S3Bucket:          configStringSelect(dotenv, "S3_BUCKET", "mybucket"),
		S3URL:             configStringSelect(dotenv, "S3_URL", ""),
		S3AccessKeyID:     configStringSelect(dotenv, "S3_ACCESS_KEY", "minioadmin"),
		S3AccessSecretKey: configStringSelect(dotenv, "S3_SECRET_KEY", "minioadmin"),

		// sentry
		SentryDSN:  configStringSelect(dotenv, "SENTRY_DSN", ""),
		Deployment: configStringSelect(dotenv, "DEPLOYMENT", "development"),

		SupportEmail: configStringSelect(dotenv, "SUPPORT_EMAIL", DefaultSupportEmail),
		FeedbackURL:  configStringSelect(dotenv, "FEEDBACK_URL", ""),
		TestSite:     mode == consts.Staging,

		StagingAdminPassword:   configStringSelect(dotenv, "STAGING_ADMIN_PASSWORD", ""),
		StagingAdminTOTPSecret: configStringSelect(dotenv, "STAGING_ADMIN_TOTP_SECRET", ""),
		GeoIPDBPath:  configStringSelect(dotenv, "GEOIP_DB_PATH", ""),

		// security
		RawKeySets: configStringSelect(dotenv, consts.KEY_SETS_ENV_KEY, ""),

		// RATE LIMITERS, from estimates of real use rather than measured traffic

		// Public RateLimits
		GlobalRateLimit:        configIntSelect(dotenv, "GLOBAL_RATE_LIMIT", 100000), // Stop DDOS doing downloads
		GlobalRateBurst:        configIntSelect(dotenv, "GLOBAL_RATE_BURST", 100000),
		GlobalIPPerMinuteLimit: configIntSelect(dotenv, "GLOBAL_IP_PER_MINUTE_LIMIT", SITE_GET_LIMIT*4), // Single IP can fully manage 4 sites
		GlobalIPPerMinuteBurst: configIntSelect(dotenv, "GLOBAL_IP_PER_MINUTE_BURST", SITE_GET_LIMIT*8),

		SiteLoginPerMinuteLimit: configIntSelect(dotenv, "SITE_LOGIN_PER_MINUTE_LIMIT", 1), // Stop credential stuffing
		SiteLoginPerMinuteBurst: configIntSelect(dotenv, "SITE_LOGIN_PER_MINUTE_BURST", 10),

		SiteUnauthPerMinuteLimit: configIntSelect(dotenv, "SITE_UNAUTH_PER_MINUTE_LIMIT", 60), // What one IP may send a site before it authenticates, so bad tokens cannot spend the site's limit
		SiteUnauthPerMinuteBurst: configIntSelect(dotenv, "SITE_UNAUTH_PER_MINUTE_BURST", 120),

		// Site Create Rate Limits
		SiteCreateGlobalPerMinuteLimit: configIntSelect(dotenv, "SITE_CREATE_GLOBAL_PER_MINUTE_LIMIT", 7), // If maxed out this filles up in a day
		SiteCreateGlobalPerMinuteBurst: configIntSelect(dotenv, "SITE_CREATE_GLOBAL_PER_MINUTE_BURST", 20),
		SiteCreateIPPerMinuteLimit:     configFloatSelect(dotenv, "SITE_CREATE_IP_PER_MINUTE_LIMIT", 0.1), // 1 event every 10 minutes
		SiteCreateIPPerMinuteBurst:     configIntSelect(dotenv, "SITE_CREATE_IP_PER_MINUTE_BURST", 1),

		// Authed
		// Site Rate Limits
		SiteGETPerMinuteLimit:        configIntSelect(dotenv, "SITE_GET_PER_MINUTE_LIMIT", SITE_GET_LIMIT),
		SiteGETPerMinuteBurst:        configIntSelect(dotenv, "SITE_GET_PER_MINUTE_BURST", SITE_GET_LIMIT*2),
		SitePOSTPerMinuteLimit:       configIntSelect(dotenv, "SITE_POST_PER_MINUTE_LIMIT", SITE_POST_LIMIT),
		SitePOSTPerMinuteBurst:       configIntSelect(dotenv, "SITE_POST_PER_MINUTE_BURST", SITE_POST_LIMIT*4),
		SiteUpdateSyncPerMinuteLimit: configIntSelect(dotenv, "SITE_UPDATE_SYNC_PER_MINUTE_LIMIT", 1), // Stop sites from updating sync settings too much
		SiteUpdateSyncPerMinuteBurst: configIntSelect(dotenv, "SITE_UPDATE_SYNC_PER_MINUTE_BURST", 4),

		SiteIdentityCreatePerMinuteLimit: configIntSelect(dotenv, "SITE_IDENTITY_CREATE_PER_MINUTE_LIMIT", 1), // Stop creating and deleting identities using up the CPU with bcrypt
		SiteIdentityCreatePerMinuteBurst: configIntSelect(dotenv, "SITE_IDENTITY_CREATE_PER_MINUTE_BURST", 5),

		SiteDownloadPerMinuteLimit: configIntSelect(dotenv, "SITE_DOWNLOAD_PER_MINUTE_LIMIT", 1),
		SiteDownloadPerMinuteBurst: configIntSelect(dotenv, "SITE_DOWNLOAD_PER_MINUTE_BURST", 3),
		SitePushPerMinuteLimit:     configIntSelect(dotenv, "SITE_PUSH_PER_MINUTE_LIMIT", 10),
		SitePushPerMinuteBurst:     configIntSelect(dotenv, "SITE_PUSH_PER_MINUTE_BURST", 30),

		//
		// Site Limits
		MaxSites:             configIntSelect(dotenv, "MAX_SITES", 10000),
		MaxIdentitiesPerSite: configIntSelect(dotenv, "MAX_IDENTITIES_PER_USER", 10),

		// Truncate Config
		TruncateIntervalInSeconds: configIntSelect(dotenv, "TRUNCATE_INTERVAL_IN_SECONDS", 60),
		LogsDBMaxRowsPerTable:     configIntSelect(dotenv, "LOGS_DB_MAX_ROWS_PER_TABLE", 100000),
		SlowQueryMS:               configIntSelect(dotenv, "SLOW_QUERY_MS", 100),
		QuerySamplePercent:        configIntSelect(dotenv, "QUERY_SAMPLE_PERCENT", 10),

		// Archiving limits
		DeleteAfterInactivityMins:  configIntSelect(dotenv, "DELETE_AFTER_INACTIVITY_MINS", 60*24*7), // 1 week
		DisableAfterInactivityDays: configIntSelect(dotenv, "DISABLE_AFTER_INACTIVITY_DAYS", 7),
		TruncateRecordsAfterDays:   configIntSelect(dotenv, "TRUNCATE_RECORDS_AFTER_DAYS", 30),

		// Resource alerts
		MemoryAlertPercent:    configIntSelect(dotenv, "MEMORY_ALERT_PERCENT", 60),
		DiskAlertPercent:      configIntSelect(dotenv, "DISK_ALERT_PERCENT", 60),
		DiskAlertFreeGB:       configIntSelect(dotenv, "DISK_ALERT_FREE_GB", 100),
		CPUPeggedAlertSeconds: configIntSelect(dotenv, "CPU_PEGGED_ALERT_SECONDS", 20),

		// Backup Config
		ControlBackupIntervalInSeconds: configIntSelect(dotenv, "CONTROL_BACKUP_INTERVAL_IN_SECONDS", 30),
		BatchBackupIntervalInSeconds:   configIntSelect(dotenv, "BATCH_BACKUP_INTERVAL_IN_SECONDS", 300), // 5 minutes
		SiteBackupRetentionDays:        configIntSelect(dotenv, "SITE_BACKUP_RETENTION_DAYS", 30),
		ControlBackupRetentionDays:     configIntSelect(dotenv, "CONTROL_BACKUP_RETENTION_DAYS", 90),

		// Sync Config
		BackgroundSyncWorkers: configIntSelect(dotenv, "BACKGROUND_SYNC_WORKERS", 5),
		SyncIntervalInSeconds: configIntSelect(dotenv, "SYNC_INTERVAL_IN_SECONDS", 5), // check syncs every 5 seconds
		BackfillDays:          configIntSelect(dotenv, "BACKFILL_DAYS", 7),

		// Socket.IO Config
		SocketPingIntervalInSeconds: configIntSelect(dotenv, "SOCKET_PING_INTERVAL_IN_SECONDS", 25), // Nightscout's, which v2 clients expect
		SocketPingTimeoutInSeconds:  configIntSelect(dotenv, "SOCKET_PING_TIMEOUT_IN_SECONDS", 20),
		SocketMaxPayloadBytes:       configIntSelect(dotenv, "SOCKET_MAX_PAYLOAD_BYTES", 100000),  // The largest POST body or websocket message
		SocketMaxSessionsPerSite:    configIntSelect(dotenv, "SOCKET_MAX_SESSIONS_PER_SITE", 100), // Stop one site holding every connection
		SocketMaxBufferedPackets:    configIntSelect(dotenv, "SOCKET_MAX_BUFFERED_PACKETS", 1000), // A client too slow to collect its packets is closed
		SocketMaxBufferedBytes:      configIntSelect(dotenv, "SOCKET_MAX_BUFFERED_BYTES", 2000000),
		SocketMaxPacketsPerSecond:   configFloatSelect(dotenv, "SOCKET_MAX_PACKETS_PER_SECOND", 20), // A client sending faster, past a burst of 5 seconds' worth, is closed
	}

	var err error
	serverConfig.TrustedProxies, err = ParseTrustedProxies(configStringSelect(dotenv, "TRUSTED_PROXIES", "127.0.0.1,::1"))
	if err != nil {
		return serverConfig, err
	}
	// A list that does not read fails the server, rather than letting every country in
	serverConfig.AllowedCountries, err = parseCountries(configStringSelect(dotenv, "ALLOWED_COUNTRIES", ""))
	return serverConfig, err
}

// rejectUnknownVariables fails if any variable is left in dotenv after the config is read, e.g. a misspelt one
func rejectUnknownVariables(dotenv map[string]string) error {
	if len(dotenv) > 0 {
		slog.Error("Unknown .env variables: ", "variables", func() []string {
			var keys []string
			for k := range dotenv {
				keys = append(keys, k)
			}
			return keys
		}())
		return fmt.Errorf("unknown .env variables")
	}
	return nil
}

// setHostURLs precomputes the host's URLs from its protocol, name and port
func (config *ServerConfig) setHostURLs() {
	config.HostNameWithPort = fmt.Sprintf("%s:%d", config.HostName, config.HostPort)
	config.HostURL = fmt.Sprintf("%s://%s", config.HostProtocol, config.HostName)
	config.HostURLWithPort = fmt.Sprintf("%s://%s:%d", config.HostProtocol, config.HostName, config.HostPort)
}

// startServices builds what the server shares across requests from the config, e.g. its rate limiters and keys
func (config *ServerConfig) startServices() {
	config.RateLimiters = newRateLimiters(*config)
	config.RecentClients = authclients.New(10_000)
	config.GeoIP = geoip.New(config.GeoIPDBPath)
	config.Events = events.NewSiteBus()
	config.Usage = usage.New()

	config.Security = security.NewSecurityConfig(config.RawKeySets, config.Mode, config.HostName, config.IsHTTPS())
}

func (config *ServerConfig) IsDev() bool {
	return config.Mode == consts.Development
}

func (config *ServerConfig) IsProd() bool {
	return config.Mode == consts.Production
}

func (config *ServerConfig) IsTest() bool {
	return config.Mode == consts.Test
}

func (config *ServerConfig) IsStaging() bool {
	return config.Mode == consts.Staging
}

func (config *ServerConfig) IsDevOrTest() bool {
	return config.IsDev() || config.IsTest()
}

func (config *ServerConfig) DBDir() string {
	return fmt.Sprintf("%s/db", config.DataDir)
}

func (config *ServerConfig) BackupDir() string {
	return fmt.Sprintf("%s/backup", config.DataDir)
}

func (config *ServerConfig) GetBcryptCost() int {
	if config.IsDevOrTest() {
		return bcrypt.MinCost
	}
	return bcrypt.DefaultCost
}

func (config *ServerConfig) ValidateClean() error {
	config.HostProtocol = strings.ToLower(config.HostProtocol)
	if config.HostProtocol != "http" && config.HostProtocol != "https" {
		return fmt.Errorf("invalid host protocol: %s", config.HostProtocol)
	}

	if config.HostProtocol == "http" && !config.IsDev() && !config.IsTest() {
		return fmt.Errorf("http is only allowed in development or test mode")
	}

	if config.IsStaging() {
		return validateStagingAdmin(config.StagingAdminPassword, config.StagingAdminTOTPSecret)
	}

	return nil
}

// validateStagingAdmin refuses a staging admin whose password is short, or whose TOTP secret is not base32 of at
// least 16 characters (80 bits), the least authenticator apps take
func validateStagingAdmin(password, totpSecret string) error {
	if len(password) < 16 {
		return fmt.Errorf("staging needs STAGING_ADMIN_PASSWORD of at least 16 characters")
	}
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.TrimRight(totpSecret, "=")))
	if err != nil || len(key) < 10 {
		return fmt.Errorf("staging needs STAGING_ADMIN_TOTP_SECRET of at least 16 base32 characters")
	}
	return nil
}

// parseCountries reads comma separated ISO 3166-1 alpha-2 country codes, e.g. "NZ,AU"
func parseCountries(raw string) ([]string, error) {
	countries := []string{}
	for _, part := range strings.Split(raw, ",") {
		code := strings.ToUpper(strings.TrimSpace(part))
		if code == "" {
			continue
		}
		if len(code) != 2 || code[0] < 'A' || code[0] > 'Z' || code[1] < 'A' || code[1] > 'Z' {
			return nil, fmt.Errorf("invalid country code in ALLOWED_COUNTRIES: %q", part)
		}
		countries = append(countries, code)
	}
	return countries, nil
}

// configSelect is envVar parsed from the environment, then dotenv, then the default. A value that does not parse
// keeps the default, as e.g. a 0 would turn a limit off. envVar is deleted from dotenv so we can find unknown keys later
func configSelect[T any](dotenv map[string]string, envVar string, defaultValue T, parse func(string) (T, error)) T {
	raw, ok := dotenv[envVar]
	delete(dotenv, envVar)
	if val, inEnv := os.LookupEnv(envVar); inEnv {
		raw, ok = val, true
	}
	if !ok {
		return defaultValue
	}

	value, err := parse(raw)
	if err != nil {
		slog.Error("Cannot parse config value, using the default", "envVar", envVar, "value", raw, "default", defaultValue, "error", err)
		return defaultValue
	}
	return value
}

func configStringSelect(dotenv map[string]string, envVar string, defaultValue string) string {
	return configSelect(dotenv, envVar, defaultValue, func(raw string) (string, error) { return raw, nil })
}

func configIntSelect(dotenv map[string]string, envVar string, defaultValue int) int {
	return configSelect(dotenv, envVar, defaultValue, strconv.Atoi)
}

func configFloatSelect(dotenv map[string]string, envVar string, defaultValue float64) float64 {
	return configSelect(dotenv, envVar, defaultValue, func(raw string) (float64, error) { return strconv.ParseFloat(raw, 64) })
}

// 1. check explicit settings, the flag and envar
// 2. Guess, if go is running test
// 3. return default dev
func findMode(flags map[string]string) consts.ServerMode {
	if modeStr, ok := flags["MODE"]; ok {
		mode, err := consts.ServerMode(modeStr).Normalize()
		if err != nil {
			slog.Error("Invalid mode in flags", "mode", modeStr, "error", err)
			return consts.Development
		}
		return mode
	}

	if modeStr, ok := os.LookupEnv("MODE"); ok {
		mode, err := consts.ServerMode(modeStr).Normalize()
		if err != nil {
			slog.Error("Invalid mode in env", "mode", modeStr, "error", err)
			return consts.Development
		}
		return mode
	}

	// If go test is running then it is test mode
	if testing.Testing() {
		return consts.Test
	}

	return consts.Development
}

func compileFlags() map[string]string {
	var mode string
	var hostName string
	var hostPort int
	var hostProtocol string

	flag.StringVar(&mode, "MODE", "", "development or production")
	flag.StringVar(&hostName, "HOST_NAME", "", "Host name for the server")
	flag.IntVar(&hostPort, "HOST_PORT", 0, "Host port for the server")
	flag.StringVar(&hostProtocol, "HOST_PROTOCOL", "", "Host protocol for the server (http or https)")

	flag.Parse()

	ret := map[string]string{}
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "MODE":
			ret["MODE"] = mode
		case "HOST_NAME":
			ret["HOST_NAME"] = hostName
		case "HOST_PORT":
			ret["HOST_PORT"] = fmt.Sprintf("%d", hostPort)
		case "HOST_PROTOCOL":
			ret["HOST_PROTOCOL"] = hostProtocol
		}
	})
	return ret
}

func compileDotEnvFile(mode consts.ServerMode) (map[string]string, error) {
	filename := ".env"
	switch mode {
	case consts.Development:
		filename = ".env.dev"
	case consts.Test:
		projectRoot := utils.FindProjectRoot()
		filename = filepath.Join(projectRoot, ".env.test")
	case consts.Production:
		filename = ".env"
	}

	return readDotEnvFile(filename)
}

func readDotEnvFile(filename string) (map[string]string, error) {
	envMap := make(map[string]string)

	file, err := os.Open(filename)
	if err != nil {
		slog.Warn(fmt.Sprintf("Could not open ENV file: %s", filename), "error", err)
		return envMap, nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Skip malformed lines
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		envMap[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return envMap, nil
}

func printServerConfig(config ServerConfig) {
	// This is a concise printing of all non-secure information at startup
	slog.Info("Server Config",
		"Mode", map[consts.ServerMode]string{consts.Production: "Production", consts.Development: "Development", consts.Test: "Test", consts.Staging: "Staging"}[config.Mode],
		"Deployment", config.Deployment,
		"Host", fmt.Sprintf("%s://%s:%d", config.HostProtocol, config.HostName, config.HostPort),
		"S3", fmt.Sprintf("%s/%s", config.S3URL, config.S3Bucket),
		"TrustedProxies", config.TrustedProxies,
		"MaxRequestBodyBytes", config.MaxRequestBodyBytes,
		"GlobalRateLimit r/s", config.GlobalRateLimit,
		"GlobalRateBurst", config.GlobalRateBurst,
		"SiteGETPerMinuteLimit r/s", config.SiteGETPerMinuteLimit,
		"SiteGETPerMinuteBurst", config.SiteGETPerMinuteBurst,
		"SitePOSTPerMinuteLimit r/s", config.SitePOSTPerMinuteLimit,
		"SitePOSTPerMinuteBurst", config.SitePOSTPerMinuteBurst,
		"SiteCreateGlobalPerMinuteLimit r/s", config.SiteCreateGlobalPerMinuteLimit,
		"SiteCreateGlobalPerMinuteBurst", config.SiteCreateGlobalPerMinuteBurst,
		"SiteCreateIPPerMinuteLimit r/s", config.SiteCreateIPPerMinuteLimit,
		"SiteCreateIPPerMinuteBurst", config.SiteCreateIPPerMinuteBurst,
		"MaxSites", config.MaxSites,
		"MaxIdentitiesPerSite", config.MaxIdentitiesPerSite,
		"DeleteAfterInactivityMins", config.DeleteAfterInactivityMins,
		"DisableAfterInactivityDays", config.DisableAfterInactivityDays,
		"TruncateRecordsAfterDays", config.TruncateRecordsAfterDays,
		"LogsDBMaxRowsPerTable", config.LogsDBMaxRowsPerTable,
		"ControlBackupIntervalInSeconds", config.ControlBackupIntervalInSeconds,
		"BatchBackupIntervalInSeconds", config.BatchBackupIntervalInSeconds,
		"SiteBackupRetentionDays", config.SiteBackupRetentionDays,
		"ControlBackupRetentionDays", config.ControlBackupRetentionDays,
		"SocketMaxSessionsPerSite", config.SocketMaxSessionsPerSite,
		"SocketMaxBufferedBytes", config.SocketMaxBufferedBytes,
		"SocketMaxPacketsPerSecond", config.SocketMaxPacketsPerSecond,
	)
}

func (config *ServerConfig) IsHTTPS() bool {
	// This should only be false in development
	return config.HostProtocol == "https"
}

// RecoverURL is the public page a site's backup key gets a new admin login on
func (config *ServerConfig) RecoverURL() string {
	return fmt.Sprintf("%s://%s:%d/recover", config.HostProtocol, config.HostName, config.HostPort)
}

// SiteURL is the address of the site with the subdomain
func (config *ServerConfig) SiteURL(subdomain string) string {
	return fmt.Sprintf("%s://%s.%s:%d", config.HostProtocol, subdomain, config.HostName, config.HostPort)
}

// SetupURL is the control site's page a new server is set up on
func (config *ServerConfig) SetupURL() string {
	return fmt.Sprintf("%s://control.%s:%d/setup", config.HostProtocol, config.HostName, config.HostPort)
}

// ParseTrustedProxies reads comma separated IPs or CIDR ranges, e.g. "127.0.0.1,::1,172.30.0.0/24"
func ParseTrustedProxies(raw string) ([]netip.Prefix, error) {
	prefixes := []netip.Prefix{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if addr, err := netip.ParseAddr(part); err == nil {
			prefixes = append(prefixes, netip.PrefixFrom(addr, addr.BitLen()))
			continue
		}
		prefix, err := netip.ParsePrefix(part)
		if err != nil {
			return nil, fmt.Errorf("invalid TRUSTED_PROXIES entry %q: %w", part, err)
		}
		prefixes = append(prefixes, prefix.Masked())
	}
	return prefixes, nil
}

func MergeMaps(a, b map[string]string) map[string]string {
	result := make(map[string]string, len(a)+len(b))

	for k, v := range a {
		result[k] = v
	}

	for k, v := range b {
		result[k] = v // overrides existing keys
	}

	return result
}
