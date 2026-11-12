package s3utils

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/config"

	"github.com/pierrec/lz4/v4"
)

////
// Most of this is borrowed from LiteStream

// //
type S3Client struct {
	s3           *s3.Client // s3 service
	ServerConfig *config.ServerConfig

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
}

// MaxKeys is the number of keys S3 can operate on per batch.
const MaxKeys = 1000

// DefaultRegion is the region used if one is not specified.
const DefaultRegion = "us-east-1"

// Retention is the tag an upload carries so the bucket's lifecycle rules expire the versions it replaces
type Retention string

const (
	SiteRetention    Retention = "site"
	ControlRetention Retention = "control"
)

const retentionTag = "nighttroop-retention"

func NewS3Client(config *config.ServerConfig) *S3Client {
	c := &S3Client{
		ServerConfig:    config,
		AccessKeyID:     config.S3AccessKeyID,
		SecretAccessKey: config.S3AccessSecretKey,
		ForcePathStyle:  true,
		Bucket:          config.S3Bucket,
		Endpoint:        config.S3URL,
		Region:          DefaultRegion,
		HostName:        config.HostName,
	}

	slog.Info("S3Client Being Created", "bucket", c.Bucket, "region", c.Region, "endpoint", c.Endpoint)

	c.s3 = s3.NewFromConfig(aws.Config{Region: c.Region}, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(c.Endpoint)
		o.Credentials = credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, "")
		o.UsePathStyle = c.ForcePathStyle
	})
	c.uploader = manager.NewUploader(c.s3)

	return c
}

// startWaits are how long Start waits after each failed test before testing again, about half a minute in all, so a
// network still coming up as the server starts is not a failure
var startWaits = []time.Duration{2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second}

// Start turns S3 on when S3_URL is set: it tests uploading to the bucket, again after each of startWaits, and gives the
// last error when none works, as a server that cannot reach its backups must not start, possibly empty, and upload
// that over them. Without S3_URL, S3 is off: nothing is tried, restored or backed up.
func (c *S3Client) Start(ctx context.Context) error {
	return c.start(ctx, c.TestS3Client, startWaits)
}

func (c *S3Client) start(ctx context.Context, test func(context.Context) error, waits []time.Duration) error {
	if c.Endpoint == "" {
		slog.Info("S3_URL is not set, so S3 is off: nothing is restored on start or backed up")
		return nil
	}
	err := test(ctx)
	for _, wait := range waits {
		if err == nil {
			break
		}
		slog.Warn("S3 is not working yet, testing it again", "wait", wait, "error", err)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
		err = test(ctx)
	}
	if err != nil {
		return fmt.Errorf("S3_URL is set but S3 is not working: %w", err)
	}
	c.ServerConfig.Status.S3Working = true
	// Without versioning each backup replaces the last, which is worth an error but not stopping backups
	if err := c.ConfigureVersioning(ctx); err != nil {
		slog.Error("S3 versioning is not set up, so each backup replaces the last", "error", err)
	}
	return nil
}

func (c *S3Client) TestS3Client(ctx context.Context) error {
	testKey := c.TestObjectPath()
	randomGen := utils.NextSiteSubdomain(false)
	testContent := fmt.Sprintf("This istest object %s", randomGen)

	_, err := c.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:  aws.String(c.Bucket),
		Key:     aws.String(testKey),
		Body:    strings.NewReader(testContent),
		Tagging: aws.String(retentionTag + "=" + string(SiteRetention)),
	})

	// Testing the Raw S3Client
	if err != nil {
		return fmt.Errorf("s3Client test upload failed: %w", err)
	}

	output, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
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
	return nil
}

// ConfigureVersioning turns on the bucket's versioning, so an upload keeps the object it replaces as an old version,
// and sets the lifecycle rules that expire those old versions
func (c *S3Client) ConfigureVersioning(ctx context.Context) error {
	_, err := c.s3.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
		Bucket:                  aws.String(c.Bucket),
		VersioningConfiguration: &types.VersioningConfiguration{Status: types.BucketVersioningStatusEnabled},
	})
	if err != nil {
		return fmt.Errorf("cannot enable bucket versioning: %w", err)
	}

	// Lifecycle rules replace the bucket's whole configuration, so keep the rules this server does not own
	rules := c.RetentionRules()
	existing, err := c.s3.GetBucketLifecycleConfiguration(ctx, &s3.GetBucketLifecycleConfigurationInput{Bucket: aws.String(c.Bucket)})
	var apiErr smithy.APIError
	if err != nil && !(errors.As(err, &apiErr) && apiErr.ErrorCode() == "NoSuchLifecycleConfiguration") {
		return fmt.Errorf("cannot get bucket lifecycle: %w", err)
	}
	if existing != nil {
		for _, rule := range existing.Rules {
			if !strings.HasPrefix(aws.ToString(rule.ID), retentionTag) {
				rules = append(rules, rule)
			}
		}
	}

	_, err = c.s3.PutBucketLifecycleConfiguration(ctx, &s3.PutBucketLifecycleConfigurationInput{
		Bucket:                 aws.String(c.Bucket),
		LifecycleConfiguration: &types.BucketLifecycleConfiguration{Rules: rules},
	})
	if err != nil {
		return fmt.Errorf("cannot set bucket lifecycle: %w", err)
	}
	return nil
}

// RetentionRules expire the old versions of the objects tagged with each Retention after its days in the config
func (c *S3Client) RetentionRules() []types.LifecycleRule {
	rules := []types.LifecycleRule{}
	for retention, days := range map[Retention]int{
		SiteRetention:    c.ServerConfig.SiteBackupRetentionDays,
		ControlRetention: c.ServerConfig.ControlBackupRetentionDays,
	} {
		rules = append(rules, types.LifecycleRule{
			ID:                          aws.String(retentionTag + "-" + string(retention)),
			Status:                      types.ExpirationStatusEnabled,
			Filter:                      &types.LifecycleRuleFilter{Tag: &types.Tag{Key: aws.String(retentionTag), Value: aws.String(string(retention))}},
			NoncurrentVersionExpiration: &types.NoncurrentVersionExpiration{NoncurrentDays: aws.Int32(int32(days))},
		})
	}
	return rules
}

// //
// PATHS
// //

func (c *S3Client) TestObjectPath() string {
	return fmt.Sprintf("%s/test.txt", c.HostName)
}

////
// END OF PATHS
////

////
// UPLOAD METHODS
////

// UploadFunc and DownloadFunc are UploadFilesToS3 and DownloadFilesFromS3, which tests can stand in for
type UploadFunc func(ctx context.Context, key string, retention Retention, paths []string) error
type DownloadFunc func(ctx context.Context, key string, outpath string) (int64, bool, error)

func (c *S3Client) UploadFilesToS3(ctx context.Context, key string, retention Retention, paths []string) error {
	if !c.ServerConfig.Status.S3Working {
		return nil
	}
	// The upload finishes even if what asked for it is cancelled, logging with its job, request or site
	ctx = context.WithoutCancel(ctx)

	// Creating the pipes
	pr, pw := io.Pipe()
	defer pr.Close()

	// Create a goroutine to write to the pipe the zip
	go func() {
		defer pw.Close()
		// An upload logs only when it fails, as every backup is one
		_, err := c.createTarLZ4(paths, pw, key)
		if err != nil {
			slog.ErrorContext(ctx, "Error creating tar LZ4", "key", key, "error", err)
			// Important: propagate error to the reader side
			pw.CloseWithError(err)
		}
	}()

	err := c.uploadStreamToS3(ctx, key, retention, pr)
	if err != nil {
		slog.ErrorContext(ctx, "Error uploading stream to s3", "key", key, "error", err)
		return fmt.Errorf("cannot upload stream to s3: %w", err)
	}

	//wg.Wait() // Wait for all uploads to finish
	return nil
}

func (c *S3Client) uploadStreamToS3(ctx context.Context, key string, retention Retention, stream io.Reader) error {
	_, err := c.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:  aws.String(c.Bucket),
		Key:     aws.String(key),
		Body:    stream,
		Tagging: aws.String(retentionTag + "=" + string(retention)),
	})

	if err != nil {
		slog.ErrorContext(ctx, "Failed to upload Stream", "key", key, "error", err)
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

func (c *S3Client) DownloadFilesFromS3(ctx context.Context, key string, outpath string) (int64, bool, error) {
	if !c.ServerConfig.Status.S3Working {
		return 0, false, nil
	}
	info := ""
	defer aslog.TrackTime(ctx, time.Now(), "DownloadFilesFromS3 "+key, nil, &info)

	ctx = context.WithoutCancel(ctx)
	out, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var noSuchKeyError *types.NoSuchKey
		if errors.As(err, &noSuchKeyError) {
			info = "Not Found"
			return 0, false, nil
		}
		info = "Error"
		return 0, false, fmt.Errorf("cannot get object from s3: %w", err)
	}
	defer out.Body.Close()

	// Extract beside outpath and move in only once all of it is there, so a cut off download leaves nothing behind
	outpath = filepath.Clean(outpath)
	if err := utils.MkParentDirs(outpath); err != nil {
		return 0, true, err
	}
	staging, err := os.MkdirTemp(filepath.Dir(outpath), "."+filepath.Base(outpath)+"-restore-")
	if err != nil {
		return 0, true, fmt.Errorf("cannot create staging dir: %w", err)
	}
	defer os.RemoveAll(staging)

	cr := &utils.CountingReader{R: out.Body}
	err = c.ExtractTarLZ4(cr, staging, key)

	info = fmt.Sprintf("downloaded and extracted %d bytes", cr.Bytes)
	if err != nil {
		return cr.Bytes, true, fmt.Errorf("cannot extract tar LZ4: %w", err)
	}

	if err := moveInto(staging, outpath); err != nil {
		return cr.Bytes, true, fmt.Errorf("cannot move download into place: %w", err)
	}
	return cr.Bytes, true, nil
}

// moveInto renames dir to outpath, or when outpath is already a folder, each of dir's entries into it
func moveInto(dir string, outpath string) error {
	if !utils.FileExists(outpath) {
		return os.Rename(dir, outpath)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := os.Rename(filepath.Join(dir, entry.Name()), filepath.Join(outpath, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

////
// END OF DOWNLOAD METHODS
////

// createTarLZ4 writes files to writer as a tar, compressed with LZ4 and encrypted with the server's keys, and is how
// many bytes it wrote
func (c *S3Client) createTarLZ4(files []string, writer io.Writer, objectKey string) (int64, error) {
	cw := &utils.CountingWriter{W: writer}
	// Wrap all up in encryption
	encWriter, err := c.ServerConfig.Security.EncryptingWriterFor(cw, objectKey)
	if err != nil {
		return 0, fmt.Errorf("cannot create encrypting writer: %w", err)
	}

	// Wrap in LZ4 writer
	lz4Writer := lz4.NewWriter(encWriter)

	// Create tar writer
	tarWriter := tar.NewWriter(lz4Writer)

	for _, file := range files {
		if err := addFileToTar(tarWriter, file); err != nil {
			tarWriter.Close()
			lz4Writer.Close()
			encWriter.Close()
			return cw.Bytes, fmt.Errorf("adding %s to tar: %w", file, err)
		}
	}

	// Replace the defers with this at the end of createTarLZ4:
	if err := tarWriter.Close(); err != nil {
		return cw.Bytes, fmt.Errorf("closing tar writer: %w", err)
	}
	if err := lz4Writer.Close(); err != nil {
		return cw.Bytes, fmt.Errorf("closing lz4 writer: %w", err)
	}
	if err := encWriter.Close(); err != nil {
		return cw.Bytes, fmt.Errorf("closing enc writer: %w", err)
	}

	return cw.Bytes, nil
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

// createFile is os.Create, which tests can stand in for to fail a file's close
var createFile = func(name string) (io.WriteCloser, error) { return os.Create(name) }

// ExtractTarLZ4 decrypts, decompresses and untars the backup kept at the object key from reader into outputDir,
// refusing an entry that would be written outside it
func (c *S3Client) ExtractTarLZ4(reader io.Reader, outputDir string, objectKey string) error {

	// Wrap with decryption
	decReader, err := c.ServerConfig.Security.DecryptingReaderFor(reader, objectKey)
	if err != nil {
		return fmt.Errorf("cannot create decrypting reader: %w", err)
	}

	// Wrap with LZ4 reader
	lz4Reader := lz4.NewReader(decReader)

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

		////
		// Security Check: dont let the tar file escape the output directory
		// e.g. if header.Name is "../../etc/passwd" or "/absolute/path/to/file"
		////
		name := filepath.Clean(header.Name)
		if filepath.IsAbs(name) {
			return fmt.Errorf("archive contains absolute path: %s", header.Name)
		}

		targetPath := filepath.Join(outputDir, name)

		rel, err := filepath.Rel(outputDir, targetPath)
		if err != nil {
			return err
		}
		if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("archive entry escapes destination: %s", header.Name)
		}

		////
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("creating dir %s: %w", targetPath, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("creating parent dir: %w", err)
			}

			outFile, err := createFile(targetPath)
			if err != nil {
				return fmt.Errorf("creating file: %w", err)
			}

			// Closed here rather than deferred, as this loops over every file; the write's error is the one returned
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("writing file: %w", err)
			}

			// Close can be the first to report a failed write, such as a full disk, leaving the file cut short.
			// Not logged, as a restore extracts every site's DB; an error is returned instead
			if err := outFile.Close(); err != nil {
				return fmt.Errorf("closing file: %w", err)
			}

			// Restore original file permissions
			if err := os.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("chmod: %w", err)
			}
			if err := os.Chtimes(targetPath, header.ModTime, header.ModTime); err != nil {
				return fmt.Errorf("chtimes: %w", err)
			}
		default:
			slog.Warn("Skipping unsupported type", "type", header.Typeflag)
		}
	}

	return nil
}
