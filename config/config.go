package config

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
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
	S3Bucket    string
	S3URL       string
	S3AccessKey string
	S3SecretKey string

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
	MaxUsers            int
	MaxUserDBSizeMB     int
	UserCreateRateLimit int
	UserCreateRateBurst int
}

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
		S3Bucket:    configStringSelect(dotenvdev, dotenv, "S3_BUCKET", "mybucket"),
		S3URL:       configStringSelect(dotenvdev, dotenv, "S3_URL", "http://localhost:9000"),
		S3AccessKey: configStringSelect(dotenvdev, dotenv, "S3_ACCESS_KEY", "minioadmin"),
		S3SecretKey: configStringSelect(dotenvdev, dotenv, "S3_SECRET_KEY", "minioadmin"),

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
		MaxUsers:            configIntSelect(dotenvdev, dotenv, "MAX_USERS", 10000),
		MaxUserDBSizeMB:     configIntSelect(dotenvdev, dotenv, "MAX_USER_DB_SIZE_MB", 100),
		UserCreateRateLimit: configIntSelect(dotenvdev, dotenv, "USER_CREATE_RATE_LIMIT", 10),
		UserCreateRateBurst: configIntSelect(dotenvdev, dotenv, "USER_CREATE_RATE_BURST", 100),
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
	flag.StringVar(&serverConfig.S3AccessKey, "s3accesskey", serverConfig.S3AccessKey, "S3 Access Key")
	flag.StringVar(&serverConfig.S3SecretKey, "s3secretkey", serverConfig.S3SecretKey, "S3 Secret Key")

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
	flag.IntVar(&serverConfig.UserCreateRateLimit, "usercreateratelimit", serverConfig.UserCreateRateLimit, "User Create Rate Limit (requests per second)")
	flag.IntVar(&serverConfig.UserCreateRateBurst, "usercreaterateburst", serverConfig.UserCreateRateBurst, "User Create Rate Burst (maximum number of requests in a burst)")

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

	return serverConfig, nil
}

func configStringSelect(dotenvdev map[string]string, dotenv map[string]string, envVar string, defaultValue string) string {
	// First
	if val, ok := os.LookupEnv(envVar); ok {
		return val
	}

	if val, ok := dotenv[envVar]; ok {
		return val
	}
	if val, ok := dotenvdev[envVar]; ok {
		return val
	}
	return defaultValue
}

func configIntSelect(dotenvdev map[string]string, dotenv map[string]string, envVar string, defaultValue int) int {
	// First
	if val, ok := os.LookupEnv(envVar); ok {
		intVal, err := strconv.Atoi(val)
		if err == nil {

			return intVal
		} else {
			fmt.Printf("Error parsing ENV var for %s: %s\n", envVar, err)
		}
	}

	if val, ok := dotenv[envVar]; ok {
		intVal, err := strconv.Atoi(val)
		if err == nil {
			return intVal
		} else {
			fmt.Printf("Error parsing .env var for %s: %s\n", envVar, err)
		}
	}
	if val, ok := dotenvdev[envVar]; ok {
		intVal, err := strconv.Atoi(val)
		if err == nil {
			return intVal
		} else {
			fmt.Printf("Error parsing .env.dev var for %s: %s\n", envVar, err)
		}
	}
	return defaultValue
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
