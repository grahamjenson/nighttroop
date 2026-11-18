package s3utils

// lock.go: the lock that keeps two servers from backing up to the same host's folder of a bucket, where each would
// replace the other's backups. It is an object written with S3's conditional writes: a new one only if there is none,
// and a held one only if it is as this server last wrote it.

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// The holder writes the lock again every lockHeartbeat. A lock not written for lockStaleAfter is of a server that
// stopped without releasing it, as in a crash, and is taken over. A server waiting for it looks every lockPoll.
var lockHeartbeat, lockStaleAfter, lockPoll = 20 * time.Second, 60 * time.Second, 2 * time.Second

// ErrLockLost is an upload refused as another server took this server's lock, so that server's backups are kept
var ErrLockLost = errors.New("another server took this server's S3 lock, so this server backs up no more")

func (c *S3Client) lockKey() string {
	return c.HostName + "/server.lock"
}

// putLock writes the lock, new each time so its ETag changes: only if there is none when ifMatch is empty, or else only
// if its ETag is still ifMatch. It gives the new ETag, or false when S3 refuses as the condition does not hold.
func (c *S3Client) putLock(ctx context.Context, ifMatch string) (string, bool, error) {
	input := &s3.PutObjectInput{Bucket: aws.String(c.Bucket), Key: aws.String(c.lockKey()),
		Body:    strings.NewReader(time.Now().Format(time.RFC3339Nano)),
		Tagging: aws.String(retentionTag + "=" + string(LockRetention))}
	if ifMatch == "" {
		input.IfNoneMatch = aws.String("*")
	} else {
		input.IfMatch = aws.String(ifMatch)
	}
	output, err := c.s3.PutObject(ctx, input)
	var response *awshttp.ResponseError
	if errors.As(err, &response) && (response.HTTPStatusCode() == 412 || response.HTTPStatusCode() == 409) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return aws.ToString(output.ETag), true, nil
}

// Lock takes the lock, and keeps it until Unlock. While another server holds it, it waits: for that server to stop and
// release it, as in a deploy, or for lockStaleAfter without that server writing it, as after a crash. It ends when ctx
// does. Without S3, or with one that does not refuse a conditional write, there is no lock.
func (c *S3Client) Lock(ctx context.Context) error {
	if !c.ServerConfig.Status.S3Working {
		return nil
	}
	seen, since := "", time.Now()
	for {
		etag, held, err := c.putLock(ctx, "")
		if err != nil {
			return fmt.Errorf("cannot take the S3 lock: %w", err)
		}
		if held {
			if _, again, _ := c.putLock(ctx, ""); again {
				slog.Error("S3 does not refuse a conditional write, so there is no lock: do not run two servers on this bucket")
				return nil
			}
		} else if head, err := c.s3.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(c.Bucket), Key: aws.String(c.lockKey())}); err == nil {
			if current := aws.ToString(head.ETag); current != seen {
				seen, since = current, time.Now()
				slog.Warn("Another server holds the S3 lock, waiting for it to stop", "key", c.lockKey())
			} else if time.Since(since) >= lockStaleAfter {
				if etag, held, err = c.putLock(ctx, seen); err != nil {
					return fmt.Errorf("cannot take over the S3 lock: %w", err)
				}
			}
		}
		if held {
			keep, stop := context.WithCancel(context.Background())
			c.unlock, c.lockDone = stop, make(chan struct{})
			go c.keepLock(keep, etag)
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(lockPoll):
		}
	}
}

// keepLock writes the lock every lockHeartbeat until ctx ends, then deletes it. If another server took it, it stops
// this server's uploads, and leaves the lock to that server.
func (c *S3Client) keepLock(ctx context.Context, etag string) {
	defer close(c.lockDone)
	for {
		select {
		case <-ctx.Done():
			if _, err := c.s3.DeleteObject(context.Background(), &s3.DeleteObjectInput{Bucket: aws.String(c.Bucket), Key: aws.String(c.lockKey())}); err != nil {
				slog.Error("Cannot release the S3 lock, so the next start waits for it", "error", err)
			}
			return
		case <-time.After(lockHeartbeat):
		}
		next, held, err := c.putLock(ctx, etag)
		switch {
		case err != nil:
			slog.Warn("Cannot write the S3 lock, trying again", "error", err)
		case !held:
			c.lockLost.Store(true)
			slog.Error("Another server took this server's S3 lock, so this server backs up no more: stop one of them")
			return
		default:
			etag = next
		}
	}
}

// Unlock releases the lock, after the last backup, so the next server to start takes it at once
func (c *S3Client) Unlock() {
	if c.unlock != nil {
		c.unlock()
		<-c.lockDone
	}
}
