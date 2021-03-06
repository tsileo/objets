package objets

import (
	_ "crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/dkumor/acmewrapper"
	"github.com/tsileo/s3layer"
	"github.com/tsileo/s3layer/multipart"
)

func init() {
	// Enable additional debug output
	s3layer.Debug = true
}

var (
	// Default subdirectory for storing buckets
	bucketDir = "buckets"
)

type Objets struct {
	tmpDir string
	conf   *Config
	acl    *ACL
	s4     *s3layer.S4
	*multipart.MultipartUploadHandler
}

// MulitpartList     func(uploadID string, maxParts, partNumberMarker int) ([]*UploadPartResponse, error) // FIXME(tsileo): handle a time.Time in the struct

func New(confPath string) (*Objets, error) {
	conf, err := newConfig(confPath)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(conf.DataDir(), bucketDir), os.ModeDir|0755); err != nil {
		return nil, err
	}
	acl, err := newACL(filepath.Join(conf.DataDir(), "acl.db"))
	if err != nil {
		return nil, err
	}
	dir, err := ioutil.TempDir("", "objets_multipart")
	if err != nil {
		panic(err)
	}

	objets := &Objets{
		tmpDir: dir,
		conf:   conf,
		acl:    acl,
	}

	objets.MultipartUploadHandler = multipart.New(dir, objets.PutObject)
	objets.s4 = &s3layer.S4{
		S3Layer: objets,
	}
	return objets, nil
}

func (o *Objets) Close() error {
	if err := os.RemoveAll(o.tmpDir); err != nil {
		return err
	}

	return o.acl.Close()
}

func (o *Objets) StatObject(bucket, key string) (bool, s3layer.CannedACL, error) {
	_, err := os.Stat(filepath.Join(o.conf.DataDir(), bucketDir, bucket, key))
	if os.IsNotExist(err) {
		return false, s3layer.Empty, nil
	}
	if err != nil {
		return false, s3layer.Empty, err
	}
	acl, err := o.acl.Get(bucket, key)
	if err != nil {
		return true, s3layer.Empty, err
	}
	return true, acl, nil
}

func (o *Objets) Buckets() ([]*s3layer.Bucket, error) {
	res := []*s3layer.Bucket{}
	root, err := os.Open(filepath.Join(o.conf.DataDir(), bucketDir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, s3layer.ErrKeyNotFound
		}
		return nil, err
	}

	fis, err := root.Readdir(-1)
	if err != nil {
		return nil, err
	}
	for _, fi := range fis {
		if fi.IsDir() {
			res = append(res, &s3layer.Bucket{Name: fi.Name(), CreationDate: fi.ModTime()})
		}
	}
	return res, nil
}

func (o *Objets) Cred(accessKeyID string) (string, error) {
	if o.conf.AccessKeyID != accessKeyID {
		return "", s3layer.ErrUnknownAccessKeyID
	}
	return o.conf.SecretAccessKey, nil
}

func (o *Objets) DeleteObject(bucket, key string) error {
	if err := o.acl.Remove(bucket, key); err != nil {
		return err
	}
	return os.Remove(filepath.Join(o.conf.DataDir(), bucketDir, bucket, key))
}

func (o *Objets) DeleteBucket(bucket string) error {
	bpath := filepath.Join(o.conf.DataDir(), bucketDir, bucket)
	if _, err := os.Stat(bpath); os.IsNotExist(err) {
		return s3layer.ErrBucketNotFound
	}
	delFunc := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			p := strings.Replace(path, filepath.Join(o.conf.DataDir(), bucketDir, bucket), "", -1)
			return o.acl.Remove(bucket, p[1:])
		}
		return nil
	}
	if err := filepath.Walk(filepath.Join(o.conf.DataDir(), bucketDir, bucket), delFunc); err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(o.conf.DataDir(), bucketDir, bucket))
}

func (o *Objets) PutBucket(bucket string) error {
	return os.MkdirAll(filepath.Join(o.conf.DataDir(), bucketDir, bucket), os.ModeDir|0755)
}

func (o *Objets) PutObjectAcl(bucket, key string, acl s3layer.CannedACL) error {
	exist, _, err := o.StatObject(bucket, key)
	if err != nil {
		return err
	}
	if !exist {
		return s3layer.ErrKeyNotFound
	}
	return o.acl.Set(bucket, key, acl)
}

func (o *Objets) ListBucket(bucket, prefix string) ([]*s3layer.ListBucketResultContent, []*s3layer.ListBucketResultPrefix, error) {
	if containsDotDot(prefix) || containsDotDot(bucket) {
		return nil, nil, fmt.Errorf("invalid prefix/bucket")
	}
	contents := []*s3layer.ListBucketResultContent{}
	prefixes := []*s3layer.ListBucketResultPrefix{}

	path := filepath.Join(o.conf.DataDir(), bucketDir, bucket, prefix)
	dir, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, s3layer.ErrBucketNotFound
		}
		return nil, nil, err
	}

	fis, err := dir.Readdir(-1)
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

func (o *Objets) GetObject(bucket, key string) (io.Reader, s3layer.CannedACL, error) {
	if containsDotDot(key) || containsDotDot(bucket) {
		return nil, s3layer.Empty, fmt.Errorf("invalid key/bucket")
	}
	path := filepath.Join(o.conf.DataDir(), bucketDir, bucket, key)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, s3layer.Empty, s3layer.ErrKeyNotFound
		}
		return nil, s3layer.Empty, err
	}
	acl, err := o.acl.Get(bucket, key)
	if err != nil {
		return nil, s3layer.Empty, err
	}
	return f, acl, err
}

func (o *Objets) PutObject(bucket, key string, reader io.Reader, acl s3layer.CannedACL) error {
	if containsDotDot(key) || containsDotDot(bucket) {
		return fmt.Errorf("invalid key/bucket")
	}
	if err := os.MkdirAll(filepath.Join(o.conf.DataDir(), bucketDir, bucket, filepath.Dir(key)), os.ModeDir|0755); err != nil {
		return err
	}
	path := filepath.Join(o.conf.DataDir(), bucketDir, bucket, key)
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
	if err := o.acl.Set(bucket, key, acl); err != nil {
		return err
	}
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
