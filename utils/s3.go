package utils

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"
	"time"

	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"nighttroop.t1d.nz/config"

	"github.com/pierrec/lz4/v4"
	"github.com/tink-crypto/tink-go/v2/keyset"
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

	// Common Config Varaibles Used In Paths
	HostName string
	kh       *keyset.Handle
}

// MaxKeys is the number of keys S3 can operate on per batch.
const MaxKeys = 1000

// DefaultRegion is the region used if one is not specified.
const DefaultRegion = "us-east-1"

func NewS3Client(config config.ServerConfig) *S3Client {
	c := &S3Client{
		AccessKeyID:     config.S3AccessKeyID,
		SecretAccessKey: config.S3AccessSecretKey,
		ForcePathStyle:  true,
		Bucket:          config.S3Bucket,
		Endpoint:        config.S3URL,
		Region:          DefaultRegion,
		HostName:        config.HostName,
	}

	fmt.Println("S3Client Being Created", c.Bucket, c.Region, c.Endpoint)

	c.s3 = s3.NewFromConfig(aws.Config{Region: c.Region}, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(c.Endpoint)
		o.Credentials = credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, "")
		o.UsePathStyle = c.ForcePathStyle
	})
	c.uploader = manager.NewUploader(c.s3)

	return c
}

func (c *S3Client) TestS3Client() error {
	testKey := c.TestObjectPath()
	randomGen := NextUserID()
	testContent := fmt.Sprintf("This istest object %s", randomGen)

	_, err := c.s3.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(testKey),
		Body:   strings.NewReader(testContent),
	})

	// Testing the Raw S3Client
	if err != nil {
		return fmt.Errorf("s3Client test upload failed: %w", err)
	}

	output, err := c.s3.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(testKey),
	})

	if err != nil {
		return fmt.Errorf("s3Client test download failed: %w", err)
	}

	defer output.Body.Close()
	bodyBytes, err := io.ReadAll(output.Body)
	if err != nil {
		return fmt.Errorf("s3Client test read failed: %w", err)
	}

	bodyString := string(bodyBytes)
	if bodyString != testContent {
		return fmt.Errorf("s3Client test content mismatch: expected %s, got %s", testContent, bodyString)
	}

	// Testing the upload and download stuff, which means key encryption and streaming
	// TODO test encryption and decryption of files
	// err = c.UploadFilesToS3(testKey, []string{})
	// if err != nil {
	// 	return fmt.Errorf("s3Client test empty upload failed: %w", err)
	// }
	return nil
}

// //
// PATHS
// //
func (c *S3Client) ControlS3Path() string {
	return fmt.Sprintf("%s/backup/control.sqlite", c.HostName)
}

func (c *S3Client) UserBatchS3Path(char string) string {
	return fmt.Sprintf("%s/backup/user/%s.lz4", c.HostName, char)
}

func (c *S3Client) UserArchiveS3Path(subdomain string) string {
	char := subdomain[0]
	return fmt.Sprintf("%s/archive/user/%s/%s.sqlite", c.HostName, string(char), subdomain)
}

func (c *S3Client) TestObjectPath() string {
	return fmt.Sprintf("%s/test.txt", c.HostName)
}

////
// END OF PATHS
////

////
// UPLOAD METHODS
////

func (c *S3Client) UploadFilesToS3(key string, paths []string) error {
	defer TrackTime(time.Now(), "UploadFilesToS3 "+key, nil, nil)
	if len(paths) == 0 {
		fmt.Printf("No files to sync for %s\n", key)
		return nil
	}
	fmt.Printf("Syncing all files to single LZ4 %v\n", len(paths))
	ctx := context.TODO()

	// Creating the pipes
	pr, pw := io.Pipe()
	defer pr.Close()

	// Create a goroutine to write to the pipe the zip
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

	return nil
}

////
// END OF UPLOAD METHODS
////

////
// DOWNLOAD METHODS
////

func (c *S3Client) DownloadFilesFromS3(key string, outpath string) (bool, error) {
	// This function downloads a path to a user
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

////
// END OF DOWNLOAD METHODS
////

// //
// TODO SLOP TO BE REVIEWED
// //
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
	header.ModTime = info.ModTime()

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
			if err := os.Chtimes(targetPath, header.ModTime, header.ModTime); err != nil {
				return fmt.Errorf("chtimes: %w", err)
			}
		default:
			fmt.Printf("Skipping unsupported type: %v\n", header.Typeflag)
		}
	}

	fmt.Println("✅ Extraction complete")
	return nil
}
