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

	// security
	SecretsEncryptionKey string

	// paths
	SqlitePath string

	// Unauthenticated Rate Limits
	GlobalRateLimit int
	GlobalRateBurst int

	GlobalIPPerMinuteLimit int
	GlobalIPPerMinuteBurst int

	UserLoginPerMinuteLimit int
	UserLoginPerMinuteBurst int

	// Authenticated Rate Limits
	UserGETPerMinuteLimit  int
	UserGETPerMinuteBurst  int
	UserPOSTPerMinuteLimit int
	UserPOSTPerMinuteBurst int

	// User Create Rate Limits
	UserCreateGlobalPerMinuteLimit int
	UserCreateGlobalPerMinuteBurst int
	UserCreateIPPerMinuteLimit     float64
	UserCreateIPPerMinuteBurst     int

	// User Limits
	MaxUsers                  int
	MaxDevicesPerUser         int
	MaxUserLastSeenInMins     int
	MaxUserLastWrittenInMins  int
	MaxUserNeverSeenInMins    int
	MaxUserNeverWrittenInMins int

	MaxUserDBSizeKB    int
	MaxControlDBSizeKB int

	// Backup Config
	ControlBackupIntervalInSeconds int
	BatchBackupIntervalInSeconds   int

	RateLimiters *RateLimiters

	// These keys are used to encrypt secrets in the DB and S3 at rest
	// This has the option to add additional keys
	Security *security.SecurityConfig

	// Logging
	Logger *slog.Logger
}

const USER_GET_LIMIT = 60
const USER_POST_LIMIT = 20

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
		Mode:   consts.Development,

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
		SecretsEncryptionKey: configStringSelect(dotenvdev, dotenv, "SECRETS_ENCRYPTION_KEY", "mysecretencryptionkey"),

		// paths
		SqlitePath: configStringSelect(dotenvdev, dotenv, "SQLITE_PATH", "./data"),

		// RATE LIMTERS TODO fine tune these a bit more

		// Public RateLimits
		GlobalRateLimit:        configIntSelect(dotenvdev, dotenv, "GLOBAL_RATE_LIMIT", 100000), // Stop DDOS doing downloads
		GlobalRateBurst:        configIntSelect(dotenvdev, dotenv, "GLOBAL_RATE_BURST", 100000),
		GlobalIPPerMinuteLimit: configIntSelect(dotenvdev, dotenv, "GLOBAL_IP_PER_MINUTE_LIMIT", 100), // Stop DOS
		GlobalIPPerMinuteBurst: configIntSelect(dotenvdev, dotenv, "GLOBAL_IP_PER_MINUTE_BURST", 200),

		UserLoginPerMinuteLimit: configIntSelect(dotenvdev, dotenv, "USER_LOGIN_PER_MINUTE_LIMIT", 1), // Stop credential stuffing
		UserLoginPerMinuteBurst: configIntSelect(dotenvdev, dotenv, "USER_LOGIN_PER_MINUTE_BURST", 10),

		// User Create Rate Limits
		UserCreateGlobalPerMinuteLimit: configIntSelect(dotenvdev, dotenv, "USER_CREATE_GLOBAL_PER_MINUTE_LIMIT", 7), // If maxed out this filles up in a day
		UserCreateGlobalPerMinuteBurst: configIntSelect(dotenvdev, dotenv, "USER_CREATE_GLOBAL_PER_MINUTE_BURST", 20),
		UserCreateIPPerMinuteLimit:     configFloatSelect(dotenvdev, dotenv, "USER_CREATE_IP_PER_MINUTE_LIMIT", 0.1), // 1 event every 10 minutes
		UserCreateIPPerMinuteBurst:     configIntSelect(dotenvdev, dotenv, "USER_CREATE_IP_PER_MINUTE_BURST", 1),

		// Authed User Rate Limits
		UserGETPerMinuteLimit: configIntSelect(dotenvdev, dotenv, "USER_GET_PER_MINUTE_LIMIT", USER_GET_LIMIT),
		UserGETPerMinuteBurst: configIntSelect(dotenvdev, dotenv, "USER_GET_PER_MINUTE_BURST", USER_GET_LIMIT*2),

		UserPOSTPerMinuteLimit: configIntSelect(dotenvdev, dotenv, "USER_POST_PER_MINUTE_LIMIT", USER_POST_LIMIT),
		UserPOSTPerMinuteBurst: configIntSelect(dotenvdev, dotenv, "USER_POST_PER_MINUTE_BURST", USER_POST_LIMIT*4),

		// User Limits
		MaxUsers:           configIntSelect(dotenvdev, dotenv, "MAX_USERS", 10000),
		MaxDevicesPerUser:  configIntSelect(dotenvdev, dotenv, "MAX_DEVICES_PER_USER", 10),
		MaxUserDBSizeKB:    configIntSelect(dotenvdev, dotenv, "MAX_USER_DB_SIZE_KB", 3000),
		MaxControlDBSizeKB: configIntSelect(dotenvdev, dotenv, "MAX_CONTROL_DB_SIZE_KB", 5000),

		// Archiving limits
		MaxUserLastSeenInMins:    configIntSelect(dotenvdev, dotenv, "MAX_USER_LAST_SEEN_IN_MINS", 60*24*30),   // 1 month
		MaxUserLastWrittenInMins: configIntSelect(dotenvdev, dotenv, "MAX_USER_LAST_WRITTEN_IN_MINS", 60*24*7), // 1 week

		MaxUserNeverSeenInMins:    configIntSelect(dotenvdev, dotenv, "MAX_USER_NEVER_SEEN_IN_MINS", 30),       // 30 minutes to login first time
		MaxUserNeverWrittenInMins: configIntSelect(dotenvdev, dotenv, "MAX_USER_NEVER_WRITTEN_IN_MINS", 60*24), // 1 day to write something

		// Backup Config
		ControlBackupIntervalInSeconds: configIntSelect(dotenvdev, dotenv, "CONTROL_BACKUP_INTERVAL_IN_SECONDS", 10),
		BatchBackupIntervalInSeconds:   configIntSelect(dotenvdev, dotenv, "BATCH_BACKUP_INTERVAL_IN_SECONDS", 60), // 1 minute
	}

	serverConfig.RateLimiters = newRateLimiters(serverConfig)

	mode := configStringSelect(dotenvdev, dotenv, "MODE", "development")

	keySetsRaw := configStringSelect(dotenvdev, dotenv, consts.KEY_SETS_ENV_KEY, "")
	serverConfig.Security = security.NewSecurityConfig(keySetsRaw, serverConfig.Mode, serverConfig.HostName, serverConfig.IsHTTPS())

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

	// Override with flags
	flagmode := ""
	flag.StringVar(&flagmode, "mode", "", "developmen or production")
	// host config
	flag.StringVar(&serverConfig.HostName, "host", serverConfig.HostName, "Host name for the server")
	flag.IntVar(&serverConfig.HostPort, "port", serverConfig.HostPort, "Host port for the server")
	flag.StringVar(&serverConfig.HostProtocol, "protocol", serverConfig.HostProtocol, "Host protocol for the server (http or https)")

	flag.Parse()

	// CLEAN UP A FEW THINGS
	if flagmode != "" {
		mode = flagmode
	}
	if err := serverConfig.ValidateClean(mode); err != nil {
		return ServerConfig{}, err
	}

	// Precompute some Host URLs
	serverConfig.HostNameWithPort = fmt.Sprintf("%s:%d", serverConfig.HostName, serverConfig.HostPort)
	serverConfig.HostURL = fmt.Sprintf("%s://%s", serverConfig.HostProtocol, serverConfig.HostName)
	serverConfig.HostURLWithPort = fmt.Sprintf("%s://%s:%d", serverConfig.HostProtocol, serverConfig.HostName, serverConfig.HostPort)

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

func (config *ServerConfig) ValidateClean(mode string) error {
	// Deal with mode

	switch strings.ToLower(mode) {
	case "production":
		config.Mode = consts.Production
	case "development":
		config.Mode = consts.Development
	case "test":
		config.Mode = consts.Test
	default:
		return fmt.Errorf("invalid mode: %s", mode)
	}

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
		"UserGETPerMinuteLimit r/s", config.UserGETPerMinuteLimit,
		"UserGETPerMinuteBurst", config.UserGETPerMinuteBurst,
		"UserPOSTPerMinuteLimit r/s", config.UserPOSTPerMinuteLimit,
		"UserPOSTPerMinuteBurst", config.UserPOSTPerMinuteBurst,
		"UserCreateGlobalPerMinuteLimit r/s", config.UserCreateGlobalPerMinuteLimit,
		"UserCreateGlobalPerMinuteBurst", config.UserCreateGlobalPerMinuteBurst,
		"UserCreateIPPerMinuteLimit r/s", config.UserCreateIPPerMinuteLimit,
		"UserCreateIPPerMinuteBurst", config.UserCreateIPPerMinuteBurst,
		"MaxUsers", config.MaxUsers,
		"MaxDevicesPerUser", config.MaxDevicesPerUser,
		"MaxUserLastSeenInMins", config.MaxUserLastSeenInMins,
		"MaxUserLastWrittenInMins", config.MaxUserLastWrittenInMins,
		"MaxUserDBSizeKB", config.MaxUserDBSizeKB,
		"ControlBackupIntervalInSeconds", config.ControlBackupIntervalInSeconds,
		"BatchBackupIntervalInSeconds", config.BatchBackupIntervalInSeconds,
	)
}

func (config *ServerConfig) IsHTTPS() bool {
	// This should only be false in development
	return config.HostProtocol == "https"
}
