package config

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/tink-crypto/tink-go/v2/jwt"
	"github.com/tink-crypto/tink-go/v2/keyset"

	"github.com/tink-crypto/tink-go/v2/tink"
)

type ServerMode int

const (
	Production ServerMode = iota
	Development
)

type ServerConfig struct {
	Mode ServerMode

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

	// User Rate Limits
	UserReadRateLimit  int
	UserReadRateBurst  int
	UserWriteRateLimit int
	UserWriteRateBurst int

	// User Limits
	MaxUsers              int
	MaxUserInactiveInMins int
	MaxUserDBSizeMB       int
	UserCreateRateLimit   int
	UserCreateRateBurst   int

	// Backup Config
	ControlBackupIntervalInSeconds int
	BatchBackupIntervalInSeconds   int

	// These keys are used to encrypt secrets in the DB and S3 at rest
	// This has the option to add additional keys
	EncryptionKeySet       *keyset.Handle
	StreamEncryptionKeySet *keyset.Handle // for files
	JWTKeySet              *keyset.Handle

	aeadEnc       tink.AEAD
	streamEnc     tink.StreamingAEAD
	jwtMac        jwt.MAC
	jwwtValidator *jwt.Validator
}

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
		Mode: Development,

		// Host config
		HostName:     configStringSelect(dotenvdev, dotenv, "HOST_NAME", "nt.local.t1d.nz"),
		HostPort:     configIntSelect(dotenvdev, dotenv, "HOST_PORT", 3000),
		HostProtocol: configStringSelect(dotenvdev, dotenv, "HOST_PROTOCOL", "http"),

		// s3 config
		S3Bucket:          configStringSelect(dotenvdev, dotenv, "S3_BUCKET", "mybucket"),
		S3URL:             configStringSelect(dotenvdev, dotenv, "S3_URL", "http://localhost:9000"),
		S3AccessKeyID:     configStringSelect(dotenvdev, dotenv, "S3_ACCESS_KEY", "minioadmin"),
		S3AccessSecretKey: configStringSelect(dotenvdev, dotenv, "S3_SECRET_KEY", "minioadmin"),

		// security
		SecretsEncryptionKey: configStringSelect(dotenvdev, dotenv, "SECRETS_ENCRYPTION_KEY", "mysecretencryptionkey"),

		// paths
		SqlitePath: configStringSelect(dotenvdev, dotenv, "SQLITE_PATH", "./data"),

		// User Rate Limits
		UserReadRateLimit:  configIntSelect(dotenvdev, dotenv, "USER_READ_RATE_LIMIT", 100),
		UserReadRateBurst:  configIntSelect(dotenvdev, dotenv, "USER_READ_RATE_BURST", 200),
		UserWriteRateLimit: configIntSelect(dotenvdev, dotenv, "USER_WRITE_RATE_LIMIT", 50),
		UserWriteRateBurst: configIntSelect(dotenvdev, dotenv, "USER_WRITE_RATE_BURST", 100),

		// User Limits
		MaxUsers:              configIntSelect(dotenvdev, dotenv, "MAX_USERS", 10000),
		MaxUserInactiveInMins: configIntSelect(dotenvdev, dotenv, "MAX_USER_INACTIVE_IN_MINS", 60*24*30), // 30 days
		MaxUserDBSizeMB:       configIntSelect(dotenvdev, dotenv, "MAX_USER_DB_SIZE_MB", 100),

		UserCreateRateLimit: configIntSelect(dotenvdev, dotenv, "USER_CREATE_RATE_LIMIT", 10),
		UserCreateRateBurst: configIntSelect(dotenvdev, dotenv, "USER_CREATE_RATE_BURST", 100),

		// Backup Config
		ControlBackupIntervalInSeconds: configIntSelect(dotenvdev, dotenv, "CONTROL_BACKUP_INTERVAL_IN_SECONDS", 10),
		BatchBackupIntervalInSeconds:   configIntSelect(dotenvdev, dotenv, "BATCH_BACKUP_INTERVAL_IN_SECONDS", 60), // 1 minute
	}

	err = serverConfig.ConfigCreateEncryptionKeys(dotenvdev, dotenv)
	if err != nil {
		return ServerConfig{}, fmt.Errorf("cannot create encryption key sets: %w", err)
	}

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
	mode := ""
	flag.StringVar(&mode, "mode", "", "developmen or production")
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

	// User Rate Limits
	flag.IntVar(&serverConfig.UserReadRateLimit, "userreadratelimit", serverConfig.UserReadRateLimit, "User Read Rate Limit (requests per second)")
	flag.IntVar(&serverConfig.UserReadRateBurst, "userreadrateburst", serverConfig.UserReadRateBurst, "User Read Rate Burst (maximum number of requests in a burst)")
	flag.IntVar(&serverConfig.UserWriteRateLimit, "userwriteratelimit", serverConfig.UserWriteRateLimit, "User Write Rate Limit (requests per second)")
	flag.IntVar(&serverConfig.UserWriteRateBurst, "userwriterateburst", serverConfig.UserWriteRateBurst, "User Write Rate Burst (maximum number of requests in a burst)")

	// User Limits
	flag.IntVar(&serverConfig.MaxUsers, "maxusers", serverConfig.MaxUsers, "Maximum number of users")
	flag.IntVar(&serverConfig.MaxUserDBSizeMB, "maxuserdbsizemb", serverConfig.MaxUserDBSizeMB, "Maximum user DB size in MB")
	flag.IntVar(&serverConfig.MaxUserInactiveInMins, "maxuserinactiveinmins", serverConfig.MaxUserInactiveInMins, "Maximum user inactive time in minutes (after this they are archived)")

	flag.IntVar(&serverConfig.UserCreateRateLimit, "usercreateratelimit", serverConfig.UserCreateRateLimit, "User Create Rate Limit (requests per second)")
	flag.IntVar(&serverConfig.UserCreateRateBurst, "usercreaterateburst", serverConfig.UserCreateRateBurst, "User Create Rate Burst (maximum number of requests in a burst)")

	// Backup Config
	flag.IntVar(&serverConfig.ControlBackupIntervalInSeconds, "controlbackupinterval", serverConfig.ControlBackupIntervalInSeconds, "Control DB Backup Interval in seconds")
	flag.IntVar(&serverConfig.BatchBackupIntervalInSeconds, "batchbackupinterval", serverConfig.BatchBackupIntervalInSeconds, "Batch Backup Interval in seconds")

	flag.Parse()

	// Deal with mode
	envmode := configStringSelect(dotenvdev, dotenv, "MODE", "development")
	if mode != "" {
		envmode = mode
	}
	switch strings.ToLower(envmode) {
	case "production":
		serverConfig.Mode = Production
	case "development":
		serverConfig.Mode = Development
	default:
		return ServerConfig{}, fmt.Errorf("invalid mode: %s", envmode)
	}

	printServerConfig(serverConfig)
	return serverConfig, nil
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
	fmt.Printf("Mode: %s\n", map[ServerMode]string{Production: "Production", Development: "Development"}[config.Mode])
	fmt.Printf("Host: %s://%s:%d\n", config.HostProtocol, config.HostName, config.HostPort)
	fmt.Printf("S3: %s@%s/%s\n", config.S3AccessKeyID, config.S3URL, config.S3Bucket)
	fmt.Printf("SQLite Path: %s\n", config.SqlitePath)
	fmt.Printf("User Rate Limit r/w req/s: %d/%d burst %d/%d\n", config.UserReadRateLimit, config.UserWriteRateLimit, config.UserReadRateBurst, config.UserWriteRateBurst)
	fmt.Printf("Max Users/User DB Size/Age (mins): %d / %d / %d\n", config.MaxUsers, config.MaxUserDBSizeMB, config.MaxUserInactiveInMins)
	fmt.Printf("User Create Rate Limit: %d req/s burst %d\n", config.UserCreateRateLimit, config.UserCreateRateBurst)
	fmt.Printf("Backup Interval (Control/Batch) in seconds: %d / %d\n", config.ControlBackupIntervalInSeconds, config.BatchBackupIntervalInSeconds)
}
