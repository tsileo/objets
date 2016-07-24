package objets

import (
	"fmt"
	"os"
	"sync"

	"github.com/cznic/kv"
	"github.com/tsileo/s3layer"
)

var (
	keyACLFmt = "acl:%s:%s" // acl:<bucket>:<key> => ACL
)

// ACL holds a key => ACL reference
type ACL struct {
	db   *kv.DB
	path string
	mu   *sync.Mutex
}

// New creates a new database.
func newACL(path string) (*ACL, error) {
	createOpen := kv.Open
	if _, err := os.Stat(path); os.IsNotExist(err) {
		createOpen = kv.Create
	}
	db, err := createOpen(path, &kv.Options{})
	if err != nil {
		return nil, err
	}
	return &ACL{
		db:   db,
		path: path,
		mu:   new(sync.Mutex),
	}, nil
}

func (acl *ACL) Set(bucket, key string, cacl s3layer.CannedACL) error {
	return acl.db.Set([]byte(fmt.Sprintf(keyACLFmt, bucket, key)), []byte(cacl))
}

func (acl *ACL) Get(bucket, key string) (s3layer.CannedACL, error) {
	racl, err := acl.db.Get(nil, []byte(fmt.Sprintf(keyACLFmt, bucket, key)))
	if err != nil {
		return "", err
	}
	return s3layer.CannedACL(racl), nil
}

func (acl *ACL) Remove(bucket, key string) error {
	return acl.db.Delete([]byte(fmt.Sprintf(keyACLFmt, bucket, key)))
}

func (acl *ACL) Close() error {
	return acl.db.Close()
}
