package s3utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
)

// S3Client is the server's S3 bucket, for the replicas of its DBs
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

// retentionRule starts the ID of each lifecycle rule this server owns
const retentionRule = "nighttroop-retention"

// ReplicaPrefix is where the replicas of the server's DBs are kept: the control DB's at <prefix>control and each
// site's at <prefix>site/<site ID>
func (c *S3Client) ReplicaPrefix() string {
	return c.HostName + "/litestream/"
}

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
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(testKey),
		Body:   strings.NewReader(testContent),
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
			if !strings.HasPrefix(aws.ToString(rule.ID), retentionRule) {
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

// RetentionRules expire the old versions of the control DB's and the sites' replicas after their days in the config
func (c *S3Client) RetentionRules() []types.LifecycleRule {
	rules := []types.LifecycleRule{}
	for name, days := range map[string]int{"site": c.ServerConfig.SiteBackupRetentionDays, "control": c.ServerConfig.ControlBackupRetentionDays} {
		rules = append(rules, types.LifecycleRule{
			ID:                          aws.String(retentionRule + "-" + name),
			Status:                      types.ExpirationStatusEnabled,
			Filter:                      &types.LifecycleRuleFilter{Prefix: aws.String(c.ReplicaPrefix() + name)},
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
// OBJECT METHODS
////

// transferTimeout is the longest a request to S3 takes: an S3 that stops answering mid transfer would otherwise hold
// the DB's replication until a restart
var transferTimeout = 30 * time.Minute

// singlePutBytes is the largest object Put sends in one request, as a DB's changes nearly always are, so most uploads
// hold no more than it; a larger one, such as a DB's snapshot, is sent in parts by the uploader
const singlePutBytes = 1 << 20

// Put uploads body to key, and is how many bytes it sent
func (c *S3Client) Put(ctx context.Context, key string, body io.Reader) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, transferTimeout)
	defer cancel()
	head, err := io.ReadAll(io.LimitReader(body, singlePutBytes))
	if err != nil {
		return 0, err
	}
	if len(head) < singlePutBytes {
		_, err = c.s3.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(c.Bucket), Key: aws.String(key), Body: bytes.NewReader(head)})
		return int64(len(head)), err
	}
	cr := &utils.CountingReader{R: io.MultiReader(bytes.NewReader(head), body)}
	_, err = c.uploader.Upload(ctx, &s3.PutObjectInput{Bucket: aws.String(c.Bucket), Key: aws.String(key), Body: cr})
	return cr.Bytes, err
}

// Get is key's object, which the caller closes, or os.ErrNotExist when there is none
func (c *S3Client) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	ctx, cancel := context.WithTimeout(ctx, transferTimeout)
	out, err := c.s3.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(c.Bucket), Key: aws.String(key)})
	var noSuchKey *types.NoSuchKey
	if errors.As(err, &noSuchKey) {
		err = os.ErrNotExist
	}
	if err != nil {
		cancel()
		return nil, err
	}
	return &body{ReadCloser: out.Body, cancel: cancel}, nil
}

// body is an object's body, whose request's timeout is cancelled once it is closed
type body struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (b *body) Close() error {
	defer b.cancel()
	return b.ReadCloser.Close()
}

// List is the objects under prefix
func (c *S3Client) List(ctx context.Context, prefix string) ([]types.Object, error) {
	ctx, cancel := context.WithTimeout(ctx, transferTimeout)
	defer cancel()
	objects := []types.Object{}
	pages := s3.NewListObjectsV2Paginator(c.s3, &s3.ListObjectsV2Input{Bucket: aws.String(c.Bucket), Prefix: aws.String(prefix)})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		objects = append(objects, page.Contents...)
	}
	return objects, nil
}

// Delete deletes the objects at keys, MaxKeys a request
func (c *S3Client) Delete(ctx context.Context, keys []string) error {
	ctx, cancel := context.WithTimeout(ctx, transferTimeout)
	defer cancel()
	for chunk := range slices.Chunk(keys, MaxKeys) {
		ids := make([]types.ObjectIdentifier, len(chunk))
		for i, key := range chunk {
			ids[i] = types.ObjectIdentifier{Key: aws.String(key)}
		}
		out, err := c.s3.DeleteObjects(ctx, &s3.DeleteObjectsInput{Bucket: aws.String(c.Bucket), Delete: &types.Delete{Objects: ids, Quiet: aws.Bool(true)}})
		if err == nil && len(out.Errors) > 0 {
			err = fmt.Errorf("cannot delete %s: %s", aws.ToString(out.Errors[0].Key), aws.ToString(out.Errors[0].Message))
		}
		if err != nil {
			return err
		}
	}
	return nil
}
