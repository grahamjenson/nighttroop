package replica

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path"
	"time"

	"github.com/benbjohnson/litestream"
	"github.com/superfly/ltx"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/security"
)

// client keeps one DB's replica in the server's bucket, under prefix, each LTX file encrypted with the server's keys
// for its object key. Every DB's client shares the server's one S3 client and its connections, where Litestream's own
// would make a client, and keep connections, for each.
type client struct {
	s3       *s3utils.S3Client
	security *security.SecurityConfig
	prefix   string
}

var _ litestream.ReplicaClient = (*client)(nil)

func (c *client) Type() string               { return "s3" }
func (c *client) Init(context.Context) error { return nil }
func (c *client) SetLogger(*slog.Logger)     {}
func (c *client) key(level int, min, max ltx.TXID) string {
	return fmt.Sprintf("%s/%04x/%s", c.prefix, level, ltx.FormatFilename(min, max))
}

// LTXFiles are the level's files from seek, in order
func (c *client) LTXFiles(ctx context.Context, level int, seek ltx.TXID, _ bool) (ltx.FileIterator, error) {
	objects, err := c.s3.List(ctx, fmt.Sprintf("%s/%04x/", c.prefix, level))
	if err != nil {
		return nil, err
	}
	infos := []*ltx.FileInfo{}
	for _, object := range objects {
		min, max, err := ltx.ParseFilename(path.Base(*object.Key))
		if err != nil || min < seek {
			continue
		}
		infos = append(infos, &ltx.FileInfo{Level: level, MinTXID: min, MaxTXID: max, Size: plainSize(*object.Size), CreatedAt: object.LastModified.UTC()})
	}
	return ltx.NewFileInfoSliceIterator(infos), nil
}

// plainSize is at most the size of the LTX file that encrypted is size bytes: Tink's streaming encryption adds a
// header of at most 40 bytes, and a tag of at most 32 bytes to each segment of at least 4 KB. Litestream reads a file
// that ends before its size as one cut off, so it is never more; and it is at least an LTX header, as every file has.
func plainSize(size int64) int64 {
	return max(ltx.HeaderSize, size-40-32*((size+4095)/4096))
}

// OpenLTXFile is the file decrypted, from offset and size bytes long, or all of it when size is 0
func (c *client) OpenLTXFile(ctx context.Context, level int, min, max ltx.TXID, offset, size int64) (io.ReadCloser, error) {
	key := c.key(level, min, max)
	body, err := c.s3.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	plain, err := c.security.DecryptingReaderFor(body, key)
	if err == nil {
		_, err = io.CopyN(io.Discard, plain, offset)
	}
	if err != nil {
		body.Close()
		return nil, err
	}
	if size > 0 {
		plain = io.LimitReader(plain, size)
	}
	return struct {
		io.Reader
		io.Closer
	}{plain, body}, nil
}

// WriteLTXFile encrypts and uploads the file
func (c *client) WriteLTXFile(ctx context.Context, level int, min, max ltx.TXID, r io.Reader) (*ltx.FileInfo, error) {
	key := c.key(level, min, max)
	pr, pw := io.Pipe()
	encrypted := make(chan struct{})
	go func() {
		defer close(encrypted)
		enc, err := c.security.EncryptingWriterFor(pw, key)
		if err == nil {
			_, err = io.Copy(enc, r)
		}
		if err == nil {
			err = enc.Close()
		}
		pw.CloseWithError(err)
	}()
	size, err := c.s3.Put(ctx, key, pr)
	// Litestream may reuse r once this returns, so it waits for the encryption to stop reading it
	pr.CloseWithError(fmt.Errorf("upload stopped: %w", err))
	<-encrypted
	if err != nil {
		return nil, err
	}
	return &ltx.FileInfo{Level: level, MinTXID: min, MaxTXID: max, Size: plainSize(size), CreatedAt: time.Now().UTC()}, nil
}

func (c *client) DeleteLTXFiles(ctx context.Context, infos []*ltx.FileInfo) error {
	keys := make([]string, len(infos))
	for i, info := range infos {
		keys[i] = c.key(info.Level, info.MinTXID, info.MaxTXID)
	}
	return c.s3.Delete(ctx, keys)
}

// DeleteAll deletes the whole replica
func (c *client) DeleteAll(ctx context.Context) error {
	objects, err := c.s3.List(ctx, c.prefix+"/")
	if err != nil {
		return err
	}
	keys := make([]string, len(objects))
	for i, object := range objects {
		keys[i] = *object.Key
	}
	return c.s3.Delete(ctx, keys)
}
