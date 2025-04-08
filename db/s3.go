package db

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"path/filepath"
	"time"

	"os"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/pierrec/lz4/v4"
)

////
// Most of this is borrowed from LiteStream

// //
type S3Client struct {
	s3       *s3.Client // s3 service
	uploader *manager.Uploader

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

func NewS3Client(endpoint, bucket string) (*S3Client, error) {
	c := &S3Client{
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		ForcePathStyle:  true,
		Bucket:          bucket,
		Endpoint:        endpoint,
		Region:          DefaultRegion,
	}

	fmt.Println("Parsed S3 URL", c.Bucket, c.Region, c.Endpoint)
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

	c.s3 = s3.NewFromConfig(aws.Config{Region: c.Region}, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(c.Endpoint)
		o.Credentials = credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, "")
		o.UsePathStyle = c.ForcePathStyle
	})
	c.uploader = manager.NewUploader(c.s3)

	fmt.Println("Created New S3 Client", *c)
	return c, nil
}

func (c *S3Client) SyncAllFiles(paths []string) error {
	ctx := context.TODO()

	for _, path := range paths {
		c.uploadToS3(ctx, path, path)
	}

	return nil
}

func (c *S3Client) SyncAllFilesToSingleLZ4(key string, paths []string) error {
	defer trackTime(time.Now(), "SyncAllFilesToSingleLZ4 "+key)
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
	err := c.uploadStreamToS3(ctx, key, pr)
	if err != nil {
		fmt.Println("Ennot upload stream to s", err)
		return fmt.Errorf("cannot upload stream to s3: %w", err)
	}

	//wg.Wait() // Wait for all uploads to finish
	return nil
}

func (c *S3Client) DownloadAndExtractLZ4(key string, outpath string) (bool, error) {
	// This function downloads a path to a tenant
	ctx := context.TODO()
	out, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var noSuchKeyError *types.NoSuchKey
		if errors.As(err, &noSuchKeyError) {
			return false, nil
		}
		return false, fmt.Errorf("cannot get object from s3: %w", err)
	}
	defer out.Body.Close()

	err = extractTarLZ4(out.Body, outpath)
	if err != nil {
		return true, fmt.Errorf("cannot extract tar LZ4: %w", err)
	}

	return true, nil
}

func (c *S3Client) uploadToS3(ctx context.Context, filePath, key string) error {

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Failed to open file %s: %v\n", filePath, err)
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	_, err = c.uploader.Upload(ctx, &s3.PutObjectInput{
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
	_, err := c.uploader.Upload(ctx, &s3.PutObjectInput{
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

func (c *S3Client) DownloadFile(s3path string, outpath string) (bool, error) {
	// This function downloads a path to a tenant
	ctx := context.TODO()
	out, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(s3path),
	})
	if err != nil {
		var noSuchKeyError *types.NoSuchKey
		if errors.As(err, &noSuchKeyError) {
			return false, nil
		}

		return false, fmt.Errorf("cannot get object from s3: %w", err)
	}
	defer out.Body.Close()

	file, err := os.Create(outpath)
	if err != nil {
		return true, fmt.Errorf("cannot create file: %w", err)
	}
	defer file.Close()
	_, err = io.Copy(file, out.Body)
	if err != nil {
		return true, fmt.Errorf("cannot copy file: %w", err)
	}
	fmt.Printf("Restored %s successfully\n", s3path)
	return true, nil
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

func extractTarLZ4(reader io.Reader, outputDir string) error {
	// Wrap with LZ4 reader
	lz4Reader := lz4.NewReader(reader)

	// Create tar reader from decompressed stream
	tarReader := tar.NewReader(lz4Reader)

	for {

		header, err := tarReader.Next()

		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("reading tar entry: %w", err)
		}

		targetPath := filepath.Join(outputDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("creating dir %s: %w", targetPath, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("creating parent dir: %w", err)
			}

			outFile, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("creating file: %w", err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("writing file: %w", err)
			}
			fmt.Println("Extraction of ", targetPath, " complete")
			outFile.Close()

			// Restore original file permissions
			if err := os.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("chmod: %w", err)
			}
		default:
			fmt.Printf("Skipping unsupported type: %v\n", header.Typeflag)
		}
	}

	fmt.Println("✅ Extraction complete")
	return nil
}
