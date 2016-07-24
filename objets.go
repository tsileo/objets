package objets

import (
	_ "crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/dkumor/acmewrapper"
	"github.com/tsileo/s3layer"
	"github.com/tsileo/s3layer/multipart"
)

// FIXME(tsileo):
// - store the acl in a kv file (graceful shutdown)

type Objets struct {
	conf *Config
	s4   *s3layer.S4
	*multipart.MultipartUploadHandler
}

// MulitpartList     func(uploadID string, maxParts, partNumberMarker int) ([]*UploadPartResponse, error) // FIXME(tsileo): handle a time.Time in the struct

func New(confPath string) (*Objets, error) {
	conf, err := newConfig(confPath)
	if err != nil {
		return nil, err
	}
	objets := &Objets{
		conf: conf,
	}
	dir, err := ioutil.TempDir("", "objets_multipart")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	objets.MultipartUploadHandler = multipart.New(dir, objets.PutObject)
	objets.s4 = &s3layer.S4{
		S3Layer: objets,
	}
	return objets, nil
}

func (o *Objets) Buckets() ([]*s3layer.Bucket, error) {
	res := []*s3layer.Bucket{&s3layer.Bucket{Name: "testing", CreationDate: time.Now()}}
	// 2006-02-03T16:45:09.000Z
	return res, nil
}

func (o *Objets) Cred(accessKeyID string) (string, error) {
	log.Printf("Cred(%s)\n", accessKeyID)
	if o.conf.AccessKeyID != accessKeyID {
		return "", s3layer.ErrUnknownAccessKeyID
	}
	return o.conf.SecretAccessKey, nil
}

func (o *Objets) DeleteObject(bucket, key string) error                        { return nil }
func (o *Objets) DeleteBucket(bucket string) error                             { return nil }
func (o *Objets) PutBucket(bucket string, acl s3layer.CannedACL) error         { return nil }
func (o *Objets) PutObjectAcl(bucket, key string, acl s3layer.CannedACL) error { return nil }

func (o *Objets) ListBucket(bucket, prefix string) ([]*s3layer.ListBucketResultContent, []*s3layer.ListBucketResultPrefix, error) {
	log.Printf("ListBucket(%s, %s)\n", bucket, prefix)
	if containsDotDot(prefix) || containsDotDot(bucket) {
		return nil, nil, fmt.Errorf("invalid prefix/bucket")
	}
	contents := []*s3layer.ListBucketResultContent{}
	prefixes := []*s3layer.ListBucketResultPrefix{}

	path := filepath.Join(o.conf.DataDir, bucket, prefix)
	log.Printf("ListBucket path=%s\n", path)
	dir, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	fis, err := dir.Readdir(-1)
	log.Printf("ListBucket fis=%+v, err=%v\n", fis, err)
	if err != nil {
		return nil, nil, err
	}

	for _, fi := range fis {
		if fi.IsDir() {
			prefixes = append(prefixes, &s3layer.ListBucketResultPrefix{filepath.Join(prefix, fi.Name()) + "/"})
		} else {
			contents = append(contents, &s3layer.ListBucketResultContent{
				Key:          filepath.Join(prefix, fi.Name()),
				Size:         int(fi.Size()),
				LastModified: fi.ModTime().Format(s3layer.S3Date),
			})
		}
	}

	return contents, prefixes, nil
}

func (o *Objets) GetObject(bucket, key string) (io.Reader, error) {
	log.Printf("GetObject(%s, %s)\n", bucket, key)
	if containsDotDot(key) || containsDotDot(bucket) {
		return nil, fmt.Errorf("invalid key/bucket")
	}
	path := filepath.Join(o.conf.DataDir, bucket, key)
	log.Printf("GetObject path=%s\n", path)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return f, err
}

func (o *Objets) PutObject(bucket, key string, reader io.Reader, acl s3layer.CannedACL) error {
	log.Printf("PutObject(%s, %s, <reader>)\n", bucket, key, reader)
	if containsDotDot(key) || containsDotDot(bucket) {
		return fmt.Errorf("invalid key/bucket")
	}
	if err := os.MkdirAll(filepath.Join(o.conf.DataDir, bucket, filepath.Dir(key)), 0644); err != nil {
		return err
	}
	path := filepath.Join(o.conf.DataDir, bucket, key)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, reader); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	log.Printf("PutObject path=%s\n", path)
	return f.Close()
}

// borrowed from net/http
func containsDotDot(v string) bool {
	if !strings.Contains(v, "..") {
		return false
	}
	for _, ent := range strings.FieldsFunc(v, isSlashRune) {
		if ent == ".." {
			return true
		}
	}
	return false
}

func isSlashRune(r rune) bool { return r == '/' || r == '\\' }