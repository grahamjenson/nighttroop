package db

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"sync"

	"os"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	s3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

////
// Most of this is borrowed from LiteStream

// //
type S3Client struct {
	s3       *s3.S3 // s3 service
	uploader *s3manager.Uploader

	// AWS authentication keys.
	AccessKeyID     string
	SecretAccessKey string

	// S3 bucket information
	Region         string
	Bucket         string
	Path           string
	Endpoint       string
	ForcePathStyle bool
}

// MaxKeys is the number of keys S3 can operate on per batch.
const MaxKeys = 1000

// DefaultRegion is the region used if one is not specified.
const DefaultRegion = "us-east-1"

func NewS3Client(s3URL string) (*S3Client, error) {
	c := &S3Client{
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		ForcePathStyle:  true,
	}

	c.Bucket, c.Region, c.Endpoint = parseS3URL(s3URL)

	// // Look up region if not specified and no endpoint is used.
	// // Endpoints are typically used for non-S3 object stores and do not
	// // necessarily require a region.
	// region := c.Region
	// if region == "" {
	// 	if c.Endpoint == "" {
	// 		if region, err = c.findBucketRegion(ctx, c.Bucket); err != nil {
	// 			return fmt.Errorf("cannot lookup bucket region: %w", err)
	// 		}
	// 	} else {
	// 		region = DefaultRegion // default for non-S3 object stores
	// 	}
	// }

	// Create new AWS session.
	config := c.config()

	sess, err := session.NewSession(config)
	if err != nil {
		return nil, fmt.Errorf("cannot create aws session: %w", err)
	}
	c.s3 = s3.New(sess)
	c.uploader = s3manager.NewUploader(sess)

	fmt.Println("Created New S3 Client", *c)
	return c, nil
}

func (c *S3Client) config() *aws.Config {
	config := &aws.Config{}

	if c.AccessKeyID != "" || c.SecretAccessKey != "" {
		config.Credentials = credentials.NewStaticCredentials(c.AccessKeyID, c.SecretAccessKey, "")
	}
	if c.Endpoint != "" {
		config.Endpoint = aws.String(c.Endpoint)
	}
	if c.ForcePathStyle {
		config.S3ForcePathStyle = aws.Bool(c.ForcePathStyle)
	}

	if c.Region != "" {
		config.Region = aws.String(c.Region)
	}

	return config
}

func (c *S3Client) SyncAllFiles(paths []string) error {
	ctx := context.TODO()
	var wg sync.WaitGroup

	for _, path := range paths {
		wg.Add(1)
		c.uploadToS3(ctx, path, path, &wg) // Concurrent upload
	}

	//wg.Wait() // Wait for all uploads to finish
	return nil
}

func (c *S3Client) uploadToS3(ctx context.Context, filePath, key string, wg *sync.WaitGroup) error {
	defer wg.Done()

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Failed to open file %s: %v\n", filePath, err)
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	_, err = c.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(key),
		Body:   file,
	})

	if err != nil {
		log.Printf("Failed to upload %s: %v\n", filePath, err)
		return fmt.Errorf("failed to upload %s: %w", filePath, err)
	}

	fmt.Printf("Uploaded %s successfully\n", filePath)
	return nil
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

func parseS3URL(s3URL string) (bucket, region, endpoint string) {
	var host, port string
	// Extract port if one is specified.
	parsedURL, err := url.Parse(s3URL)
	if err != nil {
		host = s3URL
	} else {
		host = parsedURL.Hostname()
		port = parsedURL.Port()
		endpoint = parsedURL.Host
	}

	fmt.Printf("HFIHAIOFASOI %s -- %s -- %s -- \n", host, port, err)

	// Extract fields from provider-specific host formats.
	scheme := "https"
	if a := localhostRegex.FindStringSubmatch(host); a != nil {
		fmt.Println("HFIHAIOFASOI")
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

	return bucket, region, endpoint
}

var (
	localhostRegex = regexp.MustCompile(`^(?:(.+)\.)?localhost$`)
)
