package db

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"path/filepath"
	"sync"
	"time"

	"os"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	s3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pierrec/lz4/v4"
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
		c.uploadToS3(ctx, path, path)
		wg.Done() // Concurrent upload
	}

	//wg.Wait() // Wait for all uploads to finish
	return nil
}

func (c *S3Client) SyncAllFilesToSingleLZ4(paths []string) error {
	defer trackTime(time.Now(), "SyncAllFilesToSingleLZ4")
	fmt.Printf("Syncing all files to single LZ4 %v\n", len(paths))
	ctx := context.TODO()

	fmt.Println("Creating pipes")
	pr, pw := io.Pipe()
	defer pr.Close()

	fmt.Println("Creating tar LZ4")

	go func() {
		defer pw.Close()
		err := createTarLZ4(paths, pw)
		if err != nil {
			fmt.Println("Error creating tar LZ4", err)
			// Important: propagate error to the reader side
			pw.CloseWithError(err)
		}
	}()

	fmt.Println("Streaming to S3")
	err := c.uploadStreamToS3(ctx, "backup.tar.lz4", pr)
	if err != nil {
		fmt.Println("Ennot upload stream to s", err)
		return fmt.Errorf("cannot upload stream to s3: %w", err)
	}

	//wg.Wait() // Wait for all uploads to finish
	return nil
}

func (c *S3Client) uploadToS3(ctx context.Context, filePath, key string) error {

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

func (c *S3Client) uploadStreamToS3(ctx context.Context, key string, stream io.Reader) error {
	_, err := c.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(key),
		Body:   stream,
	})

	if err != nil {
		log.Printf("Failed to upload Stream: %v\n", err)
		return fmt.Errorf("failed to upload Stream: %w", err)
	}

	fmt.Printf("Uploaded Stream to S3 successfully\n")
	return nil
}

func (c *S3Client) DownloadFile(s3path string, outpath string) error {
	// This function downloads a path to a tenant
	ctx := context.TODO()
	out, err := c.s3.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(s3path),
	})
	if err != nil {
		return fmt.Errorf("cannot get object from s3: %w", err)
	}
	defer out.Body.Close()
	file, err := os.Create(outpath)
	if err != nil {
		return fmt.Errorf("cannot create file: %w", err)
	}
	defer file.Close()
	_, err = io.Copy(file, out.Body)
	if err != nil {
		return fmt.Errorf("cannot copy file: %w", err)
	}
	fmt.Printf("Restored %s successfully\n", s3path)
	return nil
}

///
// Files
///

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

///// SLOP

func createTarLZ4(files []string, writer io.Writer) error {
	// Wrap in LZ4 writer
	lz4Writer := lz4.NewWriter(writer)
	defer lz4Writer.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(lz4Writer)
	defer tarWriter.Close()

	for _, file := range files {
		if err := addFileToTar(tarWriter, file); err != nil {
			return fmt.Errorf("adding %s to tar: %w", file, err)
		}
	}

	fmt.Printf("✅ Created archive")
	return nil
}

func addFileToTar(tw *tar.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return fmt.Errorf("header: %w", err)
	}

	header.Name = filepath.Base(filePath)

	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	_, err = io.Copy(tw, file)
	return err
}
