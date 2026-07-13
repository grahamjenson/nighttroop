package config

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/security"
)

type ServerConfig struct {
	Mode consts.ServerMode

	HostName     string
	HostPort     int
	HostProtocol string // http or https, used for generating links.

	// Some common Host URLs cached here so we dont fmt.Sprintf them all the time
	HostNameWithPort string
	HostURL          string
	HostURLWithPort  string

	// s3 config
	S3Bucket          string
	S3URL             string
	S3AccessKeyID     string
	S3AccessSecretKey string

	// sentry
	SentryDSN string

	// paths
	SqlitePath string

	// Unauthenticated Rate Limits
	GlobalRateLimit int
	GlobalRateBurst int

	GlobalIPPerMinuteLimit int
	GlobalIPPerMinuteBurst int

	SiteLoginPerMinuteLimit int
	SiteLoginPerMinuteBurst int

	// Authenticated Rate Limits
	SiteGETPerMinuteLimit        int
	SiteGETPerMinuteBurst        int
	SitePOSTPerMinuteLimit       int
	SitePOSTPerMinuteBurst       int
	SiteUpdateSyncPerMinuteLimit int
	SiteUpdateSyncPerMinuteBurst int

	//
	// Site Create Rate Limits
	SiteCreateGlobalPerMinuteLimit int
	SiteCreateGlobalPerMinuteBurst int
	SiteCreateIPPerMinuteLimit     float64
	SiteCreateIPPerMinuteBurst     int

	//
	// Site Limits
	MaxSites                  int
	MaxIdentitiesPerSite      int
	MaxSiteLastSeenInMins     int
	MaxSiteLastWrittenInMins  int
	MaxSiteNeverSeenInMins    int
	MaxSiteNeverWrittenInMins int

	// Backup Config
	ControlBackupIntervalInSeconds int
	BatchBackupIntervalInSeconds   int
	ArchiveIntervalInSeconds       int

	// Sync Config
	BackgroundSyncWorkers int
	SyncIntervalInSeconds int

	// Truncate Config
	TruncateIntervalInSeconds int
	MaxSiteDBSizeKB           int
	MaxControlDBSizeKB        int

	RateLimiters *RateLimiters

	// These keys are used to encrypt secrets in the DB and S3 at rest
	// This has the option to add additional keys
	Security   *security.SecurityConfig
	RawKeySets string

	// Logging
	Logger *slog.Logger

	// Dexcom Client ID and Secret
	DexcomClientID     string
	DexcomClientSecret string
}

const SITE_GET_LIMIT = 180
const SITE_POST_LIMIT = 60

// Priority is
// 1. Args
// 2. ENVAR
// 3. .env.dev
// 4. .env
// 5. Default
func NewServerConfig(logger *slog.Logger) (ServerConfig, error) {

	dotenvdev, err := readDotEnvFile(".env.dev")
	if err != nil {
		return ServerConfig{}, err
	}

	// If .env.dev exists then isDev = true
	// Override with .env
	dotenv, err := readDotEnvFile(".env")
	if err != nil {
		return ServerConfig{}, err
	}

	// DEFAULTS
	serverConfig := ServerConfig{
		Logger: logger,
		Mode:   consts.ServerMode(configStringSelect(dotenvdev, dotenv, "MODE", "development")),

		// Host config
		HostName:     configStringSelect(dotenvdev, dotenv, "HOST_NAME", "nt.local.t1d.nz"),
		HostPort:     configIntSelect(dotenvdev, dotenv, "HOST_PORT", 3000),
		HostProtocol: configStringSelect(dotenvdev, dotenv, "HOST_PROTOCOL", "https"),

		// s3 config
		S3Bucket:          configStringSelect(dotenvdev, dotenv, "S3_BUCKET", "mybucket"),
		S3URL:             configStringSelect(dotenvdev, dotenv, "S3_URL", "http://localhost:9000"),
		S3AccessKeyID:     configStringSelect(dotenvdev, dotenv, "S3_ACCESS_KEY", "minioadmin"),
		S3AccessSecretKey: configStringSelect(dotenvdev, dotenv, "S3_SECRET_KEY", "minioadmin"),

		// sentry
		SentryDSN: configStringSelect(dotenvdev, dotenv, "SENTRY_DSN", ""),

		// security
		RawKeySets: configStringSelect(dotenvdev, dotenv, consts.KEY_SETS_ENV_KEY, ""),

		// paths
		SqlitePath: configStringSelect(dotenvdev, dotenv, "SQLITE_PATH", "./data"),

		// RATE LIMTERS TODO fine tune these a bit more

		// Public RateLimits
		GlobalRateLimit:        configIntSelect(dotenvdev, dotenv, "GLOBAL_RATE_LIMIT", 100000), // Stop DDOS doing downloads
		GlobalRateBurst:        configIntSelect(dotenvdev, dotenv, "GLOBAL_RATE_BURST", 100000),
		GlobalIPPerMinuteLimit: configIntSelect(dotenvdev, dotenv, "GLOBAL_IP_PER_MINUTE_LIMIT", SITE_GET_LIMIT*4), // Single IP can fully manage 4 sites
		GlobalIPPerMinuteBurst: configIntSelect(dotenvdev, dotenv, "GLOBAL_IP_PER_MINUTE_BURST", SITE_GET_LIMIT*8),

		SiteLoginPerMinuteLimit: configIntSelect(dotenvdev, dotenv, "SITE_LOGIN_PER_MINUTE_LIMIT", 1), // Stop credential stuffing
		SiteLoginPerMinuteBurst: configIntSelect(dotenvdev, dotenv, "SITE_LOGIN_PER_MINUTE_BURST", 10),

		// Site Create Rate Limits
		SiteCreateGlobalPerMinuteLimit: configIntSelect(dotenvdev, dotenv, "SITE_CREATE_GLOBAL_PER_MINUTE_LIMIT", 7), // If maxed out this filles up in a day
		SiteCreateGlobalPerMinuteBurst: configIntSelect(dotenvdev, dotenv, "SITE_CREATE_GLOBAL_PER_MINUTE_BURST", 20),
		SiteCreateIPPerMinuteLimit:     configFloatSelect(dotenvdev, dotenv, "SITE_CREATE_IP_PER_MINUTE_LIMIT", 0.1), // 1 event every 10 minutes
		SiteCreateIPPerMinuteBurst:     configIntSelect(dotenvdev, dotenv, "SITE_CREATE_IP_PER_MINUTE_BURST", 1),

		// Authed
		// Site Rate Limits
		SiteGETPerMinuteLimit:        configIntSelect(dotenvdev, dotenv, "SITE_GET_PER_MINUTE_LIMIT", SITE_GET_LIMIT),
		SiteGETPerMinuteBurst:        configIntSelect(dotenvdev, dotenv, "SITE_GET_PER_MINUTE_BURST", SITE_GET_LIMIT*2),
		SitePOSTPerMinuteLimit:       configIntSelect(dotenvdev, dotenv, "SITE_POST_PER_MINUTE_LIMIT", SITE_POST_LIMIT),
		SitePOSTPerMinuteBurst:       configIntSelect(dotenvdev, dotenv, "SITE_POST_PER_MINUTE_BURST", SITE_POST_LIMIT*4),
		SiteUpdateSyncPerMinuteLimit: configIntSelect(dotenvdev, dotenv, "SITE_UPDATE_SYNC_PER_MINUTE_LIMIT", 1), // Stop sites from updating sync settings too much
		SiteUpdateSyncPerMinuteBurst: configIntSelect(dotenvdev, dotenv, "SITE_UPDATE_SYNC_PER_MINUTE_BURST", 4),

		//
		// Site Limits
		MaxSites:             configIntSelect(dotenvdev, dotenv, "MAX_SITES", 10000),
		MaxIdentitiesPerSite: configIntSelect(dotenvdev, dotenv, "MAX_IDENTITIES_PER_USER", 10),

		// Truncate Config
		MaxSiteDBSizeKB:           configIntSelect(dotenvdev, dotenv, "MAX_SITE_DB_SIZE_KB", 3000),
		MaxControlDBSizeKB:        configIntSelect(dotenvdev, dotenv, "MAX_CONTROL_DB_SIZE_KB", 5000),
		TruncateIntervalInSeconds: configIntSelect(dotenvdev, dotenv, "TRUNCATE_INTERVAL_IN_SECONDS", 60),

		// Archiving limits
		MaxSiteLastSeenInMins:    configIntSelect(dotenvdev, dotenv, "MAX_SITE_LAST_SEEN_IN_MINS", 60*24*30),   // 1 month
		MaxSiteLastWrittenInMins: configIntSelect(dotenvdev, dotenv, "MAX_SITE_LAST_WRITTEN_IN_MINS", 60*24*7), // 1 week

		MaxSiteNeverSeenInMins:    configIntSelect(dotenvdev, dotenv, "MAX_SITE_NEVER_SEEN_IN_MINS", 30),       // 30 minutes to login first time
		MaxSiteNeverWrittenInMins: configIntSelect(dotenvdev, dotenv, "MAX_SITE_NEVER_WRITTEN_IN_MINS", 60*24), // 1 day to write something

		// Backup Config
		ControlBackupIntervalInSeconds: configIntSelect(dotenvdev, dotenv, "CONTROL_BACKUP_INTERVAL_IN_SECONDS", 60),
		BatchBackupIntervalInSeconds:   configIntSelect(dotenvdev, dotenv, "BATCH_BACKUP_INTERVAL_IN_SECONDS", 60), // 1 minute
		ArchiveIntervalInSeconds:       configIntSelect(dotenvdev, dotenv, "ARCHIVE_INTERVAL_IN_SECONDS", 3600),    // 1 hour

		// Sync Config
		BackgroundSyncWorkers: configIntSelect(dotenvdev, dotenv, "BACKGROUND_SYNC_WORKERS", 5),
		SyncIntervalInSeconds: configIntSelect(dotenvdev, dotenv, "SYNC_INTERVAL_IN_SECONDS", 5), // check syncs every 5 seconds

		// Dexcom Config
		DexcomClientID:     configStringSelect(dotenvdev, dotenv, "DEXCOM_CLIENT_ID", ""),
		DexcomClientSecret: configStringSelect(dotenvdev, dotenv, "DEXCOM_CLIENT_SECRET", ""),
	}

	// Error out if unknown ENVARS are found in .env.dev or .env
	if len(dotenvdev) > 0 {
		slog.Error("Unknown .env.dev variables: ", "variables", func() []string {
			var keys []string
			for k := range dotenvdev {
				keys = append(keys, k)
			}
			return keys
		}())
		return ServerConfig{}, fmt.Errorf("unknown .env.dev variables")
	}

	if len(dotenv) > 0 {
		slog.Error("Unknown .env variables: ", "variables", func() []string {
			var keys []string
			for k := range dotenv {
				keys = append(keys, k)
			}
			return keys
		}())
		return ServerConfig{}, fmt.Errorf("unknown .env variables")
	}

	// Override with flags some common settings
	flagmode := ""
	flag.StringVar(&flagmode, "mode", "", "developmen or production")
	// host config
	flag.StringVar(&serverConfig.HostName, "host", serverConfig.HostName, "Host name for the server")
	flag.IntVar(&serverConfig.HostPort, "port", serverConfig.HostPort, "Host port for the server")
	flag.StringVar(&serverConfig.HostProtocol, "protocol", serverConfig.HostProtocol, "Host protocol for the server (http or https)")

	flag.Parse()

	if flagmode != "" {
		serverConfig.Mode = consts.ServerMode(flagmode)
	}

	// Validate the config
	if err := serverConfig.ValidateClean(); err != nil {
		return ServerConfig{}, err
	}

	// Precompute some Host URLs
	serverConfig.HostNameWithPort = fmt.Sprintf("%s:%d", serverConfig.HostName, serverConfig.HostPort)
	serverConfig.HostURL = fmt.Sprintf("%s://%s", serverConfig.HostProtocol, serverConfig.HostName)
	serverConfig.HostURLWithPort = fmt.Sprintf("%s://%s:%d", serverConfig.HostProtocol, serverConfig.HostName, serverConfig.HostPort)

	serverConfig.RateLimiters = newRateLimiters(serverConfig)

	serverConfig.Security = security.NewSecurityConfig(serverConfig.RawKeySets, serverConfig.Mode, serverConfig.HostName, serverConfig.IsHTTPS())

	printServerConfig(serverConfig)
	return serverConfig, nil
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

func (config *ServerConfig) ValidateClean() error {
	// Deal with mode
	mode, err := config.Mode.Normalize()
	if err != nil {
		return err
	}
	config.Mode = mode

	config.HostProtocol = strings.ToLower(config.HostProtocol)
	if config.HostProtocol != "http" && config.HostProtocol != "https" {
		return fmt.Errorf("invalid host protocol: %s", config.HostProtocol)
	}

	if config.HostProtocol == "http" && !config.IsDev() && !config.IsTest() {
		return fmt.Errorf("http is only allowed in development or test mode")
	}

	return nil
}

func configStringSelect(dotenvdev map[string]string, dotenv map[string]string, envVar string, defaultValue string) string {
	returnval := defaultValue

	if val, ok := dotenvdev[envVar]; ok {
		returnval = val
		delete(dotenvdev, envVar) // delete so we can find unknown keys later
	}

	if val, ok := dotenv[envVar]; ok {
		returnval = val
		delete(dotenv, envVar) // delete so we can find unknown keys later
	}

	if val, ok := os.LookupEnv(envVar); ok {
		returnval = val
	}

	return returnval
}

func configIntSelect(dotenvdev map[string]string, dotenv map[string]string, envVar string, defaultValue int) int {

	returnval := fmt.Sprintf("%d", defaultValue)

	if val, ok := dotenv[envVar]; ok {
		returnval = val
		delete(dotenv, envVar) // delete so we can find unknown keys later
	}

	if val, ok := dotenvdev[envVar]; ok {
		returnval = val
		delete(dotenvdev, envVar) // delete so we can find unknown keys later
	}

	if val, ok := os.LookupEnv(envVar); ok {
		returnval = val
	}

	intVal, err := strconv.Atoi(returnval)
	if err == nil {
		return intVal
	} else {
		slog.Error("Error parsing %s for %s: %s", "returnval", returnval, "envVar", envVar, "error", err)
	}

	return intVal
}

func configFloatSelect(dotenvdev map[string]string, dotenv map[string]string, envVar string, defaultValue float64) float64 {

	returnval := fmt.Sprintf("%f", defaultValue)

	if val, ok := dotenv[envVar]; ok {
		returnval = val
		delete(dotenv, envVar) // delete so we can find unknown keys later
	}

	if val, ok := dotenvdev[envVar]; ok {
		returnval = val
		delete(dotenvdev, envVar) // delete so we can find unknown keys later
	}

	if val, ok := os.LookupEnv(envVar); ok {
		returnval = val
	}

	floatVal, err := strconv.ParseFloat(returnval, 64)
	if err == nil {
		return floatVal
	} else {
		slog.Error("Error parsing %s for %s: %s", "returnval", returnval, "envVar", envVar, "error", err)
	}

	return floatVal
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
		"Mode", map[consts.ServerMode]string{consts.Production: "Production", consts.Development: "Development", consts.Test: "Test"}[config.Mode],
		"Host", fmt.Sprintf("%s://%s:%d", config.HostProtocol, config.HostName, config.HostPort),
		"S3", fmt.Sprintf("%s/%s", config.S3URL, config.S3Bucket),
		"SQLitePath", config.SqlitePath,
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
		"MaxSiteLastSeenInMins", config.MaxSiteLastSeenInMins,
		"MaxSiteLastWrittenInMins", config.MaxSiteLastWrittenInMins,
		"MaxSiteDBSizeKB", config.MaxSiteDBSizeKB,
		"ControlBackupIntervalInSeconds", config.ControlBackupIntervalInSeconds,
		"BatchBackupIntervalInSeconds", config.BatchBackupIntervalInSeconds,
	)
}

func (config *ServerConfig) IsHTTPS() bool {
	// This should only be false in development
	return config.HostProtocol == "https"
}

func ptr[T any](v T) *T {
	return &v
}
