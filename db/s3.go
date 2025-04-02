package db

import (
	"fmt"
	"net"
	"os"
	"regexp"
)

type S3Config struct {
	Bucket    string
	Region    string
	Endpoint  string
	AccessKey string
	SecretKey string
}

///
// Files
///

func SetupFileStructure() error {
	var err error

	err = os.MkdirAll("data/db/tenant", 0755)
	if err != nil {
		return fmt.Errorf("cannot create tenant data directory: %w", err)
	}

	// where the backups sit
	err = os.MkdirAll("data/backup/tenant", 0755)
	if err != nil {
		return fmt.Errorf("cannot create tenant data directory: %w", err)
	}

	return nil
}

////
// Borrowed from litestream
////

// ParseHost extracts data from a hostname depending on the service provider.
func CreateS3Config(s3URL string, accessKey string, secretKey string) *S3Config {
	// Extract port if one is specified.
	host, port, err := net.SplitHostPort(s3URL)
	if err != nil {
		host = s3URL
	}

	var bucket, region, endpoint string

	// Extract fields from provider-specific host formats.
	scheme := "https"
	if a := localhostRegex.FindStringSubmatch(host); a != nil {
		bucket, region = a[1], "us-east-1"
		scheme, endpoint = "http", "localhost"
	} else {
		bucket = host
	}

	// Add port back to endpoint, if available.
	if endpoint != "" && port != "" {
		endpoint = net.JoinHostPort(endpoint, port)
	}

	// Prepend scheme to endpoint.
	if endpoint != "" {
		endpoint = scheme + "://" + endpoint
	}

	return &S3Config{Bucket: bucket, Region: region, Endpoint: endpoint, AccessKey: accessKey, SecretKey: secretKey}
}

var (
	localhostRegex = regexp.MustCompile(`^(?:(.+)\.)?localhost$`)
)
