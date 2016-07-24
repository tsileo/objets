package objets

import (
	"os"
	"sync"

	"github.com/cznic/kv"
	"github.com/tsileo/s3layer"
)

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
	return nil
}

func (acl *ACL) Get(bucket, key string) (s3layer.CannedACL, error) {
	return s3layer.Private, nil
}

func (acl *ACL) Remove(bucket, key string) error {
	return nil
}

func (acl *ACL) Close() error {
	return acl.db.Close()
}
