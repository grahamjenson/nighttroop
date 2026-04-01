package config

import (
	"bufio"
	"flag"
	"fmt"
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

	// s3 config
	S3Bucket          string
	S3URL             string
	S3AccessKeyID     string
	S3AccessSecretKey string

	// security
	SecretsEncryptionKey string

	// paths
	SqlitePath string

	// Unauthenticated Rate Limits
	GlobalRateLimit int
	GlobalRateBurst int

	GlobalIPPerMinuteLimit int
	GlobalIPPerMinuteBurst int

	// Authenticated Rate Limits
	UserGETPerMinuteLimit  int
	UserGETPerMinuteBurst  int
	UserPOSTPerMinuteLimit int
	UserPOSTPerMinuteBurst int

	// User Create Rate Limits
	UserCreateGlobalPerMinuteLimit int
	UserCreateGlobalPerMinuteBurst int
	UserCreateIPPerMinuteLimit     int
	UserCreateIPPerMinuteBurst     int

	// User Limits
	MaxUsers              int
	MaxDevicesPerUser     int
	MaxUserInactiveInMins int
	MaxUserDBSizeMB       int

	// Backup Config
	ControlBackupIntervalInSeconds int
	BatchBackupIntervalInSeconds   int

	RateLimiters *RateLimiters

	// These keys are used to encrypt secrets in the DB and S3 at rest
	// This has the option to add additional keys
	Security *security.SecurityConfig
}

const USER_GET_LIMIT = 60
const USER_POST_LIMIT = 20

// Priority is
// 1. Args
// 2. ENVAR
// 3. .env.dev
// 4. .env
// 5. Default
func NewServerConfig() (ServerConfig, error) {

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
		Mode: consts.Development,

		// Host config
		HostName:     configStringSelect(dotenvdev, dotenv, "HOST_NAME", "nt.local.t1d.nz"),
		HostPort:     configIntSelect(dotenvdev, dotenv, "HOST_PORT", 3000),
		HostProtocol: configStringSelect(dotenvdev, dotenv, "HOST_PROTOCOL", "https"),

		// s3 config
		S3Bucket:          configStringSelect(dotenvdev, dotenv, "S3_BUCKET", "mybucket"),
		S3URL:             configStringSelect(dotenvdev, dotenv, "S3_URL", "http://localhost:9000"),
		S3AccessKeyID:     configStringSelect(dotenvdev, dotenv, "S3_ACCESS_KEY", "minioadmin"),
		S3AccessSecretKey: configStringSelect(dotenvdev, dotenv, "S3_SECRET_KEY", "minioadmin"),

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

		// User Create Rate Limits
		UserCreateGlobalPerMinuteLimit: configIntSelect(dotenvdev, dotenv, "USER_CREATE_GLOBAL_PER_MINUTE_LIMIT", 20),
		UserCreateGlobalPerMinuteBurst: configIntSelect(dotenvdev, dotenv, "USER_CREATE_GLOBAL_PER_MINUTE_BURST", 20),
		UserCreateIPPerMinuteLimit:     configIntSelect(dotenvdev, dotenv, "USER_CREATE_IP_PER_MINUTE_LIMIT", 1),
		UserCreateIPPerMinuteBurst:     configIntSelect(dotenvdev, dotenv, "USER_CREATE_IP_PER_MINUTE_BURST", 1),

		// Authed User Rate Limits
		UserGETPerMinuteLimit:  configIntSelect(dotenvdev, dotenv, "USER_GET_PER_MINUTE_LIMIT", USER_GET_LIMIT),
		UserGETPerMinuteBurst:  configIntSelect(dotenvdev, dotenv, "USER_GET_PER_MINUTE_BURST", USER_GET_LIMIT*2),
		UserPOSTPerMinuteLimit: configIntSelect(dotenvdev, dotenv, "USER_POST_PER_MINUTE_LIMIT", USER_POST_LIMIT),
		UserPOSTPerMinuteBurst: configIntSelect(dotenvdev, dotenv, "USER_POST_PER_MINUTE_BURST", USER_POST_LIMIT*4),

		// User Limits
		MaxUsers:              configIntSelect(dotenvdev, dotenv, "MAX_USERS", 10000),
		MaxDevicesPerUser:     configIntSelect(dotenvdev, dotenv, "MAX_DEVICES_PER_USER", 10),
		MaxUserInactiveInMins: configIntSelect(dotenvdev, dotenv, "MAX_USER_INACTIVE_IN_MINS", 60*24*30), // 30 days
		MaxUserDBSizeMB:       configIntSelect(dotenvdev, dotenv, "MAX_USER_DB_SIZE_MB", 100),

		// Backup Config
		ControlBackupIntervalInSeconds: configIntSelect(dotenvdev, dotenv, "CONTROL_BACKUP_INTERVAL_IN_SECONDS", 10),
		BatchBackupIntervalInSeconds:   configIntSelect(dotenvdev, dotenv, "BATCH_BACKUP_INTERVAL_IN_SECONDS", 60), // 1 minute
	}

	serverConfig.RateLimiters = newRateLimiters(serverConfig)

	mode := configStringSelect(dotenvdev, dotenv, "MODE", "development")

	keySetsRaw := configStringSelect(dotenvdev, dotenv, "KEY_SETS", "")
	serverConfig.Security = security.NewSecurityConfig(keySetsRaw, serverConfig.Mode, serverConfig.HostName, serverConfig.IsHTTPS())

	if len(dotenvdev) > 0 {
		fmt.Printf("Unknown .env.dev variables: ")
		for k := range dotenvdev {
			fmt.Printf("%s,", k)

		}
		fmt.Printf("\n")
		return ServerConfig{}, fmt.Errorf("unknown .env.dev variables")
	}

	if len(dotenv) > 0 {
		fmt.Printf("Unknown .env variables: ")
		for k := range dotenv {
			fmt.Printf("%s,", k)

		}
		fmt.Printf("\n")
		return ServerConfig{}, fmt.Errorf("unknown .env variables")
	}

	// Override with flags
	flagmode := ""
	flag.StringVar(&flagmode, "mode", "", "developmen or production")
	// host config
	flag.StringVar(&serverConfig.HostName, "host", serverConfig.HostName, "Host name for the server")
	flag.IntVar(&serverConfig.HostPort, "port", serverConfig.HostPort, "Host port for the server")
	flag.StringVar(&serverConfig.HostProtocol, "protocol", serverConfig.HostProtocol, "Host protocol for the server (http or https)")

	// s3 config
	flag.StringVar(&serverConfig.S3Bucket, "s3bucket", serverConfig.S3Bucket, "S3 bucket name")
	flag.StringVar(&serverConfig.S3URL, "s3url", serverConfig.S3URL, "S3 URL")
	flag.StringVar(&serverConfig.S3AccessKeyID, "s3accesskey", serverConfig.S3AccessKeyID, "S3 Access Key ID")
	flag.StringVar(&serverConfig.S3AccessSecretKey, "s3secretkey", serverConfig.S3AccessSecretKey, "S3 Access Secret Key")

	// security
	flag.StringVar(&serverConfig.SecretsEncryptionKey, "secretsencryptionkey", serverConfig.SecretsEncryptionKey, "Secrets Encryption Key")

	// paths
	flag.StringVar(&serverConfig.SqlitePath, "sqlitedbpath", serverConfig.SqlitePath, "Path to store sqlite DB files")

	// NO FLAGS FOR RATE LIMITERS

	// User Limits
	flag.IntVar(&serverConfig.MaxUsers, "maxusers", serverConfig.MaxUsers, "Maximum number of users")
	flag.IntVar(&serverConfig.MaxDevicesPerUser, "maxdevicesperuser", serverConfig.MaxDevicesPerUser, "Maximum number of devices per user")
	flag.IntVar(&serverConfig.MaxUserDBSizeMB, "maxuserdbsizemb", serverConfig.MaxUserDBSizeMB, "Maximum user DB size in MB")
	flag.IntVar(&serverConfig.MaxUserInactiveInMins, "maxuserinactiveinmins", serverConfig.MaxUserInactiveInMins, "Maximum user inactive time in minutes (after this they are archived)")

	// Backup Config
	flag.IntVar(&serverConfig.ControlBackupIntervalInSeconds, "controlbackupinterval", serverConfig.ControlBackupIntervalInSeconds, "Control DB Backup Interval in seconds")
	flag.IntVar(&serverConfig.BatchBackupIntervalInSeconds, "batchbackupinterval", serverConfig.BatchBackupIntervalInSeconds, "Batch Backup Interval in seconds")

	flag.Parse()

	// CLEAN UP A FEW THINGS
	if flagmode != "" {
		mode = flagmode
	}
	if err := serverConfig.ValidateClean(mode); err != nil {
		return ServerConfig{}, err
	}

	printServerConfig(serverConfig)
	return serverConfig, nil
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
		fmt.Errorf("invalid host protocol: %s", config.HostProtocol)
	}

	if config.HostProtocol == "http" && config.Mode != consts.Development && config.Mode != consts.Test {
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
		fmt.Printf("Error parsing %s for %s: %s\n", returnval, envVar, err)
	}

	return intVal
}

func readDotEnvFile(filename string) (map[string]string, error) {
	envMap := make(map[string]string)

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
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
	fmt.Printf("Server Config:\n")
	fmt.Printf("Mode: %s\n", map[consts.ServerMode]string{consts.Production: "Production", consts.Development: "Development", consts.Test: "Test"}[config.Mode])
	fmt.Printf("Host: %s://%s:%d\n", config.HostProtocol, config.HostName, config.HostPort)
	fmt.Printf("S3: %s@%s/%s\n", config.S3AccessKeyID, config.S3URL, config.S3Bucket)
	fmt.Printf("SQLite Path: %s\n", config.SqlitePath)
	fmt.Printf("Global Rate Limit r/w req/s: %d/%d burst %d/%d\n", config.GlobalRateLimit, config.GlobalRateLimit, config.GlobalRateBurst, config.GlobalRateBurst)
	fmt.Printf("User Rate Limit r/w req/s: %d/%d burst %d/%d\n", config.UserGETPerMinuteLimit, config.UserPOSTPerMinuteLimit, config.UserGETPerMinuteBurst, config.UserPOSTPerMinuteBurst)
	fmt.Printf("User Create Rate Limit req/s: %d burst %d\n", config.UserCreateGlobalPerMinuteLimit, config.UserCreateGlobalPerMinuteBurst)
	fmt.Printf("User Create IP Rate Limit req/s: %d burst %d\n", config.UserCreateIPPerMinuteLimit, config.UserCreateIPPerMinuteBurst)

	fmt.Printf("Max Users/Max Devices, User DB Size/Age (mins): %d / %d / %d / %d\n", config.MaxUsers, config.MaxDevicesPerUser, config.MaxUserDBSizeMB, config.MaxUserInactiveInMins)
	fmt.Printf("Backup Interval (Control/Batch) in seconds: %d / %d\n", config.ControlBackupIntervalInSeconds, config.BatchBackupIntervalInSeconds)
}

func (config *ServerConfig) GetHostURL() string {
	return fmt.Sprintf("%s://%s:%d", config.HostProtocol, config.HostName, config.HostPort)
}

func (config *ServerConfig) IsHTTPS() bool {
	// This should only be false in development
	return config.HostProtocol == "https"
}
